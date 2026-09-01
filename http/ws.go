package fbhttp

import (
	"encoding/json"
	"net/http"
	"path/filepath"
	"sync"
	"time"

	"github.com/gorilla/websocket"

	"github.com/filebrowser/filebrowser/v2/files"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Origin checked via auth token
	},
}

type wsMessage struct {
	Action string `json:"action"`
	Path   string `json:"path"`
}

type wsResponse struct {
	Type string `json:"type"`
	Path string `json:"path,omitempty"`
	Op   string `json:"op,omitempty"`
}

func wsHandler() handleFunc {
	return withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return http.StatusBadRequest, err
		}
		defer conn.Close()

		watcher, err := files.GetGlobalWatcher()
		if err != nil {
			return http.StatusInternalServerError, err
		}

		var (
			currentDir    string
			currentSubCh  chan files.FSEvent
			subMu         sync.Mutex
			closeCh       = make(chan struct{})
			writeMu       sync.Mutex
		)

		defer func() {
			close(closeCh)
			subMu.Lock()
			if currentSubCh != nil && currentDir != "" {
				watcher.Unsubscribe(currentDir, currentSubCh)
			}
			subMu.Unlock()
		}()

		sendJSON := func(v interface{}) error {
			writeMu.Lock()
			defer writeMu.Unlock()
			_ = conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
			return conn.WriteJSON(v)
		}

		// Reader loop
		go func() {
			for {
				_, msgBytes, err := conn.ReadMessage()
				if err != nil {
					return
				}

				var msg wsMessage
				if err := json.Unmarshal(msgBytes, &msg); err != nil {
					continue
				}

				switch msg.Action {
				case "ping":
					_ = sendJSON(wsResponse{Type: "pong"})
				case "watch":
					// Resolve requested virtual path to real filesystem path within user scope
					cleanPath := filepath.Clean(filepath.Join("/", msg.Path))
					realPath := filepath.Join(d.user.Scope, cleanPath)

					subMu.Lock()
					if currentSubCh != nil && currentDir != "" {
						watcher.Unsubscribe(currentDir, currentSubCh)
					}
					currentDir = realPath
					currentSubCh = watcher.Subscribe(realPath)
					subCh := currentSubCh
					subMu.Unlock()

					// Event broadcaster loop for this subscription
					go func(ch chan files.FSEvent, watchedPath, virtPath string) {
						for {
							select {
							case <-closeCh:
								return
							case ev, ok := <-ch:
								if !ok {
									return
								}
								_ = sendJSON(wsResponse{
									Type: "change",
									Path: virtPath,
									Op:   ev.Op,
								})
							}
						}
					}(subCh, realPath, cleanPath)
				}
			}
		}()

		// Keep connection alive with pings
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			if err := sendJSON(wsResponse{Type: "ping"}); err != nil {
				break
			}
		}

		return 0, nil
	})
}
