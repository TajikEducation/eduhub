.PHONY: run test test-race test-integration lint vet vuln migrate-up migrate-down seed

# Запустить API-сервер локально.
run:
	go run ./cmd/api

# Прогнать unit-тесты по всем пакетам.
test:
	go test ./...

# Прогнать unit-тесты с включённым race detector.
test-race:
	go test -race ./...

# Прогнать интеграционные тесты (нужны запущенные сервисы docker-compose, см. docker-compose.yml).
test-integration:
	go test -tags=integration ./...

# Прогнать golangci-lint по конфигу репозитория (.golangci.yml).
lint:
	golangci-lint run

# Прогнать go vet по всем пакетам.
vet:
	go vet ./...

# Проверить зависимости на известные уязвимости.
vuln:
	govulncheck ./...

# Применить все непроведённые миграции БД.
migrate-up:
	goose -dir migrations postgres "$$DATABASE_URL" up

# Откатить последнюю миграцию БД.
migrate-down:
	goose -dir migrations postgres "$$DATABASE_URL" down

# Засеять демо-данные (только dev/test, см. cmd/devseed).
seed:
	go run ./cmd/devseed
