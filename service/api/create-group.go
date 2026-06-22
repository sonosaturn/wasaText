package api

import (
	"encoding/json"
	"net/http"
	// "strconv" <--- REMOVED: No longer needed because groupID is already a string

	"github.com/julienschmidt/httprouter"
	"github.com/sonosaturn/wasatext/service/api/reqcontext"
)

type createGroupRequest struct {
	Name    string   `json:"name"`
	Members []string `json:"members"` // Strings (UUID)
}

type groupResponse struct {
	ConversationID string `json:"conversationId"`
}

// createGroup handles POST /groups
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

	// Add yourself (creator) to the members if you're not already there
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

	// Call the database
	groupID, err := rt.db.CreateGroup(req.Name, req.Members)
	if err != nil {
		ctx.Logger.WithError(err).Error("create group failed")
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	// Return the group ID
	// FIX: groupID is already a string (UUID), strconv.FormatInt isn't needed
	res := groupResponse{
		ConversationID: groupID,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(res)
}
