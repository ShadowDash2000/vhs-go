package middleware

import (
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/hook"
)

func AuthorizeGet() *hook.Handler[*core.RequestEvent] {
	return &hook.Handler[*core.RequestEvent]{
		Id:       "authorizeGet",
		Func:     authorizeGet(),
		Priority: apis.DefaultRateLimitMiddlewarePriority - 25,
	}
}

func authorizeGet() func(*core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		token := e.Request.URL.Query().Get("token")
		if token != "" {
			e.Request.Header.Set("Authorization", "Bearer "+token)
		}

		return e.Next()
	}
}
