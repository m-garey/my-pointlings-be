package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/pointlings/backend/internal/models"
)

type PointSpendRepository struct {
	db *sql.DB
	tx *sql.Tx
}

func NewPointSpendRepository(db *sql.DB) *PointSpendRepository {
	return &PointSpendRepository{db: db}
}

func (r *PointSpendRepository) txWrapper() interface {
	QueryRow(query string, args ...interface{}) *sql.Row
	Exec(query string, args ...interface{}) (sql.Result, error)
	Query(query string, args ...interface{}) (*sql.Rows, error)
} {
	if r.tx != nil {
		return r.tx
	}
	return r.db
}

func (r *PointSpendRepository) InTransaction(fn func(models.PointSpendRepository) error) error {
	tx, err := r.db.BeginTx(context.Background(), nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}

	txRepo := &PointSpendRepository{db: r.db, tx: tx}

	if err := fn(txRepo); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("rollback failed: %v (original error: %v)", rbErr, err)
		}
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}
	return nil
}

func (r *PointSpendRepository) Create(spend *models.PointSpend) error {
	query := `
		INSERT INTO public.point_spend (
			user_id, item_id, points_spent
		) VALUES ($1, $2, $3)
		RETURNING spend_id, spend_ts`

	err := r.txWrapper().QueryRow(
		query,
		spend.UserID,
		spend.ItemID,
		spend.PointsSpent,
	).Scan(&spend.SpendID, &spend.SpendTS)

	if err != nil {
		return fmt.Errorf("create spend record: %w", err)
	}
	return nil
}

func (r *PointSpendRepository) GetByUser(userID int64, limit, offset int) ([]*models.PointSpend, error) {
	query := `
		SELECT ps.spend_id, ps.user_id, ps.item_id, ps.points_spent, ps.spend_ts,
			   i.category, i.slot, i.asset_id, i.name, i.rarity, i.price_points, i.unlock_level
		FROM public.point_spend ps
		JOIN public.items i ON i.item_id = ps.item_id
		WHERE ps.user_id = $1
		ORDER BY ps.spend_ts DESC
		LIMIT $2 OFFSET $3`

	rows, err := r.txWrapper().Query(query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("get user spends query: %w", err)
	}
	defer rows.Close()

	var spends []*models.PointSpend
	for rows.Next() {
		spend := &models.PointSpend{
			Item: &models.Item{},
		}
		err := rows.Scan(
			&spend.SpendID,
			&spend.UserID,
			&spend.ItemID,
			&spend.PointsSpent,
			&spend.SpendTS,
			&spend.Item.Category,
			&spend.Item.Slot,
			&spend.Item.AssetID,
			&spend.Item.Name,
			&spend.Item.Rarity,
			&spend.Item.PricePoints,
			&spend.Item.UnlockLevel,
		)
		if err != nil {
			return nil, fmt.Errorf("scan spend record: %w", err)
		}
		spend.Item.ItemID = spend.ItemID
		spends = append(spends, spend)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate spends: %w", err)
	}
	return spends, nil
}

func (r *PointSpendRepository) GetTotalSpentByUser(userID int64) (int64, error) {
	query := `
		SELECT COALESCE(SUM(points_spent), 0)
		FROM public.point_spend
		WHERE user_id = $1`

	var total int64
	err := r.txWrapper().QueryRow(query, userID).Scan(&total)
	if err != nil {
		return 0, fmt.Errorf("get total spent: %w", err)
	}
	return total, nil
}

func (r *PointSpendRepository) SpendPoints(userID int64, itemID int64, points int) error {
	// First check and update user balance atomically
	balanceQuery := `
		UPDATE public.users
		SET point_balance = point_balance - $2
		WHERE user_id = $1
		AND point_balance >= $2
		RETURNING point_balance`

	var newBalance int64
	err := r.txWrapper().QueryRow(balanceQuery, userID, points).Scan(&newBalance)
	if err == sql.ErrNoRows {
		return models.ErrInsufficientBalance
	}
	if err != nil {
		return fmt.Errorf("update balance: %w", err)
	}

	// Create spend record
	spend := &models.PointSpend{
		UserID:      userID,
		ItemID:      itemID,
		PointsSpent: points,
	}
	if err := r.Create(spend); err != nil {
		return fmt.Errorf("create spend record: %w", err)
	}

	return nil
}
