package api

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gofrs/uuid"
	"github.com/julienschmidt/httprouter"
	"github.com/sonosaturn/wasatext/service/api/reqcontext"
)

// uploadPhoto handles POST /photos
func (rt *_router) uploadPhoto(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, "invalid form", http.StatusBadRequest)
		return
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "missing file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	photoID, err := uuid.NewV4()
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	photosDir := filepath.Join(".", "photos")
	_ = os.MkdirAll(photosDir, 0755)

	path := filepath.Join(photosDir, photoID.String()+".jpg")
	dst, err := os.Create(path)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]string{"id": photoID.String()})
}

// getPhoto handles GET /photos/:photoId
func (rt *_router) getPhoto(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	photoID := ps.ByName("photoId")

	path := filepath.Join(".", "photos", photoID+".jpg")
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	http.ServeFile(w, r, path)
}
