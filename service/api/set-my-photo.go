package api

import (
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/julienschmidt/httprouter"
	"github.com/sonosaturn/wasatext/service/api/reqcontext"
)

// setMyPhoto gestisce l'upload della foto profilo
// PUT /users/:userId/photo
func (rt *_router) setMyPhoto(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	idStr := ps.ByName("userId")
	
	// Parse Multipart Form
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		ctx.Logger.WithError(err).Error("Error parsing multipart form")
		http.Error(w, "invalid form", http.StatusBadRequest)
		return
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		ctx.Logger.WithError(err).Error("Error retrieving file from form")
		http.Error(w, "missing file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Crea la cartella se non esiste
	avatarDir := filepath.Join(".", "avatars")
	_ = os.MkdirAll(avatarDir, 0755)

	// Percorso file: ./avatars/USERID.jpg
	filename := idStr + ".jpg"
	path := filepath.Join(avatarDir, filename)

	dst, err := os.Create(path)
	if err != nil {
		ctx.Logger.WithError(err).Error("Error creating file")
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		ctx.Logger.WithError(err).Error("Error saving file")
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	// >>> FIX IMPORTANTE: Aggiorniamo il Database! <<<
	// Senza questo, al prossimo login la foto tornerà quella di default.
	photoURL := "/avatars/" + filename
	if err := rt.db.SetUserPhoto(idStr, photoURL); err != nil {
		ctx.Logger.WithError(err).Error("Error updating user photo in DB")
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}
    
	w.WriteHeader(http.StatusNoContent)
}