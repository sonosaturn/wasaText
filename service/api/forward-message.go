package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/julienschmidt/httprouter"
	"github.com/sonosaturn/wasatext/service/api/reqcontext"
)

func (rt *_router) forwardMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	destConversationID := ps.ByName("conversationId")

	if ctx.UserID == "" {
		authHeader := r.Header.Get("Authorization")
		if strings.HasPrefix(authHeader, "Bearer ") {
			ctx.UserID = strings.TrimPrefix(authHeader, "Bearer ")
		}
	}

	var body struct {
		SourceMessageID string `json:"messageId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	content, photoURL, err := rt.db.GetMessageForForwarding(body.SourceMessageID, ctx.UserID)
	if err != nil {
		ctx.Logger.WithError(err).Error("fetching source message for forward")
		http.Error(w, "message not found or user not authorized", http.StatusNotFound)
		return
	}

	// FIX: Pass TRUE as the last parameter to indicate it's Forwarded
	forwardedMsg, err := rt.db.SendMessage(destConversationID, ctx.UserID, content, photoURL, "", true)
	if err != nil {
		ctx.Logger.WithError(err).Error("forwarding message")
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(forwardedMsg)
}
