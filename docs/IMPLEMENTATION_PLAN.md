# Implementation Plan for Remaining Features

## 1. Pointling Core System

### Models and Repositories

```go
// models/pointling.go
type Pointling struct {
    PointlingID   int64     `json:"pointling_id" db:"pointling_id"`
    UserID        int64     `json:"user_id" db:"user_id"`
    Nickname      *string   `json:"nickname" db:"nickname"`
    Level         int       `json:"level" db:"level"`
    CurrentXP     int       `json:"current_xp" db:"current_xp"`
    RequiredXP    int       `json:"required_xp" db:"required_xp"`
    PersonalityID *int      `json:"personality_id" db:"personality_id"`
    LookJSON      JSONMap   `json:"look_json" db:"look_json"`
    CreatedAt     time.Time `json:"created_at" db:"created_at"`
}

type PointlingRepository interface {
    Create(pointling *Pointling) error
    GetByID(id int64) (*Pointling, error)
    GetByUserID(userID int64) ([]*Pointling, error)
    UpdateLook(id int64, look JSONMap) error
    UpdateXP(id int64, currentXP, requiredXP int) error
    UpdateLevel(id int64, level int) error
}
```

### API Endpoints

```
POST /api/v1/pointlings
- Create new pointling
- Requires: user_id, optional nickname
- Returns: 201 with pointling object

GET /api/v1/pointlings/{id}
- Get pointling details
- Returns: 200 with pointling object

GET /api/v1/users/{userId}/pointlings
- List user's pointlings
- Returns: 200 with array of pointlings
```

## 2. XP and Leveling System

### Models and Repositories

```go
// models/xp_event.go
type XPEventSource string

const (
    XPSourceReceipt  XPEventSource = "RECEIPT"
    XPSourcePlay     XPEventSource = "PLAY"
    XPSourceDaily    XPEventSource = "DAILY"
)

type XPEvent struct {
    EventID     int64         `json:"event_id" db:"event_id"`
    PointlingID int64         `json:"pointling_id" db:"pointling_id"`
    Source      XPEventSource `json:"source" db:"source"`
    XPAmount    int          `json:"xp_amount" db:"xp_amount"`
    EventTS     time.Time    `json:"event_ts" db:"event_ts"`
}

type XPRepository interface {
    AddXP(event *XPEvent) error
    GetEventsByPointling(pointlingID int64, limit int) ([]*XPEvent, error)
    GetDailyXPTotal(pointlingID int64) (int, error)
}
```

### API Endpoints

```
POST /api/v1/pointlings/{id}/xp
- Add XP from activity
- Body: {"source": "RECEIPT|PLAY|DAILY", "amount": number}
- Validates daily limits
- Handles level-up logic
- Returns: 200 with updated pointling

GET /api/v1/pointlings/{id}/xp/history
- Get XP gain history
- Query params: limit (default 50)
- Returns: 200 with array of XP events
```

## 3. Items and Customization

### Models and Repositories

