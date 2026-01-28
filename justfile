# Default recipe - list all available commands
default:
    @just --list

# Run the rbg CLI
run *ARGS:
    go run . {{ARGS}}

# Run all tests
test:
    go test ./...

# Generate Go code from SQL queries
sqlc:
    sqlc generate

# Start PostgreSQL
pg-start:
    brew services start postgresql@18

# Stop PostgreSQL
pg-stop:
    brew services stop postgresql@18

# Check PostgreSQL status
pg-status:
    brew services info postgresql@18

# Run all migrations up
db-up:
    goose -dir sql/schema postgres "$(go run . db-url)" up

# Roll back one migration
db-down:
    goose -dir sql/schema postgres "$(go run . db-url)" down

# Reset DB (down all, then up)
db-reset:
    goose -dir sql/schema postgres "$(go run . db-url)" reset
    goose -dir sql/schema postgres "$(go run . db-url)" up
