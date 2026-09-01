package fbhttp

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"

	fberrors "github.com/filebrowser/filebrowser/v2/errors"
)

type ChecksumResponse struct {
	Path   string `json:"path"`
	Size   int64  `json:"size"`
	MD5    string `json:"md5"`
	SHA1   string `json:"sha1"`
	SHA256 string `json:"sha256"`
}

var checksumHandler = withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	filePath := slashClean(r.URL.Path)
	if !d.Check(filePath) {
		return http.StatusForbidden, nil
	}

	info, err := d.user.Fs.Stat(filePath)
	if err != nil {
		return errToStatus(err), err
	}
	if info.IsDir() {
		return http.StatusBadRequest, fberrors.ErrInvalidDataType
	}

	file, err := d.user.Fs.Open(filePath)
	if err != nil {
		return errToStatus(err), err
	}
	defer file.Close()

	hMD5 := md5.New()
	hSHA1 := sha1.New()
	hSHA256 := sha256.New()

	writer := io.MultiWriter(hMD5, hSHA1, hSHA256)
	written, err := io.Copy(writer, file)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	return renderJSON(w, r, &ChecksumResponse{
		Path:   filePath,
		Size:   written,
		MD5:    hex.EncodeToString(hMD5.Sum(nil)),
		SHA1:   hex.EncodeToString(hSHA1.Sum(nil)),
		SHA256: hex.EncodeToString(hSHA256.Sum(nil)),
	})
})
