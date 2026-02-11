package api

import (
	"encoding/json"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/sonosaturn/wasatext/service/api/reqcontext"
	"github.com/sonosaturn/wasatext/service/database"
)

// getConversationMembers gestisce GET /conversations/:conversationId/members
func (rt *_router) getConversationMembers(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	conversationID := ps.ByName("conversationId")

	// Recupera i membri dal DB
	members, err := rt.db.GetGroupMembers(conversationID)
	if err != nil {
		ctx.Logger.WithError(err).Error("Error getting group members")
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	// Inizializza array vuoto se nil
	if members == nil {
		members = []database.User{}
	}

	// FIX LINTER: Wrappiamo la lista in un oggetto JSON
	type Response struct {
		Members []database.User `json:"members"`
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(Response{Members: members})
}