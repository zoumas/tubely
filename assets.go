package main

import (
	"bytes"
	"errors"
	"fmt"
	"mime"
	"mime/multipart"
	"os"
	"os/exec"
	"strings"
)

func (cfg apiConfig) ensureAssetsDir() error {
	if _, err := os.Stat(cfg.assetsRoot); os.IsNotExist(err) {
		return os.Mkdir(cfg.assetsRoot, 0755)
	}
	return nil
}

func getImageExtension(h *multipart.FileHeader) (string, error) {
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

func getVideoExtension(h *multipart.FileHeader) (string, error) {
	contentType := h.Header.Get("Content-Type")
	mediaType, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		return "", err
	}
	if mediaType != "video/mp4" {
		return "", fmt.Errorf("unsupported media type %s", mediaType)
	}
	return "mp4", nil
}

const (
	horizontalAspectRatio = "16:9"
	verticalAspectRation  = "9:16"
)

func getVideoAspectRatio(filename string) (string, error) {
	s := fmt.Sprintf(`ffmpeg.ffprobe -v error -print_format json -show_streams %s 2>/dev/null | jq '.streams[] | select(.display_aspect_ratio != null) | .display_aspect_ratio'`,
		filename)

	cmd := exec.Command("bash", "-c", s)

	var b bytes.Buffer
	cmd.Stdout = &b
	if err := cmd.Run(); err != nil {
		return "", err
	}

	aspectRatio := strings.Trim(strings.TrimSpace(b.String()), `"`)

	switch aspectRatio {
	case horizontalAspectRatio:
		return "landscape", nil
	case verticalAspectRation:
		return "portrait", nil
	}
	return "other", nil
}

func processVideoForFastStart(filename string) (string, error) {
	outFilepath := "processing-" + filename
	s := fmt.Sprintf(`ffmpeg -i %s -c copy -movflags faststart -f mp4 %s`, filename, outFilepath)
	cmd := exec.Command("bash", "-c", s)
	if err := cmd.Run(); err != nil {
		return "", err
	}
	return outFilepath, nil
}
