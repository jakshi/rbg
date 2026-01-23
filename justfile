# Default recipe - list all available commands
default:
    @just --list

# Run the rbg CLI
run *ARGS:
    go run . {{ARGS}}

# Run all tests
test:
    go test ./...

# Start PostgreSQL
pg-start:
    brew services start postgresql@18

# Stop PostgreSQL
pg-stop:
    brew services stop postgresql@18
