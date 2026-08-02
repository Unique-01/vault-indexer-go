# vault-indexer-go

A concurrent Go blockchain event indexer with a REST API and SIWE (Sign-In With Ethereum) authentication, built to index events from the [SecureVault](#https://github.com/Unique-01/SecureVault-Contract) contract. 

This repo depends on `SecureVault` being deployed somewhere reachable (see [Prerequisites](#prerequisites)); the contract itself has no dependency on this indexer and can be used independently.

## Prerequisites

- A deployed instance of [`SecureVault`](#https://github.com/Unique-01/SecureVault-Contract) 
- Its ABI, exported and placed at `internal/indexer/abi/vault.json` in this repo.

## What it does

- **Indexes** every event emitted by a deployed `SecureVault` contract, from a configurable starting block, then continues polling for new blocks indefinitely.
- **Persists** indexed events to Postgres, alongside indexing progress, so the indexer can resume safely after a restart.
- **Serves** indexed events over a paginated, filterable HTTP API, gated behind SIWE-based wallet authentication.

## Architecture

The indexer is a five-stage concurrent pipeline, each stage a pool of goroutines connected by buffered channels:

```
rangeProducer → fetchWorkers → parseWorker → sequencer → saveWorker
```

- **`rangeProducer`** — determines unindexed block ranges (backfill on startup, then polls for new blocks at a configurable interval).
- **`fetchWorker`** — fetches logs and block headers for a range, with retry/backoff on transient RPC failures.
- **`parseWorker`** — decodes raw logs into typed `VaultEvent`s using the contract's ABI.
- **`sequencer`** — reorders results back into strict block order, since the fetch/parse stages are concurrent and complete out of order.
- **`saveWorker`** — persists each range's events and advances the indexer's watermark atomically, in a single transaction.

The API and indexer are two independent binaries (`cmd/api`, `cmd/indexer`) sharing one Postgres database — they can be deployed, scaled, and restarted independently.

## Tech stack

- Go, `net/http` 
- Postgres, `database/sql` + `pgx`
- `go-ethereum` for chain interaction and ABI decoding
- `spruceid/siwe-go` for SIWE message parsing/verification
- `golang-jwt` for session tokens
- `golang-migrate` for schema migrations

## Running it

### Option 1 — Docker Compose (recommended)

Brings up Postgres, a local Anvil node, the indexer, and the API with one command.

```shell
$ cp .env.docker.example .env.docker   # fill in JWT_SECRET, etc.
$ docker compose up --build
```

Once running, deploy `SecureVault` (from its own repo — see its README) against `http://localhost:8545` (this project's Anvil container publishes that port to the host):

```shell
$ forge create src/SecureVault.sol:SecureVault --rpc-url http://localhost:8545 --private-key <anvil_key> --broadcast
```

Set the printed address as `VAULT_ADDRESS` in `.env.docker`, then restart the indexer and API so they pick up the new value:

```shell
$ docker compose up --build indexer api
```

Run `script/SimulateEvents.s.sol` from the `SecureVault` repo (see its README) to generate real events, then check they were indexed:

```shell
$ curl "http://localhost:8000/health"
```

### Option 2 — Running locally

Requires Go, Postgres, and a running Anvil node (see the contract repo).

```shell
$ cp .env.example .env   # fill in values, pointing at your local Postgres/Anvil

$ migrate -database "$DATABASE_URL" -path migrations up

$ go run ./cmd/indexer
$ go run ./cmd/api   # in a second terminal
```

## Configuration

All configuration is via environment variables (see `.env.example`):

| Variable | Description |
|---|---|
| `RPC_URL` | Ethereum JSON-RPC endpoint |
| `VAULT_ADDRESS` | Deployed `SecureVault` contract address |
| `DATABASE_URL` | Postgres connection string |
| `BATCH_SIZE` | Blocks fetched per `eth_getLogs` call |
| `POLL_INTERVAL` | How often to check for new blocks once backfill is caught up |
| `HTTP_ADDR` | Address the API listens on |
| `JWT_SECRET` | Signing key for session tokens |
| `JWT_TOKEN_EXPIRY` | Session token lifetime |
| `SIWE_DOMAIN` / `SIWE_URI` | Domain/URI embedded in SIWE challenge messages |
| `CHAIN_ID` | Chain ID embedded in SIWE challenge messages |

## Authentication

Wallet authentication follows SIWE (EIP-4361), with the server — not the client — constructing the full challenge message, to remove any client-side influence over its content:

```
GET  /auth/challenge?walletAddress=0x...   → { "message": "<full SIWE message text>" }
                                              (sign this with the wallet, unmodified)
POST /auth/verify   { walletAddress, signature } → { "token": "<jwt>" }
```

The returned JWT is passed as `Authorization: Bearer <token>` on subsequent requests.

## API

```
GET /health

GET /events?eventType=&cursor=&limit=
Authorization: Bearer <token>
```

- `eventType` — optional, one of `UserDeposited`, `UserRequestedWithdrawal`, `UserWithdrawn`, `UserModifiedPendingWithdrawal`, `UserCancelledPendingWithdrawal`
- `cursor` — optional, opaque pagination token from a previous response's `nextCursor`
- `limit` — optional, defaults to 20, capped at 100

Response:

```json
{
  "events": [ { "walletAddress": "0x...", "eventType": "UserDeposited", "amount": "1000000000000000000", "...": "..." } ],
  "nextCursor": "MTAwOjM="
}
```

`nextCursor` is `null` on the last page. The wallet address is taken from the authenticated session, not a request parameter.

## Testing

```shell
$ go test ./... -v
```

## Project structure

```
cmd/
    indexer/       entrypoint for the indexing binary
    api/           entrypoint for the API binary
internal/
    indexer/       pipeline stages, ABI parsing, domain types (Store, BlockchainClient interfaces)
    api/           HTTP handlers, middleware, pagination (EventReader interface)
    auth/          SIWE challenge/verification, JWT sessions (NonceStore interface)
    repository/    concrete Store implementations (Postgres)
    config/        environment-based configuration
    bootstrap/     shared startup logic (config + logger + DB, used by both binaries)
migrations/        golang-migrate schema migrations
