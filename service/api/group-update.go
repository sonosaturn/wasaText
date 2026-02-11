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

// setGroupName gestisce PUT /conversations/:id/name
func (rt *_router) setGroupName(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	conversationID := ps.ByName("conversationId")

	var body struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	// FIX: Chiamata corretta
	if err := rt.db.SetConversationName(conversationID, body.Name); err != nil {
		ctx.Logger.WithError(err).Error("updating group name")
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// setGroupPhoto gestisce PUT /conversations/:id/photo
func (rt *_router) setGroupPhoto(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	conversationID := ps.ByName("conversationId")

	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		http.Error(w, "error parsing form", http.StatusBadRequest)
		return
	}

	file, handler, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "missing file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	imgUUID, _ := uuid.NewV4()
	filename := "group_" + imgUUID.String() + filepath.Ext(handler.Filename)
	saveDir := filepath.Join(".", "images")
	savePath := filepath.Join(saveDir, filename)

	_ = os.MkdirAll(saveDir, 0755)
	dst, err := os.Create(savePath)
	if err != nil {
		http.Error(w, "error saving file", http.StatusInternalServerError)
		return
	}
	defer dst.Close()
	if _, err := io.Copy(dst, file); err != nil {
		http.Error(w, "error writing file", http.StatusInternalServerError)
		return
	}

	photoURL := "/images/" + filename
	// FIX: Chiamata corretta
	if err := rt.db.SetConversationPhoto(conversationID, photoURL); err != nil {
		ctx.Logger.WithError(err).Error("updating group photo db")
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"photo_url": photoURL})
}