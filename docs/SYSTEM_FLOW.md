# System Flow Diagrams

## Pointling Creation and Customization Flow

```mermaid
sequenceDiagram
    participant C as Client
    participant API as API Server
    participant DB as Database

    C->>API: POST /api/v1/pointlings
    Note over API: Validate user exists
    API->>DB: Create pointling record
    API->>DB: Add default color
    API-->>C: Return new pointling

    C->>API: GET /api/v1/items?category=ACCESSORY
    API->>DB: Query available items
    API-->>C: Return item catalog

    C->>API: POST /pointlings/{id}/items/{itemId}
    Note over API: Validate points & requirements
    API->>DB: Deduct user points
    API->>DB: Add item to pointling
    API-->>C: Return updated pointling
```

## XP and Leveling Flow

```mermaid
sequenceDiagram
    participant C as Client
    participant API as API Server
    participant DB as Database

    C->>API: POST /pointlings/{id}/xp
    Note over API: Validate daily limits
    API->>DB: Record XP event
    API->>DB: Update current XP

    alt Level Up Required
        API->>DB: Update level
        API->>DB: Query level rewards
        API-->>C: Return level-up options
        C->>API: POST /pointlings/{id}/rewards
        API->>DB: Grant chosen reward
    end

    API-->>C: Return updated pointling
```

## Item Purchase Flow

```mermaid
sequenceDiagram
    participant C as Client
    participant API as API Server
    participant DB as Database

    C->>API: GET /api/v1/items
    API->>DB: Query available items
    API-->>C: Return filtered items

    C->>API: POST /users/{id}/points/spend
    Note over API: Validate points balance
    API->>DB: Begin transaction
    API->>DB: Deduct points
    API->>DB: Record purchase
    API->>DB: Add item to inventory
    API->>DB: Commit transaction
    API-->>C: Return success
```

## Data Model Relationships

```mermaid
erDiagram
    users ||--o{ pointlings : "owns"
    pointlings ||--o{ pointling_items : "has"
    pointlings ||--o{ pointling_colors : "has"
    pointlings ||--o{ xp_events : "earns"
    items ||--o{ pointling_items : "equipped_as"
    users ||--o{ point_spend : "makes"
    items ||--o{ point_spend : "purchased_in"
```

## Key Business Rules

1. XP and Leveling

   - Daily XP caps per source type
   - Linear XP growth (3 to 120)
   - Level rewards (3 options, pick 1)
   - Color unlocks every 5 levels

2. Points and Items

   - Points can only be spent on available items
   - Items may require specific levels
   - Each slot can have one equipped item
   - Colors don't count as equipped items

3. Customization

   - Features are permanent once unlocked
   - Accessories can be toggled
   - Base customization is always available
   - Some items may be limited time only

4. Transaction Safety
   - Point spends must be atomic
   - XP calculations must be race-condition safe
   - Item equipped status must be consistent
   - Level-up rewards must be exactly one choice
