package api

import (
	"encoding/json"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/sonosaturn/wasatext/service/api/reqcontext"
)

type conversationSummaryResponse struct {
	ID                  string `json:"id"`
	IsGroup             bool   `json:"is_group"`
	Title               string `json:"title"`
	PhotoURL            string `json:"photo_url"`
	OtherUserID         string `json:"other_user_id"`
	LastMessageAt       string `json:"last_message_at"`
	LastMessagePreview  string `json:"last_message_preview"`
	LastMessageSenderID string `json:"last_message_sender_id"`
	LastMessageStatus   int    `json:"last_message_status"` // <--- NUOVO
	UnreadCount         int    `json:"unread_count"`
}

func (rt *_router) getMyConversations(w http.ResponseWriter, r *http.Request, _ httprouter.Params, ctx reqcontext.RequestContext) {
	convs, err := rt.db.GetMyConversations(ctx.UserID)
	if err != nil {
		ctx.Logger.WithError(err).Error("listing conversations failed")
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	out := make([]conversationSummaryResponse, 0, len(convs))
	for _, c := range convs {
		out = append(out, conversationSummaryResponse{
			ID:                  c.ID,
			IsGroup:             c.IsGroup,
			Title:               c.Title,
			PhotoURL:            c.PhotoURL,
			OtherUserID:         c.OtherUserID,
			LastMessageAt:       c.LastMessageAt,
			LastMessagePreview:  c.LastMessagePreview,
			LastMessageSenderID: c.LastMessageSenderID,
			LastMessageStatus:   c.LastMessageStatus, // MAPPARE QUESTO
			UnreadCount:         c.UnreadCount,
		})
	}

	type Response struct {
		Conversations []conversationSummaryResponse `json:"conversations"`
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(Response{Conversations: out})
}
