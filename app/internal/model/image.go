package model

import "github.com/gabriel-vasile/mimetype"

const (
	MimeTypeJpeg = "image/jpeg"
	MimeTypePng  = "image/png"
)

var SupportedImageMimeTypes = map[string]struct{}{
	MimeTypeJpeg: {},
	MimeTypePng:  {},
}

func IsImageTypeAllowed(data []byte) bool {
	mime := mimetype.Detect(data).String()
	_, ok := SupportedImageMimeTypes[mime]
	return ok
}
