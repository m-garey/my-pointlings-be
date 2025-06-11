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

type MockPointSpendRepository struct {
	mock.Mock
}

func (m *MockPointSpendRepository) Create(spend *models.PointSpend) error {
	args := m.Called(spend)
	if args.Get(0) == nil {
		spend.SpendID = 1
		spend.SpendTS = time.Now()
	}
	return args.Error(0)
}

func (m *MockPointSpendRepository) GetByUser(userID int64, limit, offset int) ([]*models.PointSpend, error) {
	args := m.Called(userID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.PointSpend), args.Error(1)
}

func (m *MockPointSpendRepository) GetTotalSpentByUser(userID int64) (int64, error) {
	args := m.Called(userID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockPointSpendRepository) InTransaction(fn func(models.PointSpendRepository) error) error {
	args := m.Called(fn)
	if args.Get(0) == nil {
		return fn(m)
	}
	return args.Error(0)
}

func (m *MockPointSpendRepository) SpendPoints(userID int64, itemID int64, points int) error {
	args := m.Called(userID, itemID, points)
	return args.Error(0)
}

func TestPointHandler_SpendPoints(t *testing.T) {
	tests := []struct {
		name           string
		userID         string
		requestBody    interface{}
		setupMocks     func(*MockPointSpendRepository, *MockItemRepository, *MockPointlingItemRepository)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:   "successful purchase",
			userID: "123",
			requestBody: spendPointsRequest{
				ItemID:      456,
				PointlingID: 789,
			},
			setupMocks: func(pointRepo *MockPointSpendRepository, itemRepo *MockItemRepository, piRepo *MockPointlingItemRepository) {
				points := 100
				item := &models.Item{
					ItemID:      456,
					Name:        "Test Item",
					PricePoints: &points,
				}

				itemRepo.On("GetByID", int64(456)).Return(item, nil)
				pointRepo.On("InTransaction", mock.Anything).Return(nil)
				pointRepo.On("SpendPoints", int64(123), int64(456), 100).Return(nil)
				piRepo.On("AddItem", int64(789), int64(456)).Return(nil)
				pointRepo.On("GetTotalSpentByUser", int64(123)).Return(int64(100), nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: `{
				"item_id": 456,
				"points_spent": 100,
				"new_balance": 100
			}`,
		},
		{
			name:   "insufficient points",
			userID: "123",
			requestBody: spendPointsRequest{
				ItemID:      456,
				PointlingID: 789,
			},
			setupMocks: func(pointRepo *MockPointSpendRepository, itemRepo *MockItemRepository, piRepo *MockPointlingItemRepository) {
				points := 100
				item := &models.Item{
					ItemID:      456,
					PricePoints: &points,
				}

				itemRepo.On("GetByID", int64(456)).Return(item, nil)
				pointRepo.On("InTransaction", mock.Anything).Return(models.ErrInsufficientBalance)
			},
			expectedStatus: http.StatusPaymentRequired,
			expectedBody:   `{"error":"Insufficient points"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pointRepo := &MockPointSpendRepository{}
			itemRepo := &MockItemRepository{}
			piRepo := &MockPointlingItemRepository{}
			tt.setupMocks(pointRepo, itemRepo, piRepo)

			handler := NewPointHandler(pointRepo, itemRepo, piRepo)

			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/api/v1/users/"+tt.userID+"/points/spend", bytes.NewReader(body))
			rec := httptest.NewRecorder()

			r := chi.NewRouter()
			handler.RegisterRoutes(r)
			r.ServeHTTP(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)
			assert.JSONEq(t, tt.expectedBody, rec.Body.String())

			pointRepo.AssertExpectations(t)
			itemRepo.AssertExpectations(t)
			piRepo.AssertExpectations(t)
		})
	}
}

func TestPointHandler_GetSpendHistory(t *testing.T) {
	tests := []struct {
		name           string
		userID         string
		queryParams    string
		setupMocks     func(*MockPointSpendRepository)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:        "successful history fetch",
			userID:      "123",
			queryParams: "?limit=1",
			setupMocks: func(repo *MockPointSpendRepository) {
				spend := &models.PointSpend{
					SpendID:     1,
					UserID:      123,
					ItemID:      456,
					PointsSpent: 100,
					SpendTS:     time.Date(2025, 6, 10, 0, 0, 0, 0, time.UTC),
					Item: &models.Item{
						ItemID:   456,
						Name:     "Test Item",
						AssetID:  "test_1",
						Category: models.CategoryAccessory,
						Rarity:   models.RarityCommon,
					},
				}
				repo.On("GetByUser", int64(123), 1, 0).Return([]*models.PointSpend{spend}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: `[{
				"spend_id": 1,
				"user_id": 123,
				"item_id": 456,
				"points_spent": 100,
				"spend_ts": "2025-06-10T00:00:00Z",
				"item": {
					"item_id": 456,
					"name": "Test Item",
					"asset_id": "test_1",
					"category": "ACCESSORY",
					"rarity": "COMMON"
				}
			}]`,
		},
		{
			name:        "no history",
			userID:      "123",
			queryParams: "",
			setupMocks: func(repo *MockPointSpendRepository) {
				repo.On("GetByUser", int64(123), 50, 0).Return([]*models.PointSpend{}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `[]`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &MockPointSpendRepository{}
			tt.setupMocks(repo)

			handler := NewPointHandler(repo, nil, nil)

			req := httptest.NewRequest(http.MethodGet, "/api/v1/users/"+tt.userID+"/points/history"+tt.queryParams, nil)
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
