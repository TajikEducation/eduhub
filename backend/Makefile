.PHONY: run test test-race test-integration lint vet vuln migrate-up migrate-down seed

# Start the API server locally.
run:
	go run ./cmd/api

# Run unit tests across all packages.
test:
	go test ./...

# Run unit tests with the race detector enabled.
test-race:
	go test -race ./...

# Run integration tests (require docker-compose services running, see docker-compose.yml).
test-integration:
	go test -tags=integration ./...

# Run golangci-lint using the repo config (.golangci.yml).
lint:
	golangci-lint run

# Run go vet across all packages.
vet:
	go vet ./...

# Check dependencies for known vulnerabilities.
vuln:
	govulncheck ./...

# Apply all pending database migrations.
migrate-up:
	goose -dir migrations postgres "$$DATABASE_URL" up

# Roll back the most recent database migration.
migrate-down:
	goose -dir migrations postgres "$$DATABASE_URL" down

# Seed demo data (dev/test only, see cmd/devseed).
seed:
	go run ./cmd/devseed
