package core

import (
	"github.com/devthicket/willowui/internal/render"
	"github.com/devthicket/willowui/internal/sg"
)

// ---------------------------------------------------------------------------
// ComponentState
// ---------------------------------------------------------------------------

// ComponentState represents the visual state of a component.
type ComponentState uint8

const (
	StateDefault ComponentState = iota
	StateHover
	StateActive
	StateDisabled
	StateFocus
	StateFocusHover
	StateFocusActive
	StateFocusDisabled
	StateCount // exported sentinel for array sizing
)

// StateFallbacks defines the fallback chain for each state. Each entry is
// tried left-to-right; the first non-zero value is used. StateDefault has
// no fallbacks (it is the terminal).
var StateFallbacks = [StateCount][]ComponentState{
	StateDefault:       {},
	StateHover:         {StateDefault},
	StateActive:        {StateHover, StateDefault},
	StateDisabled:      {StateDefault},
	StateFocus:         {StateActive, StateHover, StateDefault},
	StateFocusHover:    {StateHover, StateFocus, StateDefault},
	StateFocusActive:   {StateFocus, StateActive, StateDefault},
	StateFocusDisabled: {StateDisabled, StateDefault},
}

// ComputeState maps component boolean flags to a ComponentState.
func ComputeState(enabled, focused, hovered, active bool) ComponentState {
	switch {
	case !enabled && focused:
		return StateFocusDisabled
	case !enabled:
		return StateDisabled
	case active && focused:
		return StateFocusActive
	case active:
		return StateActive
	case hovered && focused:
		return StateFocusHover
	case hovered:
		return StateHover
	case focused:
		return StateFocus
	default:
		return StateDefault
	}
}

// ---------------------------------------------------------------------------
// Background
// ---------------------------------------------------------------------------

// BackgroundType tags the kind of background rendering.
type BackgroundType uint8

const (
	BgNone BackgroundType = iota
	BgSolid
	BgNineSlice
	BgGradient
)

// Background describes how to render a component's background.
type Background struct {
	Type     BackgroundType
	Color    sg.Color
	Slice    *render.NineSlice
	Gradient *render.GradientColors
}

// SolidBackground creates a solid-color background.
func SolidBackground(c sg.Color) Background {
	return Background{Type: BgSolid, Color: c}
}

// SliceBackground creates a nine-slice background.
func SliceBackground(s *render.NineSlice) Background {
	return Background{Type: BgNineSlice, Slice: s}
}

// GradientBackground creates a gradient background.
func GradientBackground(g *render.GradientColors) Background {
	return Background{Type: BgGradient, Gradient: g}
}

// ---------------------------------------------------------------------------
// Insets / padding utilities
// ---------------------------------------------------------------------------

// AutoPadding is the sentinel Insets value meaning "use the component's
// built-in default padding". All four fields are -1.
var AutoPadding = render.Insets{Top: -1, Right: -1, Bottom: -1, Left: -1}

// ResolveAutoInsets returns fallback if i is auto, otherwise returns i unchanged.
func ResolveAutoInsets(i, fallback render.Insets) render.Insets {
	if i.IsAuto() {
		return fallback
	}
	return i
}

// Default auto-padding values per component type.
var (
	DefaultButtonPadding    = render.Insets{Top: 8, Right: 16, Bottom: 8, Left: 16}
	DefaultTextInputPadding = render.Insets{Top: 8, Right: 8, Bottom: 8, Left: 8}
	DefaultTextAreaPadding  = render.Insets{Top: 8, Right: 8, Bottom: 8, Left: 8}
	DefaultBarPadding       = render.Insets{Top: 4, Right: 4, Bottom: 4, Left: 4}
)

// ClampDim clamps v between min and max. A zero max means no upper bound.
func ClampDim(v, min, max float64) float64 {
	if v < min {
		v = min
	}
	if max > 0 && v > max {
		v = max
	}
	return v
}
