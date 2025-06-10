# Pointlings Backend Service

Backend service for the MyPointling project - a gamified feature within the Fetch app allowing users to interact with and customize their own Pointling creatures.

## Features

- User management with point balance tracking
- Pointling creation and customization
- Item/accessory system with rarity levels
- XP-based leveling system
- Integration with Supabase for data persistence

## API Endpoints

### Users

```
GET /api/v1/users
- List users with pagination
- Query params: limit (default: 50, max: 100), offset
- Response: 200 OK with array of users

POST /api/v1/users
- Create new user
- Body: {"user_id": number, "display_name": string}
- Response: 201 Created with user object

GET /api/v1/users/{userID}
- Get user by ID
- Response: 200 OK with user object or 404 Not Found

PATCH /api/v1/users/{userID}/points
- Update user's point balance
- Body: {"new_balance": number}
- Response: 204 No Content or 404 Not Found
```

### Pointlings (To Be Implemented)

```
POST /api/v1/pointlings
- Create new pointling for user
- Body: {"user_id": number, "nickname": string}

GET /api/v1/pointlings/{pointlingID}
- Get pointling details
- Response: Current level, XP, equipped items, etc.

PATCH /api/v1/pointlings/{pointlingID}/xp
- Add XP from activities
- Body: {"xp_amount": number, "source": string}

GET /api/v1/pointlings/{pointlingID}/items
- List pointling's items/accessories
- Query params: equipped (boolean)

POST /api/v1/pointlings/{pointlingID}/items/{itemID}/equip
- Equip/unequip item
- Body: {"equipped": boolean}
```

### Items/Shop (To Be Implemented)

```
GET /api/v1/items
- List available items
- Query params: category, rarity, slot

POST /api/v1/items/purchase
- Purchase item with points
- Body: {"user_id": number, "item_id": number}
```

## Project Structure

```
/cmd/server          - Entry point
/internal
  /models           - Domain models
  /repository       - Data access layer
  /handlers         - HTTP handlers
/pkg/config         - Configuration
/docs               - Documentation
```

## Setup

1. Clone the repository
2. Copy `.env.example` to `.env` and configure:
   ```
   SUPABASE_URL=https://najpaslwftzfnycafwcv.supabase.co
   SUPABASE_SERVICE_KEY=your-key-here
   HTTP_ADDR=:8080
   ```
3. Install dependencies:
   ```
   go mod tidy
   ```
4. Run the server:
   ```
   make run
   ```

## Development

- Build: `go build ./cmd/server`
- Test: `make test`
- Lint: `make lint`
- Docker build: `make docker-build`

## Testing

The project uses table-driven tests with mocked dependencies. Run tests with:

```bash
make test
```

Coverage requirements:

- Handlers: 90%+
- Repositories: 90%+
- Models: 80%+

## Docker

Build the container:

```bash
make docker-build
```

Run with environment variables:

```bash
docker run -p 8080:8080 \
  -e SUPABASE_URL=https://najpaslwftzfnycafwcv.supabase.co \
  -e SUPABASE_SERVICE_KEY=your-key-here \
  pointlings-backend
```
