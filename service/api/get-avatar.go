package api

import (
	"errors"
	"net/http"
	"os"
	"path/filepath"

	"github.com/julienschmidt/httprouter"
	"github.com/sonosaturn/wasatext/service/api/reqcontext"
)

// getAvatar gestisce la richiesta per visualizzare una foto profilo
// URL: GET /avatars/:filename
func (rt *_router) getAvatar(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	// 1. Recuperiamo il nome del file dall'URL (es. "10.jpg")
	filename := ps.ByName("filename")

	// 2. Costruiamo il percorso completo (es. "./avatars/10.jpg")
	// Usa filepath.Join per sicurezza (evita hack tipo ../../)
	avatarPath := filepath.Join("avatars", filename)

	// 3. Controlliamo se il file esiste
	if _, err := os.Stat(avatarPath); errors.Is(err, os.ErrNotExist) {
		// Se non esiste, restituiamo 404 Not Found
		http.Error(w, "Image not found", http.StatusNotFound)
		return
	}

	// 4. Serviamo il file direttamente
	// http.ServeFile gestisce da solo Content-Type, caching e stream dei byte
	http.ServeFile(w, r, avatarPath)
}
