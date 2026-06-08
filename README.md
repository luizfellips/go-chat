# Go Chat

real-time 1:1 chat portfolio project built with **Go**, **React**, **PostgreSQL**, and **WebSocket**.

everything runs via **Docker**, no need to install Go, Node, or PostgreSQL locally.

## Stack

| Layer | Tech |
|-------|------|
| Backend | Go, Chi, pgx, JWT, WebSocket Hub |
| Frontend | React, TypeScript, Vite, TanStack Query, Zustand, Tailwind |
| Database | PostgreSQL 16 |
| Infra | Docker Compose |

## Quick Start (Docker only)

### Prerequisites

- [Docker](https://www.docker.com)

### Run

```powershell
# From project root
copy .env.example .env
docker compose up --build
```

Open:

- **Frontend:** http://localhost:5173
- **Backend API:** http://localhost:8080/api/v1
- **Health:** http://localhost:8080/health
- **Metrics:** http://localhost:8080/metrics

### Demo users (seeded on first start)

| Email | Password |
|-------|----------|
| alice@example.com | password123 |
| bob@example.com | password123 |

**Try it:** log in as Alice, start a chat with username `bob`, send messages. Open another browser/incognito as Bob to see real-time delivery, online status, and read receipts.

## Development mode (hot reload)

```powershell
docker compose -f docker-compose.yml -f docker-compose.dev.yml up --build
```

- Backend: Air hot reload (Go)
- Frontend: Vite dev server on port 5173

## Useful commands

```powershell
# Stop
docker compose down

# Stop and remove database volume
docker compose down -v

# View logs
docker compose logs -f backend

# Run backend tests inside Docker (uses builder image)
docker compose build backend
```

## Architecture

```
frontend (React) ──REST/JWT──► backend (Go Clean Architecture) ──► PostgreSQL
              └──WebSocket──► Hub (in-memory presence + realtime)
```

### Backend packages (domain-driven)

- `auth/` — JWT, login, register, refresh tokens
- `users/` — profiles and user search
- `conversations/` — 1:1 chat threads
- `messages/` — message history, send, read receipts
- `websocket/` — Hub/Client realtime events
- `apperr/`, `httpx/`, `health/`, `requestctx/` — shared errors, HTTP helpers, health/metrics, request context

### Frontend state

- **Zustand:** auth session, WebSocket status, online presence
- **TanStack Query:** conversations, messages (REST cache)

## API overview

| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/v1/auth/register` | Register |
| POST | `/api/v1/auth/login` | Login |
| POST | `/api/v1/auth/refresh` | Refresh tokens |
| POST | `/api/v1/auth/logout` | Logout |
| GET | `/api/v1/users/me` | Profile |
| GET | `/api/v1/users/search?username=` | Find user |
| GET | `/api/v1/conversations` | List chats |
| POST | `/api/v1/conversations` | Start 1:1 chat |
| GET | `/api/v1/conversations/:id/messages` | Message history |
| POST | `/api/v1/conversations/:id/messages` | Send message |
| POST | `/api/v1/ws/ticket` | Issue short-lived WebSocket ticket (auth required) |
| GET | `/ws/connect?ticket=` | WebSocket (one-time ticket from `/ws/ticket`) |

## Production checklist

- Set `APP_ENV=production` and strong `JWT_ACCESS_SECRET` (32+ chars)
- Set `SEED_DEMO_USERS=false`
- Do not expose Postgres to the host; use TLS for database connections
- Set `METRICS_TOKEN` to protect `/metrics`
- Deploy frontend and backend together after WS ticket changes

## WebSocket events

`connection`, `message_sent`, `message_received`, `message_read`, `user_online`, `user_offline`, `typing_start`, `typing_stop`

## Environment variables

See [`.env.example`](.env.example).

## License

MIT
