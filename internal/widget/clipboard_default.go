//go:build !js

package widget

import "github.com/devthicket/willowui/internal/core"

func clipboardRead() (string, error) {
	return core.ClipboardRead()
}

func clipboardWrite(s string) {
	core.ClipboardWrite(s)
}
