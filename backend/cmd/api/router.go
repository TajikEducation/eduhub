package main

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/abdulhalim/eduhub/backend/internal/auth/googleoauth"
	"github.com/abdulhalim/eduhub/backend/internal/auth/jwt"
	"github.com/abdulhalim/eduhub/backend/internal/auth/password"
	authhttp "github.com/abdulhalim/eduhub/backend/internal/auth/transport/http"
	authusecase "github.com/abdulhalim/eduhub/backend/internal/auth/usecase"
	"github.com/abdulhalim/eduhub/backend/internal/catalog/domain"
	cataloghttp "github.com/abdulhalim/eduhub/backend/internal/catalog/transport/http"
	catalogusecase "github.com/abdulhalim/eduhub/backend/internal/catalog/usecase"
	"github.com/abdulhalim/eduhub/backend/internal/platform/clock"
	"github.com/abdulhalim/eduhub/backend/internal/platform/httpx"
)

// readyzTimeout — сколько ждём ответа от каждой зависимости в /readyz.
const readyzTimeout = 2 * time.Second

// cacheTTL — время жизни закэшированной страницы листинга каталога.
const cacheTTL = 60 * time.Second

// accessTokenTTL — срок жизни access-токена (E2.3/E2.4), не конфигурируется — тот же
// прецедент, что cacheTTL.
const accessTokenTTL = 15 * time.Minute

// refreshTokenTTL — срок жизни refresh-токена (E2.3/E2.4), не конфигурируется.
const refreshTokenTTL = 30 * 24 * time.Hour

// catalogService — то, что нужно транспорту каталога от usecase-слоя: и голому *Service,
// и *CachedService реализуют этот контракт, что позволяет подставлять кэш прозрачно.
type catalogService interface {
	List(ctx context.Context, f domain.Filter) (domain.ListResult, error)
	Get(ctx context.Context, id uuid.UUID) (domain.Institution, error)
}

// authUserRepo — то, что нужно auth-usecase слою от репозитория пользователей: и порт
// AccountService (UserRepo), и порт SessionService (UserRoleLookup) одновременно —
// internal/auth/repo/postgres.UserRepo реализует оба.
type authUserRepo interface {
	authusecase.UserRepo
	authusecase.UserRoleLookup
}

// newHandler собирает полный HTTP-хендлер приложения: маршруты + сквозная цепочка middleware.
// Вынесено из main() отдельной функцией ради тестируемости — задача 29 подставляет фейковый
// catalogRepo вместо реального Postgres, чтобы прогнать сквозной smoke-тест без БД.
// cache == nil (например, в тестах без Redis) — осознанная деградация до некэширующего Service,
// а не ошибка: кэш не обязательная зависимость (см. cmd/api/main.go — Redis не в /readyz).
func newHandler(
	log *slog.Logger,
	corsOrigins []string,
	readyDeps []httpx.Dependency,
	catalogRepo catalogusecase.InstitutionRepo,
	cache catalogusecase.CacheClient,
	userRepo authUserRepo,
	refreshTokenRepo authusecase.RefreshTokenRepo,
	oauthRepo authusecase.OAuthIdentityRepo,
	verificationCodeRepo authusecase.VerificationCodeRepo,
	hasher *password.Hasher,
	jwtSecret []byte,
	clk clock.Clock,
	googleClientID string,
) http.Handler {
	baseSvc := catalogusecase.New(catalogRepo)

	var catalogSvc catalogService
	if cache != nil {
		catalogSvc = catalogusecase.NewCachedService(baseSvc, cache, cacheTTL, log)
	} else {
		catalogSvc = baseSvc
	}

	jwtIssuer := jwt.NewIssuer(jwtSecret, accessTokenTTL, clk)
	sessionSvc := authusecase.NewSessionService(refreshTokenRepo, userRepo, jwtIssuer, clk, refreshTokenTTL)
	googleVerifier := googleoauth.NewVerifier(googleClientID)
	accountSvc := authusecase.NewAccountService(userRepo, hasher, sessionSvc, clk, googleVerifier, oauthRepo)
	verificationSvc := authusecase.NewVerificationService(userRepo, verificationCodeRepo, sessionSvc, hasher, clk, log)

	router := httpx.NewRouter(log)
	router.Handle("GET /healthz", httpx.Healthz(log))
	router.Handle("GET /readyz", httpx.Readyz(log, readyzTimeout, readyDeps...))
	router.Handle("GET /api/v1/institutions", cataloghttp.ListHandler(catalogSvc, log))
	router.Handle("GET /api/v1/institutions/{id}", cataloghttp.GetHandler(catalogSvc, log))
	router.Handle("POST /auth/register", authhttp.RegisterHandler(accountSvc, log))
	router.Handle("POST /auth/login", authhttp.LoginHandler(accountSvc, log))
	router.Handle("POST /auth/refresh", authhttp.RefreshHandler(sessionSvc, log))
	router.Handle("POST /auth/logout", authhttp.LogoutHandler(sessionSvc, log))
	router.Handle("GET /auth/me", authhttp.RequireAuth(jwtIssuer, log)(authhttp.MeHandler(accountSvc, log)))
	// Регистрируется ВСЕГДА, даже если googleClientID == "": googleoauth.Verifier с пустым
	// clientID просто отклоняет любые токены как невалидные (oidc.Config{ClientID: ""} не
	// матчит ни одну аудиторию) — ожидаемая деградация для окружений без настроенного Google
	// OAuth, не баг.
	router.Handle("POST /auth/oauth/google", authhttp.OAuthGoogleHandler(accountSvc, log))
	router.Handle("POST /auth/verify", authhttp.VerifyEmailHandler(verificationSvc, log))
	router.Handle("POST /auth/verify/resend", authhttp.ResendVerificationHandler(verificationSvc, log))
	router.Handle("POST /auth/password/reset-request", authhttp.PasswordResetRequestHandler(verificationSvc, log))
	router.Handle("POST /auth/password/reset-confirm", authhttp.PasswordResetConfirmHandler(verificationSvc, log))
	router.Handle("POST /auth/consent", authhttp.RequireAuth(jwtIssuer, log)(authhttp.ConsentHandler(accountSvc, log)))
	router.Handle("DELETE /auth/me", authhttp.RequireAuth(jwtIssuer, log)(authhttp.DeleteMeHandler(accountSvc, log)))

	return httpx.Chain(httpx.WithRequestID, httpx.AccessLog(log), httpx.CORS(corsOrigins), httpx.Recover(log))(router)
}
