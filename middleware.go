package main

import (
	"errors"
	"log"
	"net/http"

	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/auth"
	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/database"
	"github.com/google/uuid"
)

type videoHandler func(w http.ResponseWriter, r *http.Request, user *database.User, video database.Video)

func (cfg *apiConfig) middlewareVideo(h videoHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token, err := auth.GetBearerToken(r.Header)
		if err != nil {
			respondWithError(w, http.StatusUnauthorized, "getting bearer jwt token", err)
			return
		}
		userID, err := auth.ValidateJWT(token, cfg.jwtSecret)
		if err != nil {
			respondWithError(w, http.StatusUnauthorized, "validating jwt", err)
			return
		}
		user, err := cfg.db.GetUser(userID)
		if err != nil {
			respondWithError(w, http.StatusUnauthorized, "getting user", err)
			return
		}

		videoID, err := uuid.Parse(r.PathValue("videoID"))
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "invalid video id", err)
			return
		}
		video, err := cfg.db.GetVideo(videoID)
		if err != nil {
			respondWithError(w, http.StatusNotFound, "getting video", err)
			return
		}

		if video.UserID != userID {
			respondWithError(w, http.StatusUnauthorized, "user is not video owner", errors.New("ownership"))
			return
		}

		h(w, r, user, video)
	}
}

type StatusCodeResponseWriter struct {
	http.ResponseWriter
	StatusCode int
}

func NewStatusCodeResponseWriter(w http.ResponseWriter) *StatusCodeResponseWriter {
	return &StatusCodeResponseWriter{ResponseWriter: w}
}

func (w *StatusCodeResponseWriter) WriteHeader(statusCode int) {
	w.ResponseWriter.WriteHeader(statusCode)
	w.StatusCode = statusCode
}

func middlewareLogger(h http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		wr := NewStatusCodeResponseWriter(w)
		h.ServeHTTP(wr, r)
		log.Println(r.Method, r.URL.Path, wr.StatusCode)
	}
}
