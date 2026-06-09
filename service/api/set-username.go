package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/julienschmidt/httprouter"
	"github.com/sonosaturn/wasatext/service/api/reqcontext"
)

func (rt *_router) setUsername(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	idStr := ps.ByName("userId")

	// Fallback: se il middleware non ha settato ctx.UserID, proviamo a leggerlo dall'header
	if ctx.UserID == "" {
		authHeader := r.Header.Get("Authorization")
		if strings.HasPrefix(authHeader, "Bearer ") {
			ctx.UserID = strings.TrimPrefix(authHeader, "Bearer ")
		}
	}

	// Verifica che l'utente stia modificando se stesso
	if idStr != ctx.UserID {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	var body struct {
		Username string `json:"username"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	if len(body.Username) < 3 || len(body.Username) > 16 {
		http.Error(w, "invalid username length", http.StatusBadRequest)
		return
	}

	// Chiama il DB
	success, err := rt.db.SetUsername(idStr, body.Username)
	if err != nil {
		ctx.Logger.WithError(err).Error("setting username")
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	if !success {
		http.Error(w, "username already taken", http.StatusConflict) // 409 Conflict
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	_, _ = w.Write([]byte("OK"))
}
