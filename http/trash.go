package fbhttp

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/filebrowser/filebrowser/v2/files"
	"github.com/spf13/afero"
)

const (
	TrashDirName      = ".filebrowser_trash"
	TrashMetadataFile = ".filebrowser_trash/metadata.json"
	TrashFilesDir     = ".filebrowser_trash/files"
)

// TrashItem represents an item moved to the trash bin.
type TrashItem struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	OriginalPath string    `json:"originalPath"`
	DeletedAt    time.Time `json:"deletedAt"`
	Size         int64     `json:"size"`
	IsDir        bool      `json:"isDir"`
	TrashPath    string    `json:"trashPath"`
}

type restoreRequest struct {
	IDs []string `json:"ids"`
}

type trashDeleteRequest struct {
	IDs []string `json:"ids"`
}

func generateTrashID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return fmt.Sprintf("%d_%s", time.Now().UnixNano(), hex.EncodeToString(b))
}

func loadTrashMetadata(fs afero.Fs) ([]TrashItem, error) {
	exists, err := afero.Exists(fs, TrashMetadataFile)
	if err != nil || !exists {
		return []TrashItem{}, nil
	}

	data, err := afero.ReadFile(fs, TrashMetadataFile)
	if err != nil {
		return []TrashItem{}, err
	}

	var items []TrashItem
	if err := json.Unmarshal(data, &items); err != nil {
		return []TrashItem{}, nil
	}

	return items, nil
}

func saveTrashMetadata(fs afero.Fs, items []TrashItem) error {
	_ = fs.MkdirAll(TrashDirName, 0755)
	data, err := json.MarshalIndent(items, "", "  ")
	if err != nil {
		return err
	}
	return afero.WriteFile(fs, TrashMetadataFile, data, 0644)
}

func moveToTrash(fs afero.Fs, relPath string, fi *files.FileInfo) (*TrashItem, error) {
	relPath = slashClean(relPath)
	if relPath == "" || relPath == "/" || relPath == "." {
		return nil, fmt.Errorf("cannot delete root")
	}

	// If already in trash, just remove permanently
	if strings.HasPrefix(relPath, "/"+TrashDirName) || strings.HasPrefix(relPath, TrashDirName) {
		return nil, fs.RemoveAll(relPath)
	}

	if err := fs.MkdirAll(TrashFilesDir, 0755); err != nil {
		return nil, err
	}

	id := generateTrashID()
	destPath := path.Join(TrashFilesDir, id)

	// Move file/folder to trash storage
	if err := fs.Rename(relPath, destPath); err != nil {
		// Fallback: if rename across boundaries fails, copy and delete
		return nil, err
	}

	item := TrashItem{
		ID:           id,
		Name:         fi.Name,
		OriginalPath: relPath,
		DeletedAt:    time.Now(),
		Size:         fi.Size,
		IsDir:        fi.IsDir,
		TrashPath:    destPath,
	}

	items, _ := loadTrashMetadata(fs)
	items = append(items, item)
	_ = saveTrashMetadata(fs, items)

	return &item, nil
}

// trashListHandler returns all items currently in the trash bin.
var trashListHandler = withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	items, err := loadTrashMetadata(d.user.Fs)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	// Sort newest first
	sort.Slice(items, func(i, j int) bool {
		return items[i].DeletedAt.After(items[j].DeletedAt)
	})

	return renderJSON(w, r, items)
})

// trashRestoreHandler restores selected trash items to their original location.
var trashRestoreHandler = withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	if !d.user.Perm.Create || !d.user.Perm.Modify {
		return http.StatusForbidden, nil
	}

	var req restoreRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return http.StatusBadRequest, err
	}

	if len(req.IDs) == 0 {
		return renderJSON(w, r, map[string]interface{}{"restored": []string{}})
	}

	idSet := make(map[string]bool)
	for _, id := range req.IDs {
		idSet[id] = true
	}

	items, err := loadTrashMetadata(d.user.Fs)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	var remaining []TrashItem
	var restored []string

	for _, item := range items {
		if idSet[item.ID] {
			targetDir := path.Dir(item.OriginalPath)
			if targetDir != "" && targetDir != "/" && targetDir != "." {
				_ = d.user.Fs.MkdirAll(targetDir, 0755)
			}

			// If target already exists, resolve collision
			targetPath := item.OriginalPath
			if exists, _ := afero.Exists(d.user.Fs, targetPath); exists {
				ext := filepath.Ext(targetPath)
				base := strings.TrimSuffix(targetPath, ext)
				targetPath = fmt.Sprintf("%s_restored_%d%s", base, time.Now().Unix(), ext)
			}

			if err := d.user.Fs.Rename(item.TrashPath, targetPath); err != nil {
				// Keep in trash if restore failed
				remaining = append(remaining, item)
				continue
			}

			restored = append(restored, item.OriginalPath)
		} else {
			remaining = append(remaining, item)
		}
	}

	_ = saveTrashMetadata(d.user.Fs, remaining)

	return renderJSON(w, r, map[string]interface{}{
		"restored": restored,
	})
})

// trashDeleteHandler permanently deletes items from the trash bin or empties the entire trash.
var trashDeleteHandler = withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	if !d.user.Perm.Delete {
		return http.StatusForbidden, nil
	}

	all := r.URL.Query().Get("all") == "true"

	if all {
		_ = d.user.Fs.RemoveAll(TrashDirName)
		return http.StatusNoContent, nil
	}

	var req trashDeleteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return http.StatusBadRequest, err
	}

	if len(req.IDs) == 0 {
		return http.StatusNoContent, nil
	}

	idSet := make(map[string]bool)
	for _, id := range req.IDs {
		idSet[id] = true
	}

	items, err := loadTrashMetadata(d.user.Fs)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	var remaining []TrashItem
	for _, item := range items {
		if idSet[item.ID] {
			_ = d.user.Fs.RemoveAll(item.TrashPath)
		} else {
			remaining = append(remaining, item)
		}
	}

	_ = saveTrashMetadata(d.user.Fs, remaining)

	return http.StatusNoContent, nil
})