```go
// models/item.go
type ItemCategory string
type ItemSlot string
type ItemRarity string

const (
    CategoryAccessory ItemCategory = "ACCESSORY"
    CategoryFeature   ItemCategory = "FEATURE"

    SlotHat      ItemSlot = "HAT"
    SlotShoes    ItemSlot = "SHOES"
    SlotFace     ItemSlot = "FACE"
    SlotWings    ItemSlot = "WINGS"

    RarityCommon    ItemRarity = "COMMON"
    RarityRare      ItemRarity = "RARE"
    RarityEpic      ItemRarity = "EPIC"
    RarityLegendary ItemRarity = "LEGENDARY"
)

type Item struct {
    ItemID      int64        `json:"item_id" db:"item_id"`
    Category    ItemCategory `json:"category" db:"category"`
    Slot        *ItemSlot    `json:"slot,omitempty" db:"slot"`
    AssetID     string       `json:"asset_id" db:"asset_id"`
    Name        string       `json:"name" db:"name"`
    Rarity      ItemRarity   `json:"rarity" db:"rarity"`
    PricePoints *int         `json:"price_points,omitempty" db:"price_points"`
    UnlockLevel *int         `json:"unlock_level,omitempty" db:"unlock_level"`
}

type ItemRepository interface {
    Create(item *Item) error
    GetByID(id int64) (*Item, error)
    List(category *ItemCategory, rarity *ItemRarity, slot *ItemSlot) ([]*Item, error)
    GetUnlocksForLevel(level int) ([]*Item, error)
}

// models/pointling_item.go
type PointlingItem struct {
    PointlingID int64     `json:"pointling_id" db:"pointling_id"`
    ItemID      int64     `json:"item_id" db:"item_id"`
    AcquiredAt  time.Time `json:"acquired_at" db:"acquired_at"`
    Equipped    bool      `json:"equipped" db:"equipped"`
}

type PointlingItemRepository interface {
    AddItem(pointlingID, itemID int64) error
    GetItems(pointlingID int64, equipped *bool) ([]*PointlingItem, error)
    ToggleEquipped(pointlingID, itemID int64, equipped bool) error
}
```

### API Endpoints

```
GET /api/v1/items
- List available items
- Query params: category, rarity, slot
- Returns: 200 with array of items

POST /api/v1/pointlings/{id}/items/{itemId}
- Purchase/acquire item
- Validates point cost or level requirement
- Returns: 200 with updated point balance

PATCH /api/v1/pointlings/{id}/items/{itemId}/equip
- Toggle item equipped status
- Body: {"equipped": boolean}
- Returns: 200 with updated pointling look

GET /api/v1/pointlings/{id}/items
- List pointling's items
- Query params: equipped (boolean, optional)
- Returns: 200 with array of items
```

## 4. Point Management

### Models and Repositories

```go
// models/point_spend.go
type PointSpend struct {
    SpendID     int64     `json:"spend_id" db:"spend_id"`
    UserID      int64     `json:"user_id" db:"user_id"`
    ItemID      int64     `json:"item_id" db:"item_id"`
    PointsSpent int       `json:"points_spent" db:"points_spent"`
    SpendTS     time.Time `json:"spend_ts" db:"spend_ts"`
}

type PointSpendRepository interface {
    Create(spend *PointSpend) error
    GetByUser(userID int64) ([]*PointSpend, error)
}
```

### API Endpoints

```
POST /api/v1/users/{id}/points/spend
- Spend points on item
- Body: {"item_id": number, "points": number}
- Validates sufficient balance
- Creates point_spend record
- Returns: 200 with updated balance

GET /api/v1/users/{id}/points/history
- Get point spend history
- Query params: limit, offset
- Returns: 200 with array of transactions
```

## Implementation Order

1. Core Pointling Management

   - Pointling model and repository
   - Basic CRUD endpoints
   - Pointling creation flow

2. XP and Leveling System

   - XP event tracking
   - Level-up calculations
   - XP source validation
   - Daily limits

3. Item System

   - Item catalog
   - Purchase validation
   - Inventory management
   - Equipment system

4. Point Transactions
   - Spend tracking
   - Balance updates
   - Purchase history

Each component should follow the established patterns:

- Table-driven tests
- Mocked repositories
- Proper error handling
- Input validation
- Structured responses
- Clear documentation

## Testing Strategy

1. Unit Tests

   - Repository methods
   - Business logic
   - Input validation
   - Error handling

2. Integration Tests

   - API endpoints
   - Database operations
   - Transaction handling

3. Load Tests
   - Concurrent requests
   - Database performance
   - Cache effectiveness

## Documentation Updates

After implementation:

1. Update README.md with all endpoints
2. Add sequence diagrams for key flows
3. Document rate limits and validation rules
4. Add example requests/responses
