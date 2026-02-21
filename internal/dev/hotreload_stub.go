//go:build !hotreload

package dev

import (
	"github.com/devthicket/willowui/internal/template"
	"github.com/devthicket/willowui/internal/widget"
)

// HotReloader is a no-op stub when the hotreload build tag is absent.
type HotReloader struct{}

// NewHotReloader returns an error-free no-op when the hotreload build tag is absent.
func NewHotReloader(_ *template.TemplateRegistry, _ *widget.Screen, _ widget.Controller, _, _ string) (*HotReloader, error) {
	return &HotReloader{}, nil
}

// NewHotReloaderDirect returns a no-op stub.
func NewHotReloaderDirect(_ *template.TemplateRegistry, _ *widget.Screen, _ widget.Controller, _, _ string) *HotReloader {
	return &HotReloader{}
}

// WatchTheme is a no-op when the hotreload build tag is absent.
func (hr *HotReloader) WatchTheme(_ string) error { return nil }

// Stop is a no-op when the hotreload build tag is absent.
func (hr *HotReloader) Stop() {}

// Reload is a no-op when the hotreload build tag is absent.
func (hr *HotReloader) Reload() error { return nil }

// ReloadTheme is a no-op when the hotreload build tag is absent.
func (hr *HotReloader) ReloadTheme() error { return nil }
