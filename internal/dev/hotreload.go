//go:build hotreload

package dev

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/devthicket/willowui/internal/template"
	"github.com/devthicket/willowui/internal/theme"
	"github.com/devthicket/willowui/internal/widget"
)

// HotReloader watches XML template and JSON theme files, live-reloading
// when they change. Only available with the "hotreload" build tag.
type HotReloader struct {
	registry     *template.TemplateRegistry
	screen       *widget.Screen
	controller   widget.Controller
	templateName string
	xmlPath      string
	watcher      *fsnotify.Watcher
	stopCh       chan struct{}
	mu           sync.Mutex
	lastContent  []byte

	// Theme reload state.
	themePath        string
	lastThemeContent []byte
	watchedDirs      map[string]bool
}

// NewHotReloader creates a hot reloader that watches xmlPath for changes and
// recompiles the template, swapping the live component tree on the screen.
func NewHotReloader(reg *template.TemplateRegistry, screen *widget.Screen, ctrl widget.Controller, templateName, xmlPath string) (*HotReloader, error) {
	absPath, err := filepath.Abs(xmlPath)
	if err != nil {
		return nil, fmt.Errorf("hotreload: abs path: %w", err)
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("hotreload: create watcher: %w", err)
	}
	// Watch the directory, not the file. Editors that do atomic saves
	// (write-to-temp + rename) delete the original inode, which kills a
	// file-level watch. Watching the directory catches Create events for
	// the replacement file.
	dir := filepath.Dir(absPath)
	if err := watcher.Add(dir); err != nil {
		watcher.Close()
		return nil, fmt.Errorf("hotreload: watch %s: %w", dir, err)
	}

	// Compute initial content hash so the first save-without-change is skipped.
	initialData, err := os.ReadFile(absPath)
	if err != nil {
		watcher.Close()
		return nil, fmt.Errorf("hotreload: read %s: %w", absPath, err)
	}

	hr := &HotReloader{
		registry:     reg,
		screen:       screen,
		controller:   ctrl,
		templateName: templateName,
		xmlPath:      absPath,
		watcher:      watcher,
		stopCh:       make(chan struct{}),
		lastContent:  initialData,
		watchedDirs:  map[string]bool{dir: true},
	}

	go hr.watchLoop()
	return hr, nil
}

// NewHotReloaderDirect creates a HotReloader without starting a file watcher.
// Intended for unit tests that call Reload() directly.
func NewHotReloaderDirect(reg *template.TemplateRegistry, screen *widget.Screen, ctrl widget.Controller, templateName, xmlPath string) *HotReloader {
	absPath, _ := filepath.Abs(xmlPath)
	return &HotReloader{
		registry:     reg,
		screen:       screen,
		controller:   ctrl,
		templateName: templateName,
		xmlPath:      absPath,
		stopCh:       make(chan struct{}),
		watchedDirs:  map[string]bool{},
	}
}

// WatchTheme adds a JSON theme file to the watch set. When the file changes,
// the theme is recompiled and applied to all live components on the screen.
func (hr *HotReloader) WatchTheme(jsonPath string) error {
	absPath, err := filepath.Abs(jsonPath)
	if err != nil {
		return fmt.Errorf("hotreload: abs path: %w", err)
	}

	initialData, err := os.ReadFile(absPath)
	if err != nil {
		return fmt.Errorf("hotreload: read %s: %w", absPath, err)
	}

	hr.mu.Lock()
	hr.themePath = absPath
	hr.lastThemeContent = initialData
	hr.mu.Unlock()

	// Watch the theme file's directory if not already watched.
	dir := filepath.Dir(absPath)
	if hr.watcher != nil && !hr.watchedDirs[dir] {
		if err := hr.watcher.Add(dir); err != nil {
			return fmt.Errorf("hotreload: watch %s: %w", dir, err)
		}
		hr.watchedDirs[dir] = true
	}

	return nil
}

// Stop shuts down the file watcher and stops the reload loop.
func (hr *HotReloader) Stop() {
	close(hr.stopCh)
	if hr.watcher != nil {
		hr.watcher.Close()
	}
}

