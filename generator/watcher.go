package generator

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/fsnotify/fsnotify"
	"golang.org/x/sync/errgroup"
)

type Watcher struct {
	isWorking   bool
	isWorkingMu sync.RWMutex
	eventCh     chan fsnotify.Event
	handler     func(context.Context, fsnotify.Event)
	watcher     *fsnotify.Watcher
}

func NewWatcher() (*Watcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	return &Watcher{
		eventCh: make(chan fsnotify.Event),
		handler: func(context.Context, fsnotify.Event) {},
		watcher: watcher,
	}, nil
}

func (w *Watcher) Close() {
	w.watcher.Close()
}

func (w *Watcher) SetHandler(handler func(ctx context.Context, event fsnotify.Event)) {
	w.handler = handler
}

func (w *Watcher) SetWatchPath(path ...string) error {
	return w.setWatchPathRecursive(path...)
}

func (w *Watcher) setWatchPathRecursive(path ...string) error {
	pathMap := make(map[string]struct{})
	for _, p := range path {
		if err := filepath.Walk(p, func(path string, info os.FileInfo, err error) error {
			if info == nil {
				return nil
			}
			if !info.IsDir() {
				return nil
			}
			if strings.HasPrefix(path, ".") {
				return nil
			}
			pathMap[path] = struct{}{}
			return nil
		}); err != nil {
			return err
		}
	}
	for path := range pathMap {
		if err := w.watcher.Add(path); err != nil {
			return err
		}
	}
	return nil
}

func (w *Watcher) Run(ctx context.Context) error {
	eg, ctx := errgroup.WithContext(ctx)
	eg.Go(func() error {
		w.sendEventLoop()
		return nil
	})
	eg.Go(func() error {
		w.receiveEventLoop(ctx)
		return nil
	})
	if err := eg.Wait(); err != nil {
		return err
	}
	return nil
}

func (w *Watcher) sendEventLoop() {
	for {
		select {
		case event, ok := <-w.watcher.Events:
			if !ok {
				continue
			}
			w.handleEvent(event)
		case err, ok := <-w.watcher.Errors:
			if !ok {
				continue
			}
			log.Println("error:", err)
		}
	}
}

func (w *Watcher) handleEvent(event fsnotify.Event) {
	if w.IsWorking() {
		return
	}
	if !w.isModified(event) {
		return
	}
	if w.ignoreFilePath(event.Name) {
		return
	}
	w.eventCh <- event
}

func (w *Watcher) isModified(ev fsnotify.Event) bool {
	return ev.Has(fsnotify.Write) ||
		ev.Has(fsnotify.Create) ||
		ev.Has(fsnotify.Remove) ||
		ev.Has(fsnotify.Rename)
}

func (w *Watcher) ignoreFilePath(path string) bool {
	if strings.HasPrefix(path, "#") {
		return true
	}
	if strings.HasPrefix(path, ".") {
		return true
	}
	if filepath.Ext(path) != ".proto" {
		return true
	}
	return false
}

func (w *Watcher) receiveEventLoop(ctx context.Context) {
	for {
		event := <-w.eventCh
		w.setWorking(true)
		w.handler(ctx, event)
		w.setWorking(false)
	}
}

func (w *Watcher) IsWorking() bool {
	w.isWorkingMu.RLock()
	defer w.isWorkingMu.RUnlock()
	return w.isWorking
}

func (w *Watcher) setWorking(working bool) {
	w.isWorkingMu.Lock()
	defer w.isWorkingMu.Unlock()
	w.isWorking = working
}
