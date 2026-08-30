package main

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/abdulhalim/eduhub/backend/internal/catalog/domain"
	cataloghttp "github.com/abdulhalim/eduhub/backend/internal/catalog/transport/http"
	catalogusecase "github.com/abdulhalim/eduhub/backend/internal/catalog/usecase"
	"github.com/abdulhalim/eduhub/backend/internal/platform/httpx"
)

// readyzTimeout — сколько ждём ответа от каждой зависимости в /readyz.
const readyzTimeout = 2 * time.Second

// cacheTTL — время жизни закэшированной страницы листинга каталога.
const cacheTTL = 60 * time.Second

// catalogService — то, что нужно транспорту каталога от usecase-слоя: и голому *Service,
// и *CachedService реализуют этот контракт, что позволяет подставлять кэш прозрачно.
type catalogService interface {
	List(ctx context.Context, f domain.Filter) (domain.ListResult, error)
	Get(ctx context.Context, id uuid.UUID) (domain.Institution, error)
}

// newHandler собирает полный HTTP-хендлер приложения: маршруты + сквозная цепочка middleware.
// Вынесено из main() отдельной функцией ради тестируемости — задача 29 подставляет фейковый
// catalogRepo вместо реального Postgres, чтобы прогнать сквозной smoke-тест без БД.
// cache == nil (например, в тестах без Redis) — осознанная деградация до некэширующего Service,
// а не ошибка: кэш не обязательная зависимость (см. cmd/api/main.go — Redis не в /readyz).
func newHandler(log *slog.Logger, corsOrigins []string, readyDeps []httpx.Dependency, catalogRepo catalogusecase.InstitutionRepo, cache catalogusecase.CacheClient) http.Handler {
	baseSvc := catalogusecase.New(catalogRepo)

	var catalogSvc catalogService
	if cache != nil {
		catalogSvc = catalogusecase.NewCachedService(baseSvc, cache, cacheTTL, log)
	} else {
		catalogSvc = baseSvc
	}

	router := httpx.NewRouter(log)
	router.Handle("GET /healthz", httpx.Healthz(log))
	router.Handle("GET /readyz", httpx.Readyz(log, readyzTimeout, readyDeps...))
	router.Handle("GET /api/v1/institutions", cataloghttp.ListHandler(catalogSvc, log))
	router.Handle("GET /api/v1/institutions/{id}", cataloghttp.GetHandler(catalogSvc, log))

	return httpx.Chain(httpx.WithRequestID, httpx.AccessLog(log), httpx.CORS(corsOrigins), httpx.Recover(log))(router)
}
