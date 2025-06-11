package service

import (
	"context"
	"fmt"
	"my-pointlings-be/internal/repository"
	model "my-pointlings-be/internal/service/model"
	"strconv"
)

type API interface {
	ListUsers(c context.Context) (model.UserListResponse, error)
	CreateUser(c context.Context, user model.CreateUserRequest) (model.SuccessResponse, error)
	GetUser(c context.Context, userID string) (model.User, error)
	UpdateUserPoints(c context.Context, updateUserPoints model.UpdateUserPointsRequest) (model.SuccessResponse, error)
	CreatePointling(c context.Context, pointling model.CreatePointlingRequest) (model.SuccessResponse, error)
	GetPointling(c context.Context, pointlingID string) (model.Pointling, error)
	AddXP(c context.Context, xpUpdate model.XPUpdateRequest) (model.XPUpdateResponse, error)
	GetXPHistory(c context.Context, pointlingID string) (model.XPHistoryResponse, error)
	UpdateNickname(c context.Context, updateNickname model.UpdateNicknameRequest) (model.SuccessResponse, error)
	ListUserPointlings(c context.Context, pointlingID string) (model.PointlingListResponse, error)
	ListItems(c context.Context) (model.ItemListResponse, error)
	GetItem(c context.Context, itemID string) (model.Item, error)
	CreateItem(c context.Context, item model.CreateItemRequest) (model.SuccessResponse, error)
	GetInventory(c context.Context, pointlingID string) (model.InventoryResponse, error)
	AcquireItem(c context.Context, acquire model.AcquireItemRequest) (model.SuccessResponse, error)
	ToggleEquipped(c context.Context, toggle model.ToggleEquippedRequest) (model.Pointling, error)
	SpendPoints(c context.Context, spendPoints model.SpendPointsRequest) (model.SuccessResponse, error)
	GetSpendHistory(c context.Context, userID string) (model.SpendHistoryResponse, error)
}

type PointlingService struct {
	PointlingRepo *repository.API
}

func New(pointlingRepo *repository.API) *PointlingService {
	return &PointlingService{PointlingRepo: pointlingRepo}
}

func (s *PointlingService) ListUsers(c context.Context) (model.UserListResponse, error) {
	users, err := s.PointlingRepo.ListUsers(100, 0)
	if err != nil {
		return model.UserListResponse{}, fmt.Errorf("list users: %w", err)
	}
	var result model.UserListResponse
	for _, u := range users {
		result.Users = append(result.Users, model.User{
			UserID:      strconv.FormatInt(u.UserID, 10),
			UserName:    u.DisplayName,
			PointAmount: int(u.PointBalance),
		})
	}
	return result, nil
}
