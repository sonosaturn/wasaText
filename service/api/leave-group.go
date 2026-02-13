package api

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/sonosaturn/wasatext/service/api/reqcontext"
)

func (rt *_router) leaveGroup(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	conversationID := ps.ByName("conversationId")
	targetUserID := ps.ByName("userId")

	// FIX: Chiamata corretta all'interfaccia
	err := rt.db.LeaveConversation(conversationID, targetUserID)

	if err != nil {
		// Nota: se il DB restituisce sql.ErrNoRows o altro, gestiscilo qui se vuoi 404
		ctx.Logger.Error("Errore uscita gruppo: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
