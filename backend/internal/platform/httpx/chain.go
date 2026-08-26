package httpx

import "net/http"

// Middleware оборачивает хендлер дополнительным поведением (логирование, recover, request_id и т.п.).
type Middleware func(http.Handler) http.Handler

// Chain собирает middleware в один слой: первый в списке — самый внешний, выполняется первым
// и последним (оборачивает остальные). Свёртка справа налево: mw1(mw2(...(h))).
func Chain(mws ...Middleware) Middleware {
	return func(h http.Handler) http.Handler {
		for i := len(mws) - 1; i >= 0; i-- {
			h = mws[i](h)
		}
		return h
	}
}
