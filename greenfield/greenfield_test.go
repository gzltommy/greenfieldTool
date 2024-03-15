package greenfield

import (
	exterrors "github.com/pkg/errors"
	"os"
	"path"
	"path/filepath"
	"testing"
)

const (
	privateKey  = "3" // 3
	privateKey1 = "1" // 1
	bucketName1 = "defution1"
)

func TestUploadResource(t *testing.T) {
	err := Init([]string{privateKey}, bucketName1, true)
	if err != nil {
		t.Logf("%T", exterrors.Cause(err))
		t.Logf("%v", exterrors.Cause(err))
		t.Logf("%v", err)
		t.Logf("%+v", err)
		return
	}

	fName := "./aa.png"
	fileData, err := os.ReadFile(fName)
	if err != nil {
		t.Fatal(err)
	}

	ext := path.Ext(filepath.Base(fName)) //文件扩展名
	sPath, err := UploadResource(fileData, ext, "nft/image")
	if err != nil {
		t.Logf("%T", exterrors.Cause(err))
		t.Logf("%v", exterrors.Cause(err))
		t.Logf("%v", err)
		t.Logf("%+v", err)
		return
	}
	t.Logf("11111---%s", sPath)
}
