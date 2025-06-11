package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"my-pointlings-be/internal/models"
)

// MockItemRepository mocks the ItemRepository interface
type MockItemRepository struct {
	mock.Mock
}

func (m *MockItemRepository) GetByID(id int64) (*models.Item, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Item), args.Error(1)
}

func (m *MockItemRepository) List(category *models.ItemCategory, rarity *models.ItemRarity, slot *models.ItemSlot) ([]*models.Item, error) {
	args := m.Called(category, rarity, slot)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Item), args.Error(1)
}

func (m *MockItemRepository) GetUnlocksForLevel(level int) ([]*models.Item, error) {
	args := m.Called(level)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Item), args.Error(1)
}

func (m *MockItemRepository) Create(item *models.Item) error {
	args := m.Called(item)
	return args.Error(0)
}

// MockPointlingItemRepository mocks the PointlingItemRepository interface
type MockPointlingItemRepository struct {
	mock.Mock
}

func (m *MockPointlingItemRepository) AddItem(pointlingID, itemID int64) error {
	args := m.Called(pointlingID, itemID)
	return args.Error(0)
}

func (m *MockPointlingItemRepository) GetItems(pointlingID int64, equipped *bool) ([]*models.PointlingItem, error) {
	args := m.Called(pointlingID, equipped)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.PointlingItem), args.Error(1)
}

func (m *MockPointlingItemRepository) ToggleEquipped(pointlingID, itemID int64, equipped bool) error {
	args := m.Called(pointlingID, itemID, equipped)
	return args.Error(0)
}

func (m *MockPointlingItemRepository) GetEquippedInSlot(pointlingID int64, slot models.ItemSlot) (*models.PointlingItem, error) {
	args := m.Called(pointlingID, slot)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.PointlingItem), args.Error(1)
}

func (m *MockPointlingItemRepository) InTransaction(fn func(models.PointlingItemRepository) error) error {
	args := m.Called(fn)
	if args.Get(0) == nil {
		return fn(m)
	}
	return args.Error(0)
}

