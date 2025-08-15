package watcher

import (
	"fmt"
	"gophercheck/internal/config"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
)

type FileWatcher struct {
	watcher     *fsnotify.Watcher
	config      *config.Config
	watchedDirs map[string]bool
	debouncer   *debouncer
}

type FileChangeEvent struct {
	Path      string
	Operation string
	Timestamp time.Time
}

type FileChangeHandler func([]string) error

func NewFileWatcher(cfg *config.Config) (*FileWatcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("failed to create file watcher: %w", err)
	}
	fw := &FileWatcher{
		watcher:     watcher,
		config:      cfg,
		watchedDirs: make(map[string]bool),
		debouncer:   newDebouncer(500 * time.Millisecond), // 500ms debounce
	}
	return fw, nil
}

func (fw *FileWatcher) Watch(paths []string, handler FileChangeHandler) error {
	for _, path := range paths {
		if err := fw.addPath(path); err != nil {
			return fmt.Errorf("failed to watch path %s: %w", path, err)
		}
	}
	go fw.eventLoop(handler)
	return nil
}

func (fw *FileWatcher) addPath(path string) error {
	return filepath.Walk(path, func(walkPath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			return nil
		}
		if fw.shouldSkipDir(walkPath) {
			return filepath.SkipDir
		}
		if !fw.watchedDirs[walkPath] {
			if err := fw.watcher.Add(walkPath); err != nil {
				return fmt.Errorf("failed to add directory %s to watcher: %w", walkPath, err)
			}
			fw.watchedDirs[walkPath] = true
		}
		return nil
	})
}

func (fw *FileWatcher) eventLoop(handler FileChangeHandler) {
	for {
		select {
		case event, ok := <-fw.watcher.Events:
			if !ok {
				return
			}
			fw.handleEvent(event, handler)
		case err, ok := <-fw.watcher.Errors:
			if !ok {
				return
			}
			fmt.Printf("File watcher error: %v\n", err)
		}
	}
}

func (fw *FileWatcher) handleEvent(event fsnotify.Event, handler FileChangeHandler) {
	if !fw.isGoFile(event.Name) {
		return
	}
	if fw.shouldSkipFile(event.Name) {
		return
	}
	changeEvent := FileChangeEvent{
		Path:      event.Name,
		Operation: fw.eventOpToString(event.Op),
		Timestamp: time.Now(),
	}
	fw.debouncer.add(changeEvent, handler)
}

func (fw *FileWatcher) isGoFile(path string) bool {
	if !strings.HasSuffix(path, ".go") {
		return false
	}
	if strings.HasSuffix(path, "_test.go") {
		return fw.config != nil && fw.config.Files.IncludeTests
	}
	return true
}

func (fw *FileWatcher) shouldSkipDir(path string) bool {
	defaultExclusions := []string{
		"vendor", ".git", "node_modules", ".vscode", ".idea", "build", "dist", "tmp", "temp",
	}
	dirName := filepath.Base(path)
	for _, excluded := range defaultExclusions {
		if dirName == excluded {
			return true
		}
	}
	if fw.config != nil {
		for _, pattern := range fw.config.Files.Exclude {
			matched, _ := filepath.Match(pattern, path)
			if matched {
				return true
			}
		}
	}
	return false
}

func (fw *FileWatcher) shouldSkipFile(path string) bool {
	filename := filepath.Base(path)
	if strings.HasPrefix(filename, ".") {
		return true
	}
	if strings.HasSuffix(filename, ".tmp") || strings.HasSuffix(filename, "~") {
		return true
	}
	if strings.HasSuffix(filename, ".swp") || strings.HasSuffix(filename, ".swo") {
		return true
	}
	return false
}

func (fw *FileWatcher) eventOpToString(op fsnotify.Op) string {
	switch {
	case op&fsnotify.Create == fsnotify.Create:
		return "CREATE"
	case op&fsnotify.Write == fsnotify.Write:
		return "WRITE"
	case op&fsnotify.Remove == fsnotify.Remove:
		return "REMOVE"
	case op&fsnotify.Rename == fsnotify.Rename:
		return "RENAME"
	case op&fsnotify.Chmod == fsnotify.Chmod:
		return "CHMOD"
	default:
		return "UNKNOWN"
	}
}

func (fw *FileWatcher) Close() error {
	fw.debouncer.stop()
	return fw.watcher.Close()
}

func (fw *FileWatcher) GetWatchedPaths() []string {
	paths := make([]string, 0, len(fw.watchedDirs))
	for path := range fw.watchedDirs {
		paths = append(paths, path)
	}
	return paths
}
