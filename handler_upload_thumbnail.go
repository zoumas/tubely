package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/database"
)

func (cfg *apiConfig) handlerUploadThumbnail(w http.ResponseWriter, r *http.Request, user *database.User, video database.Video) {
	fmt.Println("uploading thumbnail for video", video.ID, "by user", user.ID)

	const maxMemory = 10 << 20
	if err := r.ParseMultipartForm(maxMemory); err != nil {
		respondWithError(w, http.StatusBadRequest, "parse multipart form", err)
		return
	}

	const formkey = "thumbnail"
	file, header, err := r.FormFile(formkey)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "form file", err)
		return
	}
	defer file.Close()

	ext, err := getImageFileExtension(header)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "parsing media type", err)
		return
	}

	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		respondWithError(w, http.StatusInternalServerError, "rand read", err)
		return
	}

	filename := filepath.Join(cfg.assetsRoot, base64.RawURLEncoding.EncodeToString(b)+"."+ext)
	dst, err := os.Create(filename)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "creating asset", err)
		return
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		respondWithError(w, http.StatusInternalServerError, "copying to asset", err)
		return
	}

	thumbnailURL := fmt.Sprintf("http://localhost:%s/%s", cfg.port, filename)
	video.ThumbnailURL = &thumbnailURL
	if err := cfg.db.UpdateVideo(video); err != nil {
		respondWithError(w, http.StatusInternalServerError, "update video", err)
		return
	}

	respondWithJSON(w, http.StatusOK, video)
}