func TestItemHandler_ListItems(t *testing.T) {
	tests := []struct {
		name           string
		queryParams    string
		setupMocks     func(*MockItemRepository)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:        "list all items",
			queryParams: "",
			setupMocks: func(repo *MockItemRepository) {
				repo.On("List", (*models.ItemCategory)(nil), (*models.ItemRarity)(nil), (*models.ItemSlot)(nil)).Return([]*models.Item{
					{
						ItemID:   1,
						Category: models.CategoryAccessory,
						Name:     "Test Hat",
						AssetID:  "hat_1",
						Rarity:   models.RarityCommon,
					},
				}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: `[{
				"item_id": 1,
				"category": "ACCESSORY",
				"name": "Test Hat",
				"asset_id": "hat_1",
				"rarity": "COMMON"
			}]`,
		},
		{
			name:        "filter by category",
			queryParams: "?category=ACCESSORY",
			setupMocks: func(repo *MockItemRepository) {
				category := models.CategoryAccessory
				repo.On("List", &category, (*models.ItemRarity)(nil), (*models.ItemSlot)(nil)).Return([]*models.Item{}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "[]",
		},
		{
			name:           "invalid category",
			queryParams:    "?category=INVALID",
			setupMocks:     func(repo *MockItemRepository) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"Invalid category"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			itemRepo := &MockItemRepository{}
			tt.setupMocks(itemRepo)
			handler := NewItemHandler(itemRepo, nil, nil)

			req := httptest.NewRequest(http.MethodGet, "/api/v1/items"+tt.queryParams, nil)
			rec := httptest.NewRecorder()

			r := chi.NewRouter()
			handler.RegisterRoutes(r)
			r.ServeHTTP(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)
			if tt.expectedStatus == http.StatusOK {
				if tt.expectedBody == "[]" {
					assert.JSONEq(t, tt.expectedBody, rec.Body.String())
				} else {
					var response []*models.Item
					err := json.Unmarshal(rec.Body.Bytes(), &response)
					assert.NoError(t, err)
					assert.Equal(t, 1, len(response))
					assert.Equal(t, int64(1), response[0].ItemID)
					assert.Equal(t, models.CategoryAccessory, response[0].Category)
					assert.Equal(t, "Test Hat", response[0].Name)
					assert.Equal(t, "hat_1", response[0].AssetID)
					assert.Equal(t, models.RarityCommon, response[0].Rarity)
				}
			} else {
				assert.JSONEq(t, tt.expectedBody, rec.Body.String())
			}
			itemRepo.AssertExpectations(t)
		})
	}
}

func TestItemHandler_AcquireItem(t *testing.T) {
	tests := []struct {
		name           string
		pointlingID    string
		itemID         string
		setupMocks     func(*MockItemRepository, *MockPointlingRepository, *MockPointlingItemRepository)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:        "successful acquisition",
			pointlingID: "123",
			itemID:      "456",
			setupMocks: func(itemRepo *MockItemRepository, pointlingRepo *MockPointlingRepository, pointlingItemRepo *MockPointlingItemRepository) {
				item := &models.Item{
					ItemID:   456,
					Category: models.CategoryAccessory,
					Name:     "Test Item",
					AssetID:  "test_1",
					Rarity:   models.RarityCommon,
				}
				pointling := &models.Pointling{
					PointlingID: 123,
					Level:       1,
				}

				itemRepo.On("GetByID", int64(456)).Return(item, nil)
				pointlingRepo.On("GetByID", int64(123)).Return(pointling, nil)
				pointlingItemRepo.On("AddItem", int64(123), int64(456)).Return(nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: `{
				"item_id": 456,
				"category": "ACCESSORY",
				"name": "Test Item",
				"asset_id": "test_1",
				"rarity": "COMMON"
			}`,
		},
		{
			name:        "level requirement not met",
			pointlingID: "123",
			itemID:      "456",
			setupMocks: func(itemRepo *MockItemRepository, pointlingRepo *MockPointlingRepository, pointlingItemRepo *MockPointlingItemRepository) {
				level := 5
				item := &models.Item{
					ItemID:      456,
					UnlockLevel: &level,
				}
				pointling := &models.Pointling{
					PointlingID: 123,
					Level:       1,
				}

				itemRepo.On("GetByID", int64(456)).Return(item, nil)
				pointlingRepo.On("GetByID", int64(123)).Return(pointling, nil)
			},
			expectedStatus: http.StatusForbidden,
			expectedBody:   `{"error":"Level requirement not met"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			itemRepo := &MockItemRepository{}
			pointlingRepo := &MockPointlingRepository{}
			pointlingItemRepo := &MockPointlingItemRepository{}
			tt.setupMocks(itemRepo, pointlingRepo, pointlingItemRepo)

			handler := NewItemHandler(itemRepo, pointlingRepo, pointlingItemRepo)

			path := "/api/v1/pointlings/" + tt.pointlingID + "/items/" + tt.itemID
			req := httptest.NewRequest(http.MethodPost, path, nil)
			rec := httptest.NewRecorder()

			r := chi.NewRouter()
			handler.RegisterRoutes(r)
			r.ServeHTTP(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)
			if tt.expectedStatus == http.StatusOK {
				var response models.Item
				err := json.Unmarshal(rec.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, int64(456), response.ItemID)
				assert.Equal(t, models.CategoryAccessory, response.Category)
				assert.Equal(t, "Test Item", response.Name)
				assert.Equal(t, "test_1", response.AssetID)
				assert.Equal(t, models.RarityCommon, response.Rarity)
			} else {
				assert.JSONEq(t, tt.expectedBody, rec.Body.String())
			}

			itemRepo.AssertExpectations(t)
			pointlingRepo.AssertExpectations(t)
			pointlingItemRepo.AssertExpectations(t)
		})
	}
}
