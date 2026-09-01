package files

import (
	"log"
	"path/filepath"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

// FSEvent represents a file system change notification
type FSEvent struct {
	Path string `json:"path"`
	Op   string `json:"op"`
}

// DirWatcher watches directories on the host filesystem and broadcasts changes.
type DirWatcher struct {
	watcher     *fsnotify.Watcher
	mu          sync.RWMutex
	subscribers map[string][]chan FSEvent
	watchedDirs map[string]int
	stopCh      chan struct{}
}

var (
	globalWatcher     *DirWatcher
	globalWatcherOnce sync.Once
)

// GetGlobalWatcher returns a singleton directory watcher.
func GetGlobalWatcher() (*DirWatcher, error) {
	var err error
	globalWatcherOnce.Do(func() {
		w, wErr := fsnotify.NewWatcher()
		if wErr != nil {
			err = wErr
			return
		}
		globalWatcher = &DirWatcher{
			watcher:     w,
			subscribers: make(map[string][]chan FSEvent),
			watchedDirs: make(map[string]int),
			stopCh:      make(chan struct{}),
		}
		go globalWatcher.loop()
	})
	return globalWatcher, err
}

func (dw *DirWatcher) loop() {
	debounceMap := make(map[string]time.Time)
	var debounceMu sync.Mutex

	for {
		select {
		case <-dw.stopCh:
			return
		case event, ok := <-dw.watcher.Events:
			if !ok {
				return
			}

			// Debounce rapid events on same file within 100ms
			debounceMu.Lock()
			last, exists := debounceMap[event.Name]
			if exists && time.Since(last) < 100*time.Millisecond {
				debounceMu.Unlock()
				continue
			}
			debounceMap[event.Name] = time.Now()
			debounceMu.Unlock()

			dir := filepath.Dir(event.Name)
			opStr := event.Op.String()

			dw.mu.RLock()
			// Notify subscribers of the directory and the file itself
			for _, targetDir := range []string{dir, event.Name} {
				if chans, ok := dw.subscribers[targetDir]; ok {
					fsEv := FSEvent{Path: event.Name, Op: opStr}
					for _, ch := range chans {
						select {
						case ch <- fsEv:
						default:
						}
					}
				}
			}
			dw.mu.RUnlock()

		case err, ok := <-dw.watcher.Errors:
			if !ok {
				return
			}
			log.Printf("file watcher error: %v", err)
		}
	}
}

// Subscribe listens to changes in a directory.
func (dw *DirWatcher) Subscribe(dirPath string) chan FSEvent {
	dw.mu.Lock()
	defer dw.mu.Unlock()

	ch := make(chan FSEvent, 10)
	dw.subscribers[dirPath] = append(dw.subscribers[dirPath], ch)

	dw.watchedDirs[dirPath]++
	if dw.watchedDirs[dirPath] == 1 {
		_ = dw.watcher.Add(dirPath)
	}

	return ch
}

// Unsubscribe removes a subscriber channel.
func (dw *DirWatcher) Unsubscribe(dirPath string, ch chan FSEvent) {
	dw.mu.Lock()
	defer dw.mu.Unlock()

	if list, ok := dw.subscribers[dirPath]; ok {
		var updated []chan FSEvent
		for _, c := range list {
			if c != ch {
				updated = append(updated, c)
			}
		}
		if len(updated) == 0 {
			delete(dw.subscribers, dirPath)
		} else {
			dw.subscribers[dirPath] = updated
		}
	}

	if dw.watchedDirs[dirPath] > 0 {
		dw.watchedDirs[dirPath]--
		if dw.watchedDirs[dirPath] == 0 {
			delete(dw.watchedDirs, dirPath)
			_ = dw.watcher.Remove(dirPath)
		}
	}
}
