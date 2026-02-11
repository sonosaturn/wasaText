package api

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/sonosaturn/wasatext/service/api/reqcontext"
)

// deleteMessage gestisce DELETE /conversations/:conversationId/messages/:messageId
func (rt *_router) deleteMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	// conversationID := ps.ByName("conversationId") // Non strettamente necessario per la delete se l'ID messaggio è univoco, ma fa parte del path
	messageID := ps.ByName("messageId")
	
	err := rt.db.DeleteMessage(messageID, ctx.UserID)
	if err != nil {
		ctx.Logger.WithError(err).Error("Can't delete message")
		// Se l'errore è "forbidden", potremmo ritornare 403, ma per semplicità gestiamo genericamente
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}