# Architecture

Technical reference for the SMS gateway's structure, message flow and data model. For
installation and day-to-day commands see the [README](../README.md).

## Overview

The gateway accepts SMS requests over HTTP, charges the user, persists the message and hands
it to RabbitMQ. Delivery is asynchronous and split across two independent lanes:

- **Normal lane** — the default path. The caller supplies the source number; the message is
  published to `sms.normal` and consumed by normal workers.
- **Express lane** — a low-latency path with its own queue (`sms.express`) and its own worker
  binary. The caller does *not* supply a source number: the express worker resolves a stable
  operator target at send time. Operators reserve a slice of their throughput for this lane
  (`ExpressReservedTPS`), so express traffic cannot be starved by normal traffic.

The lane is decided by the **endpoint**, not by a field in the payload — `POST /sms/send`
versus `POST /sms/send/express`. Each handler overwrites `SmsType` on the request before it
reaches the service, so a client cannot pick the lane by setting the field itself.

The system ships as three binaries that share one MySQL database and one RabbitMQ broker:

| Binary | Role |
| --- | --- |
| `smsgateway` | HTTP API, migration runner |
| `normal-sms-worker` | Consumes `sms.normal` |
| `express-sms-worker` | Consumes `sms.express` |

## Layering

Dependencies point in one direction: `delivery → service → repository → adapter`. A layer
never imports the layer above it.

```
delivery/httpserver/     Echo handlers. Bind, validate, call a service, map errors to HTTP.
        │
        ▼
service/                 Business logic. Owns transactions and the send pipeline.
  service/             UserService, OperatorService
  service/smsservice/  SMSService (send normal, send express, report)
        │
        ▼
repository/              Persistence. One interface per repository, MySQL implementation
        │                under a /mysql subpackage.
        ▼
adapter/                 Infrastructure clients: mysql, rabbitmq, sms_operator.
```

Supporting packages:

| Package | Responsibility |
| --- | --- |
| `config/` | Loads configuration from the environment (and an optional `.env`). |
| `entity/` | Domain types and enums. No behaviour beyond small helpers. |
| `param/` | Request and response structs, grouped by feature. |
| `contract/sms/` | Protobuf messages used as the queue wire format. |
| `validator/` | Request validation, including the DB-backed checks. |
| `pkg/richerror/` | Error type carrying an operation, a kind and metadata. |
| `pkg/httpmsg/` | Maps a `richerror.Kind` onto an HTTP status and message. |
| `pkg/errmsg/` | Shared error strings. |
| `internal/worker/smsworker/` | The two worker implementations. |

### Interfaces are declared by the consumer

Each layer defines the interface it needs from the layer below, and the implementation
asserts conformance at compile time. For example `service/smsservice` declares
`SMSRepository` and `MessageBroker`, and `repository/smsrepository/mysql` asserts:

```go
var _ smsrepository.SMSRepository = &MysqlSMSRepository{}
```

The concrete MySQL repositories are wired in `main.go` for the API and in each worker's
`cmd/` entrypoint. There is no DI framework — construction is explicit.

### Services

**`UserService`** — balance top-up. Currently a direct balance increase; the TODOs in
`service/user.go` describe the intended flow (create a `balance_request`, redirect to a
payment gateway, then credit the balance and link the transaction once the gateway confirms).

**`OperatorService`** — operator selection. `GetStableTargets` returns, for every active
operator, one active source number belonging to it:

```
GetActives()                      → operators where is_active = 1
  for each operator:
    GetActiveNumberByOperator(id) → one active number, or skip the operator
```

An operator with no active number is silently skipped rather than failing the batch, so one
misconfigured operator does not take down express delivery.

**`SMSService`** — the send pipeline and reporting.

### Workers

`internal/worker/smsworker` holds both implementations. They share the `MessageBrokerAdapter`
and `SmsOperator` interfaces and the `MaxAttempts` / `backoff` retry policy.

| | Normal (`worker.go`) | Express (`express_worker.go`) |
| --- | --- | --- |
| Payload | `sms.SmsRequest` | `sms.ExpressSmsRequest` |
| Source number | from the payload | resolved per attempt via `GetStableTargets` |
| Delivery | single target from the payload | tries each target in turn |
| Status writes | `SMSService.UpdateState` | `SMSService.UpdateState` |

Both write the delivery outcome back through `SMSService.UpdateState`, so a delivered message
becomes `sent` and an exhausted one becomes `failed` (with a refund). The express worker
resolves its own source numbers but shares the same status-reporting path.

## Message flow

### Normal lane

