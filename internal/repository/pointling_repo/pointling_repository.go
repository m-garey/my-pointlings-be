package pointling_repo

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"my-pointlings-be/internal/models"
)

var _ API = (*PointlingRepository)(nil)

type PointlingRepository struct {
	db *sql.DB
	tx *sql.Tx
}

type API interface {
	// InTransaction executes operations in a transaction
	InTransaction(fn func(*PointlingRepository) error) error

	// Create stores a new pointling
	CreatePointling(pointling *models.Pointling) error

	// GetByID retrieves a pointling by its ID
	GetPointlingByID(id int64) (*models.Pointling, error)

	// GetByUserID retrieves all pointlings owned by a user
	GetPointlingByUserID(userID int64) ([]*models.Pointling, error)

	// UpdateLook updates a pointling's appearance
	UpdatePointlingLook(id int64, look models.JSONMap) error

	// UpdateXP updates a pointling's XP and required XP values
	UpdatePointlingXP(id int64, currentXP, requiredXP int) error

	// UpdateLevel updates a pointling's level
	UpdatePointlingLevel(id int64, level int) error

	// UpdateNickname sets a pointling's nickname
	UpdatePointlingNickname(id int64, nickname *string) error

	// AddXP creates a new XP event and updates the pointling's XP
	AddXP(event *models.XPEvent) error

	// GetEventsByPointling retrieves recent XP events for a pointling
	GetEventsByPointling(pointlingID int64, limit int) ([]*models.XPEvent, error)

	// GetDailyXPBySource gets total XP gained from a source today
	GetDailyXPBySource(pointlingID int64, source models.XPEventSource) (int, error)

	// GetUser retrieves a user by ID
	GetUser(userID int64) (*models.User, error)

	// CreateUser creates a new user
	CreateUser(user *models.User) error

	// UpdatePointBalance updates a user's point balance
	UpdatePointBalance(userID int64, newBalance int64) error

	// ListUsers retrieves all users with optional limit/offset pagination
	ListUsers(limit, offset int) ([]*models.User, error)

	// Create records a new point spend transaction
	CreatePointSpend(spend *models.PointSpend) error

	// GetTotalSpentByUser gets total points spent by a user
	GetTotalSpentByUser(userID int64) (int64, error)

	// SpendPoints atomically updates user balance and creates spend record
	SpendPoints(userID int64, itemID int64, points int) error

	// GetByID retrieves an item by its ID
	GetItemByID(id int64) (*models.Item, error)

	// List retrieves items with optional filters
	ListItems(category *models.ItemCategory, rarity *models.ItemRarity, slot *models.ItemSlot) ([]*models.Item, error)

	// GetUnlocksForLevel gets items available at a specific level
	GetUnlocksForLevel(level int) ([]*models.Item, error)

	// Create creates a new item (admin only)
	CreateItem(item *models.Item) error

	// AddItem gives an item to a pointling
	AddItem(pointlingID, itemID int64) error

	// GetItems lists items owned by a pointling
	GetItems(pointlingID int64, equipped *bool) ([]*models.PointlingItem, error)

	// ToggleEquipped equips or unequips an item
	ToggleEquipped(pointlingID, itemID int64, equipped bool) error

	// GetEquippedInSlot gets the currently equipped item in a slot
	GetEquippedInSlot(pointlingID int64, slot models.ItemSlot) (*models.PointlingItem, error)
}

func New(db *sql.DB) *PointlingRepository {
	return &PointlingRepository{db: db}
}

func (r *PointlingRepository) txWrapper() interface {
	QueryRow(query string, args ...interface{}) *sql.Row
	Exec(query string, args ...interface{}) (sql.Result, error)
	Query(query string, args ...interface{}) (*sql.Rows, error)
} {
	if r.tx != nil {
		return r.tx
	}
	return r.db
}

