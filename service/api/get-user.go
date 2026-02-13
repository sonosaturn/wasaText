package api

import (
	"encoding/json"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/sonosaturn/wasatext/service/api/reqcontext"
)

func (rt *_router) getUserById(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	userId := ps.ByName("userId")

	// FIX: Chiamata corretta
	user, err := rt.db.GetUser(userId)
	if err != nil {
		ctx.Logger.Error("Errore recupero utente: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if user.ID == "" {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(user)
}
