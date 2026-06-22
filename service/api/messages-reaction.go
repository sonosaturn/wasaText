package api

import (
	"encoding/json"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/sonosaturn/wasatext/service/api/reqcontext"
)

// commentMessage handles PUT /conversations/:conversationId/messages/:messageId/reaction
func (rt *_router) commentMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	conversationID := ps.ByName("conversationId")
	messageID := ps.ByName("messageId")

	var body struct {
		Emoji string `json:"emoji"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	// If emoji is an empty string, remove the reaction
	if body.Emoji == "" {
		err := rt.db.UnreactToMessage(messageID, ctx.UserID)
		if err != nil {
			ctx.Logger.WithError(err).Error("db unreact error")
			http.Error(w, "db error", http.StatusInternalServerError)
			return
		}
	} else {
		// Otherwise add/update (Now also passing conversationID)
		err := rt.db.ReactToMessage(conversationID, messageID, ctx.UserID, body.Emoji)
		if err != nil {
			ctx.Logger.WithError(err).Error("db react error")
			http.Error(w, "db error", http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "text/plain")
	_, _ = w.Write([]byte("OK"))
}

// uncommentMessage handles DELETE /conversations/:conversationId/messages/:messageId/reaction
func (rt *_router) uncommentMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	messageID := ps.ByName("messageId")

	if err := rt.db.UnreactToMessage(messageID, ctx.UserID); err != nil {
		ctx.Logger.WithError(err).Error("db unreact error")
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
