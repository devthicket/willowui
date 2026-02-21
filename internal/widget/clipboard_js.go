//go:build js

package widget

// Clipboard is not available in WASM — provide no-op stubs so the
// library compiles. Copy/paste shortcuts silently do nothing.

import "github.com/devthicket/willowui/internal/core"

func clipboardRead() (string, error) {
	return core.ClipboardRead()
}

func clipboardWrite(s string) {
	core.ClipboardWrite(s)
}