```
client ──POST /sms/send──▶ handler
                             │ bind, read Idempotency-Key header, force SmsType=normal
                             ▼
                          validator ──▶ user has enough balance? source number active?
                             ▼
                          SMSService.SendNormalSMS
                             │ 1. look up the idempotency key
                             │ 2. CreateMessage: charge balance + insert message + insert
                             │    transaction, in ONE MySQL transaction
                             │ 3. proto.Marshal(sms.SmsRequest)
                             │ 4. publish to sms.normal (publisher confirms)
                             │ 5. status → queued
                             ▼
                          normal-sms-worker ──▶ operator API
                             │ 200 → status sent, ack
                             │ retryable → backoff, retry (max 3 attempts)
                             │ exhausted → status failed (balance refunded) → sms.dlq
```

### Express lane

Identical up to step 3, then:

```
                          SMSService.SendExpressSMS
                             │ publish proto.Marshal(sms.ExpressSmsRequest) to sms.express
                             ▼
                          express-sms-worker
                             │ 1. GetStableTargets → [(operator, number), ...]
                             │ 2. for each target: SendSMS with retries
                             │    first success → status sent, ack
                             │ 3. no target succeeded → status failed (balance refunded)
                             │    → nack → sms.dlq
```

### Idempotency

`messages.idempotency_key` has a unique index. Before creating anything, the service calls
`GetByIdempotencyKey`; if a row exists it returns the stored status and content instead of
re-sending. This is what makes a client retry after a timeout safe.

### Failure handling

| Situation | Behaviour |
| --- | --- |
| Malformed protobuf payload | `Nack(requeue=false)` → dead-letter queue |
| Operator returns 200 | status `sent`, `Ack` |
| Retryable status (5xx, 429) | exponential backoff, up to 3 attempts |
| Non-retryable status | stop retrying → status `failed`, balance refunded → dead-letter queue |
| All attempts exhausted | status `failed`, balance refunded → dead-letter queue |
| Publish to RabbitMQ fails | message marked failed; balance is refunded |
| Status write fails after delivery | `Nack(requeue=true)` so the status is retried |
| State update for an already-terminal message | ignored — no status change, no second refund |

`UpdateState` routes on the status: `failed` goes through `UpdateStateToFailed` (which refunds
the charge and writes a `reversal` transaction), anything else through the plain
`UpdateState`. Both the publish failure in the service and the delivery failure in a worker
end up in the same refund path.

Backoff is `1s, 2s, 4s, …` capped at 30s (`backoff` in `internal/worker/smsworker/worker.go`).

### Refunds

`UpdateStateToFailed` is not a plain status update. In one transaction it locks the message
row, sets the status, credits the charged amount back to the user's balance, and writes a
`reversal` transaction. It reads the amount to refund from the original `sms_charge`
transaction row. If the message is already terminal the transaction commits without
refunding, so a redelivery cannot refund the same charge twice.

## Broker topology

Declared in `adapter/rabbitmq/adapter.go` on every connection (the adapter's `setup()` is
idempotent):

```
sms.exchange (direct, durable)
   ├── routing key "normal"   ──▶ sms.normal    ──┐
   └── routing key "express"  ──▶ sms.express   ──┤  x-dead-letter-exchange: dlx.exchange
                                                  │  x-dead-letter-routing-key: failed
dlx.exchange (direct, durable)                    │
   └── routing key "failed"   ──▶ sms.dlq     ◀───┘
```

Publishing uses **publisher confirms** (`ch.Confirm(false)` + `NotifyPublish`), so a publish
only returns successfully once the broker has acknowledged it. Messages are persistent
(`DeliveryMode: amqp.Persistent`). Consumption uses manual ack with a per-consumer prefetch
(`RABBITMQ_PREFETCH_COUNT`), which bounds how many unacked messages a worker holds.

One consequence worth knowing: `setup()` declares **both** queues, so either worker binary
will create both `sms.normal` and `sms.express` on connect. The lane separation happens at
consume time, not at declare time.

## Data model

Schema lives in `repository/mysql/migrations` and is applied by `sql-migrate` on API startup.

```
users ──┬──< messages ──< transactions
        ├──< balance_requests ──< transactions
        └──< wages

operators ──< sms_numbers
```

| Table | Purpose |
| --- | --- |
| `users` | Account and `balance` (DECIMAL). |
| `messages` | One row per SMS. Unique `idempotency_key`; `type` normal/express; `status` initiated/queued/sent/failed. |
| `transactions` | Ledger. `amount` plus a `type` of sms_charge, refund, reversal or manual_adjustment. Optionally linked to a message and/or a balance request. |
| `wages` | Price per user and lane. Unique on `(type, user_id)`. |
| `operators` | Throughput budget: `total_tps` and `express_reserved_tps`, with a CHECK that the reservation fits within the total. |
| `sms_numbers` | Source numbers, each belonging to an operator, with an `is_active` flag. |
| `balance_requests` | Top-up intents and their approval status. |

