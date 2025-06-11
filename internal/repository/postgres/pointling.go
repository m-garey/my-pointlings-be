package postgres

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"my-pointlings-be/internal/models"
)

type PointlingRepository struct {
	db *sql.DB
}

func NewPointlingRepository(db *sql.DB) *PointlingRepository {
	return &PointlingRepository{db: db}
}

func (r *PointlingRepository) Create(pointling *models.Pointling) error {
	query := `
		INSERT INTO public.pointlings (
			user_id, nickname, level, current_xp, required_xp,
			personality_id, look_json
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING pointling_id, created_at`

	lookJSON, err := json.Marshal(pointling.LookJSON)
	if err != nil {
		return fmt.Errorf("marshal look_json: %w", err)
	}

	err = r.db.QueryRow(
		query,
		pointling.UserID,
		pointling.Nickname,
		pointling.Level,
		pointling.CurrentXP,
		pointling.RequiredXP,
		pointling.PersonalityID,
		lookJSON,
	).Scan(&pointling.PointlingID, &pointling.CreatedAt)

	if err != nil {
		return fmt.Errorf("create pointling: %w", err)
	}
	return nil
}

func (r *PointlingRepository) GetByID(id int64) (*models.Pointling, error) {
	query := `
		SELECT pointling_id, user_id, nickname, level, current_xp,
			required_xp, personality_id, look_json, created_at
		FROM public.pointlings
		WHERE pointling_id = $1`

	pointling := &models.Pointling{}
	var lookJSON []byte

	err := r.db.QueryRow(query, id).Scan(
		&pointling.PointlingID,
		&pointling.UserID,
		&pointling.Nickname,
		&pointling.Level,
		&pointling.CurrentXP,
		&pointling.RequiredXP,
		&pointling.PersonalityID,
		&lookJSON,
		&pointling.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get pointling: %w", err)
	}

	if err := json.Unmarshal(lookJSON, &pointling.LookJSON); err != nil {
		return nil, fmt.Errorf("unmarshal look_json: %w", err)
	}

	return pointling, nil
}

func (r *PointlingRepository) GetByUserID(userID int64) ([]*models.Pointling, error) {
	query := `
		SELECT pointling_id, user_id, nickname, level, current_xp,
			required_xp, personality_id, look_json, created_at
		FROM public.pointlings
		WHERE user_id = $1
		ORDER BY created_at DESC`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("query pointlings: %w", err)
	}
	defer rows.Close()

	var pointlings []*models.Pointling
	for rows.Next() {
		pointling := &models.Pointling{}
		var lookJSON []byte

		err := rows.Scan(
			&pointling.PointlingID,
			&pointling.UserID,
			&pointling.Nickname,
			&pointling.Level,
			&pointling.CurrentXP,
			&pointling.RequiredXP,
			&pointling.PersonalityID,
			&lookJSON,
			&pointling.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan pointling: %w", err)
		}

		if err := json.Unmarshal(lookJSON, &pointling.LookJSON); err != nil {
			return nil, fmt.Errorf("unmarshal look_json: %w", err)
		}

		pointlings = append(pointlings, pointling)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate pointlings: %w", err)
	}

	return pointlings, nil
}

func (r *PointlingRepository) UpdateLook(id int64, look models.JSONMap) error {
	query := `
		UPDATE public.pointlings
		SET look_json = $2
		WHERE pointling_id = $1`

	lookJSON, err := json.Marshal(look)
	if err != nil {
		return fmt.Errorf("marshal look_json: %w", err)
	}

	result, err := r.db.Exec(query, id, lookJSON)
	if err != nil {
		return fmt.Errorf("update look: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("pointling not found: %d", id)
	}

	return nil
}

func (r *PointlingRepository) UpdateXP(id int64, currentXP, requiredXP int) error {
	query := `
		UPDATE public.pointlings
		SET current_xp = $2, required_xp = $3
		WHERE pointling_id = $1`

	result, err := r.db.Exec(query, id, currentXP, requiredXP)
	if err != nil {
		return fmt.Errorf("update xp: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("pointling not found: %d", id)
	}

	return nil
}

func (r *PointlingRepository) UpdateLevel(id int64, level int) error {
	query := `
		UPDATE public.pointlings
		SET level = $2
		WHERE pointling_id = $1`

	result, err := r.db.Exec(query, id, level)
	if err != nil {
		return fmt.Errorf("update level: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("pointling not found: %d", id)
	}

	return nil
}

func (r *PointlingRepository) UpdateNickname(id int64, nickname *string) error {
	query := `
		UPDATE public.pointlings
		SET nickname = $2
		WHERE pointling_id = $1`

	result, err := r.db.Exec(query, id, nickname)
	if err != nil {
		return fmt.Errorf("update nickname: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("pointling not found: %d", id)
	}

	return nil
}
