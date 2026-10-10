# SMS Gateway

## Binaries

| Binary | Source | Role | Queue |
| --- | --- | --- | --- |
| `smsgateway` | [main.go](main.go) | HTTP API, applies DB migrations on startup | — |
| `normal-sms-worker` | [cmd/normal-sms-worker](cmd/normal-sms-worker/main.go) | Normal SMS delivery | `sms.normal` |
| `express-sms-worker` | [cmd/express-sms-worker](cmd/express-sms-worker/main.go) | Express SMS delivery | `sms.express` |

The queue a worker consumes is **fixed in its binary** — no command-line arguments. This is
why the two workers can be deployed and scaled independently.

## Configuration

Configuration is read from environment variables. Locally, a `.env` file at the repository
root is loaded automatically when present (it is optional — in containers the variables are
injected by the runtime). See [.env.example](.env.example) for the full list.

| Variable | Default | Description |
| --- | --- | --- |
| `HTTP_HOST` | `127.0.0.1` | API bind address. Use `0.0.0.0` inside a container. |
| `HTTP_PORT` | `8080` | API port. |
| `DB_HOST` / `DB_PORT` | `127.0.0.1` / `3306` | MySQL address. |
| `DB_NAME` | `smsgateway` | Database name. |
| `DB_USERNAME` / `DB_PASSWORD` | `root` / `smsgw@1234` | MySQL credentials. |
| `RABBITMQ_HOST` / `RABBITMQ_PORT` | `127.0.0.1` / `5672` | RabbitMQ address. |
| `RABBITMQ_USERNAME` / `RABBITMQ_PASSWORD` | `guest` / `guest` | RabbitMQ credentials. |
| `RABBITMQ_VHOST` | `/` | RabbitMQ vhost. |
| `RABBITMQ_HEARTBEAT` | `10s` | Connection heartbeat. |
| `RABBITMQ_DIAL_TIMEOUT` | `30s` | Connection dial timeout. |
| `RABBITMQ_PREFETCH_COUNT` | `1` | Messages prefetched per worker. |
| `FIRST_OPERATOR_BASE_URL` | `https://api.first-operator.ir` | Operator API base URL. |
| `FIRST_OPERATOR_API_KEY` | — | Operator API key. |
| `FIRST_OPERATOR_TIMEOUT` | `10s` | Operator request timeout. |
| `LOG_LEVEL` | `info` | `debug`, `info`, `warn` or `error`. |
| `LOG_DIR` | `./logs` | Directory for the rotating JSON log file. Logs also always go to stdout. |

## Logs

Every service logs JSON to stdout, which is what `docker compose logs` and the container
runtime collect:

```bash
docker compose logs -f normal_sms_worker
```

Outside a container the same lines are also written to `$LOG_DIR/log.json`, rotated at 10 MB
and kept for 30 days. Set `LOG_LEVEL=debug` for more detail.

## Running with Docker Compose (recommended)

Compose starts MySQL, RabbitMQ and all three services. Service environment is filled from
your shell and from `.env` when present; the defaults already point at the compose services
(`mysql`, `rabbitmq`), so it works out of the box.

```bash
docker compose up -d --build
```

The API is then on `http://localhost:8080` and the RabbitMQ management UI on
`http://localhost:15672` (default login `admin` / `admin123`).

Check status and logs:

```bash
docker compose ps
```

```bash
docker compose logs -f app
```

```bash
docker compose logs -f normal_sms_worker express_sms_worker
```

### Scaling a worker independently

Because the workers are separate binaries, each one scales on its own:

```bash
docker compose up -d --scale normal_sms_worker=3
```

```bash
docker compose up -d --scale express_sms_worker=5
```

### Rebuilding one service

```bash
docker compose up -d --build normal_sms_worker
```

### Stopping

```bash
docker compose down
```

Add `-v` to also drop the RabbitMQ volume. The MySQL data lives in `~/Docker/mysql0` and is
not removed by `down`.

## Running without Docker

Requires Go 1.25+ and a reachable MySQL and RabbitMQ.

1. Create your local config and adjust the connection settings:

```bash
cp .env.example .env
```

2. Run the API (it applies migrations on startup, so it must start first):

```bash
go run .
```

3. In separate terminals, run each worker:

```bash
go run ./cmd/normal-sms-worker
```

```bash
go run ./cmd/express-sms-worker
```

## Building the binaries

```bash
go build -o bin/smsgateway .
```

```bash
go build -o bin/normal-sms-worker ./cmd/normal-sms-worker
```

```bash
go build -o bin/express-sms-worker ./cmd/express-sms-worker
```

## Building the images directly

Every Dockerfile expects the **repository root** as its build context:

```bash
docker build -t smsgateway-app -f Dockerfile .
```

```bash
docker build -t smsgateway-normal-worker -f deploy/normal_sms_worker/Dockerfile .
```

```bash
docker build -t smsgateway-express-worker -f deploy/express_sms_wowrker/Dockerfile .
```
## Database migrations

Migrations live in [repository/mysql/migrations](repository/mysql/migrations) and are applied
automatically by the API on startup via `sql-migrate`. The API must therefore start before
the workers, which is what the compose `depends_on` ordering expresses.