// Reload recompiles the XML file and swaps the component tree. Exported
// for testing; normally called automatically by the file watcher.
func (hr *HotReloader) Reload() error {
	hr.mu.Lock()
	defer hr.mu.Unlock()

	// Retry briefly — editors doing atomic saves (rename-over) may have a
	// tiny window where the target path does not yet exist.
	var data []byte
	var readErr error
	for attempt := 0; attempt < 3; attempt++ {
		data, readErr = os.ReadFile(hr.xmlPath)
		if readErr == nil {
			break
		}
		time.Sleep(50 * time.Millisecond)
	}
	if readErr != nil {
		return fmt.Errorf("hotreload: read %s: %w", hr.xmlPath, readErr)
	}

	// Skip reload if content hasn't actually changed.
	if bytes.Equal(data, hr.lastContent) {
		return nil
	}
	hr.lastContent = data

	if err := hr.registry.RegisterXML(hr.templateName, data); err != nil {
		return fmt.Errorf("hotreload: compile: %w", err)
	}

	hr.screen.ClearTemplateTree()

	comp, err := hr.registry.Instantiate(hr.templateName, hr.controller, hr.screen)
	if err != nil {
		return fmt.Errorf("hotreload: instantiate: %w", err)
	}

	hr.screen.Add(comp)
	return nil
}

// ReloadTheme recompiles the theme JSON file and applies it to all live
// components on the screen. Exported for testing.
func (hr *HotReloader) ReloadTheme() error {
	hr.mu.Lock()
	defer hr.mu.Unlock()

	if hr.themePath == "" {
		return fmt.Errorf("hotreload: no theme file configured")
	}

	var data []byte
	var readErr error
	for attempt := 0; attempt < 3; attempt++ {
		data, readErr = os.ReadFile(hr.themePath)
		if readErr == nil {
			break
		}
		time.Sleep(50 * time.Millisecond)
	}
	if readErr != nil {
		return fmt.Errorf("hotreload: read %s: %w", hr.themePath, readErr)
	}

	if bytes.Equal(data, hr.lastThemeContent) {
		return nil
	}
	hr.lastThemeContent = data

	t, err := theme.LoadThemeFromFile(hr.themePath)
	if err != nil {
		return fmt.Errorf("hotreload: theme compile: %w", err)
	}

	// Update the registry so future instantiations use the new theme.
	hr.registry.SetTheme(t)

	// Apply to all live components on the screen.
	for _, child := range hr.screen.Children() {
		child.SetTheme(t)
	}

	return nil
}

func (hr *HotReloader) watchLoop() {
	// Separate debounce timers so a rapid XML+theme change doesn't drop one.
	var xmlDebounce, themeDebounce *time.Timer
	for {
		select {
		case <-hr.stopCh:
			if xmlDebounce != nil {
				xmlDebounce.Stop()
			}
			if themeDebounce != nil {
				themeDebounce.Stop()
			}
			return
		case event, ok := <-hr.watcher.Events:
			if !ok {
				return
			}
			if !(event.Has(fsnotify.Write) || event.Has(fsnotify.Create) || event.Has(fsnotify.Rename)) {
				continue
			}
			evPath, err := filepath.Abs(event.Name)
			if err != nil {
				continue
			}

			switch evPath {
			case hr.xmlPath:
				if xmlDebounce != nil {
					xmlDebounce.Stop()
				}
				xmlDebounce = time.AfterFunc(200*time.Millisecond, func() {
					if err := hr.Reload(); err != nil {
						fmt.Fprintf(os.Stderr, "hotreload: %v\n", err)
					} else {
						fmt.Fprintf(os.Stderr, "hotreload: reloaded %s\n", hr.xmlPath)
					}
				})
			case hr.themePath:
				if themeDebounce != nil {
					themeDebounce.Stop()
				}
				themeDebounce = time.AfterFunc(200*time.Millisecond, func() {
					if err := hr.ReloadTheme(); err != nil {
						fmt.Fprintf(os.Stderr, "hotreload: %v\n", err)
					} else {
						fmt.Fprintf(os.Stderr, "hotreload: reloaded %s\n", hr.themePath)
					}
				})
			}
		case err, ok := <-hr.watcher.Errors:
			if !ok {
				return
			}
			fmt.Fprintf(os.Stderr, "hotreload: watcher error: %v\n", err)
		}
	}
}
