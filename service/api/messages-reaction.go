package api

import (
	"encoding/json"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/sonosaturn/wasatext/service/api/reqcontext"
)

// reactToMessage gestisce PUT /conversations/:conversationId/messages/:messageId/reaction
func (rt *_router) reactToMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	conversationID := ps.ByName("conversationId")
	messageID := ps.ByName("messageId")

	var body struct {
		Emoji string `json:"emoji"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	// Se emoji è stringa vuota, rimuoviamo la reazione
	if body.Emoji == "" {
		err := rt.db.UnreactToMessage(messageID, ctx.UserID)
		if err != nil {
			ctx.Logger.WithError(err).Error("db unreact error")
			http.Error(w, "db error", http.StatusInternalServerError)
			return
		}
	} else {
		// Altrimenti aggiungiamo/aggiorniamo (Passiamo anche conversationID ora)
		err := rt.db.ReactToMessage(conversationID, messageID, ctx.UserID, body.Emoji)
		if err != nil {
			ctx.Logger.WithError(err).Error("db react error")
			http.Error(w, "db error", http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusNoContent)
}
