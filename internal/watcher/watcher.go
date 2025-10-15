package watcher

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
)

// ReloadFunc is called when proto files change
type ReloadFunc func()

// Watcher monitors a directory for .proto file changes
type Watcher struct {
	watcher    *fsnotify.Watcher
	root       string
	reloadFunc ReloadFunc
	debounce   time.Duration
}

// New creates a new file watcher for the given directory
func New(root string, reloadFunc ReloadFunc) (*Watcher, error) {
	fsw, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	w := &Watcher{
		watcher:    fsw,
		root:       root,
		reloadFunc: reloadFunc,
		debounce:   300 * time.Millisecond,
	}

	// Add the root directory and all subdirectories
	if err := w.addRecursive(root); err != nil {
		fsw.Close()
		return nil, err
	}

	return w, nil
}

// addRecursive adds the directory and all subdirectories to the watcher
func (w *Watcher) addRecursive(path string) error {
	return filepath.Walk(path, func(walkPath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// Add directories to watch (not files, since fsnotify watches dirs)
		if info != nil && info.IsDir() {
			if err := w.watcher.Add(walkPath); err != nil {
				return err
			}
		}
		return nil
	})
}

// Start begins watching for file changes
func (w *Watcher) Start(ctx context.Context) {
	var debounceTimer *time.Timer

	for {
		select {
		case <-ctx.Done():
			return
		case event, ok := <-w.watcher.Events:
			if !ok {
				return
			}
			// Only care about .proto files
			if !strings.HasSuffix(strings.ToLower(event.Name), ".proto") {
				continue
			}
			// Watch for create, write, remove, rename operations
			if event.Op&(fsnotify.Write|fsnotify.Create|fsnotify.Remove|fsnotify.Rename) != 0 {
				log.Printf("Proto file changed: %s (%s)", event.Name, event.Op)

				// Debounce: reset timer on each event
				if debounceTimer != nil {
					debounceTimer.Stop()
				}
				debounceTimer = time.AfterFunc(w.debounce, func() {
					log.Println("Reloading proto files...")
					w.reloadFunc()
					log.Println("Proto files reloaded successfully")
				})
			}
		case err, ok := <-w.watcher.Errors:
			if !ok {
				return
			}
			log.Printf("Watcher error: %v", err)
		}
	}
}

// Close stops the watcher
func (w *Watcher) Close() error {
	return w.watcher.Close()
}
