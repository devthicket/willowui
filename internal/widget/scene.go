package widget

import (
	"github.com/devthicket/willowui/internal/core"
	"github.com/devthicket/willowui/internal/sg"
)

// activeScene is a package-level alias pointing at the core ActiveScene var.
// textinput.go and textarea.go access it directly; they remain in this package.
var activeScene *sg.Scene

// fallbackScene keeps the last non-nil scene seen via SetScene. When activeScene
// is nil, text widgets can still read injected input from this scene.
var fallbackScene *sg.Scene

// SetScene registers the active scene so UI components can read injected
// keyboard input from test runners. The most recently registered non-nil scene
// is retained as a fallback when active scene is nil.
func SetScene(s *sg.Scene) {
	activeScene = s
	if s != nil {
		fallbackScene = s
	}
	core.SetScene(s)
	DefaultTooltipManager.setScene(s)
	DefaultMenuPopupManager.setScene(s)
	DefaultColorPickerManager.setScene(s)
	DefaultToastManager.setScene(s)
	DefaultPopoverManager.setScene(s)
}

// currentScene returns the explicitly active scene when set, otherwise the
// last non-nil scene registered via SetScene.
func currentScene() *sg.Scene {
	if activeScene != nil {
		return activeScene
	}
	return fallbackScene
}
