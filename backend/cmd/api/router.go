package main

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"

	applicantshttp "github.com/abdulhalim/eduhub/backend/internal/applicants/transport/http"
	applicantsusecase "github.com/abdulhalim/eduhub/backend/internal/applicants/usecase"
	authdomain "github.com/abdulhalim/eduhub/backend/internal/auth/domain"
	"github.com/abdulhalim/eduhub/backend/internal/auth/rbac"
	authhttp "github.com/abdulhalim/eduhub/backend/internal/auth/transport/http"
	authusecase "github.com/abdulhalim/eduhub/backend/internal/auth/usecase"
	"github.com/abdulhalim/eduhub/backend/internal/catalog/domain"
	cataloghttp "github.com/abdulhalim/eduhub/backend/internal/catalog/transport/http"
	catalogusecase "github.com/abdulhalim/eduhub/backend/internal/catalog/usecase"
	chathttp "github.com/abdulhalim/eduhub/backend/internal/communications/transport/http"
	chatusecase "github.com/abdulhalim/eduhub/backend/internal/communications/usecase"
	moderationhttp "github.com/abdulhalim/eduhub/backend/internal/moderation/transport/http"
	moderationusecase "github.com/abdulhalim/eduhub/backend/internal/moderation/usecase"
	"github.com/abdulhalim/eduhub/backend/internal/platform/httpx"
	reviewsdomain "github.com/abdulhalim/eduhub/backend/internal/reviews/domain"
	reviewshttp "github.com/abdulhalim/eduhub/backend/internal/reviews/transport/http"
	reviewsusecase "github.com/abdulhalim/eduhub/backend/internal/reviews/usecase"
	vacanciesdomain "github.com/abdulhalim/eduhub/backend/internal/vacancies/domain"
	vacancieshttp "github.com/abdulhalim/eduhub/backend/internal/vacancies/transport/http"
	vacanciesusecase "github.com/abdulhalim/eduhub/backend/internal/vacancies/usecase"
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

// authServicePort — то, что нужно транспорту auth от usecase-слоя (структурно совпадает с
// unexported authService в internal/auth/transport/http — *authusecase.Service удовлетворяет
// обоим благодаря структурной типизации интерфейсов Go).
type authServicePort interface {
	Register(ctx context.Context, in authdomain.RegisterInput) (authdomain.User, authusecase.TokenPair, error)
	Login(ctx context.Context, email, password string) (authdomain.User, authusecase.TokenPair, error)
	Refresh(ctx context.Context, rawToken string) (authusecase.TokenPair, error)
	Me(ctx context.Context, userID uuid.UUID) (authdomain.User, error)
}

// reviewsServicePort — то, что нужно транспорту reviews от usecase-слоя.
type reviewsServicePort interface {
	Create(ctx context.Context, userID uuid.UUID, in reviewsdomain.CreateReviewInput) (reviewsdomain.Review, error)
	ListApproved(ctx context.Context, institutionID uuid.UUID) ([]reviewsdomain.Review, error)
	ListMine(ctx context.Context, institutionID uuid.UUID) ([]reviewsdomain.Review, error)
	Reply(ctx context.Context, reviewID uuid.UUID, actorUserID uuid.UUID, isPrivileged bool, reply string) (reviewsdomain.Review, error)
	IsOwnerOfReview(ctx context.Context, reviewID uuid.UUID, userID uuid.UUID) (bool, error)
}

