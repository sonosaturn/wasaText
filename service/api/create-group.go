package api

import (
	"encoding/json"
	"net/http"
	// "strconv" <--- RIMOSSO: Non serve più perché groupID è già stringa

	"github.com/julienschmidt/httprouter"
	"github.com/sonosaturn/wasatext/service/api/reqcontext"
)

type createGroupRequest struct {
	Name    string   `json:"name"`
	Members []string `json:"members"` // Stringhe (UUID)
}

type groupResponse struct {
	ConversationID string `json:"conversationId"`
}

// createGroup gestisce POST /groups
func (rt *_router) createGroup(w http.ResponseWriter, r *http.Request, _ httprouter.Params, ctx reqcontext.RequestContext) {
	var req createGroupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		http.Error(w, "group name is required", http.StatusBadRequest)
		return
	}

	// Aggiungi te stesso (creatore) ai membri se non ci sei già
	found := false
	for _, m := range req.Members {
		if m == ctx.UserID {
			found = true
			break
		}
	}
	if !found {
		req.Members = append(req.Members, ctx.UserID)
	}

	// Chiama il database
	groupID, err := rt.db.CreateGroup(req.Name, req.Members)
	if err != nil {
		ctx.Logger.WithError(err).Error("create group failed")
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	// Restituisci l'ID del gruppo
	// FIX: groupID è già una stringa (UUID), non serve strconv.FormatInt
	res := groupResponse{
		ConversationID: groupID,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(res)
}