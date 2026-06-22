package api

import (
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/julienschmidt/httprouter"
	"github.com/sonosaturn/wasatext/service/api/reqcontext"
)

// setMyPhoto handles the profile photo upload
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

	// Create the folder if it doesn't exist
	avatarDir := filepath.Join(".", "avatars")
	_ = os.MkdirAll(avatarDir, 0755)

	// File path: ./avatars/USERID.jpg
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

	// >>> IMPORTANT FIX: Update the Database! <<<
	// Without this, the photo will revert to the default one on the next login.
	photoURL := "/avatars/" + filename
	if err := rt.db.SetUserPhoto(idStr, photoURL); err != nil {
		ctx.Logger.WithError(err).Error("Error updating user photo in DB")
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
