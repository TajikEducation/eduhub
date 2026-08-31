package httpx

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

// Router — обёртка над http.ServeMux (Go 1.22+ pattern-routing), отдаёт JSON-контракт ошибок
// вместо стандартного text/plain на 404/405.
type Router struct {
	mux    *http.ServeMux
	logger *slog.Logger
}

// NewRouter создаёт пустой Router поверх нового http.ServeMux.
func NewRouter(logger *slog.Logger) *Router {
	return &Router{
		mux:    http.NewServeMux(),
		logger: logger,
	}
}

// Handle регистрирует хендлер под pattern (в т.ч. с методом, например "GET /items").
func (rt *Router) Handle(pattern string, handler http.Handler) {
	rt.mux.Handle(pattern, handler)
}

// ServeHTTP различает 404 и 405 через probe-запуск internal-хендлера mux'а: mux.Handler
// возвращает пустой pattern в обоих случаях (не совпал путь / не совпал метод) — единственный
// надёжный сигнал, какой из них произошёл, это статус, который internal-хендлер сам выставит.
func (rt *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	_, pattern := rt.mux.Handler(r)
	if pattern != "" {
		// Настоящий диспатч — только mux.ServeHTTP() заполняет r.Pattern/PathValue,
		// mux.Handler() документированно этого не делает (см. комментарий в net/http).
		rt.mux.ServeHTTP(w, r)
		return
	}

	probe := &statusProbe{}
	h, _ := rt.mux.Handler(r)
	h.ServeHTTP(probe, r)

	if probe.status == http.StatusMethodNotAllowed {
		rt.writeRouteError(w, r, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed")
		return
	}
	rt.writeRouteError(w, r, http.StatusNotFound, "not_found", "route not found")
}

// writeRouteError пишет JSON-ошибку роутинга напрямую (не через WriteError — тут нет apperr.Error,
// код и статус уже известны).
func (rt *Router) writeRouteError(w http.ResponseWriter, r *http.Request, status int, code, message string) {
	body := errorResponse{
		Error: errorPayload{
			Code:    code,
			Message: message,
		},
		RequestID: RequestID(r.Context()),
	}

	payload, err := json.Marshal(body)
	if err != nil {
		// Маршалинг фиксированной структуры не должен падать — но если случилось,
		// это тоже инцидент, а не бизнес-ошибка.
		rt.logger.Error("failed to marshal route error response", slog.Any("error", err))
		payload = []byte(`{"error":{"code":"internal","message":"internal server error"}}`)
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if _, writeErr := w.Write(payload); writeErr != nil {
		rt.logger.Warn("failed to write route error response body", slog.Any("error", writeErr))
	}
}

// statusProbe — минимальный http.ResponseWriter, который запоминает выставленный статус
// и отбрасывает тело; нужен только чтобы узнать, решил ли internal-хендлер mux'а ответить
// 404 или 405 — сам ответ клиенту мы формируем заново в JSON.
type statusProbe struct {
	header http.Header
	status int
}

func (p *statusProbe) Header() http.Header {
	if p.header == nil {
		p.header = make(http.Header)
	}
	return p.header
}

func (p *statusProbe) Write(b []byte) (int, error) { return len(b), nil }

func (p *statusProbe) WriteHeader(status int) { p.status = status }
