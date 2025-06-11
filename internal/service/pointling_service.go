package service

import (
	"context"
	"fmt"
	"strconv"

	"my-pointlings-be/internal/models"
	model "my-pointlings-be/internal/models"
	"my-pointlings-be/internal/repository"
)

type PointlingService struct {
	PointlingRepo repository.API
}

type API interface {
	ListUsers(c context.Context) (model.UserListResponse, error)
	CreateUser(c context.Context, user model.CreateUserRequest) (model.SuccessResponse, error)
	GetUser(c context.Context, userID string) (model.User, error)
	UpdateUserPoints(c context.Context, update model.UpdateUserPointsRequest) (model.SuccessResponse, error)
	CreatePointling(c context.Context, req model.CreatePointlingRequest) (model.SuccessResponse, error)
	GetPointling(c context.Context, pointlingID string) (model.Pointling, error)
	AddXP(c context.Context, req model.AddXPRequest) (model.XPUpdateResponse, error)
	UpdateNickname(c context.Context, req model.UpdateNicknameRequest) (model.SuccessResponse, error)
	ListUserPointlings(c context.Context, userID string) (model.PointlingListResponse, error)
	ListItems(c context.Context) (model.ItemListResponse, error)
	GetItem(c context.Context, itemID string) (model.Item, error)
	CreateItem(c context.Context, item model.CreateItemRequest) (model.SuccessResponse, error)
	GetInventory(c context.Context, pointlingID string) (model.InventoryResponse, error)
	AcquireItem(c context.Context, acquire model.AcquireItemRequest) (model.SuccessResponse, error)
	ToggleEquipped(c context.Context, toggle model.ToggleEquippedRequest) (model.Pointling, error)
	SpendPoints(c context.Context, spend model.SpendPointsRequest) (model.SuccessResponse, error)
}

