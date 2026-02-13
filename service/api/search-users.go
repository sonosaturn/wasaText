package api

import (
	"encoding/json"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/sonosaturn/wasatext/service/api/reqcontext"
)

func (rt *_router) searchUsers(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	query := r.URL.Query().Get("q")
	if len(query) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	usersDB, err := rt.db.SearchUsers(query)
	if err != nil {
		ctx.Logger.Error("Errore ricerca utenti: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Struttura di risposta per il JSON
	type UserResponse struct {
		ID       string `json:"id"`
		Username string `json:"username"`
		PhotoURL string `json:"photo_url"` // <--- AGGIUNTO QUESTO CAMPO
	}

	response := make([]UserResponse, 0)

	for _, u := range usersDB {
		response = append(response, UserResponse{
			ID:       u.ID,
			Username: u.Username,
			PhotoURL: u.PhotoURL, // <--- MAPPIAMO IL VALORE DAL DB
		})
	}

	type Wrapper struct {
		Users []UserResponse `json:"users"`
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(Wrapper{Users: response}); err != nil {
		ctx.Logger.Error("Errore encoding risposta: ", err)
	}
}
