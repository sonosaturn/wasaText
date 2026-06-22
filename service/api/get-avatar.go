package api

import (
	"errors"
	"net/http"
	"os"
	"path/filepath"

	"github.com/julienschmidt/httprouter"
	"github.com/sonosaturn/wasatext/service/api/reqcontext"
)

// getAvatar handles the request to display a profile photo
// URL: GET /avatars/:filename
func (rt *_router) getAvatar(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	// 1. Get the filename from the URL (e.g. "10.jpg")
	filename := ps.ByName("filename")

	// 2. Build the full path (e.g. "./avatars/10.jpg")
	// Use filepath.Join for safety (avoids ../../ -style hacks)
	avatarPath := filepath.Join("avatars", filename)

	// 3. Check if the file exists
	if _, err := os.Stat(avatarPath); errors.Is(err, os.ErrNotExist) {
		// If it doesn't exist, return 404 Not Found
		http.Error(w, "Image not found", http.StatusNotFound)
		return
	}

	// 4. Serve the file directly
	// http.ServeFile handles Content-Type, caching and byte streaming on its own
	http.ServeFile(w, r, avatarPath)
}
