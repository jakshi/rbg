# rbg

A small RSS/feeds CLI and backend.

This README explains prerequisites, how to install the CLI, how to set up the config file, a few useful commands, how to use the included `justfile`, and how to seed example data.

---

## Prerequisites

- Go (project targets Go 1.25+)
- PostgreSQL server (local or remote)
- `psql` (Postgres client) is useful for inspecting the DB
- (optional) `just` for convenient task shortcuts
- (optional) `docker` if you prefer running Postgres in a container

Make sure `go` and `psql` are on your PATH.

---

## Install the CLI

Install the `rbg` CLI with `go install`:

```sh
go install github.com/jakshi/rbg@latest
```

After installing, the `rbg` binary will be in your `$GOBIN` or `$GOPATH/bin`. Ensure that directory is in your `PATH`.

During development you can run commands in-place:

- With Go:
```sh
go run . <command>
```

- With `just` (recommended when available):
```sh
just run <command>
```

> Note: In this repository the `just` recipe named `run` simply forwards its arguments to `go run .`. Using `just run` is a thin convenience wrapper.

---

## Install `just` (macOS)

If you use Homebrew on macOS:

```sh
brew install just
```

List available tasks:

```sh
just --list
```

---

## Install `goose` (macOS)

`goose` is used for managing DB migrations. Install it on macOS with Homebrew:

```sh
brew install goose
```

---

## Run PostgreSQL with Docker (quick local setup)

If you prefer not to install Postgres locally, run a container for development:

```sh
docker run --name rbg-postgres \
  -e POSTGRES_USER=postgres \
  -e POSTGRES_PASSWORD=postgres \
  -e POSTGRES_DB=rbg \
  -p 5432:5432 \
  -d postgres:15
```

Example DB URL for `~/.config/rbg/config.json`:

```txt
postgres://postgres:postgres@localhost:5432/rbg?sslmode=disable
```

Connect with `psql`:

```sh
psql "postgres://postgres:postgres@localhost:5432/rbg?sslmode=disable"
```

Stop and remove the container when done:

```sh
docker stop rbg-postgres
docker rm rbg-postgres
```

---

## Configuration

Create `~/.config/rbg/config.json` with at least the `db_url` key:

```json
{
  "db_url": "postgres://user:password@localhost:5432/rbg?sslmode=disable",
  "current_user_name": "alice"
}
```

- `db_url` — Postgres connection string used by the CLI.
- `current_user_name` — optional; used by commands that require a logged-in user.

You can print the configured DB URL:

```sh
# with go run
go run . db-url

# or, if installed:
rbg db-url

# or via just:
just run db-url
```

Alternatively read it directly:

```sh
jq -r .db_url ~/.config/rbg/config.json
```

---

## Useful commands

Run the CLI with `go run . <command>`, `rbg <command>` (if installed), or `just run <command>`.

Common commands:

- `register <username>` — create a new user and log in.
- `login <username>` — set the current user in the config.
- `addfeed <name> <url>` — add a feed (auto-follows for the adding user).
- `feeds` — list feeds.
- `follow <feed_url>` — follow a feed.
- `following` — list followed feeds.
- `browse [limit] [offset]` — show posts for followed feeds (default limit: 2).
- `agg <duration>` — run the aggregator; `<duration>` is a Go duration (e.g. `1s`, `30s`, `1m`).
- `db-url` — print the DB connection URL.

Examples (both variants):

```sh
# register and login
go run . register alice
just run register alice

# add a feed
go run . addfeed "TechCrunch" "https://techcrunch.com/feed/"
just run addfeed "TechCrunch" "https://techcrunch.com/feed/"

# show posts (default limit)
go run . browse
just run browse

# run aggregator with 30s between requests
go run . agg 30s
just run agg 30s
```

---

## Migrations and the Justfile

Migrations are stored in `sql/schema` and managed with `goose`. The repository includes a `justfile` wrapper that makes running migration commands convenient.

Key `justfile` targets:

- `just db-up` — apply all pending migrations (reads DB URL from `~/.config/rbg/config.json`).
- `just db-down` — roll back the last migration.
- `just db-reset` — reset the DB (down then up).
- `just sqlc` — generate Go DB code from the SQL queries (requires `sqlc`).
- `just test` — run `go test ./...`.
- `just run ...` — forwards args to `go run .`.
- `just seed_example` — seed example data (script included at `scripts/seed_example.sh`).

Examples:

```sh
# apply migrations (reads DB URL from config)
just db-up

# regenerate generated DB code after editing SQL queries
just sqlc

# seed example data (creates testuser, feeds, follows, and a few posts)
just seed_example
```

---

## Seed script

A small idempotent script is provided at `scripts/seed_example.sh`. It:

- Creates a `testuser`.
- Adds several example feeds.
- Ensures `testuser` follows those feeds.
- Inserts a few sample posts.

Run it via `just`:

```sh
just seed_example
```

Or directly:

```sh
./scripts/seed_example.sh
```

The script reads DB URL from (in order):

1. `DB_URL` environment variable (if set),
2. `~/.config/rbg/config.json` (if `jq` is present),
3. `go run . db-url` (if available).

It requires `psql` to be present.

---

## Development & testing tips

- Run tests:

```sh
just test
# or
go test ./...
```

- Regenerate sqlc code after changing SQL queries:

```sh
just sqlc
```

- If you need to debug SQL or migrations:

```sh
psql "$(jq -r .db_url ~/.config/rbg/config.json)"
```

---

## Contributing

Open a pull request with a clear description of the change. This project uses `goose` to manage database migrations; a `justfile` wrapper is available to run migration commands without compiling the CLI.

Please run:

```sh
just sqlc
just test
```

(or `go test ./...`) before submitting changes that touch SQL or database-related code.

Install `goose` on macOS with Homebrew:

```sh
brew install goose
```

---