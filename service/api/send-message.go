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

func (rt *_router) sendMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	conversationID := ps.ByName("conversationId")

	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		http.Error(w, "error parsing form", http.StatusBadRequest)
		return
	}

	content := r.FormValue("content")
	replyToID := r.FormValue("reply_to")
	photoURL := ""

	file, handler, err := r.FormFile("file")
	if err == nil {
		defer file.Close()
		imgUUID, err := uuid.NewV4()
		if err != nil {
			ctx.Logger.WithError(err).Error("generating image UUID")
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		filename := imgUUID.String() + filepath.Ext(handler.Filename)
		savePath := filepath.Join(".", "images", filename)
		_ = os.MkdirAll("images", 0755)
		dst, err := os.Create(savePath)
		if err != nil {
			ctx.Logger.WithError(err).Error("creating image file")
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		defer dst.Close()
		if _, err = io.Copy(dst, file); err != nil {
			ctx.Logger.WithError(err).Error("saving image file")
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		photoURL = "/images/" + filename
	}

	msg, err := rt.db.SendMessage(conversationID, ctx.UserID, content, photoURL, replyToID, false)
	if err != nil {
		ctx.Logger.WithError(err).Error("sending message")
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	// FIX: Assign the username from the context, so the frontend sees it right away
	msg.SenderUsername = ctx.UserName

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(msg)
}
