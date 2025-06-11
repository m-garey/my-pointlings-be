package postgres

import (
	"database/sql"
	"fmt"

	"my-pointlings-be/internal/models"
)

type ItemRepository struct {
	db *sql.DB
}

func NewItemRepository(db *sql.DB) *ItemRepository {
	return &ItemRepository{db: db}
}

func (r *ItemRepository) GetByID(id int64) (*models.Item, error) {
	query := `
		SELECT item_id, category, slot, asset_id, name, rarity, price_points, unlock_level
		FROM public.items
		WHERE item_id = $1`

	item := &models.Item{}
	err := r.db.QueryRow(query, id).Scan(
		&item.ItemID,
		&item.Category,
		&item.Slot,
		&item.AssetID,
		&item.Name,
		&item.Rarity,
		&item.PricePoints,
		&item.UnlockLevel,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get item: %w", err)
	}
	return item, nil
}

func (r *ItemRepository) List(category *models.ItemCategory, rarity *models.ItemRarity, slot *models.ItemSlot) ([]*models.Item, error) {
	query := `
		SELECT item_id, category, slot, asset_id, name, rarity, price_points, unlock_level
		FROM public.items
		WHERE ($1::text IS NULL OR category = $1)
		AND ($2::text IS NULL OR rarity = $2)
		AND ($3::text IS NULL OR slot = $3)
		ORDER BY rarity, name`

	rows, err := r.db.Query(query, category, rarity, slot)
	if err != nil {
		return nil, fmt.Errorf("list items query: %w", err)
	}
	defer rows.Close()

	var items []*models.Item
	for rows.Next() {
		item := &models.Item{}
		err := rows.Scan(
			&item.ItemID,
			&item.Category,
			&item.Slot,
			&item.AssetID,
			&item.Name,
			&item.Rarity,
			&item.PricePoints,
			&item.UnlockLevel,
		)
		if err != nil {
			return nil, fmt.Errorf("scan item row: %w", err)
		}
		items = append(items, item)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate items: %w", err)
	}
	return items, nil
}

func (r *ItemRepository) GetUnlocksForLevel(level int) ([]*models.Item, error) {
	query := `
		SELECT item_id, category, slot, asset_id, name, rarity, price_points, unlock_level
		FROM public.items
		WHERE unlock_level = $1
		ORDER BY rarity, name`

	rows, err := r.db.Query(query, level)
	if err != nil {
		return nil, fmt.Errorf("get level unlocks query: %w", err)
	}
	defer rows.Close()

	var items []*models.Item
	for rows.Next() {
		item := &models.Item{}
		err := rows.Scan(
			&item.ItemID,
			&item.Category,
			&item.Slot,
			&item.AssetID,
			&item.Name,
			&item.Rarity,
			&item.PricePoints,
			&item.UnlockLevel,
		)
		if err != nil {
			return nil, fmt.Errorf("scan unlock item: %w", err)
		}
		items = append(items, item)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate unlock items: %w", err)
	}
	return items, nil
}

func (r *ItemRepository) Create(item *models.Item) error {
	query := `
		INSERT INTO public.items (
			category, slot, asset_id, name, rarity, price_points, unlock_level
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING item_id`

	err := r.db.QueryRow(
		query,
		item.Category,
		item.Slot,
		item.AssetID,
		item.Name,
		item.Rarity,
		item.PricePoints,
		item.UnlockLevel,
	).Scan(&item.ItemID)

	if err != nil {
		return fmt.Errorf("create item: %w", err)
	}
	return nil
}
