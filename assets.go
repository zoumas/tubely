package main

import (
	"errors"
	"mime/multipart"
	"os"
	"strings"
)

func (cfg apiConfig) ensureAssetsDir() error {
	if _, err := os.Stat(cfg.assetsRoot); os.IsNotExist(err) {
		return os.Mkdir(cfg.assetsRoot, 0755)
	}
	return nil
}

func getFileExtension(h *multipart.FileHeader) (string, error) {
	mediaType := h.Header.Get("Content-Type")
	fields := strings.Split(mediaType, "/")
	if len(fields) != 2 {
		return "", errors.New("malformed Content-Type header")
	}
	return fields[1], nil
}
