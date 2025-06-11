package pointling_service

import (
	"context"
	"my-pointlings-be/internal/repository/pointling_repo"
	model "my-pointlings-be/internal/service/model"
)

type API interface {
	listUsers(c context.Context) (model.UserListResponse, error)
	createUser(c *context.Context, user model.CreateUserRequest) (model.SuccessResponse, error)
	getUser(c *context.Context, userID string) (model.User, error)
	updateUserPoints(c *context.Context, updateUserPoints model.UpdateUserPointsRequest) (model.SuccessResponse, error)
	createPointling(c *context.Context, pointling model.CreatePointlingRequest) (model.SuccessResponse, error)
	getPointling(c *context.Context, pointlingID string) (model.Pointling, error)
	addXP(c *context.Context, xpUpdate model.XPUpdateRequest) (model.XPUpdateResponse, model.LevelUpOptionsResponse, error)
	getXPHistory(c *context.Context, pointlingID string) (model.XPHistoryResponse, error)
	updateNickname(c *context.Context, updateNickname model.UpdateNicknameRequest) (model.SuccessResponse, error)
	listUserPointlings(c *context.Context, pointlingID string) (model.PointlingListResponse, error)
	listItems(c *context.Context) (model.ItemListResponse, error)
	getItem(c *context.Context, itemID string) (model.Item, error)
	createItem(c *context.Context, item model.CreateItemRequest) (model.SuccessResponse, error)
	getInventory(c *context.Context, pointlingID string) model.InventoryResponse
	acquireItem(c *context.Context, acquire model.AcquireItemRequest) (model.SuccessResponse, error)
	toggleEquipped(c *context.Context, toggle model.ToggleEquippedRequest) (model.Pointling, error)
	spendPoints(c *context.Context, spendPoints model.SpendPointsRequest) (model.SuccessResponse, error)
	getSpendHistory(c *context.Context, userID string) (model.SpendHistoryResponse, error)
}

type PointlingService struct {
	PointlingRepo pointling_repo.API
}

func New(pointlingRepo pointling_repo.API) *PointlingService {
	return &PointlingService{PointlingRepo: pointlingRepo}
}
