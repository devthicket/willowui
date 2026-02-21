//go:build js

package core

// Clipboard is not available in WASM — provide no-op stubs so the
// library compiles. Copy/paste shortcuts silently do nothing.

// ClipboardRead returns an empty string in WASM environments.
func ClipboardRead() (string, error) {
	return "", nil
}

// ClipboardWrite is a no-op in WASM environments.
func ClipboardWrite(_ string) {}
