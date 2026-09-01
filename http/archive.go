package fbhttp

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	fberrors "github.com/filebrowser/filebrowser/v2/errors"
	"github.com/filebrowser/filebrowser/v2/files"
	"github.com/spf13/afero"
)

type extractRequest struct {
	Source      string `json:"source"`
	Destination string `json:"destination"`
}

type compressRequest struct {
	Items       []string `json:"items"`
	Destination string   `json:"destination"`
	ArchiveName string   `json:"archiveName"`
}

var archiveExtractHandler = withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	if !d.user.Perm.Create || !d.user.Perm.Modify {
		return http.StatusForbidden, nil
	}

	var req extractRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return http.StatusBadRequest, err
	}

	srcPath := files.Clean(req.Source)
	if !d.Check(srcPath) {
		return http.StatusForbidden, nil
	}

	destPath := files.Clean(req.Destination)
	if destPath == "" || destPath == "." {
		destPath = path.Dir(srcPath)
	}
	if !d.Check(destPath) {
		return http.StatusForbidden, nil
	}

	// Open source archive
	srcFile, err := d.user.Fs.Open(srcPath)
	if err != nil {
		return errToStatus(err), err
	}
	defer srcFile.Close()

	srcStat, err := srcFile.Stat()
	if err != nil {
		return errToStatus(err), err
	}

	lowerSrc := strings.ToLower(srcPath)
	if strings.HasSuffix(lowerSrc, ".zip") {
		// ZIP extraction
		zipReader, err := zip.NewReader(srcFile, srcStat.Size())
		if err != nil {
			return http.StatusBadRequest, fmt.Errorf("failed to open zip: %w", err)
		}

		for _, f := range zipReader.File {
			cleanName := filepath.Clean(f.Name)
			// Prevent ZipSlip / path traversal
			if strings.HasPrefix(cleanName, "..") || strings.HasPrefix(cleanName, "/") || strings.HasPrefix(cleanName, "\\") {
				continue
			}

			targetPath := path.Join(destPath, filepath.ToSlash(cleanName))
			if !d.Check(targetPath) {
				continue
			}

			if f.FileInfo().IsDir() {
				_ = d.user.Fs.MkdirAll(targetPath, 0755)
				continue
			}

			if err := d.user.Fs.MkdirAll(path.Dir(targetPath), 0755); err != nil {
				return http.StatusInternalServerError, err
			}

			rc, err := f.Open()
			if err != nil {
				return http.StatusInternalServerError, err
			}

			outFile, err := d.user.Fs.OpenFile(targetPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				rc.Close()
				return http.StatusInternalServerError, err
			}

			_, copyErr := io.Copy(outFile, rc)
			rc.Close()
			outFile.Close()
			if copyErr != nil {
				return http.StatusInternalServerError, copyErr
			}
		}
	} else if strings.HasSuffix(lowerSrc, ".tar.gz") || strings.HasSuffix(lowerSrc, ".tgz") || strings.HasSuffix(lowerSrc, ".tar") {
		// TAR / TAR.GZ extraction
		var tarReader *tar.Reader
		if strings.HasSuffix(lowerSrc, ".tar") {
			tarReader = tar.NewReader(srcFile)
		} else {
			gzReader, err := gzip.NewReader(srcFile)
			if err != nil {
				return http.StatusBadRequest, fmt.Errorf("failed to open gzip: %w", err)
			}
			defer gzReader.Close()
			tarReader = tar.NewReader(gzReader)
		}

		for {
			header, err := tarReader.Next()
			if err == io.EOF {
				break
			}
			if err != nil {
				return http.StatusBadRequest, fmt.Errorf("corrupt tar archive: %w", err)
			}

			cleanName := filepath.Clean(header.Name)
			// Prevent ZipSlip / path traversal
			if strings.HasPrefix(cleanName, "..") || strings.HasPrefix(cleanName, "/") || strings.HasPrefix(cleanName, "\\") {
				continue
			}

			targetPath := path.Join(destPath, filepath.ToSlash(cleanName))
			if !d.Check(targetPath) {
				continue
			}

			switch header.Typeflag {
			case tar.TypeDir:
				_ = d.user.Fs.MkdirAll(targetPath, 0755)
			case tar.TypeReg:
				if err := d.user.Fs.MkdirAll(path.Dir(targetPath), 0755); err != nil {
					return http.StatusInternalServerError, err
				}
				outFile, err := d.user.Fs.OpenFile(targetPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.FileMode(header.Mode))
				if err != nil {
					return http.StatusInternalServerError, err
				}
				if _, err := io.Copy(outFile, tarReader); err != nil {
					outFile.Close()
					return http.StatusInternalServerError, err
				}
				outFile.Close()
			}
		}
	} else {
		return http.StatusBadRequest, fmt.Errorf("unsupported archive format")
	}

	return renderJSON(w, r, map[string]interface{}{
		"success": true,
		"path":    destPath,
	})
})

