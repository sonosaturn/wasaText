package api

import (
	"database/sql"
	"net/http"
	"strings"
	"errors"

	"github.com/julienschmidt/httprouter"
	"github.com/sonosaturn/wasatext/service/api/reqcontext"
)

func (rt *_router) requireAuth(fn httpRouterHandler) func(http.ResponseWriter, *http.Request, httprouter.Params) {
	return rt.wrap(func(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
		auth := r.Header.Get("Authorization")
		if auth == "" {
			http.Error(w, "missing Authorization header", http.StatusUnauthorized)
			return
		}
		const prefix = "Bearer "
		if !strings.HasPrefix(auth, prefix) {
			http.Error(w, "invalid Authorization header", http.StatusUnauthorized)
			return
		}
		token := strings.TrimSpace(strings.TrimPrefix(auth, prefix))
		if token == "" {
			http.Error(w, "invalid Authorization header", http.StatusUnauthorized)
			return
		}

		// FIX: Chiamata corretta
		u, err := rt.db.GetUser(token)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			ctx.Logger.WithError(err).Error("auth lookup failed")
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}

		ctx.UserID = u.ID
		ctx.UserName = u.Username

		fn(w, r, ps, ctx)
	})
}