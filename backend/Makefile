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
# -p 1 — пакеты строго последовательно: несколько пакетов интеграционных тестов бьют в одну
# и ту же TEST_DATABASE_URL, а cmd/devseed коммитит реальные строки (не в rollback-транзакции,
# как остальные) — при параллельном запуске пакетов это гонка с фикстурами других пакетов.
test-integration:
	go test -tags=integration -p 1 ./...

# Прогнать golangci-lint по конфигу репозитория (.golangci.yml).
lint:
	golangci-lint run

# Прогнать go vet по всем пакетам.
vet:
	go vet ./...

# Проверить зависимости на известные уязвимости.
vuln:
	go tool govulncheck ./...

# Применить все непроведённые миграции БД.
migrate-up:
	go tool goose -dir migrations postgres "$$DATABASE_URL" up

# Откатить последнюю миграцию БД.
migrate-down:
	go tool goose -dir migrations postgres "$$DATABASE_URL" down

# Засеять демо-данные (только dev/test, см. cmd/devseed).
seed:
	go run ./cmd/devseed
