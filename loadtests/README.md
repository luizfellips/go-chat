# HTTP Load Tests (k6)

Validates the REST API before WebSocket tests: **login**, **conversation creation**, and **message retrieval**.

## Prerequisites

- Docker Desktop
- Project stack running with login rate limit disabled for load testing

## Start environment for load test

```powershell
copy .env.example .env
make load-up
```

The `docker-compose.load.yml` overlay sets `LOGIN_RATE_LIMIT_RPM=0` on the backend (default in production/dev: 5 req/min per IP).

## Run tests

| Command | Scenario | Description |
|---------|----------|-------------|
| `make load-test` | smoke | 1 VU, full flow |
| `make load-test-100` | load_100 | ramp-up 30s → 100 VUs → 2min → ramp-down |
| `make load-test-1000` | load_1000 | ramp-up 2min → 1000 VUs → 3min → ramp-down |
| `make load-test-ramp` | ramp_up | gradual 0 → 50 → 100 → 200 VUs |
| `make load-test-login` | load_100 | `POST /auth/login` only |

### Tested flow (`http-flow.js`)

1. `POST /api/v1/auth/login` (alice)
2. `POST /api/v1/conversations` (with bob)
3. `GET /api/v1/conversations`
4. `POST /api/v1/conversations/{id}/messages`
5. `GET /api/v1/conversations/{id}/messages`

Seed users: `alice@example.com` / `password123` and `bob@example.com` / `password123`.

## Variables (Docker)

| Variable | Docker default |
|----------|----------------|
| `BASE_URL` | `http://backend:8080/api/v1` |
| `SCENARIO` | `smoke` |
| `ALICE_EMAIL` | `alice@example.com` |
| `ALICE_PASSWORD` | `password123` |

## Available scenarios

Defined in `loadtests/lib/scenarios.js`:

- **smoke** — quick validation (1 iteration)
- **load_100** — 100 concurrent virtual users
- **load_1000** — 1000 virtual users
- **ramp_up** — gradual ramp-up (50 → 100 → 200)

## Manual Docker command

```powershell
docker compose -f docker-compose.yml -f docker-compose.load.yml --profile loadtest run --rm k6 run /scripts/http-flow.js -e SCENARIO=load_100
```

The `loadtest` profile prevents k6 from starting with `docker compose up`. The backend starts with `LOGIN_RATE_LIMIT_RPM=0` via the overlay.

## Other tools

The same endpoints can be exercised with **JMeter** or **Locust**. The flow and payloads are documented in the main README and in `loadtests/lib/api.js`.

## Next step

Use the **load simulator** (`backend/cmd/simulator`) for realistic chat simulation.

---

# Load Simulator (recommended)

Realistic simulation of users in 1:1 chat with latency measurement.

```
backend/cmd/
 ├─ server/
 └─ simulator/
```

## Flow

1. Creates N fake users (`simuser-0001@loadtest.local`, …)
2. Creates 1:1 conversations (user pairs)
3. Login + WebSocket for **all** users (online presence)
4. Sends random messages at the configured rate
5. Measures **delivery** latency (sender → peer) and **round-trip** latency (echo to sender)

## Default configuration

| Parameter | Default |
|-----------|---------|
| `users` | 100 |
| `conversations` | 20 (40 active users in chat) |
| `messages_per_second` | 50 |
| `duration` | 5m |

## Run (Docker)

```powershell
make load-up
make simulate              # 100 users, 20 convs, 50 msg/s, 5min
make simulate-quick        # 20 users, 5 convs, 10 msg/s, 30s
```

Local (with Go in the backend):

```powershell
cd backend
go run ./cmd/simulator -users 100 -conversations 20 -messages-per-second 50 -duration 5m
```

## Expected output

```
[progress] connected=100 sent=250 delivered=248 round_trips=248 errors=0
  delivery_latency: count=248 avg=12ms p50=10ms p95=28ms p99=45ms max=120ms
  round_trip_latency: count=248 avg=8ms p50=7ms p95=18ms p99=32ms max=95ms
```

Code in [`backend/internal/simulator/`](../backend/internal/simulator/).

---

# WebSocket Load Tests (simple bots)

Simulates **N bots** that authenticate via REST, open a WebSocket, and exchange messages with a peer user (default: **bob**).

Each bot:

1. Registers a unique user (`wsbot-0001@loadtest.local`, …)
2. Logs in → `access_token`
3. Creates a 1:1 conversation with the peer
4. Connects to `ws://…/ws/connect?token=…`
5. Sends `message_sent` at each interval
6. Counts `message_received` in the read loop

> **Important:** the hub accepts **1 WebSocket connection per user**. That is why each bot uses a different user (you cannot run 100 bots as `alice`).

## Run (Docker)

```powershell
make load-up
make ws-test        # 10 bots, 30s (quick validation)
make ws-test-100    # 100 bots, ~100 msg/s (interval=1s)
make ws-test-1000   # 1000 bots, ramp 2min, duration 5min
```

Manual command:

```powershell
docker compose -f docker-compose.yml -f docker-compose.load.yml --profile loadtest run --rm --build wsbots -bots 100 -interval 1s -ramp 30s -duration 3m
```

## Variables

| Variable / flag | Docker default | Description |
|-----------------|----------------|-------------|
| `BOTS` / `-bots` | `100` | concurrent bots |
| `INTERVAL` / `-interval` | `1s` | interval between messages per bot |
| `RAMP` / `-ramp` | `30s` | gradual connection ramp-up |
| `DURATION` / `-duration` | `3m` | duration (`0` = until Ctrl+C) |
| `API_URL` / `-api` | `http://backend:8080/api/v1` | REST |
| `WS_URL` / `-ws` | `ws://backend:8080/ws/connect` | WebSocket |
| `PEER_USERNAME` / `-peer` | `bob` | target user for conversations |

## Expected output

```
[progress] connected=100 sent=420 received=380 errors=0
[done] connected=0 sent=18000 received=17500 errors=0
```

With 100 bots and `-interval 1s`: ~**100 messages/second** sent.

Code in [`loadtests/wsbots/`](wsbots/).
