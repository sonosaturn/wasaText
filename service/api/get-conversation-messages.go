package api

import (
	"encoding/json"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/sonosaturn/wasatext/service/api/reqcontext"
	"github.com/sonosaturn/wasatext/service/database"
)

// getConversationMessages gestisce GET /conversations/:conversationId/messages
func (rt *_router) getConversationMessages(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	conversationID := ps.ByName("conversationId")

	// 1. Recupera i messaggi dal DB
	messages, err := rt.db.GetConversationMessages(conversationID, ctx.UserID)
	if err != nil {
		ctx.Logger.WithError(err).Error("Can't list messages")
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	// FIX: Assicuriamoci che non sia nil per evitare problemi nel JSON (anche se Go gestisce nil come null/empty)
	if messages == nil {
		messages = []database.Message{}
	}

	// FIX LINTER: Wrappiamo la lista in un oggetto JSON
	type Response struct {
		Messages []database.Message `json:"messages"`
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(Response{Messages: messages}); err != nil {
		ctx.Logger.WithError(err).Error("Can't encode messages")
		http.Error(w, "internal error", http.StatusInternalServerError)
	}
}