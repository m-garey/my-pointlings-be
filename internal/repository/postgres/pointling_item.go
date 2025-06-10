package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/pointlings/backend/internal/models"
)

type PointlingItemRepository struct {
	db *sql.DB
	tx *sql.Tx // For transaction support
}

func NewPointlingItemRepository(db *sql.DB) *PointlingItemRepository {
	return &PointlingItemRepository{db: db}
}

// txWrapper ensures operations use a transaction if one exists
func (r *PointlingItemRepository) txWrapper() interface {
	QueryRow(query string, args ...interface{}) *sql.Row
	Exec(query string, args ...interface{}) (sql.Result, error)
	Query(query string, args ...interface{}) (*sql.Rows, error)
} {
	if r.tx != nil {
		return r.tx
	}
	return r.db
}

func (r *PointlingItemRepository) InTransaction(fn func(models.PointlingItemRepository) error) error {
	tx, err := r.db.BeginTx(context.Background(), nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}

	// Create new repository with transaction
	txRepo := &PointlingItemRepository{db: r.db, tx: tx}

	// Execute the function
	if err := fn(txRepo); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("rollback failed: %v (original error: %v)", rbErr, err)
		}
		return err
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}

func (r *PointlingItemRepository) AddItem(pointlingID, itemID int64) error {
	query := `
		INSERT INTO public.pointling_items (pointling_id, item_id)
		VALUES ($1, $2)
		ON CONFLICT (pointling_id, item_id) DO NOTHING`

	result, err := r.txWrapper().Exec(query, pointlingID, itemID)
	if err != nil {
		return fmt.Errorf("add item: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get rows affected: %w", err)
	}
	if rows == 0 {
		return models.ErrAlreadyOwned
	}

	return nil
}

func (r *PointlingItemRepository) GetItems(pointlingID int64, equipped *bool) ([]*models.PointlingItem, error) {
	query := `
		SELECT pi.pointling_id, pi.item_id, pi.acquired_at, pi.equipped,
			   i.category, i.slot, i.asset_id, i.name, i.rarity, i.price_points, i.unlock_level
		FROM public.pointling_items pi
		JOIN public.items i ON i.item_id = pi.item_id
		WHERE pi.pointling_id = $1
		AND ($2::boolean IS NULL OR pi.equipped = $2)
		ORDER BY pi.acquired_at DESC`

	rows, err := r.txWrapper().Query(query, pointlingID, equipped)
	if err != nil {
		return nil, fmt.Errorf("get items query: %w", err)
	}
	defer rows.Close()

	var items []*models.PointlingItem
	for rows.Next() {
		pi := &models.PointlingItem{
			Item: &models.Item{},
		}
		err := rows.Scan(
			&pi.PointlingID,
			&pi.ItemID,
			&pi.AcquiredAt,
			&pi.Equipped,
			&pi.Item.Category,
			&pi.Item.Slot,
			&pi.Item.AssetID,
			&pi.Item.Name,
			&pi.Item.Rarity,
			&pi.Item.PricePoints,
			&pi.Item.UnlockLevel,
		)
		if err != nil {
			return nil, fmt.Errorf("scan pointling item: %w", err)
		}
		pi.Item.ItemID = pi.ItemID // Set the ID from the join
		items = append(items, pi)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate items: %w", err)
	}
	return items, nil
}

func (r *PointlingItemRepository) GetEquippedInSlot(pointlingID int64, slot models.ItemSlot) (*models.PointlingItem, error) {
	query := `
		SELECT pi.pointling_id, pi.item_id, pi.acquired_at, pi.equipped,
			   i.category, i.slot, i.asset_id, i.name, i.rarity, i.price_points, i.unlock_level
		FROM public.pointling_items pi
		JOIN public.items i ON i.item_id = pi.item_id
		WHERE pi.pointling_id = $1
		AND i.slot = $2
		AND pi.equipped = true`

	pi := &models.PointlingItem{
		Item: &models.Item{},
	}
	err := r.txWrapper().QueryRow(query, pointlingID, slot).Scan(
		&pi.PointlingID,
		&pi.ItemID,
		&pi.AcquiredAt,
		&pi.Equipped,
		&pi.Item.Category,
		&pi.Item.Slot,
		&pi.Item.AssetID,
		&pi.Item.Name,
		&pi.Item.Rarity,
		&pi.Item.PricePoints,
		&pi.Item.UnlockLevel,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get equipped item: %w", err)
	}
	pi.Item.ItemID = pi.ItemID
	return pi, nil
}

func (r *PointlingItemRepository) ToggleEquipped(pointlingID, itemID int64, equipped bool) error {
	if equipped {
		// If equipping, first unequip any item in the same slot
		unequipQuery := `
			UPDATE public.pointling_items pi1
			SET equipped = false
			FROM public.items i1
			WHERE pi1.item_id = i1.item_id
			AND pi1.pointling_id = $1
			AND i1.slot = (
				SELECT i2.slot
				FROM public.items i2
				WHERE i2.item_id = $2
			)
			AND pi1.equipped = true`

		if _, err := r.txWrapper().Exec(unequipQuery, pointlingID, itemID); err != nil {
			return fmt.Errorf("unequip current item: %w", err)
		}
	}

	// Now toggle the requested item
	query := `
		UPDATE public.pointling_items
		SET equipped = $3
		WHERE pointling_id = $1
		AND item_id = $2`

	result, err := r.txWrapper().Exec(query, pointlingID, itemID, equipped)
	if err != nil {
		return fmt.Errorf("toggle equipped: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("pointling does not own this item")
	}

	return nil
}
