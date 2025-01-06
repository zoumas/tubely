package main

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"log"
	"mime"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/database"
)

func (cfg *apiConfig) handlerUploadVideo(w http.ResponseWriter, r *http.Request, user *database.User, video database.Video) {
	const maxUploadSize = 1 << 30 // 1GB
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)
	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		respondWithError(w, http.StatusBadRequest, "parse multipart form", err)
		return
	}

	const formkey = "video"
	file, header, err := r.FormFile(formkey)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "form file", err)
		return
	}
	defer func() {
		if err := file.Close(); err != nil {
			log.Println(err)
		}
	}()

	contentType := header.Header.Get("Content-Type")
	mediaType, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "parsing media type", err)
		return
	}
	if mediaType != "video/mp4" {
		respondWithError(w, http.StatusBadRequest, "unsupported media type", errors.New(mediaType))
		return
	}
	ext := "mp4"

	filename := "tubely-upload" + "." + ext
	temp, err := os.Create(filename)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "creating temp file", err)
		return
	}
	// defer func() {
	// 	if err := os.Remove(temp.Name()); err != nil {
	// 		log.Println(err)
	// 	}
	// }()
	defer func() {
		if err := temp.Close(); err != nil {
			log.Println(err)
		}
	}()

	if _, err := io.Copy(temp, file); err != nil {
		respondWithError(w, http.StatusInternalServerError, "copying from multipart file to temp", err)
		return
	}

	if _, err := temp.Seek(0, io.SeekStart); err != nil {
		respondWithError(w, http.StatusInternalServerError, "reseting temp file pointer to start", err)
		return
	}

	name, err := processVideoForFastStart(temp.Name())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error(), err)
		return
	}

	processedFile, err := os.Open(name)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error(), err)
		return
	}
	defer processedFile.Close()

	log.Println(processedFile.Name())
	aspectRatio, err := getVideoAspectRatio(processedFile.Name())
	if err != nil {
		log.Println("get video aspect ratio", err)
		respondWithError(w, http.StatusInternalServerError, err.Error(), err)
		return
	}

	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		respondWithError(w, http.StatusInternalServerError, "rand read", err)
		return
	}

	s3key := fmt.Sprintf("%s/%s", aspectRatio, base64.URLEncoding.EncodeToString(b)+"."+ext)
	s3params := &s3.PutObjectInput{
		Bucket:      &cfg.s3Bucket,
		Key:         &s3key,
		Body:        processedFile,
		ContentType: &mediaType,
	}
	_, err = cfg.s3Client.PutObject(r.Context(), s3params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "s3 put object", err)
		return
	}

	videoURL := fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", cfg.s3Bucket, cfg.s3Region, s3key)
	video.VideoURL = &videoURL
	if err := cfg.db.UpdateVideo(video); err != nil {
		respondWithError(w, http.StatusInternalServerError, "updating video url", err)
		return
	}

	respondWithJSON(w, http.StatusOK, video)
}
