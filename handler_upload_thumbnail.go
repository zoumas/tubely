package main

import (
	"fmt"
	"io"
	"net/http"

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

	mediaType := header.Header.Get("Content-Type")
	if mediaType == "" {
		respondWithError(w, http.StatusBadRequest, "missing Content-Type", nil)
		return
	}

	data, err := io.ReadAll(file)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "read all", err)
		return
	}

	thumbnailURL := fmt.Sprintf("http://localhost:%s/api/thumbnails/%s", cfg.port, video.ID.String())
	video.ThumbnailURL = &thumbnailURL
	if err := cfg.db.UpdateVideo(video); err != nil {
		respondWithError(w, http.StatusInternalServerError, "update video", err)
		return
	}

	videoThumbnails[video.ID] = thumbnail{
		mediaType: mediaType,
		data:      data,
	}

	respondWithJSON(w, http.StatusOK, video)
}
