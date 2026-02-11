package api

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/sonosaturn/wasatext/service/api/reqcontext"
)

func (rt *_router) addToGroup(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	conversationID := ps.ByName("conversationId")
	targetUserID := ps.ByName("userId")

	// FIX: Chiamata corretta all'interfaccia
	err := rt.db.AddToConversation(conversationID, targetUserID)

	if err != nil {
		ctx.Logger.Error("Errore aggiunta membro al gruppo: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}