func New(pointlingRepo repository.API) *PointlingService {
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

func (s *PointlingService) CreateUser(c context.Context, user model.CreateUserRequest) (model.SuccessResponse, error) {
	newUser := &models.User{
		DisplayName:  user.UserName,
		PointBalance: 0,
	}
	err := s.PointlingRepo.CreateUser(newUser)
	if err != nil {
		return model.SuccessResponse{Success: false}, fmt.Errorf("create user: %w", err)
	}
	return model.SuccessResponse{Success: true, Message: "User created successfully"}, nil
}

func (s *PointlingService) GetUser(c context.Context, userID string) (model.User, error) {
	id, _ := strconv.ParseInt(userID, 10, 64)
	user, err := s.PointlingRepo.GetUser(id)
	if err != nil {
		return model.User{}, err
	}
	return model.User{
		UserID:      strconv.FormatInt(user.UserID, 10),
		UserName:    user.DisplayName,
		PointAmount: int(user.PointBalance),
	}, nil
}

func (s *PointlingService) UpdateUserPoints(c context.Context, update model.UpdateUserPointsRequest) (model.SuccessResponse, error) {
	id, _ := strconv.ParseInt(update.UserID, 10, 64)
	err := s.PointlingRepo.UpdatePointBalance(id, int64(update.PointAmount))
	if err != nil {
		return model.SuccessResponse{Success: false}, err
	}
	return model.SuccessResponse{Success: true}, nil
}

func (s *PointlingService) CreatePointling(c context.Context, req model.CreatePointlingRequest) (model.SuccessResponse, error) {
	pointling := &models.Pointling{
		UserID:     parseID(req.UserID),
		Nickname:   &req.Name,
		Level:      1,
		CurrentXP:  0,
		RequiredXP: 3,
	}
	err := s.PointlingRepo.CreatePointling(pointling)
	if err != nil {
		return model.SuccessResponse{Success: false}, err
	}
	return model.SuccessResponse{Success: true}, nil
}

func (s *PointlingService) GetPointling(c context.Context, pointlingID string) (model.Pointling, error) {
	id := parseID(pointlingID)
	p, err := s.PointlingRepo.GetPointlingByID(id)
	if err != nil {
		return model.Pointling{}, err
	}
	return model.Pointling{
		PointlingID:  strconv.FormatInt(p.PointlingID, 10),
		Name:         *p.Nickname,
		CurrentXP:    p.CurrentXP,
		RequiredXP:   p.RequiredXP,
		Level:        p.Level,
		AppearanceID: "",
		WardrobeID:   "",
	}, nil
}

func (s *PointlingService) AddXP(c context.Context, req model.AddXPRequest) (model.XPUpdateResponse, error) {
	return model.XPUpdateResponse{
			LeveledUp:  true,
			NewLevel:   2,
			RequiredXP: 10,
		}, model.LevelUpOptionsResponse{
			Options: []struct {
				Type string `json:"type"`
				ID   string `json:"id"`
			}{
				{"accessory", "a001"},
				{"feature", "f002"},
				{"accessory", "a005"},
			},
		}, nil
}

func (s *PointlingService) UpdateNickname(c context.Context, req model.UpdateNicknameRequest) (model.SuccessResponse, error) {
	id := parseID(req.PointlingID)
	err := s.PointlingRepo.UpdatePointlingNickname(id, &req.Nickname)
	if err != nil {
		return model.SuccessResponse{Success: false}, err
	}
	return model.SuccessResponse{Success: true}, nil
}

func (s *PointlingService) ListUserPointlings(c context.Context, userID string) (model.PointlingListResponse, error) {
	id := parseID(userID)
	plist, err := s.PointlingRepo.GetPointlingByUserID(id)
	if err != nil {
		return model.PointlingListResponse{}, err
	}
	var result model.PointlingListResponse
	for _, p := range plist {
		result.Pointlings = append(result.Pointlings, model.Pointling{
			PointlingID:  strconv.FormatInt(p.PointlingID, 10),
			Name:         *p.Nickname,
			CurrentXP:    p.CurrentXP,
			RequiredXP:   p.RequiredXP,
			Level:        p.Level,
			AppearanceID: "",
			WardrobeID:   "",
		})
	}
	return result, nil
}

func (s *PointlingService) ListItems(c context.Context) (model.ItemListResponse, error) {
	items, err := s.PointlingRepo.ListItems(nil, nil, nil)
	if err != nil {
		return model.ItemListResponse{}, err
	}
	var res model.ItemListResponse
	for _, item := range items {
		res.Items = append(res.Items, model.Item{
			ItemID: strconv.FormatInt(item.ItemID, 10),
			Name:   item.Name,
			Type:   item.Category,
			Rarity: item.Rarity,
			Cost:   item.PricePoints,
			Image:  item.AssetID,
		})
	}
	return res, nil
}

func (s *PointlingService) GetItem(c context.Context, itemID string) (model.Item, error) {
	id := parseID(itemID)
	item, err := s.PointlingRepo.GetItemByID(id)
	if err != nil {
		return model.Item{}, err
	}
	return model.Item{
		ItemID: strconv.FormatInt(item.ItemID, 10),
		Name:   item.Name,
		Type:   item.Category,
		Rarity: item.Rarity,
		Cost:   item.PricePoints,
		Image:  item.AssetID,
	}, nil
}

func (s *PointlingService) CreateItem(c context.Context, item model.CreateItemRequest) (model.SuccessResponse, error) {
	newItem := &models.Item{
		Name:        item.Name,
		Category:    item.Type,
		Rarity:      item.Rarity,
		PricePoints: item.Cost,
		AssetID:     item.Image,
	}
	err := s.PointlingRepo.CreateItem(newItem)
	if err != nil {
		return model.SuccessResponse{Success: false}, err
	}
	return model.SuccessResponse{Success: true}, nil
}

func (s *PointlingService) GetInventory(c context.Context, pointlingID string) (model.InventoryResponse, error) {
	id := parseID(pointlingID)
	items, _ := s.PointlingRepo.GetItems(id, nil)
	var res model.InventoryResponse
	for _, pi := range items {
		res.Items = append(res.Items, model.Item{
			ItemID: strconv.FormatInt(pi.Item.ItemID, 10),
			Name:   pi.Item.Name,
			Type:   pi.Item.Category,
			Rarity: pi.Item.Rarity,
			Cost:   pi.Item.PricePoints,
			Image:  pi.Item.AssetID,
		})
	}
	return res
}

func (s *PointlingService) AcquireItem(c context.Context, acquire model.AcquireItemRequest) (model.SuccessResponse, error) {
	err := s.PointlingRepo.AddItem(parseID(acquire.PointlingID), parseID(acquire.ItemID))
	if err != nil {
		return model.SuccessResponse{Success: false}, err
	}
	return model.SuccessResponse{Success: true}, nil
}

func (s *PointlingService) ToggleEquipped(c context.Context, toggle model.ToggleEquippedRequest) (model.Pointling, error) {
	err := s.PointlingRepo.ToggleEquipped(parseID(toggle.PointlingID), parseID(toggle.ItemID), toggle.Equipped)
	if err != nil {
		return model.Pointling{}, err
	}
	return s.GetPointling(c, toggle.PointlingID)
}

func (s *PointlingService) SpendPoints(c context.Context, spend model.SpendPointsRequest) (model.SuccessResponse, error) {
	err := s.PointlingRepo.SpendPoints(parseID(spend.UserID), parseID(spend.ItemID), spend.Amount)
	if err != nil {
		return model.SuccessResponse{Success: false}, err
	}
	return model.SuccessResponse{Success: true}, nil
}

// Helper function to parse ID
func parseID(id string) int64 {
	val, _ := strconv.ParseInt(id, 10, 64)
	return val
}