// newHandler собирает полный HTTP-хендлер приложения: маршруты + сквозная цепочка middleware.
// Вынесено из main() отдельной функцией ради тестируемости — задача 29 подставляет фейковый
// catalogRepo вместо реального Postgres, чтобы прогнать сквозной smoke-тест без БД.
// cache == nil (например, в тестах без Redis) — осознанная деградация до некэширующего Service,
// а не ошибка: кэш не обязательная зависимость (см. cmd/api/main.go — Redis не в /readyz).
// authSvc == nil — auth-маршруты не регистрируются (тесты каталога, которым auth не нужен).
// auditRecorder == nil — маршруты записи каталога и модерации не регистрируются (та же причина).
func newHandler(log *slog.Logger, corsOrigins []string, readyDeps []httpx.Dependency, catalogRepo catalogusecase.InstitutionRepo, cache catalogusecase.CacheClient, authSvc authServicePort, jwtSecret string, auditRecorder moderationusecase.Recorder, reviewsRepo reviewsusecase.ReviewRepo, vacanciesRepo vacanciesusecase.VacancyRepo, conversationRepo chatusecase.ConversationRepo, messageRepo chatusecase.MessageRepo, applicantRepo applicantsusecase.ApplicantRepo, applicantAchievementRepo applicantsusecase.AchievementRepo, applicationRepo applicantsusecase.ApplicationRepo) http.Handler {
	baseSvc := catalogusecase.New(catalogRepo)

	var reviewsSvc *reviewsusecase.Service
	if reviewsRepo != nil {
		reviewsSvc = reviewsusecase.New(reviewsRepo, baseSvc)
	}

	var vacanciesSvc *vacanciesusecase.Service
	if vacanciesRepo != nil {
		vacanciesSvc = vacanciesusecase.New(vacanciesRepo, baseSvc)
	}

	var chatSvc *chatusecase.Service
	if conversationRepo != nil && messageRepo != nil {
		chatSvc = chatusecase.New(conversationRepo, messageRepo, baseSvc)
	}

	var applicantsSvc *applicantsusecase.Service
	if applicantRepo != nil && applicantAchievementRepo != nil && applicationRepo != nil {
		applicantsSvc = applicantsusecase.New(applicantRepo, applicantAchievementRepo, applicationRepo)
	}

	var catalogSvc catalogService
	if cache != nil {
		catalogSvc = catalogusecase.NewCachedService(baseSvc, cache, cacheTTL, log)
	} else {
		catalogSvc = baseSvc
	}

	// institutionSummary — порт vacancies→catalog для обогащения глобального списка вакансий
	// сводкой об учреждении (см. пакетный комментарий internal/vacancies/transport/http). Только
	// composition root видит оба доменных типа одновременно.
	institutionSummary := func(ctx context.Context, id uuid.UUID) (vacanciesdomain.InstitutionSummary, error) {
		inst, err := catalogSvc.Get(ctx, id)
		if err != nil {
			return vacanciesdomain.InstitutionSummary{}, err
		}
		var city *vacanciesdomain.Bilingual
		if inst.City != nil {
			city = &vacanciesdomain.Bilingual{RU: inst.City.RU, TG: inst.City.TG}
		}
		return vacanciesdomain.InstitutionSummary{
			ID:              inst.ID,
			Name:            vacanciesdomain.Bilingual{RU: inst.Name.RU, TG: inst.Name.TG},
			Types:           inst.Types,
			Region:          inst.Region,
			District:        inst.District,
			City:            city,
			CoverPhotoS3Key: inst.CoverPhotoS3Key,
			Verified:        inst.Verified,
		}, nil
	}

	router := httpx.NewRouter(log)
	router.Handle("GET /healthz", httpx.Healthz(log))
	router.Handle("GET /readyz", httpx.Readyz(log, readyzTimeout, readyDeps...))
	router.Handle("GET /api/v1/institutions", cataloghttp.ListHandler(catalogSvc, log))
	router.Handle("GET /api/v1/institutions/{id}", cataloghttp.GetHandler(catalogSvc, log))
	router.Handle("GET /api/v1/institutions/{id}/news/published", cataloghttp.ListPublishedNewsHandler(baseSvc, log))
	router.Handle("GET /api/v1/news/{newsId}", cataloghttp.GetNewsHandler(baseSvc, log))
	router.Handle("GET /api/v1/staff/{staffId}", cataloghttp.GetStaffHandler(baseSvc, log))

	if reviewsSvc != nil {
		router.Handle("GET /api/v1/institutions/{id}/reviews", reviewshttp.ListHandler(reviewsSvc, log))
	}

	if vacanciesSvc != nil {
		router.Handle("GET /api/v1/institutions/{id}/vacancies", vacancieshttp.ListByInstitutionHandler(vacanciesSvc, log))
		router.Handle("GET /api/v1/vacancies", vacancieshttp.ListGlobalHandler(vacanciesSvc, institutionSummary, log))
		router.Handle("GET /api/v1/vacancies/{vacancyId}", vacancieshttp.GetGlobalHandler(vacanciesSvc, institutionSummary, log))
	}

	if applicantsSvc != nil {
		router.Handle("GET /api/v1/applicants", applicantshttp.ListPublicHandler(applicantsSvc, log))
		router.Handle("GET /api/v1/applicants/{id}", applicantshttp.GetHandler(applicantsSvc, log))
		router.Handle("GET /api/v1/applicants/{id}/achievements", applicantshttp.ListAchievementsHandler(applicantsSvc, log))
	}

	if authSvc != nil {
		router.Handle("POST /auth/register", authhttp.RegisterHandler(authSvc, log))
		router.Handle("POST /auth/login", authhttp.LoginHandler(authSvc, log))
		router.Handle("POST /auth/refresh", authhttp.RefreshHandler(authSvc, log))
		router.Handle("GET /auth/me", rbac.RequireAuth(jwtSecret, log)(authhttp.MeHandler(authSvc, log)))

		authed := rbac.RequireAuth(jwtSecret, log)

		router.Handle("POST /api/v1/institutions", authed(cataloghttp.CreateHandler(baseSvc, log)))
		router.Handle("PATCH /api/v1/institutions/{id}", authed(cataloghttp.UpdateHandler(baseSvc, log)))
		router.Handle("GET /api/v1/institutions/mine", authed(cataloghttp.MineHandler(baseSvc, log)))
		router.Handle("GET /api/v1/institutions/{id}/mine", authed(cataloghttp.GetMineHandler(baseSvc, log)))

		router.Handle("POST /api/v1/institutions/{id}/staff", authed(cataloghttp.CreateStaffHandler(baseSvc, log)))
		router.Handle("PATCH /api/v1/institutions/{id}/staff/{staffId}", authed(cataloghttp.UpdateStaffHandler(baseSvc, log)))
		router.Handle("DELETE /api/v1/institutions/{id}/staff/{staffId}", authed(cataloghttp.DeleteStaffHandler(baseSvc, log)))

		router.Handle("POST /api/v1/institutions/{id}/achievements", authed(cataloghttp.CreateAchievementHandler(baseSvc, log)))
		router.Handle("DELETE /api/v1/institutions/{id}/achievements/{achId}", authed(cataloghttp.DeleteAchievementHandler(baseSvc, log)))

		router.Handle("POST /api/v1/institutions/{id}/gallery", authed(cataloghttp.CreateGalleryItemHandler(baseSvc, log)))
		router.Handle("DELETE /api/v1/institutions/{id}/gallery/{itemId}", authed(cataloghttp.DeleteGalleryItemHandler(baseSvc, log)))

		router.Handle("POST /api/v1/institutions/{id}/alumni", authed(cataloghttp.CreateAlumnusHandler(baseSvc, log)))
		router.Handle("DELETE /api/v1/institutions/{id}/alumni/{alumnusId}", authed(cataloghttp.DeleteAlumnusHandler(baseSvc, log)))

		router.Handle("GET /api/v1/institutions/{id}/news", authed(cataloghttp.ListNewsHandler(baseSvc, log)))
		router.Handle("POST /api/v1/institutions/{id}/news", authed(cataloghttp.CreateNewsHandler(baseSvc, log)))
		router.Handle("PATCH /api/v1/institutions/{id}/news/{newsId}", authed(cataloghttp.UpdateNewsHandler(baseSvc, log)))
		router.Handle("DELETE /api/v1/institutions/{id}/news/{newsId}", authed(cataloghttp.DeleteNewsHandler(baseSvc, log)))

		if reviewsSvc != nil {
			router.Handle("POST /api/v1/institutions/{id}/reviews", authed(reviewshttp.CreateHandler(reviewsSvc, log)))
			router.Handle("GET /api/v1/institutions/{id}/reviews/mine", authed(reviewshttp.ListMineHandler(reviewsSvc, baseSvc.IsOwner, log)))
			router.Handle("POST /api/v1/reviews/{reviewId}/reply", authed(reviewshttp.ReplyHandler(reviewsSvc, log)))
		}

		if vacanciesSvc != nil {
			router.Handle("GET /api/v1/institutions/{id}/vacancies/mine", authed(vacancieshttp.ListMineHandler(vacanciesSvc, baseSvc.IsOwner, log)))
			router.Handle("POST /api/v1/institutions/{id}/vacancies", authed(vacancieshttp.CreateHandler(vacanciesSvc, baseSvc.IsOwner, log)))
			router.Handle("PATCH /api/v1/vacancies/{vacancyId}", authed(vacancieshttp.UpdateHandler(vacanciesSvc, log)))
			router.Handle("DELETE /api/v1/vacancies/{vacancyId}", authed(vacancieshttp.DeleteHandler(vacanciesSvc, log)))
		}

		if chatSvc != nil {
			router.Handle("POST /api/v1/conversations", authed(chathttp.CreateConversationHandler(chatSvc, log)))
			router.Handle("GET /api/v1/conversations", authed(chathttp.ListConversationsHandler(chatSvc, log)))
			router.Handle("GET /api/v1/conversations/{id}/messages", authed(chathttp.ListMessagesHandler(chatSvc, log)))
			router.Handle("POST /api/v1/conversations/{id}/messages", authed(chathttp.SendMessageHandler(chatSvc, log)))
		}

		if applicantsSvc != nil {
			router.Handle("GET /api/v1/applicants/me", authed(applicantshttp.GetMineHandler(applicantsSvc, log)))
			router.Handle("PUT /api/v1/applicants/me", authed(applicantshttp.UpsertMineHandler(applicantsSvc, log)))
			router.Handle("GET /api/v1/applicants/me/applications", authed(applicantshttp.ListMyApplicationsHandler(applicantsSvc, log)))
			router.Handle("POST /api/v1/applicants/me/achievements", authed(applicantshttp.CreateAchievementHandler(applicantsSvc, log)))
			router.Handle("DELETE /api/v1/applicants/achievements/{achId}", authed(applicantshttp.DeleteAchievementHandler(applicantsSvc, log)))
			router.Handle("POST /api/v1/vacancies/{id}/apply", authed(applicantshttp.ApplyHandler(applicantsSvc, log)))
		}
	}

	if authSvc != nil && auditRecorder != nil {
		moderationSvc := moderationusecase.New(baseSvc, auditRecorder, reviewsSvc)
		moderatorOnly := func(h http.Handler) http.Handler {
			return rbac.RequireAuth(jwtSecret, log)(rbac.RequireRole(log, "moderator", "admin")(h))
		}
		router.Handle("POST /api/v1/moderation/institutions/{id}/approve", moderatorOnly(moderationhttp.ApproveHandler(moderationSvc, log)))
		router.Handle("POST /api/v1/moderation/institutions/{id}/reject", moderatorOnly(moderationhttp.RejectHandler(moderationSvc, log)))
		router.Handle("GET /api/v1/moderation/institutions", moderatorOnly(cataloghttp.ListForModerationHandler(baseSvc, log)))

		if reviewsSvc != nil {
			router.Handle("POST /api/v1/moderation/reviews/{reviewId}/approve", moderatorOnly(moderationhttp.ApproveReviewHandler(moderationSvc, log)))
			router.Handle("POST /api/v1/moderation/reviews/{reviewId}/reject", moderatorOnly(moderationhttp.RejectReviewHandler(moderationSvc, log)))
		}
	}

	return httpx.Chain(httpx.WithRequestID, httpx.AccessLog(log), httpx.CORS(corsOrigins), httpx.Recover(log))(router)
}
