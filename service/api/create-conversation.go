package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/sonosaturn/wasatext/service/api/reqcontext"
)

type createConversationRequest struct {
	OtherUserID string `json:"otherUserId"`
}

type conversationRef struct {
	ConversationID string `json:"conversationId"`
}

// createConversation handles POST /conversations.
func (rt *_router) createConversation(w http.ResponseWriter, r *http.Request, _ httprouter.Params, ctx reqcontext.RequestContext) {
	var req createConversationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if req.OtherUserID == "" {
		http.Error(w, "otherUserId is required", http.StatusBadRequest)
		return
	}
	if req.OtherUserID == ctx.UserID {
		http.Error(w, "cannot create conversation with yourself", http.StatusBadRequest)
		return
	}

	convID, err := rt.db.CreateDirectConversation(ctx.UserID, req.OtherUserID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "user not found", http.StatusNotFound)
			return
		}
		ctx.Logger.WithError(err).Error("create conversation failed")
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(conversationRef{ConversationID: convID})
}
