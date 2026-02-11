package api

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/sonosaturn/wasatext/service/api/reqcontext"
)

func (rt *_router) addToGroup(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	// 1. Recuperiamo gli ID dall'URL
	conversationID := ps.ByName("conversationId")
	targetUserID := ps.ByName("userId") // L'utente da aggiungere

	// 2. Chiamiamo il database
	// Nota: Abbiamo rimosso requesterID perché l'interfaccia DB attuale usa solo 2 parametri.
	// (In futuro potresti voler aggiungere un controllo qui per vedere se ctx.UserID è admin del gruppo)
	err := rt.db.AddMemberToGroup(conversationID, targetUserID)

	if err != nil {
		ctx.Logger.Error("Errore aggiunta membro al gruppo: ", err)
		// Se vuoi gestire errori specifici (es. "not a group"), dovresti controllare err.Error()
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// 3. Successo
	w.WriteHeader(http.StatusNoContent)
}