`created_at` / `updated_at` are `BIGINT` unix timestamps, not datetime columns.

Message status transitions:

```
initiated ──(published)──▶ queued ──(operator 200)──▶ sent
    │
    └──(publish failed)──▶ failed  ◀──(delivery exhausted)── queued
                             │
                             └─ balance refunded via a reversal transaction
```

`sent` and `failed` are terminal. No transition leaves them, and a later attempt to set
either state again is ignored (see the write-once rule under Concurrency).

## Concurrency and consistency

- **Balance charge on create.** `CreateMessage` runs in a `READ COMMITTED` transaction: it
  reads the wage, locks the user row with `SELECT … FOR UPDATE`, decrements the balance,
  inserts the message and inserts the `sms_charge` transaction, then commits. The row lock
  serialises concurrent sends for the same user.
- **Refund on failure.** `UpdateStateToFailed` uses the same pattern: lock the user, credit
  the balance, insert a `reversal` transaction, commit.
- **Rollback.** Both use a `committed` flag plus a deferred `tx.Rollback()`, so a panic or an
  early return cannot leave a transaction open.
- **Idempotency of the send request.** Enforced by the unique index on
  `messages.idempotency_key`, with a read-before-write check in the service.
- **Terminal states are write-once.** Once a message is `sent` or `failed`, further state
  updates are no-ops (`MessageStatus.IsTerminal`). This is what makes redelivery safe: a
  worker that reprocesses a message cannot move it backwards, and cannot trigger a second
  refund. The guard is applied in two places, because the two paths carry different risk:
  - `UpdateState` adds `AND status NOT IN ('sent','failed')` to the `UPDATE` itself, so the
    check and the write are one atomic statement. A zero-row result is only an error when the
    message does not exist.
  - `UpdateStateToFailed` re-reads the row with `SELECT … FOR UPDATE` inside the transaction
    and returns early when the status is already terminal. This one guards money, so the read
    and the refund must be in the same transaction under a row lock — a plain `WHERE` guard
    would leave a window where two workers both read a non-terminal status and both refund.

## Error handling

A single error type, `richerror.RichError`, carries an operation name, a `Kind`, an optional
wrapped error and metadata. Kinds are `InvalidInput`, `Forbidden`, `Unauthenticated`,
`NotFound` and `Unexpected`.

`httpmsg.MapRichErrKindsToHttpResponse` translates the kind into a status code:

| Kind | HTTP |
| --- | --- |
| `KindInvalidInput` | 400 |
| `KindUnauthenticated` | 401 |
| `KindForbidden` | 403 |
| `KindNotFound` | 404 |
| `KindUnexpected` / unknown | 500 |

Validation returns a `map[string]string` of field errors alongside the error, which handlers
serialise as `{"message": …, "errors": {…}}`.

## Known gaps

These are visible in the code and are called out so the document matches reality:

- **No tests.** There are no `_test.go` files in the repository.
- **No logging.** Errors are returned and mapped to HTTP responses, but there is no logger;
  several `TODO` comments mark where structured logging belongs.
- **`markSent` requeues on a failed status write.** When an SMS is delivered but the status
  update fails, the worker nacks with `requeue=true` so the status is not lost. The operator
  call is therefore repeated on redelivery. This is safe only while `SendSMS` is idempotent
  per message — worth revisiting once the real operator adapter is implemented.
- **Worker retries are per-delivery, not per-message.** The attempt counter lives in the
  `handle` loop; it does not persist across redeliveries, so a message that keeps failing
  after a requeue can exceed `MaxAttempts` in total.

## Planned but not implemented

- **Capacity enforcement from `express_reserved_tps`.** `Operator.NormalTPS()` computes the
  normal-lane share, and the `operators` table stores the reservation with a CHECK constraint,
  but no code path throttles either lane by it. The reservation is data-only today: neither
  worker rate-limits against the operator's TPS budget, so the express lane is not yet
  protected from normal-lane saturation by this mechanism. Enforcing it is what the removed
  `capacity/` package was presumably for. The HTTP-level rate limits in
  `delivery/httpserver/smshandler/handler.go` are static and unrelated to this.
- Dynamic rate limiting derived from operator capacity and user SLA (currently static).
- The full top-up flow: balance request, payment gateway redirect, and transaction
  reconciliation.
- The `FirstOperatorOperator.SendSMS` adapter returns a hard-coded `200`; the real HTTP call
  to the operator API is not implemented.
