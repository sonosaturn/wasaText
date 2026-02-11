package api

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gofrs/uuid"
	"github.com/julienschmidt/httprouter"
	"github.com/sonosaturn/wasatext/service/api/reqcontext"
)

// sendMessage gestisce POST /conversations/:conversationId/messages
func (rt *_router) sendMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	conversationID := ps.ByName("conversationId")

	// Parse Multipart
	err := r.ParseMultipartForm(10 << 20) // 10MB
	if err != nil {
		http.Error(w, "error parsing form", http.StatusBadRequest)
		return
	}

	content := r.FormValue("content")
	// Leggiamo il nuovo parametro reply_to (sarà vuoto se non è una risposta)
	replyToID := r.FormValue("reply_to")

	photoURL := ""

	// Gestione Foto
	file, handler, err := r.FormFile("file")
	if err == nil {
		defer file.Close()
		
		// FIX: Controllo errore UUID
		imgUUID, err := uuid.NewV4()
		if err != nil {
			ctx.Logger.WithError(err).Error("generating image UUID")
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		
		filename := imgUUID.String() + filepath.Ext(handler.Filename)
		savePath := filepath.Join(".", "images", filename)
		
		// Ignoriamo l'errore della creazione cartella se già esiste
		_ = os.MkdirAll("images", 0755)
		
		// FIX: Controllo errore os.Create
		dst, err := os.Create(savePath)
		if err != nil {
			ctx.Logger.WithError(err).Error("creating image file")
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		defer dst.Close()
		
		// FIX LINTER: Rimosso il duplicato e aggiunto il controllo errore con corpo non vuoto
		if _, err = io.Copy(dst, file); err != nil {
			ctx.Logger.WithError(err).Error("saving image file")
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		photoURL = "/images/" + filename
	}

	// Chiama DB passando anche replyToID
	msg, err := rt.db.SendMessage(conversationID, ctx.UserID, content, photoURL, replyToID)
	if err != nil {
		ctx.Logger.WithError(err).Error("sending message")
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	// FIX: Ignora esplicitamente l'errore dell'encode per il linter
	_ = json.NewEncoder(w).Encode(msg)
}