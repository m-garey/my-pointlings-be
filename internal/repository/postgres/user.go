package postgres

import (
	"database/sql"
	"fmt"

	"github.com/pointlings/backend/internal/models"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) GetUser(userID int64) (*models.User, error) {
	query := `
		SELECT user_id, display_name, point_balance, created_at
		FROM public.users
		WHERE user_id = $1`

	user := &models.User{}
	err := r.db.QueryRow(query, userID).Scan(
		&user.UserID,
		&user.DisplayName,
		&user.PointBalance,
		&user.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}
	return user, nil
}

func (r *UserRepository) CreateUser(user *models.User) error {
	query := `
		INSERT INTO public.users (user_id, display_name, point_balance)
		VALUES ($1, $2, $3)`

	_, err := r.db.Exec(query,
		user.UserID,
		user.DisplayName,
		user.PointBalance,
	)
	if err != nil {
		return fmt.Errorf("create user: %w", err)
	}
	return nil
}

func (r *UserRepository) UpdatePointBalance(userID int64, newBalance int64) error {
	query := `
		UPDATE public.users 
		SET point_balance = $2
		WHERE user_id = $1`

	result, err := r.db.Exec(query, userID, newBalance)
	if err != nil {
		return fmt.Errorf("update point balance: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("user not found: %d", userID)
	}
	return nil
}

func (r *UserRepository) ListUsers(limit, offset int) ([]*models.User, error) {
	query := `
		SELECT user_id, display_name, point_balance, created_at
		FROM public.users
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2`

	rows, err := r.db.Query(query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list users query: %w", err)
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		user := &models.User{}
		err := rows.Scan(
			&user.UserID,
			&user.DisplayName,
			&user.PointBalance,
			&user.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan user row: %w", err)
		}
		users = append(users, user)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate users: %w", err)
	}
	return users, nil
}
