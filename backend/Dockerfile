# syntax=docker/dockerfile:1

# --- builder ---
# Версия тега фиксируется под go.mod (`go 1.26.6`), не под устаревший план (`golang:1.23`).
FROM golang:1.26.6 AS builder

WORKDIR /src

# Сначала только go.mod/go.sum — кэш слоя переиспользуется, пока зависимости не менялись.
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# CGO_ENABLED=0: статический бинарь без libc-зависимостей — нужен для запуска
# в distroless/static (там нет ни libc, ни shell).
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -o /out/api ./cmd/api

# --- runtime ---
# distroless/static: нет shell, нет package manager, нет ничего кроме CA-сертификатов
# и tzdata — минимальная площадь атаки для сервиса без прямых внешних TLS-вызовов.
FROM gcr.io/distroless/static-debian12

COPY --from=builder /out/api /api

EXPOSE 8080

ENTRYPOINT ["/api"]
