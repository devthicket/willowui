//go:build !js

package core

import "github.com/atotto/clipboard"

// ClipboardRead reads the current clipboard contents.
func ClipboardRead() (string, error) {
	return clipboard.ReadAll()
}

// ClipboardWrite writes s to the clipboard.
func ClipboardWrite(s string) {
	clipboard.WriteAll(s)
}
