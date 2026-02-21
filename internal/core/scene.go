package core

import (
	"github.com/devthicket/willowui/internal/engine"
	"github.com/devthicket/willowui/internal/sg"
)

// ActiveScene holds the current scene for input injection support.
// Set via SetScene; used by TextInput/TextArea to read injected chars/keys.
var ActiveScene *sg.Scene

// FallbackScene keeps the last non-nil scene registered with SetScene.
// It lets injected-key lookups continue to work when ActiveScene is nil.
var FallbackScene *sg.Scene

// SetScene registers the active scene so UI components can read injected
// keyboard input from test runners. The most recently registered non-nil scene
// is retained as a fallback when active scene is nil.
func SetScene(s *sg.Scene) {
	ActiveScene = s
	if s != nil {
		FallbackScene = s
	}
}

// IsKeyJustPressed returns true if the key was just pressed via real input
// or was present in the injected keys queue.
func IsKeyJustPressed(key engine.Key) bool {
	if engine.IsKeyJustPressed(key) {
		return true
	}
	if ActiveScene != nil {
		return ActiveScene.IsInjectedKeyPressed(key)
	}
	if FallbackScene != nil {
		return FallbackScene.IsInjectedKeyPressed(key)
	}
	return false
}
