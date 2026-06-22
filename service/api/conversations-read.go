package api

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/sonosaturn/wasatext/service/api/reqcontext"
)

// setConversationSeen handles PUT /conversations/:conversationId/seen
func (rt *_router) setConversationSeen(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	conversationID := ps.ByName("conversationId")

	err := rt.db.MarkConversationAsRead(conversationID, ctx.UserID)
	if err != nil {
		ctx.Logger.WithError(err).Error("marking conversation as seen")
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
