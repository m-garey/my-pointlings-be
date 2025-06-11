package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"my-pointlings-be/internal/models"
)

// MockUserRepository is a mock implementation of models.UserRepository
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) GetUser(userID int64) (*models.User, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) CreateUser(user *models.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserRepository) UpdatePointBalance(userID int64, newBalance int64) error {
	args := m.Called(userID, newBalance)
	return args.Error(0)
}

func (m *MockUserRepository) ListUsers(limit, offset int) ([]*models.User, error) {
	args := m.Called(limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.User), args.Error(1)
}

func TestUserHandler_CreateUser(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    interface{}
		mockSetup      func(*MockUserRepository)
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "successful creation",
			requestBody: createUserRequest{
				UserID:      123,
				DisplayName: "TestUser",
			},
			mockSetup: func(repo *MockUserRepository) {
				repo.On("CreateUser", mock.MatchedBy(func(u *models.User) bool {
					return u.UserID == 123 && u.DisplayName == "TestUser"
				})).Return(nil)
			},
			expectedStatus: http.StatusCreated,
			expectedBody:   `{"user_id":123,"display_name":"TestUser","point_balance":0,"created_at":"0001-01-01T00:00:00Z"}`,
		},
		{
			name: "missing required fields",
			requestBody: createUserRequest{
				UserID: 123,
				// Missing DisplayName
			},
			mockSetup:      func(repo *MockUserRepository) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"user_id and display_name are required"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &MockUserRepository{}
			tt.mockSetup(repo)
			handler := NewUserHandler(repo)

			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/api/v1/users", bytes.NewReader(body))
			rec := httptest.NewRecorder()

			// Create router to test with URL parameters
			r := chi.NewRouter()
			handler.RegisterRoutes(r)
			r.ServeHTTP(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)
			assert.JSONEq(t, tt.expectedBody, rec.Body.String())
			repo.AssertExpectations(t)
		})
	}
}

func TestUserHandler_GetUser(t *testing.T) {
	tests := []struct {
		name           string
		userID         string
		mockSetup      func(*MockUserRepository)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:   "user found",
			userID: "123",
			mockSetup: func(repo *MockUserRepository) {
				repo.On("GetUser", int64(123)).Return(&models.User{
					UserID:       123,
					DisplayName:  "TestUser",
					PointBalance: 100,
				}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"user_id":123,"display_name":"TestUser","point_balance":100,"created_at":"0001-01-01T00:00:00Z"}`,
		},
		{
			name:   "user not found",
			userID: "999",
			mockSetup: func(repo *MockUserRepository) {
				repo.On("GetUser", int64(999)).Return(nil, nil)
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   `{"error":"User not found"}`,
		},
		{
			name:           "invalid user ID",
			userID:         "invalid",
			mockSetup:      func(repo *MockUserRepository) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"Invalid user ID"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &MockUserRepository{}
			tt.mockSetup(repo)
			handler := NewUserHandler(repo)

			req := httptest.NewRequest(http.MethodGet, "/api/v1/users/"+tt.userID, nil)
			rec := httptest.NewRecorder()

			// Create router to test with URL parameters
			r := chi.NewRouter()
			handler.RegisterRoutes(r)
			r.ServeHTTP(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)
			assert.JSONEq(t, tt.expectedBody, rec.Body.String())
			repo.AssertExpectations(t)
		})
	}
}
