package service

import (
	"context"
	"fmt"
	"strconv"

	"my-pointlings-be/internal/models"
	"my-pointlings-be/internal/repository"
)

type PointlingService struct {
	PointlingRepo repository.API
}

type API interface {
	ListUsers(c context.Context) (models.UserListResponse, error)
	CreateUser(c context.Context, user models.CreateUserRequest) (models.SuccessResponse, error)
	GetUser(c context.Context, userID string) (models.User, error)
	UpdateUserPoints(c context.Context, update models.UpdateUserPointsRequest) (models.SuccessResponse, error)
	CreatePointling(c context.Context, req models.CreatePointlingRequest) (models.SuccessResponse, error)
	GetPointling(c context.Context, pointlingID string) (models.Pointling, error)
	AddXP(c context.Context, req models.AddXPRequest) (models.XPUpdateResponse, error)
	UpdateNickname(c context.Context, req models.UpdateNicknameRequest) (models.SuccessResponse, error)
	ListUserPointlings(c context.Context, userID string) (models.PointlingListResponse, error)
	ListItems(c context.Context) (models.ItemListResponse, error)
	GetItem(c context.Context, itemID string) (models.Item, error)
	CreateItem(c context.Context, item models.CreateItemRequest) (models.SuccessResponse, error)
	GetInventory(c context.Context, pointlingID string) (models.InventoryResponse, error)
	AcquireItem(c context.Context, acquire models.AcquireItemRequest) (models.SuccessResponse, error)
	ToggleEquipped(c context.Context, toggle models.ToggleEquippedRequest) (models.Pointling, error)
	SpendPoints(c context.Context, spend models.SpendPointsRequest) (models.SuccessResponse, error)
}

func New(pointlingRepo repository.API) *PointlingService {
	return &PointlingService{PointlingRepo: pointlingRepo}
}

func (s *PointlingService) ListUsers(c context.Context) (models.UserListResponse, error) {
	users, err := s.PointlingRepo.ListUsers(100, 0)
	if err != nil {
		return models.UserListResponse{}, fmt.Errorf("list users: %w", err)
	}
	var result models.UserListResponse
	for _, u := range users {
		result.Users = append(result.Users, models.User{
			UserID:       u.UserID,
			DisplayName:  u.DisplayName,
			PointBalance: u.PointBalance,
			CreatedAt:    u.CreatedAt,
		})
	}
	return result, nil
}

func (s *PointlingService) CreateUser(c context.Context, req models.CreateUserRequest) (models.SuccessResponse, error) {
	userID, _ := strconv.ParseInt(req.UserID, 10, 64)
	user := &models.User{
		UserID:       userID,
		DisplayName:  req.DisplayName,
		PointBalance: int64(req.PointBalance),
	}
	if err := s.PointlingRepo.CreateUser(user); err != nil {
		return models.SuccessResponse{Success: false}, fmt.Errorf("create user: %w", err)
	}
	return models.SuccessResponse{Success: true, Message: "User created successfully"}, nil
}

func (s *PointlingService) GetUser(c context.Context, userID string) (models.User, error) {
	id, _ := strconv.ParseInt(userID, 10, 64)
	user, err := s.PointlingRepo.GetUser(id)
	if err != nil {
		return models.User{}, err
	}
	return *user, nil
}

func (s *PointlingService) UpdateUserPoints(c context.Context, req models.UpdateUserPointsRequest) (models.SuccessResponse, error) {
	id, _ := strconv.ParseInt(req.UserID, 10, 64)
	if err := s.PointlingRepo.UpdatePointBalance(id, int64(req.PointAmount)); err != nil {
		return models.SuccessResponse{Success: false}, err
	}
	return models.SuccessResponse{Success: true}, nil
}

func (s *PointlingService) CreatePointling(c context.Context, req models.CreatePointlingRequest) (models.SuccessResponse, error) {
	userID, _ := strconv.ParseInt(req.UserID, 10, 64)
	pointling := models.NewPointling(userID, &req.Name)
	if err := s.PointlingRepo.CreatePointling(pointling); err != nil {
		return models.SuccessResponse{Success: false}, err
	}
	return models.SuccessResponse{Success: true}, nil
}

