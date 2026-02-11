package api

import (
	"encoding/json"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/sonosaturn/wasatext/service/api/reqcontext"
)

// loginRequest represents the JSON request body for the login endpoint.
type loginRequest struct {
	Name string `json:"name"`
}

// sessionResponse represents the JSON response body sent after a successful login.
type sessionResponse struct {
	Identifier string `json:"identifier"`
}

// doLogin handles POST /session.
// It validates the request body, delegates the operation to the database layer, and returns the user identifier.
func (rt *_router) doLogin(w http.ResponseWriter, r *http.Request, _ httprouter.Params, ctx reqcontext.RequestContext) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ctx.Logger.WithError(err).Warn("invalid login request body")
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if len(req.Name) < 3 || len(req.Name) > 16 {
		http.Error(w, "name must be between 3 and 16 characters", http.StatusBadRequest)
		return
	}

	id, err := rt.db.DoLogin(req.Name)
	if err != nil {
		ctx.Logger.WithError(err).Warn("login failed")
		http.Error(w, "login failed", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(sessionResponse{Identifier: id})
}
