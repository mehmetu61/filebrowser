package fbhttp

import (
	"encoding/json"
	"net/http"

	fberrors "github.com/filebrowser/filebrowser/v2/errors"
	"github.com/filebrowser/filebrowser/v2/files"
	"github.com/filebrowser/filebrowser/v2/fileutils"
)

type BatchRenameItem struct {
	From string `json:"from"`
	To   string `json:"to"`
}

type batchRenameRequest struct {
	Items []BatchRenameItem `json:"items"`
}

type batchRenameResult struct {
	From  string `json:"from"`
	To    string `json:"to"`
	Error string `json:"error,omitempty"`
}

var batchRenameHandler = withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	if !d.user.Perm.Modify {
		return http.StatusForbidden, nil
	}

	var req batchRenameRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return http.StatusBadRequest, err
	}

	if len(req.Items) == 0 {
		return http.StatusBadRequest, fberrors.ErrEmptyRequest
	}

	results := make([]batchRenameResult, 0, len(req.Items))

	for _, item := range req.Items {
		src := files.Clean(item.From)
		dst := files.Clean(item.To)

		if !d.Check(src) || !d.Check(dst) {
			results = append(results, batchRenameResult{
				From:  src,
				To:    dst,
				Error: "permission denied or blocked by rule",
			})
			continue
		}

		if src == dst {
			continue
		}

		err := fileutils.MoveFile(d.user.Fs, src, dst)
		if err != nil {
			results = append(results, batchRenameResult{
				From:  src,
				To:    dst,
				Error: err.Error(),
			})
		} else {
			results = append(results, batchRenameResult{
				From: src,
				To:   dst,
			})
		}
	}

	return renderJSON(w, r, map[string]interface{}{
		"success": true,
		"results": results,
	})
})
