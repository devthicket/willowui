package widget

import (
	"github.com/devthicket/willowui/internal/core"
	"github.com/devthicket/willowui/internal/render"
	"github.com/devthicket/willowui/internal/theme"
)

// ---------------------------------------------------------------------------
// Variant (re-exported from internal/theme)
// ---------------------------------------------------------------------------

// Variant selects a color group for a component (e.g. Primary, Danger).
type Variant = theme.Variant

const (
	Primary   = theme.Primary
	Secondary = theme.Secondary
	Accent    = theme.Accent
	Neutral   = theme.Neutral
	Danger    = theme.Danger
	Success   = theme.Success
	Warning   = theme.Warning
	Info      = theme.Info
	Custom1   = theme.Custom1
	Custom2   = theme.Custom2
	Custom3   = theme.Custom3
	Custom4   = theme.Custom4
	Custom5   = theme.Custom5
	Custom6   = theme.Custom6
	Custom7   = theme.Custom7
	Custom8   = theme.Custom8
	Custom9   = theme.Custom9
	Custom10  = theme.Custom10
	Custom11  = theme.Custom11
	Custom12  = theme.Custom12
	Custom13  = theme.Custom13
	Custom14  = theme.Custom14
	Custom15  = theme.Custom15
	Custom16  = theme.Custom16
	Custom17  = theme.Custom17
	Custom18  = theme.Custom18
	Custom19  = theme.Custom19
	Custom20  = theme.Custom20
	Custom21  = theme.Custom21
	Custom22  = theme.Custom22
	Custom23  = theme.Custom23
	Custom24  = theme.Custom24
	Custom25  = theme.Custom25
	Custom26  = theme.Custom26
	Custom27  = theme.Custom27
	Custom28  = theme.Custom28
	Custom29  = theme.Custom29
	Custom30  = theme.Custom30
	Custom31  = theme.Custom31
	Custom32  = theme.Custom32
	Custom33  = theme.Custom33
	Custom34  = theme.Custom34
	Custom35  = theme.Custom35
	Custom36  = theme.Custom36
	Custom37  = theme.Custom37
	Custom38  = theme.Custom38
	Custom39  = theme.Custom39
	Custom40  = theme.Custom40
	Custom41  = theme.Custom41
	Custom42  = theme.Custom42
	Custom43  = theme.Custom43
	Custom44  = theme.Custom44
	Custom45  = theme.Custom45
	Custom46  = theme.Custom46
	Custom47  = theme.Custom47
	Custom48  = theme.Custom48
	Custom49  = theme.Custom49
	Custom50  = theme.Custom50
	Custom51  = theme.Custom51
	Custom52  = theme.Custom52
	Custom53  = theme.Custom53
	Custom54  = theme.Custom54
	Custom55  = theme.Custom55
	Custom56  = theme.Custom56

	variantCount = theme.VariantCount
)

// ---------------------------------------------------------------------------
// ComponentState (re-exported from internal/core)
// ---------------------------------------------------------------------------

// ComponentState represents the visual state of a component.
type ComponentState = core.ComponentState

const (
	StateDefault       = core.StateDefault
	StateHover         = core.StateHover
	StateActive        = core.StateActive
	StateDisabled      = core.StateDisabled
	StateFocus         = core.StateFocus
	StateFocusHover    = core.StateFocusHover
	StateFocusActive   = core.StateFocusActive
	StateFocusDisabled = core.StateFocusDisabled
	stateCount         = core.StateCount
)

// stateFallbacks is the fallback chain for each state (delegated to core).
var stateFallbacks = core.StateFallbacks

// computeState maps component boolean flags to a ComponentState.
func computeState(enabled, focused, hovered, active bool) ComponentState {
	return core.ComputeState(enabled, focused, hovered, active)
}

// ---------------------------------------------------------------------------
// Background (re-exported from internal/core)
// ---------------------------------------------------------------------------

// BackgroundType tags the kind of background rendering.
type BackgroundType = core.BackgroundType

const (
	BgNone      = core.BgNone
	BgSolid     = core.BgSolid
	BgNineSlice = core.BgNineSlice
	BgGradient  = core.BgGradient
)

// Rect describes a rectangle with position and dimensions.
type Rect = render.Rect

// NineSlice describes a nine-slice image for use as a component background.
type NineSlice = render.NineSlice

// GradientColors defines per-corner colors for gradient backgrounds.
type GradientColors = render.GradientColors

// Background describes how to render a component's background.
type Background = core.Background

// SolidBackground creates a solid-color background.
var SolidBackground = core.SolidBackground

// SliceBackground creates a nine-slice background.
var SliceBackground = core.SliceBackground

// GradientBackground creates a gradient background.
var GradientBackground = core.GradientBackground

// ---------------------------------------------------------------------------
// Property types (re-exported from internal/theme)
// ---------------------------------------------------------------------------

// ColorProperty holds a color value for each component state.
type ColorProperty = theme.ColorProperty

// BackgroundProperty holds a background value for each component state.
type BackgroundProperty = theme.BackgroundProperty

// FloatProperty holds a float64 value for each component state.
type FloatProperty = theme.FloatProperty

// NewColorPropStates creates a ColorProperty from per-state colors.
var NewColorPropStates = theme.NewColorPropStates


// ---------------------------------------------------------------------------
// Generic Config type (re-exported from internal/theme)
// ---------------------------------------------------------------------------

// Config is a generic component config with variant group support.
type Config[G any] = theme.Config[G]

// ---------------------------------------------------------------------------
// Per-component Group types used within this package
// ---------------------------------------------------------------------------

type TextInputGroup = theme.TextInputGroup
type InputFieldGroup = theme.InputFieldGroup
type SearchBoxGroup = theme.SearchBoxGroup
type PanelGroup = theme.PanelGroup
type SortableListGroup = theme.SortableListGroup
type SortableTreeListGroup = theme.SortableTreeListGroup
type MenuBarGroup = theme.MenuBarGroup
type MenuPopupGroup = theme.MenuPopupGroup

// ---------------------------------------------------------------------------
// Theme (re-exported from internal/theme)
// ---------------------------------------------------------------------------

// Theme holds the complete visual configuration for all WillowUI components.
type Theme = theme.Theme

// DefaultTheme is the fallback theme used when no explicit theme is set.
// This variable is kept for backward compatibility, but EffectiveTheme()
// reads *theme.DefaultThemeRef so that the root package can redirect the
// canonical default by setting theme.DefaultThemeRef = &willowui.DefaultTheme.
var DefaultTheme = theme.DefaultTheme

// getDefaultTheme returns the current canonical default theme by dereferencing
// DefaultThemeRef. When the root package redirects DefaultThemeRef to its own
// DefaultTheme variable, changes to willowui.DefaultTheme are visible here.
func getDefaultTheme() *Theme {
	return *theme.DefaultThemeRef
}
