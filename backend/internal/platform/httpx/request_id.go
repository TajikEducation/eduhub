// Package httpx — транспортный слой бэкенда EduHub: роутер, цепочка
// middleware, JSON in/out, маппинг доменных ошибок (apperr) в HTTP-код.
// В отличие от остальных пакетов internal/platform, зависимость от
// net/http здесь ожидаема и необходима — это единственный пакет, которому
// разрешено знать про транспорт.
package httpx

import (
	"context"
	"crypto/rand"
	"fmt"
	"net/http"
)

// headerRequestID — имя HTTP-заголовка, в котором передаётся/возвращается
// идентификатор запроса.
const headerRequestID = "X-Request-ID"

// requestIDKey — неэкспортированный тип ключа контекста для request_id.
// Собственный тип (а не string) — стандартная защита от коллизии ключей
// контекста между разными пакетами.
type requestIDKey struct{}

// WithRequestID — middleware, обеспечивающая идентификатор запроса
// (request_id) на весь путь обработки: используется как для корреляции в
// логах (см. docs/EduHub_Backend_Architecture.md, раздел «Логирование» —
// request_id обязателен в каждой лог-записи), так и для тела ответа об
// ошибке (httpx.WriteError, отдельная задача).
//
// Если во входящем запросе уже есть заголовок X-Request-ID — используется
// как есть (доверяем клиенту/прокси, чтобы сквозная трассировка работала
// через границы сервисов). Если заголовка нет — генерируется новый UUID
// v4. В обоих случаях значение кладётся в контекст запроса (доступно через
// RequestID(ctx)) и выставляется в заголовок ответа X-Request-ID — причём
// до вызова next.ServeHTTP, так как заголовки нельзя менять после того,
// как обработчик начал писать тело ответа.
func WithRequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get(headerRequestID)
		if id == "" {
			id = newRequestID()
		}

		w.Header().Set(headerRequestID, id)

		ctx := context.WithValue(r.Context(), requestIDKey{}, id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequestID возвращает идентификатор запроса, установленный в контекст
// middleware WithRequestID. Если в контексте нет request_id (например,
// контекст не проходил через WithRequestID — фоновая задача, тест) —
// возвращает пустую строку, не паникует.
func RequestID(ctx context.Context) string {
	id, ok := ctx.Value(requestIDKey{}).(string)
	if !ok {
		return ""
	}
	return id
}

// newRequestID генерирует новый идентификатор запроса в формате UUID v4
// (RFC 4122). Собственная минимальная реализация вместо внешней
// библиотеки (google/uuid и т.п.) — тот же принцип, что и в config/logger/
// apperr: платформенные пакеты EduHub на голой stdlib, алгоритм
// тривиален — 16 случайных байт + фиксация версии/варианта в нужных битах.
func newRequestID() string {
	var b [16]byte
	// Ошибку не игнорируем через "_" (правило проекта). crypto/rand.Read
	// возвращает ошибку только если ОС не может отдать энтропию — это
	// действительно неисправимая ситуация (сломан системный генератор
	// случайности), на которую по правилам проекта допустима паника, а не
	// проброс ошибки через весь стек вызовов middleware.
	if _, err := rand.Read(b[:]); err != nil {
		panic(fmt.Errorf("httpx: crypto/rand unavailable: %w", err))
	}

	// Версия 4: старшие 4 бита 7-го байта = 0100.
	b[6] = (b[6] & 0x0f) | 0x40
	// Вариант RFC 4122: старшие 2 бита 9-го байта = 10.
	b[8] = (b[8] & 0x3f) | 0x80

	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}
