package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/julienschmidt/httprouter"
	"github.com/sonosaturn/wasatext/service/api/reqcontext"
)

// forwardMessage gestisce l'inoltro di un messaggio esistente verso un'altra chat
// POST /conversations/:conversationId/messages/forward
func (rt *_router) forwardMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	// 1. Chat di destinazione (dove voglio mandare il messaggio)
	destConversationID := ps.ByName("conversationId")

	// --- FIX: Recupero Manuale ID Utente (come in set-username) ---
	if ctx.UserID == "" {
		authHeader := r.Header.Get("Authorization")
		if strings.HasPrefix(authHeader, "Bearer ") {
			ctx.UserID = strings.TrimPrefix(authHeader, "Bearer ")
		}
	}
	
	// 2. Leggi dal body quale messaggio vogliamo inoltrare
	var body struct {
		SourceMessageID string `json:"message_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	fmt.Printf("DEBUG FORWARD: Messaggio Sorgente ID=%s\n", body.SourceMessageID)

	// 3. Recupera il contenuto originale (sicuro: controlla se l'utente può vederlo)
	content, photoURL, err := rt.db.GetMessageForForwarding(body.SourceMessageID, ctx.UserID)
	if err != nil {
		// Se l'errore persiste, vedremo esattamente perché nel terminale
		ctx.Logger.WithError(err).Error("fetching source message for forward")
		http.Error(w, "message not found or user not authorized", http.StatusNotFound)
		return
	}

	// 4. Invia il nuovo messaggio nella chat di destinazione
	forwardedMsg, err := rt.db.SendMessage(destConversationID, ctx.UserID, content, photoURL, "")
	if err != nil {
		ctx.Logger.WithError(err).Error("forwarding message")
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	// 5. Rispondi col nuovo messaggio creato
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(forwardedMsg)
}