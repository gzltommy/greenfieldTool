package file_type

import (
	"os"
	"testing"
)

func TestFileType(t *testing.T) {
	fileData, err := os.ReadFile("./cc.webp")
	if err != nil {
		t.Log(err)
		return
	}
	imageTypeSS := DetectFileType(fileData)
	t.Log(imageTypeSS)
}