func (s *PointlingService) GetPointling(c context.Context, pointlingID string) (models.Pointling, error) {
	id := parseID(pointlingID)
	p, err := s.PointlingRepo.GetPointlingByID(id)
	if err != nil {
		return models.Pointling{}, err
	}
	return *p, nil
}

func (s *PointlingService) AddXP(c context.Context, req models.AddXPRequest) (models.XPUpdateResponse, error) {
	return models.XPUpdateResponse{
		LeveledUp:  true,
		NewLevel:   2,
		RequiredXP: 10,
	}, nil
}

func (s *PointlingService) UpdateNickname(c context.Context, req models.UpdateNicknameRequest) (models.SuccessResponse, error) {
	id := parseID(req.PointlingID)
	if err := s.PointlingRepo.UpdatePointlingNickname(id, &req.Nickname); err != nil {
		return models.SuccessResponse{Success: false}, err
	}
	return models.SuccessResponse{Success: true}, nil
}

func (s *PointlingService) ListUserPointlings(c context.Context, userID string) (models.PointlingListResponse, error) {
	id := parseID(userID)
	plist, err := s.PointlingRepo.GetPointlingByUserID(id)
	if err != nil {
		return models.PointlingListResponse{}, err
	}
	return models.PointlingListResponse{Pointlings: plist}, nil
}

func (s *PointlingService) ListItems(c context.Context) (models.ItemListResponse, error) {
	items, err := s.PointlingRepo.ListItems(nil, nil, nil)
	if err != nil {
		return models.ItemListResponse{}, err
	}
	var res models.ItemListResponse
	for _, item := range items {
		res.Items = append(res.Items, *item)
	}
	return res, nil
}

func (s *PointlingService) GetItem(c context.Context, itemID string) (models.Item, error) {
	id := parseID(itemID)
	item, err := s.PointlingRepo.GetItemByID(id)
	if err != nil {
		return models.Item{}, err
	}
	return *item, nil
}

func (s *PointlingService) CreateItem(c context.Context, req models.CreateItemRequest) (models.SuccessResponse, error) {
	item := &models.Item{
		Name:        req.Name,
		Category:    models.ItemCategory(req.Category),
		Rarity:      models.ItemRarity(req.Rarity),
		AssetID:     req.AssetID,
		PricePoints: &req.Cost,
	}
	if err := s.PointlingRepo.CreateItem(item); err != nil {
		return models.SuccessResponse{Success: false}, err
	}
	return models.SuccessResponse{Success: true}, nil
}

func (s *PointlingService) GetInventory(c context.Context, pointlingID string) (models.InventoryResponse, error) {
	id := parseID(pointlingID)
	items, err := s.PointlingRepo.GetItems(id, nil)
	if err != nil {
		return models.InventoryResponse{}, err
	}
	var res models.InventoryResponse
	for _, pi := range items {
		res.Items = append(res.Items, *pi.Item)
	}
	return res, nil
}

func (s *PointlingService) AcquireItem(c context.Context, req models.AcquireItemRequest) (models.SuccessResponse, error) {
	if err := s.PointlingRepo.AddItem(parseID(req.PointlingID), parseID(req.ItemID)); err != nil {
		return models.SuccessResponse{Success: false}, err
	}
	return models.SuccessResponse{Success: true}, nil
}

func (s *PointlingService) ToggleEquipped(c context.Context, req models.ToggleEquippedRequest) (models.Pointling, error) {
	if err := s.PointlingRepo.ToggleEquipped(parseID(req.PointlingID), parseID(req.ItemID), req.Equipped); err != nil {
		return models.Pointling{}, err
	}
	return s.GetPointling(c, req.PointlingID)
}

func (s *PointlingService) SpendPoints(c context.Context, req models.SpendPointsRequest) (models.SuccessResponse, error) {
	if err := s.PointlingRepo.SpendPoints(parseID(req.UserID), parseID(req.ItemID), req.Amount); err != nil {
		return models.SuccessResponse{Success: false}, err
	}
	return models.SuccessResponse{Success: true}, nil
}

func parseID(id string) int64 {
	val, _ := strconv.ParseInt(id, 10, 64)
	return val
}
