# Default recipe - list all available commands
default:
    @just --list

# Run the rbg CLI
run:
    go run .

# Run all tests
test:
    go test ./...
