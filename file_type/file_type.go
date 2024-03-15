package file_type

import (
	"github.com/gabriel-vasile/mimetype"
)

const (
	FileContentTypeMP4  = "video/mp4"
	FileContentTypeJpeg = "image/jpeg" // jpg/jpeg
	FileContentTypePng  = "image/png"  // png
	FileContentTypeWebp = "image/webp" // webp
)

func DetectFileType(f []byte) string {
	return mimetype.Detect(f).String()
}
