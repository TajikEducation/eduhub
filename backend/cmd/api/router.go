package main

import (
	"log/slog"
	"net/http"
	"time"

	cataloghttp "github.com/abdulhalim/eduhub/backend/internal/catalog/transport/http"
	catalogusecase "github.com/abdulhalim/eduhub/backend/internal/catalog/usecase"
	"github.com/abdulhalim/eduhub/backend/internal/platform/httpx"
)

// readyzTimeout — сколько ждём ответа от каждой зависимости в /readyz.
const readyzTimeout = 2 * time.Second

// newHandler собирает полный HTTP-хендлер приложения: маршруты + сквозная цепочка middleware.
// Вынесено из main() отдельной функцией ради тестируемости — задача 29 подставляет фейковый
// catalogRepo вместо реального Postgres, чтобы прогнать сквозной smoke-тест без БД.
func newHandler(log *slog.Logger, corsOrigins []string, readyDeps []httpx.Dependency, catalogRepo catalogusecase.InstitutionRepo) http.Handler {
	catalogSvc := catalogusecase.New(catalogRepo)

	router := httpx.NewRouter(log)
	router.Handle("GET /healthz", httpx.Healthz(log))
	router.Handle("GET /readyz", httpx.Readyz(log, readyzTimeout, readyDeps...))
	router.Handle("GET /api/v1/institutions", cataloghttp.ListHandler(catalogSvc, log))
	router.Handle("GET /api/v1/institutions/{id}", cataloghttp.GetHandler(catalogSvc, log))

	return httpx.Chain(httpx.WithRequestID, httpx.AccessLog(log), httpx.CORS(corsOrigins), httpx.Recover(log))(router)
}
