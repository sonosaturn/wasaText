package api

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/sonosaturn/wasatext/service/api/reqcontext"
)

func (rt *_router) leaveGroup(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	conversationID := ps.ByName("conversationId")
	targetUserID := ps.ByName("userId")

	// FIX: Correct call to the interface
	err := rt.db.LeaveConversation(conversationID, targetUserID)

	if err != nil {
		// Note: if the DB returns sql.ErrNoRows or something else, handle it here if you want a 404
		ctx.Logger.Error("Errore uscita gruppo: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
