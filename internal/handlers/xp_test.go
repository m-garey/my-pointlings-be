package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"my-pointlings-be/internal/models"
)

type MockXPRepository struct {
	mock.Mock
}

func (m *MockXPRepository) AddXP(event *models.XPEvent) error {
	args := m.Called(event)
	if args.Get(0) == nil {
		event.EventID = 1
		event.EventTS = time.Now()
	}
	return args.Error(0)
}

func (m *MockXPRepository) GetEventsByPointling(pointlingID int64, limit int) ([]*models.XPEvent, error) {
	args := m.Called(pointlingID, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.XPEvent), args.Error(1)
}

func (m *MockXPRepository) GetDailyXPBySource(pointlingID int64, source models.XPEventSource) (int, error) {
	args := m.Called(pointlingID, source)
	return args.Int(0), args.Error(1)
}

func (m *MockXPRepository) InTransaction(fn func(models.XPRepository) error) error {
	args := m.Called(fn)
	if args.Get(0) == nil {
		return fn(m)
	}
	return args.Error(0)
}

func TestXPHandler_AddXP(t *testing.T) {
	tests := []struct {
		name           string
		pointlingID    string
		requestBody    interface{}
		setupMocks     func(*MockXPRepository, *MockPointlingRepository)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:        "successful XP gain without level up",
			pointlingID: "123",
			requestBody: struct {
				Source string `json:"source"`
				Amount int    `json:"amount"`
			}{
				Source: string(models.XPSourceReceipt),
				Amount: 10,
			},
			setupMocks: func(xpRepo *MockXPRepository, pointlingRepo *MockPointlingRepository) {
				pointling := &models.Pointling{
					PointlingID: 123,
					Level:       1,
					CurrentXP:   0,
					RequiredXP:  20,
				}

				pointlingRepo.On("GetByID", int64(123)).Return(pointling, nil)

				expectedEvent := &models.XPEvent{
					PointlingID: 123,
					Source:      models.XPSourceReceipt,
					XPAmount:    10,
				}

				xpRepo.On("InTransaction", mock.AnythingOfType("func(models.XPRepository) error")).Return(nil)
				xpRepo.On("AddXP", mock.MatchedBy(func(e *models.XPEvent) bool {
					return e.PointlingID == expectedEvent.PointlingID &&
						e.Source == expectedEvent.Source &&
						e.XPAmount == expectedEvent.XPAmount
				})).Return(nil)

				pointlingRepo.On("UpdateXP", int64(123), 10, 20).Return(nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: `{
				"xp_gained": 10,
				"new_total": 10,
				"leveled_up": false,
				"required_xp": 20,
				"pointling_id": 123,
				"event": {
					"event_id": 1,
					"pointling_id": 123,
					"source": "RECEIPT",
					"xp_amount": 10
				}
			}`,
		},
		{
			name:        "daily limit reached",
			pointlingID: "123",
			requestBody: struct {
				Source string `json:"source"`
				Amount int    `json:"amount"`
			}{
				Source: string(models.XPSourceReceipt),
				Amount: 10,
			},
			setupMocks: func(xpRepo *MockXPRepository, pointlingRepo *MockPointlingRepository) {
				pointlingRepo.On("GetByID", int64(123)).Return(&models.Pointling{
					PointlingID: 123,
					Level:       1,
					CurrentXP:   0,
					RequiredXP:  20,
				}, nil)

				xpRepo.On("InTransaction", mock.AnythingOfType("func(models.XPRepository) error")).Return(models.ErrDailyXPLimitReached)
			},
			expectedStatus: http.StatusTooManyRequests,
			expectedBody:   `{"error":"Daily XP limit reached for this source"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			xpRepo := &MockXPRepository{}
			pointlingRepo := &MockPointlingRepository{}
			tt.setupMocks(xpRepo, pointlingRepo)
			handler := NewXPHandler(xpRepo, pointlingRepo)

			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/api/v1/pointlings/"+tt.pointlingID+"/xp", bytes.NewReader(body))
			rec := httptest.NewRecorder()

			r := chi.NewRouter()
			handler.RegisterRoutes(r)
			r.ServeHTTP(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)
			if tt.expectedStatus == http.StatusOK {
				var response struct {
					XPGained   int             `json:"xp_gained"`
					NewTotal   int             `json:"new_total"`
					LeveledUp  bool            `json:"leveled_up"`
					RequiredXP int             `json:"required_xp"`
					Event      *models.XPEvent `json:"event"`
				}
				err := json.Unmarshal(rec.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, tt.requestBody.(struct {
					Source string `json:"source"`
					Amount int    `json:"amount"`
				}).Amount, response.XPGained)
				assert.NotNil(t, response.Event)
				assert.Equal(t, models.XPSourceReceipt, response.Event.Source)
				assert.Equal(t, 10, response.Event.XPAmount)
			} else {
				assert.JSONEq(t, tt.expectedBody, rec.Body.String())
			}

			xpRepo.AssertExpectations(t)
			pointlingRepo.AssertExpectations(t)
		})
	}
}
