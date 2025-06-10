package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/pointlings/backend/internal/models"
)

type XPRepository struct {
	db *sql.DB
	tx *sql.Tx // For transaction support
}

func NewXPRepository(db *sql.DB) *XPRepository {
	return &XPRepository{db: db}
}

// txWrapper ensures operations use a transaction if one exists
func (r *XPRepository) txWrapper() interface {
	QueryRow(query string, args ...interface{}) *sql.Row
	Exec(query string, args ...interface{}) (sql.Result, error)
	Query(query string, args ...interface{}) (*sql.Rows, error)
} {
	if r.tx != nil {
		return r.tx
	}
	return r.db
}

func (r *XPRepository) InTransaction(fn func(models.XPRepository) error) error {
	tx, err := r.db.BeginTx(context.Background(), nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}

	// Create new repository with transaction
	txRepo := &XPRepository{db: r.db, tx: tx}

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

func (r *XPRepository) AddXP(event *models.XPEvent) error {
	// First check if adding this XP would exceed daily limits
	currentDaily, err := r.GetDailyXPBySource(event.PointlingID, event.Source)
	if err != nil {
		return fmt.Errorf("check daily xp: %w", err)
	}

	if currentDaily+event.XPAmount > event.Source.GetMaxDailyXP() {
		return models.ErrDailyXPLimitReached
	}

	// Insert XP event
	query := `
		INSERT INTO public.xp_events (
			pointling_id, source, xp_amount
		) VALUES ($1, $2, $3)
		RETURNING event_id, event_ts`

	err = r.txWrapper().QueryRow(
		query,
		event.PointlingID,
		event.Source,
		event.XPAmount,
	).Scan(&event.EventID, &event.EventTS)

	if err != nil {
		return fmt.Errorf("insert xp event: %w", err)
	}

	// Update pointling's XP
	updateQuery := `
		UPDATE public.pointlings
		SET current_xp = current_xp + $2
		WHERE pointling_id = $1
		RETURNING current_xp`

	var newCurrentXP int
	err = r.txWrapper().QueryRow(updateQuery, event.PointlingID, event.XPAmount).Scan(&newCurrentXP)
	if err != nil {
		return fmt.Errorf("update pointling xp: %w", err)
	}

	return nil
}

func (r *XPRepository) GetEventsByPointling(pointlingID int64, limit int) ([]*models.XPEvent, error) {
	query := `
		SELECT event_id, pointling_id, source, xp_amount, event_ts
		FROM public.xp_events
		WHERE pointling_id = $1
		ORDER BY event_ts DESC
		LIMIT $2`

	rows, err := r.txWrapper().Query(query, pointlingID, limit)
	if err != nil {
		return nil, fmt.Errorf("query xp events: %w", err)
	}
	defer rows.Close()

	var events []*models.XPEvent
	for rows.Next() {
		event := &models.XPEvent{}
		err := rows.Scan(
			&event.EventID,
			&event.PointlingID,
			&event.Source,
			&event.XPAmount,
			&event.EventTS,
		)
		if err != nil {
			return nil, fmt.Errorf("scan xp event: %w", err)
		}
		events = append(events, event)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate xp events: %w", err)
	}

	return events, nil
}

func (r *XPRepository) GetDailyXPBySource(pointlingID int64, source models.XPEventSource) (int, error) {
	query := `
		SELECT COALESCE(SUM(xp_amount), 0)
		FROM public.xp_events
		WHERE pointling_id = $1
		AND source = $2
		AND event_ts >= CURRENT_DATE
		AND event_ts < CURRENT_DATE + INTERVAL '1 day'`

	var totalXP int
	err := r.txWrapper().QueryRow(query, pointlingID, source).Scan(&totalXP)
	if err != nil {
		return 0, fmt.Errorf("get daily xp: %w", err)
	}

	return totalXP, nil
}
