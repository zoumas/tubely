package main

import (
	"errors"
	"fmt"
	"mime"
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

func getImageFileExtension(h *multipart.FileHeader) (string, error) {
	contentType := h.Header.Get("Content-Type")
	mediaType, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		return "", err
	}
	if mediaType != "image/png" && mediaType != "image/jpg" {
		return "", fmt.Errorf("unsupported media type %s", mediaType)
	}
	fields := strings.Split(mediaType, "/")
	if len(fields) != 2 {
		return "", errors.New("malformed Content-Type header")
	}
	return fields[1], nil
}