func (r *PointlingRepository) InTransaction(fn func(*PointlingRepository) error) error {
	tx, err := r.db.BeginTx(context.Background(), nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}

	txRepo := &PointlingRepository{db: r.db, tx: tx}

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

func (r *PointlingRepository) CreatePointling(pointling *models.Pointling) error {
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

func (r *PointlingRepository) GetPointlingByID(id int64) (*models.Pointling, error) {
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

func (r *PointlingRepository) GetPointlingByUserID(userID int64) ([]*models.Pointling, error) {
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

func (r *PointlingRepository) UpdatePointlingLook(id int64, look models.JSONMap) error {
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

func (r *PointlingRepository) UpdatePointlingXP(id int64, currentXP, requiredXP int) error {
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

func (r *PointlingRepository) UpdatePointlingLevel(id int64, level int) error {
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

func (r *PointlingRepository) UpdatePointlingNickname(id int64, nickname *string) error {
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

func (r *PointlingRepository) GetUser(userID int64) (*models.User, error) {
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

func (r *PointlingRepository) CreateUser(user *models.User) error {
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

func (r *PointlingRepository) UpdatePointBalance(userID int64, newBalance int64) error {
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

func (r *PointlingRepository) ListUsers(limit, offset int) ([]*models.User, error) {
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

func (r *PointlingRepository) AddXP(event *models.XPEvent) error {
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

func (r *PointlingRepository) GetEventsByPointling(pointlingID int64, limit int) ([]*models.XPEvent, error) {
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

func (r *PointlingRepository) GetDailyXPBySource(pointlingID int64, source models.XPEventSource) (int, error) {
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

func (r *PointlingRepository) AddItem(pointlingID, itemID int64) error {
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

func (r *PointlingRepository) GetItems(pointlingID int64, equipped *bool) ([]*models.PointlingItem, error) {
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

func (r *PointlingRepository) GetEquippedInSlot(pointlingID int64, slot models.ItemSlot) (*models.PointlingItem, error) {
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

func (r *PointlingRepository) ToggleEquipped(pointlingID, itemID int64, equipped bool) error {
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

func (r *PointlingRepository) CreatePointSpend(spend *models.PointSpend) error {
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

func (r *PointlingRepository) GetByUser(userID int64, limit, offset int) ([]*models.PointSpend, error) {
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

func (r *PointlingRepository) GetTotalSpentByUser(userID int64) (int64, error) {
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

func (r *PointlingRepository) SpendPoints(userID int64, itemID int64, points int) error {
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
	if err := r.CreatePointSpend(spend); err != nil {
		return fmt.Errorf("create spend record: %w", err)
	}

	return nil
}

func (r *PointlingRepository) GetItemByID(id int64) (*models.Item, error) {
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

func (r *PointlingRepository) List(category *models.ItemCategory, rarity *models.ItemRarity, slot *models.ItemSlot) ([]*models.Item, error) {
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

func (r *PointlingRepository) GetUnlocksForLevel(level int) ([]*models.Item, error) {
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

func (r *PointlingRepository) CreateItem(item *models.Item) error {
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

func (r *PointlingRepository) ListItems(category *models.ItemCategory, rarity *models.ItemRarity, slot *models.ItemSlot) ([]*models.Item, error) {
	query := "SELECT item_id, category, slot, asset_id, name, rarity, price_points, unlock_level FROM items WHERE 1=1"
	args := []interface{}{}

	if category != nil {
		query += " AND category = $" + fmt.Sprint(len(args)+1)
		args = append(args, *category)
	}

	if rarity != nil {
		query += " AND rarity = $" + fmt.Sprint(len(args)+1)
		args = append(args, *rarity)
	}

	if slot != nil {
		query += " AND slot = $" + fmt.Sprint(len(args)+1)
		args = append(args, *slot)
	}

	query += " ORDER BY name ASC"

	rows, err := r.db.Query(query, args...)
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
		return nil, fmt.Errorf("iterate item rows: %w", err)
	}

	return items, nil
}
