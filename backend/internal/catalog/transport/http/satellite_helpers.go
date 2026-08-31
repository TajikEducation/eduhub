package http

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/abdulhalim/eduhub/backend/internal/auth/rbac"
	"github.com/abdulhalim/eduhub/backend/internal/platform/apperr"
	"github.com/abdulhalim/eduhub/backend/internal/platform/httpx"
)

// requirePrincipal читает Principal из контекста или пишет 401. Общая первая строка
// каждого satellite-хендлера (staff/achievements/gallery/alumni/news CRUD).
func requirePrincipal(w http.ResponseWriter, r *http.Request, logger *slog.Logger) (rbac.Principal, bool) {
	principal, ok := rbac.FromContext(r.Context())
	if !ok {
		httpx.WriteError(w, r, logger, apperr.Unauthorized("authentication required"))
		return rbac.Principal{}, false
	}
	return principal, true
}

// requireOwnerOrPrivileged проверяет доступ: moderator/admin — всегда, иначе checkOwner()
// должен вернуть true. Пишет 403 (или ошибку checkOwner) и возвращает false, если доступ запрещён.
func requireOwnerOrPrivileged(w http.ResponseWriter, r *http.Request, logger *slog.Logger, principal rbac.Principal, checkOwner func(context.Context) (bool, error)) bool {
	if isPrivilegedRole(principal.Role) {
		return true
	}
	owner, err := checkOwner(r.Context())
	if err != nil {
		httpx.WriteError(w, r, logger, err)
		return false
	}
	if !owner {
		httpx.WriteError(w, r, logger, apperr.Forbidden("доступ запрещён"))
		return false
	}
	return true
}
