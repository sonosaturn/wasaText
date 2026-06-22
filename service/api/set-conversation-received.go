package api

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/sonosaturn/wasatext/service/api/reqcontext"
)

// setConversationReceived handles PUT /conversations/:conversationId/received
func (rt *_router) setConversationReceived(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	conversationID := ps.ByName("conversationId")

	err := rt.db.MarkConversationAsReceived(conversationID, ctx.UserID)
	if err != nil {
		ctx.Logger.WithError(err).Error("marking conversation as received")
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
