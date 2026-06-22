package api

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/sonosaturn/wasatext/service/api/reqcontext"
)

// deleteMessage handles DELETE /conversations/:conversationId/messages/:messageId
func (rt *_router) deleteMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	// Retrieve both parameters from the path
	conversationID := ps.ByName("conversationId")
	messageID := ps.ByName("messageId")

	// Now pass all 3 parameters required by the DB
	err := rt.db.DeleteMessage(conversationID, messageID, ctx.UserID)
	if err != nil {
		ctx.Logger.WithError(err).Error("Can't delete message")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
