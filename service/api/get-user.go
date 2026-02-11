package api

import (
	"encoding/json"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/sonosaturn/wasatext/service/api/reqcontext"
)

func (rt *_router) getUserById(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	userId := ps.ByName("userId")

	// Cerchiamo l'utente nel DB (dobbiamo aggiungere questa funzione al DB tra poco)
	user, err := rt.db.GetUserById(userId)
	if err != nil {
		ctx.Logger.Error("Errore recupero utente: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
    
    // Se l'utente non esiste (user nil o vuoto), 404
    if user.ID == "" {
        w.WriteHeader(http.StatusNotFound)
        return
    }

	// Restituiamo il JSON dell'utente (che include PhotoURL!)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(user)
}