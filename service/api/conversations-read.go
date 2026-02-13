package api

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/sonosaturn/wasatext/service/api/reqcontext"
)

// markAsRead gestisce PUT /conversations/:conversationId/read
func (rt *_router) markAsRead(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	conversationID := ps.ByName("conversationId")

	// Chiama la funzione del DB per aggiornare il timestamp di lettura
	err := rt.db.MarkConversationAsRead(conversationID, ctx.UserID)
	if err != nil {
		ctx.Logger.WithError(err).Error("marking conversation as read")
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
