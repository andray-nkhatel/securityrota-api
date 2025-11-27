# Security Rota API

A Go/Gin REST API for managing security officer shift rotas with weekly rotation.

## Features

- Officer management (CRUD)
- Automatic weekly shift generation based on rotation rules
- Day/Night shift rotation between teams
- Special handling for Sergeant and Female officers
- Swagger API documentation

## Rotation Rules

- **Weekly Rotation**: Teams swap between day and night shifts weekly
- **Week runs**: Sunday to Saturday
- **Sergeant**: Day shift Sun-Fri, off Saturday
- **Female Officer 1**: Day shift Mon-Sat, off Sunday
- **Female Officer 2**: Day shift Sun-Fri, off Saturday
- **Sunday**: Special transition day with reduced day shift (4 officers)
- **Night Shift Mon-Thu**: 2 officers off each day (rotating)

## Setup

### 1. Start PostgreSQL

```bash
docker-compose up -d
```

### 2. Install dependencies

```bash
go mod tidy
```

### 3. Run the API

```bash
go run main.go
```

## API Endpoints

- **Swagger UI**: http://localhost:8080/swagger/index.html

### Officers
- `GET /api/v1/officers` - List all officers
- `GET /api/v1/officers/:id` - Get officer by ID
- `POST /api/v1/officers` - Create officer
- `PUT /api/v1/officers/:id` - Update officer
- `DELETE /api/v1/officers/:id` - Delete officer

### Shifts
- `GET /api/v1/shifts` - Get shifts (filter by date, officer_id, week_start)
- `POST /api/v1/shifts/generate` - Generate rota for a week
- `GET /api/v1/shifts/rotation` - Get week rotation info

## Example: Create Officers

```bash
# Sergeant
curl -X POST http://localhost:8080/api/v1/officers \
  -H "Content-Type: application/json" \
  -d '{"name":"Sgt. Smith","badge_no":"SGT001","role":"sergeant","team":1}'

# Female Officers
curl -X POST http://localhost:8080/api/v1/officers \
  -H "Content-Type: application/json" \
  -d '{"name":"Officer Jane","badge_no":"F001","role":"female","team":1}'

curl -X POST http://localhost:8080/api/v1/officers \
  -H "Content-Type: application/json" \
  -d '{"name":"Officer Mary","badge_no":"F002","role":"female","team":1}'

# Regular Officers (Team 1)
curl -X POST http://localhost:8080/api/v1/officers \
  -H "Content-Type: application/json" \
  -d '{"name":"Officer A","badge_no":"R001","role":"regular","team":1}'

# Regular Officers (Team 2)
curl -X POST http://localhost:8080/api/v1/officers \
  -H "Content-Type: application/json" \
  -d '{"name":"Officer B","badge_no":"R002","role":"regular","team":2}'
```

## Generate Weekly Rota

```bash
curl -X POST http://localhost:8080/api/v1/shifts/generate \
  -H "Content-Type: application/json" \
  -d '{"week_start":"2025-11-30"}'  # Must be a Sunday
```

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| DB_HOST | localhost | PostgreSQL host |
| DB_PORT | 5432 | PostgreSQL port |
| DB_USER | rota | Database user |
| DB_PASSWORD | rotapass | Database password |
| DB_NAME | securityrota | Database name |
| JWT_SECRET | (default) | Secret key for JWT tokens |

## Docker Deployment (Self-Hosted)

### Build the Docker Image

```bash
# Make build script executable
chmod +x build.sh

# Build (requires securityrota-client in sibling directory)
./build.sh
```

### Run with Docker Compose

```bash
# Create .env file for production secrets
cat > .env << EOF
DB_PASSWORD=your-secure-password
JWT_SECRET=your-secure-jwt-secret
EOF

# Start the application
docker-compose -f docker-compose.prod.yml up -d
```

The app will be available at `http://localhost:8080`

### Default Login Credentials

- **Admin**: `admin` / `admin123`
- **User**: `user` / `user123`

### Stop the Application

```bash
docker-compose -f docker-compose.prod.yml down
```

### View Logs

```bash
docker-compose -f docker-compose.prod.yml logs -f app
```

