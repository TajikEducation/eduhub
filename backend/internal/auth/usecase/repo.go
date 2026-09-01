// Package usecase — оркестрация сессий auth-сервиса (выпуск/ротация токенов, E2.3).
package usecase

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/abdulhalim/eduhub/backend/internal/auth/domain"
	"github.com/abdulhalim/eduhub/backend/internal/auth/googleoauth"
)

// RefreshTokenRepo — порт в БД для refresh-токенов. Реализация — internal/auth/repo/postgres.
type RefreshTokenRepo interface {
	Create(ctx context.Context, rt domain.RefreshToken) error
	// FindByHash возвращает apperr.NotFound (обёрнутый), если хеш не найден.
	FindByHash(ctx context.Context, tokenHash string) (domain.RefreshToken, error)
	// Revoke отзывает один токен; replacedBy — nil при отзыве без ротации (logout/reuse).
	Revoke(ctx context.Context, id uuid.UUID, revokedAt time.Time, replacedBy *uuid.UUID) error
	// RevokeFamily отзывает разом все ещё не отозванные токены семьи — reuse-detection.
	RevokeFamily(ctx context.Context, familyID uuid.UUID, revokedAt time.Time) error
	// RevokeAllForUser отзывает разом ВСЕ ещё не отозванные refresh-токены пользователя (все
	// семьи/устройства) — нужен для password-reset и удаления аккаунта: смена пароля/удаление
	// аккаунта должны выкидывать из всех активных сессий, не только текущей.
	RevokeAllForUser(ctx context.Context, userID uuid.UUID, revokedAt time.Time) error
}

// UserRoleLookup — порт чтения текущей роли пользователя. Нужен Rotate(): роль в новом
// access-токене должна быть АКТУАЛЬНОЙ на момент ротации, не унаследованной от исходного
// Issue() — иначе смена роли (модерация забанила/повысила) не подействует, пока не истечёт
// 30-дневный refresh и человек не перелогинится заново.
type UserRoleLookup interface {
	RoleByUserID(ctx context.Context, userID uuid.UUID) (string, error)
}

// UserRepo — порт в БД для пользователей (E2.4). Реализация — internal/auth/repo/postgres.
type UserRepo interface {
	// Create вставляет нового пользователя, возвращает его с полями, сгенерированными БД
	// (id, created_at, updated_at). Конфликт email → apperr.ConflictCode("email_taken", ...).
	Create(ctx context.Context, u domain.User) (domain.User, error)
	// FindByEmail — apperr.NotFound, если email не найден.
	FindByEmail(ctx context.Context, email string) (domain.User, error)
	// FindByID — apperr.NotFound, если не найден.
	FindByID(ctx context.Context, id uuid.UUID) (domain.User, error)
	// UpdateConsent обновляет consent_at/consent_version — переподтверждение согласия.
	UpdateConsent(ctx context.Context, userID uuid.UUID, consentVersion string, consentAt time.Time) error
	// MarkEmailVerified ставит email_verified_at и переводит status в 'active'.
	MarkEmailVerified(ctx context.Context, userID uuid.UUID, verifiedAt time.Time) error
	// UpdatePasswordHash — новый пароль после password-reset.
	UpdatePasswordHash(ctx context.Context, userID uuid.UUID, passwordHash string) error
	// SoftDelete — soft-delete + анонимизация PII (закон РТ №1537), не физический DELETE.
	SoftDelete(ctx context.Context, userID uuid.UUID, anonymizedEmail string, deletedAt time.Time) error
}

// VerificationCodeRepo — порт в БД для кодов подтверждения (email-верификация, password-reset,
// E2.4). Реализация — internal/auth/repo/postgres.
type VerificationCodeRepo interface {
	Create(ctx context.Context, vc domain.VerificationCode) error
	// FindLatestActive — самый свежий ещё не истёкший код для (userID, channel, purpose).
	// apperr.NotFound, если такого нет.
	FindLatestActive(ctx context.Context, userID uuid.UUID, channel, purpose string, now time.Time) (domain.VerificationCode, error)
	IncrementAttempts(ctx context.Context, id uuid.UUID) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// GoogleIDTokenVerifier — порт верификации Google ID-токена. Реализация — internal/auth/googleoauth.
type GoogleIDTokenVerifier interface {
	Verify(ctx context.Context, rawIDToken string) (googleoauth.Claims, error)
}

// OAuthIdentityRepo — порт в БД для связок пользователь↔внешний провайдер (E2.4).
// Реализация — internal/auth/repo/postgres.
type OAuthIdentityRepo interface {
	// FindByProvider — apperr.NotFound, если пары (provider, providerUserID) нет.
	FindByProvider(ctx context.Context, provider, providerUserID string) (domain.OAuthIdentity, error)
	// Create — UNIQUE(provider, provider_user_id) конфликт маловероятен в штатном потоке
	// (проверяется до вызова), но на всякий случай тоже apperr.ConflictCode, не голая ошибка БД.
	Create(ctx context.Context, oi domain.OAuthIdentity) error
}

// ChildRepo — порт в БД для привязок родитель↔учреждение (E2.6). Реализация —
// internal/auth/repo/postgres.
type ChildRepo interface {
	// Create вставляет новую привязку. UNIQUE(user_id, institution_id) конфликт →
	// apperr.ConflictCode("child_link_exists", ...).
	Create(ctx context.Context, c domain.Child) (domain.Child, error)

	// ListPendingByInstitution — привязки со confirmation_status='pending' для очереди подтверждения.
	ListPendingByInstitution(ctx context.Context, institutionID uuid.UUID) ([]domain.Child, error)

	// Confirm атомарно переводит confirmation_status: pending→confirmed и пишет запись
	// в moderation.audit_log в одной транзакции. apperr.NotFound, если привязки с id нет;
	// apperr.ConflictCode("child_not_pending", ...), если confirmation_status уже не 'pending'.
	Confirm(ctx context.Context, childID, actorID uuid.UUID, actorRole, requestID string) (domain.Child, error)

	// Reject — то же самое для отклонения, с обязательной структурированной причиной.
	Reject(ctx context.Context, childID, actorID uuid.UUID, actorRole, reasonCode string, reasonText *string, requestID string) (domain.Child, error)
}

// InstitutionStatusChecker — кросс-схемный порт проверки статуса модерации учреждения:
// auth не владеет схемой catalog, поэтому не прямой SQL-запрос, а интерфейс. Реализация —
// internal/catalog/repo/postgres.
type InstitutionStatusChecker interface {
	// IsApproved — apperr.NotFound, если учреждение с id не существует.
	IsApproved(ctx context.Context, institutionID uuid.UUID) (bool, error)
}
