package server

import (
	"log"

	"github.com/fsnotify/fsnotify"
	"github.com/lirrensi/looty/internal/clipboard"
)

func StartWatcher(dir string) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Printf("Failed to create watcher: %v", err)
		return
	}

	err = watcher.Add(dir)
	if err != nil {
		log.Printf("Failed to watch directory: %v", err)
		return
	}

	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Op&fsnotify.Create == fsnotify.Create ||
					event.Op&fsnotify.Write == fsnotify.Write ||
					event.Op&fsnotify.Remove == fsnotify.Remove {
					log.Printf("File changed: %s", event.Name)
					Broadcast(clipboard.NewRefreshMessage())
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Printf("Watcher error: %v", err)
			}
		}
	}()
}
