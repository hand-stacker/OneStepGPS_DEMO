package main

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
)

// Notice: Most of this code was made with claude

// max size of a device marker image (5MB)
const maxMarkerBytes = 5 << 20

// extra room for multipart boundaries and headers on top of the image
const multipartOverheadBytes = 16 << 10

// form field the frontend should put the file under, e.g. formData.append("image", file)
const markerFormField = "image"

// image types we accept, svg is left out on purpose since it can carry scripts
var allowedMarkerTypes = map[string]bool{
	"image/png":  true,
	"image/jpeg": true,
	"image/gif":  true,
	"image/webp": true,
}

var errMarkerTooLarge = fmt.Errorf("image must be %d MB or smaller", maxMarkerBytes>>20)

// reads the image out of a multipart/form-data request
// returns the raw bytes and the sniffed content type (we dont trust the client's header)
func readMarkerImage(w http.ResponseWriter, r *http.Request) ([]byte, string, error) {
	r.Body = http.MaxBytesReader(w, r.Body, maxMarkerBytes+multipartOverheadBytes)

	file, _, err := r.FormFile(markerFormField)
	if err != nil {
		if _, ok := errors.AsType[*http.MaxBytesError](err); ok {
			return nil, "", errMarkerTooLarge
		}
		return nil, "", fmt.Errorf("expected multipart form with an %q file field: %w", markerFormField, err)
	}
	defer file.Close()

	// read one extra byte so we can tell if the file went over the limit
	data, err := io.ReadAll(io.LimitReader(file, maxMarkerBytes+1))
	if err != nil {
		return nil, "", err
	}
	if len(data) > maxMarkerBytes {
		return nil, "", errMarkerTooLarge
	}
	if len(data) == 0 {
		return nil, "", errors.New("image is empty")
	}

	contentType := http.DetectContentType(data)
	if !allowedMarkerTypes[contentType] {
		return nil, "", fmt.Errorf("unsupported image type %q, use png, jpeg, gif, or webp", contentType)
	}
	return data, contentType, nil
}

// writes the stored marker image back as raw bytes so it can be used directly in an <img> src
func writeMarkerImage(w http.ResponseWriter, m *DeviceMarker) {
	w.Header().Set("Content-Type", m.ContentType)
	w.Header().Set("Content-Length", strconv.Itoa(len(m.Data)))
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Cache-Control", "no-cache")
	w.Write(m.Data)
}

// metadata returned after an upload, we leave out the bytes so the response stays small
type DeviceMarkerInfo struct {
	UserID      *int64 `json:"user_id"`
	DeviceID    string `json:"device_id"`
	ContentType string `json:"content_type"`
	Size        int    `json:"size"`
}

func markerInfo(m *DeviceMarker) DeviceMarkerInfo {
	return DeviceMarkerInfo{
		UserID:      m.UserID,
		DeviceID:    m.DeviceID,
		ContentType: m.ContentType,
		Size:        len(m.Data),
	}
}