var archiveCompressHandler = withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	if !d.user.Perm.Create {
		return http.StatusForbidden, nil
	}

	var req compressRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return http.StatusBadRequest, err
	}

	if len(req.Items) == 0 {
		return http.StatusBadRequest, fberrors.ErrEmptyRequest
	}

	archiveName := strings.TrimSpace(req.ArchiveName)
	if archiveName == "" {
		archiveName = "archive.zip"
	}
	if !strings.HasSuffix(strings.ToLower(archiveName), ".zip") {
		archiveName += ".zip"
	}

	destDir := files.Clean(req.Destination)
	if !d.Check(destDir) {
		return http.StatusForbidden, nil
	}

	archivePath := path.Join(destDir, archiveName)
	if !d.Check(archivePath) {
		return http.StatusForbidden, nil
	}

	zipFile, err := d.user.Fs.OpenFile(archivePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return http.StatusInternalServerError, err
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	for _, item := range req.Items {
		itemPath := files.Clean(item)
		if !d.Check(itemPath) {
			continue
		}

		info, err := d.user.Fs.Stat(itemPath)
		if err != nil {
			continue
		}

		baseDir := path.Dir(itemPath)
		if info.IsDir() {
			err = afero.Walk(d.user.Fs, itemPath, func(fPath string, fInfo os.FileInfo, walkErr error) error {
				if walkErr != nil || !d.Check(fPath) {
					return nil
				}
				rel, err := filepath.Rel(baseDir, fPath)
				if err != nil {
					return nil
				}
				rel = filepath.ToSlash(rel)

				if fInfo.IsDir() {
					if !strings.HasSuffix(rel, "/") {
						rel += "/"
					}
					_, _ = zipWriter.Create(rel)
					return nil
				}

				header, err := zip.FileInfoHeader(fInfo)
				if err != nil {
					return nil
				}
				header.Name = rel
				header.Method = zip.Deflate

				w, err := zipWriter.CreateHeader(header)
				if err != nil {
					return nil
				}

				file, err := d.user.Fs.Open(fPath)
				if err != nil {
					return nil
				}
				defer file.Close()
				_, _ = io.Copy(w, file)
				return nil
			})
			if err != nil {
				return http.StatusInternalServerError, err
			}
		} else {
			rel := path.Base(itemPath)
			header, err := zip.FileInfoHeader(info)
			if err != nil {
				continue
			}
			header.Name = rel
			header.Method = zip.Deflate

			w, err := zipWriter.CreateHeader(header)
			if err != nil {
				continue
			}

			file, err := d.user.Fs.Open(itemPath)
			if err != nil {
				continue
			}
			_, _ = io.Copy(w, file)
			file.Close()
		}
	}

	return renderJSON(w, r, map[string]interface{}{
		"success": true,
		"path":    archivePath,
	})
})
