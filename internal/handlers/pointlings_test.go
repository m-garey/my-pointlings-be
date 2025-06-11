package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"my-pointlings-be/internal/models"
)

// MockPointlingRepository mocks the PointlingRepository interface
type MockPointlingRepository struct {
	mock.Mock
}

func (m *MockPointlingRepository) Create(pointling *models.Pointling) error {
	args := m.Called(pointling)
	if args.Get(0) != nil {
		// Simulate DB auto-generated fields
		pointling.PointlingID = 1
		pointling.CreatedAt = time.Now()
	}
	return args.Error(0)
}

func (m *MockPointlingRepository) GetByID(id int64) (*models.Pointling, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Pointling), args.Error(1)
}

func (m *MockPointlingRepository) GetByUserID(userID int64) ([]*models.Pointling, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Pointling), args.Error(1)
}

func (m *MockPointlingRepository) UpdateLook(id int64, look models.JSONMap) error {
	args := m.Called(id, look)
	return args.Error(0)
}

func (m *MockPointlingRepository) UpdateXP(id int64, currentXP, requiredXP int) error {
	args := m.Called(id, currentXP, requiredXP)
	return args.Error(0)
}

func (m *MockPointlingRepository) UpdateLevel(id int64, level int) error {
	args := m.Called(id, level)
	return args.Error(0)
}

func (m *MockPointlingRepository) UpdateNickname(id int64, nickname *string) error {
	args := m.Called(id, nickname)
	return args.Error(0)
}

func TestPointlingHandler_CreatePointling(t *testing.T) {
	nickname := "TestPointling"
	tests := []struct {
		name           string
		requestBody    interface{}
		mockSetup      func(*MockPointlingRepository)
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "successful creation",
			requestBody: createPointlingRequest{
				UserID:   123,
				Nickname: &nickname,
			},
			mockSetup: func(repo *MockPointlingRepository) {
				repo.On("Create", mock.MatchedBy(func(p *models.Pointling) bool {
					return p.UserID == 123 && *p.Nickname == nickname
				})).Return(nil)
			},
			expectedStatus: http.StatusCreated,
			expectedBody:   `{"pointling_id":1,"user_id":123,"nickname":"TestPointling","level":1,"current_xp":0,"required_xp":3,"look_json":{}}`,
		},
		{
			name: "missing user ID",
			requestBody: createPointlingRequest{
				Nickname: &nickname,
			},
			mockSetup:      func(repo *MockPointlingRepository) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"user_id is required"}`,
		},
		{
			name: "repository error",
			requestBody: createPointlingRequest{
				UserID:   123,
				Nickname: &nickname,
			},
			mockSetup: func(repo *MockPointlingRepository) {
				repo.On("Create", mock.Anything).Return(fmt.Errorf("db error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"error":"Failed to create pointling"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &MockPointlingRepository{}
			tt.mockSetup(repo)
			handler := NewPointlingHandler(repo)

			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/api/v1/pointlings", bytes.NewReader(body))
			rec := httptest.NewRecorder()

			// Create router to test with URL parameters
			r := chi.NewRouter()
			handler.RegisterRoutes(r)
			r.ServeHTTP(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)
			if tt.expectedStatus == http.StatusCreated {
				// For creation, we only check that required fields are present and correct
				var response models.Pointling
				err := json.Unmarshal(rec.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, int64(123), response.UserID)
				assert.Equal(t, nickname, *response.Nickname)
				assert.Equal(t, 1, response.Level)
			} else {
				assert.JSONEq(t, tt.expectedBody, rec.Body.String())
			}
			repo.AssertExpectations(t)
		})
	}
}

func TestPointlingHandler_GetPointling(t *testing.T) {
	nickname := "TestPointling"
	tests := []struct {
		name           string
		pointlingID    string
		mockSetup      func(*MockPointlingRepository)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:        "pointling found",
			pointlingID: "123",
			mockSetup: func(repo *MockPointlingRepository) {
				repo.On("GetByID", int64(123)).Return(&models.Pointling{
					PointlingID:   123,
					UserID:        456,
					Nickname:      &nickname,
					Level:         1,
					CurrentXP:     0,
					RequiredXP:    3,
					PersonalityID: nil,
					LookJSON:      models.JSONMap{},
					CreatedAt:     time.Date(2025, 6, 10, 0, 0, 0, 0, time.UTC),
				}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"pointling_id":123,"user_id":456,"nickname":"TestPointling","level":1,"current_xp":0,"required_xp":3,"look_json":{},"created_at":"2025-06-10T00:00:00Z"}`,
		},
		{
			name:        "pointling not found",
			pointlingID: "999",
			mockSetup: func(repo *MockPointlingRepository) {
				repo.On("GetByID", int64(999)).Return(nil, nil)
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   `{"error":"Pointling not found"}`,
		},
		{
			name:           "invalid ID",
			pointlingID:    "invalid",
			mockSetup:      func(repo *MockPointlingRepository) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"Invalid pointling ID"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &MockPointlingRepository{}
			tt.mockSetup(repo)
			handler := NewPointlingHandler(repo)

			req := httptest.NewRequest(http.MethodGet, "/api/v1/pointlings/"+tt.pointlingID, nil)
			rec := httptest.NewRecorder()

			r := chi.NewRouter()
			handler.RegisterRoutes(r)
			r.ServeHTTP(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)
			assert.JSONEq(t, tt.expectedBody, rec.Body.String())
			repo.AssertExpectations(t)
		})
	}
}
