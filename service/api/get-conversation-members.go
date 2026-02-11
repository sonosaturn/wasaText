package api

import (
	"encoding/json"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/sonosaturn/wasatext/service/api/reqcontext"
	"github.com/sonosaturn/wasatext/service/database"
)

func (rt *_router) getConversationMembers(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	conversationID := ps.ByName("conversationId")

	// FIX: Chiamata corretta
	members, err := rt.db.GetConversationMembers(conversationID)
	if err != nil {
		ctx.Logger.WithError(err).Error("Error getting group members")
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	if members == nil {
		members = []database.User{}
	}

	type Response struct {
		Members []database.User `json:"members"`
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(Response{Members: members})
}