package api

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/sonosaturn/wasatext/service/api/reqcontext"
)

func (rt *_router) leaveGroup(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	// 1. Prendi i parametri dall'URL
	conversationID := ps.ByName("conversationId")
	targetUserID := ps.ByName("userId")

	// 2. Chiama il database
	// Rimossa requesterID per compatibilità con la firma del DB
	err := rt.db.RemoveMemberFromGroup(conversationID, targetUserID)
	
	if err != nil {
		errMsg := err.Error()
		if errMsg == "member not found" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		
		ctx.Logger.Error("Errore uscita gruppo: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// 3. Successo
	w.WriteHeader(http.StatusNoContent)
}