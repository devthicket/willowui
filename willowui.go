package willowui

import (
	_ "embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"

	"github.com/devthicket/willowui/internal/colorutil"
	"github.com/devthicket/willowui/internal/core"
	"github.com/devthicket/willowui/internal/dev"
	"github.com/devthicket/willowui/internal/markup"
	"github.com/devthicket/willowui/internal/reactive"
	"github.com/devthicket/willowui/internal/render"
	"github.com/devthicket/willowui/internal/sg"
	"github.com/devthicket/willowui/internal/template"
	"github.com/devthicket/willowui/internal/theme"
	"github.com/devthicket/willowui/internal/widget"
)

// =============================================================================
// Reactive
// =============================================================================

// Ref is a reactive reference — re-exported from internal/reactive.
type Ref[T comparable] = reactive.Ref[T]

// Computed is a reactive computed value — re-exported from internal/reactive.
type Computed[T comparable] = reactive.Computed[T]

// WatchHandle is a handle returned by WatchEffect / WatchValue.
type WatchHandle = reactive.WatchHandle

// Scheduler drives reactive flush cycles.
type Scheduler = reactive.Scheduler

// DefaultScheduler is the package-level scheduler. It points into the
// internal reactive package so that Ref.Set() and Computed.markDirty()
// enqueue to the same scheduler that external callers flush.
var DefaultScheduler = &reactive.DefaultScheduler

// NewRef creates a new reactive reference.
func NewRef[T comparable](initial T) *Ref[T] {
	return reactive.NewRef(initial)
}

// NewComputed creates a new computed reactive value.
func NewComputed[T comparable](fn func() T) *Computed[T] {
	return reactive.NewComputed(fn)
}

// WatchEffect creates a reactive effect that re-runs when dependencies change.
func WatchEffect(fn func()) WatchHandle {
	return reactive.WatchEffect(fn)
}

// WatchValue watches a specific Ref and calls fn with old and new values.
func WatchValue[T comparable](ref *Ref[T], fn func(old, new T)) WatchHandle {
	return reactive.WatchValue(ref, fn)
}

// Numeric is a constraint for types that support arithmetic operations.
type Numeric = reactive.Numeric

// Increment returns a func() that adds delta to a numeric ref.
func Increment[T Numeric](r *Ref[T], delta T) func() {
	return reactive.Increment(r, delta)
}

// ToggleRef returns a func() that flips a boolean ref.
func ToggleRef(r *Ref[bool]) func() {
	return reactive.Toggle(r)
}

// Set returns a func() that sets a ref to the given value.
func Set[T comparable](r *Ref[T], val T) func() {
	return reactive.Set(r, val)
}

// BindFormatter returns a *Ref[string] that stays in sync with source,
// converting values via fmt.Sprint. Use with BindText.
//
// The returned WatchHandle must be stopped when the binding is no longer
// needed. Pass it to screen.TrackRef or call h.Stop() explicitly.
func BindFormatter[T comparable](source *Ref[T]) (*Ref[string], WatchHandle) {
	return reactive.BindFormatter(source)
}

// BindFormatterf returns a *Ref[string] that stays in sync with source,
// converting values via fmt.Sprintf. Use with BindText.
//
// The returned WatchHandle must be stopped when the binding is no longer
// needed. Pass it to screen.TrackRef or call h.Stop() explicitly.
func BindFormatterf[T comparable](source *Ref[T], format string) (*Ref[string], WatchHandle) {
	return reactive.BindFormatterf(source, format)
}

// Array[T] is a reactive ordered collection — re-exported from internal/reactive.
type Array[T any] = reactive.Array[T]

// Record is a reactive key-value object — re-exported from internal/reactive.
type Record = reactive.Record

// NewArray creates an empty reactive array.
func NewArray[T any]() *Array[T] {
	return reactive.NewArray[T]()
}

// NewArrayFrom creates a reactive array copied from the given slice.
func NewArrayFrom[T any](items []T) *Array[T] {
	return reactive.NewArrayFrom(items)
}

// NewArrayFromAny creates a reactive Array[any] from a typed slice, boxing each element.
func NewArrayFromAny[T any](items []T) *Array[any] {
	return reactive.NewArrayFromAny(items)
}

// NewArrayWithCap creates an empty reactive array with the given initial capacity.
func NewArrayWithCap[T any](cap int) *Array[T] {
	return reactive.NewArrayWithCap[T](cap)
}

// ArrayMap transforms each element of a using fn and returns a plain []U.
func ArrayMap[T, U any](a *Array[T], fn func(T) U) []U {
	return reactive.ArrayMap(a, fn)
}

// ArrayReduce folds a into a single U value using fn, starting from init.
func ArrayReduce[T, U any](a *Array[T], fn func(U, T) U, init U) U {
	return reactive.ArrayReduce(a, fn, init)
}

// ArraySort sorts a in ascending order. T must satisfy cmp.Ordered.
func ArraySort[T interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr |
		~float32 | ~float64 | ~string
}](a *Array[T]) {
	reactive.ArraySort(a)
}

// ArraySortDesc sorts a in descending order. T must satisfy cmp.Ordered.
func ArraySortDesc[T interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr |
		~float32 | ~float64 | ~string
}](a *Array[T]) {
	reactive.ArraySortDesc(a)
}

// ArraySortFold sorts a case-insensitively using key to extract a string from
// each element. Comparison uses Unicode simple case-folding so "Abc" == "abc" == "ABC".
func ArraySortFold[T any](a *Array[T], key func(T) string) {
	reactive.ArraySortFold(a, key)
}

// IndexOf returns the index of item in a, or -1 if absent.
func IndexOf[T comparable](a *Array[T], item T) int {
	return reactive.IndexOf(a, item)
}

// Includes reports whether item is present in a.
func Includes[T comparable](a *Array[T], item T) bool {
	return reactive.Includes(a, item)
}

// NewRecord creates an empty reactive Record.
func NewRecord() *Record {
	return reactive.NewRecord()
}

// NewRecordFrom creates a reactive Record pre-populated from fields (copied).
func NewRecordFrom(fields map[string]any) *Record {
	return reactive.NewRecordFrom(fields)
}

// =============================================================================
// Markup
// =============================================================================

// TextSpan represents a styled segment of text within a RichText component.
// Fields left at their zero values inherit from the parent RichText.
type TextSpan = markup.TextSpan

// Outline defines a text stroke rendered behind the fill.
type Outline = markup.Outline

var (
	// ParseMarkup parses XML-like markup into TextSpan slices.
	ParseMarkup = markup.ParseMarkup

	// ParseColor parses a color string in any supported format.
	ParseColor = markup.ParseColor
)

// =============================================================================
// Render
// =============================================================================

// Render — test-facing re-exports from internal/widget (backed by internal/render).

// SubRegion extracts a sub-region from a texture atlas region.
var SubRegion = widget.SubRegion

// CreateNineSliceNodes builds 9 sprite nodes for a nine-slice background.
var CreateNineSliceNodes = widget.CreateNineSliceNodes

// LayoutNineSlice positions and scales the 9 sprites of a nine-slice.
var LayoutNineSlice = widget.LayoutNineSlice

// RoundedRectPoints returns the outline points for a rounded rectangle.
var RoundedRectPoints = widget.RoundedRectPoints

// RoundedRectBorderMesh builds a mesh for a rounded rectangle border.
var RoundedRectBorderMesh = widget.RoundedRectBorderMesh

// LerpColor linearly interpolates between two colors.
var LerpColor = widget.LerpColor

// RoundedRectGradientMesh builds a gradient-filled rounded rectangle mesh.
var RoundedRectGradientMesh = widget.RoundedRectGradientMesh

// =============================================================================
// Core types
// =============================================================================

// ComponentState represents the visual state of a component.
type ComponentState = core.ComponentState

const (
	StateDefault       = core.StateDefault       // Normal idle state.
	StateHover         = core.StateHover         // Pointer is over the component.
	StateActive        = core.StateActive        // Component is being pressed.
	StateDisabled      = core.StateDisabled      // Component is non-interactive.
	StateFocus         = core.StateFocus         // Component has keyboard focus.
	StateFocusHover    = core.StateFocusHover    // Focused and hovered.
	StateFocusActive   = core.StateFocusActive   // Focused and pressed.
	StateFocusDisabled = core.StateFocusDisabled // Focused but disabled.
)

// BackgroundType tags the kind of background rendering.
type BackgroundType = core.BackgroundType

const (
	BgNone      = core.BgNone      // No background rendered.
	BgSolid     = core.BgSolid     // Flat solid-color fill.
	BgNineSlice = core.BgNineSlice // Nine-slice image background.
	BgGradient  = core.BgGradient  // Per-corner gradient fill.
)

// Rect describes a rectangle with position and dimensions.
// It is an alias for render.Rect.
type Rect = render.Rect

// NineSlice describes a nine-slice image for use as a component background.
// It is an alias for render.NineSlice.
type NineSlice = render.NineSlice

// GradientColors defines per-corner colors for gradient backgrounds.
// It is an alias for render.GradientColors.
type GradientColors = render.GradientColors

// Background describes how to render a component's background.
type Background = core.Background

// SolidBackground creates a solid-color background.
var SolidBackground = core.SolidBackground

// SliceBackground creates a nine-slice background.
var SliceBackground = core.SliceBackground

// GradientBackground creates a gradient background.
var GradientBackground = core.GradientBackground

// =============================================================================
// Theme
// =============================================================================

// Variant selects a color group for a component (e.g. Primary, Danger).
type Variant = theme.Variant

const (
	Primary   = theme.Primary   // Primary action or branding color.
	Secondary = theme.Secondary // Secondary or supporting color.
	Accent    = theme.Accent    // Accent highlight color.
	Neutral   = theme.Neutral   // Neutral/muted color.
	Danger    = theme.Danger    // Destructive or error actions.
	Success   = theme.Success   // Positive confirmation.
	Warning   = theme.Warning   // Caution or non-blocking alert.
	Info      = theme.Info      // Informational or help context.
	Custom1   = theme.Custom1   // Application-defined variant 1.
	Custom2   = theme.Custom2   // Application-defined variant 2.
	Custom3   = theme.Custom3   // Application-defined variant 3.
	Custom4   = theme.Custom4   // Application-defined variant 4.
	Custom5   = theme.Custom5   // Application-defined variant 5.
	Custom6   = theme.Custom6   // Application-defined variant 6.
	Custom7   = theme.Custom7   // Application-defined variant 7.
	Custom8   = theme.Custom8   // Application-defined variant 8.
	Custom9   = theme.Custom9   // Application-defined variant 9.
	Custom10  = theme.Custom10  // Application-defined variant 10.
	Custom11  = theme.Custom11  // Application-defined variant 11.
	Custom12  = theme.Custom12  // Application-defined variant 12.
	Custom13  = theme.Custom13  // Application-defined variant 13.
	Custom14  = theme.Custom14  // Application-defined variant 14.
	Custom15  = theme.Custom15  // Application-defined variant 15.
	Custom16  = theme.Custom16  // Application-defined variant 16.
	Custom17  = theme.Custom17  // Application-defined variant 17.
	Custom18  = theme.Custom18  // Application-defined variant 18.
	Custom19  = theme.Custom19  // Application-defined variant 19.
	Custom20  = theme.Custom20  // Application-defined variant 20.
	Custom21  = theme.Custom21  // Application-defined variant 21.
	Custom22  = theme.Custom22  // Application-defined variant 22.
	Custom23  = theme.Custom23  // Application-defined variant 23.
	Custom24  = theme.Custom24  // Application-defined variant 24.
	Custom25  = theme.Custom25  // Application-defined variant 25.
	Custom26  = theme.Custom26  // Application-defined variant 26.
	Custom27  = theme.Custom27  // Application-defined variant 27.
	Custom28  = theme.Custom28  // Application-defined variant 28.
	Custom29  = theme.Custom29  // Application-defined variant 29.
	Custom30  = theme.Custom30  // Application-defined variant 30.
	Custom31  = theme.Custom31  // Application-defined variant 31.
	Custom32  = theme.Custom32  // Application-defined variant 32.
	Custom33  = theme.Custom33  // Application-defined variant 33.
	Custom34  = theme.Custom34  // Application-defined variant 34.
	Custom35  = theme.Custom35  // Application-defined variant 35.
	Custom36  = theme.Custom36  // Application-defined variant 36.
	Custom37  = theme.Custom37  // Application-defined variant 37.
	Custom38  = theme.Custom38  // Application-defined variant 38.
	Custom39  = theme.Custom39  // Application-defined variant 39.
	Custom40  = theme.Custom40  // Application-defined variant 40.
	Custom41  = theme.Custom41  // Application-defined variant 41.
	Custom42  = theme.Custom42  // Application-defined variant 42.
	Custom43  = theme.Custom43  // Application-defined variant 43.
	Custom44  = theme.Custom44  // Application-defined variant 44.
	Custom45  = theme.Custom45  // Application-defined variant 45.
	Custom46  = theme.Custom46  // Application-defined variant 46.
	Custom47  = theme.Custom47  // Application-defined variant 47.
	Custom48  = theme.Custom48  // Application-defined variant 48.
	Custom49  = theme.Custom49  // Application-defined variant 49.
	Custom50  = theme.Custom50  // Application-defined variant 50.
	Custom51  = theme.Custom51  // Application-defined variant 51.
	Custom52  = theme.Custom52  // Application-defined variant 52.
	Custom53  = theme.Custom53  // Application-defined variant 53.
	Custom54  = theme.Custom54  // Application-defined variant 54.
	Custom55  = theme.Custom55  // Application-defined variant 55.
	Custom56  = theme.Custom56  // Application-defined variant 56.
)

// ColorProperty holds a color value for each component state.
type ColorProperty = theme.ColorProperty

// BackgroundProperty holds a background value for each component state.
type BackgroundProperty = theme.BackgroundProperty

// FloatProperty holds a float64 value for each component state.
type FloatProperty = theme.FloatProperty

var (
	// NewColorPropUniform creates a ColorProperty with the same color for all states.
	NewColorPropUniform = theme.NewColorPropUniform
	// NewColorPropStates creates a ColorProperty with per-state colors.
	NewColorPropStates = theme.NewColorPropStates
	// NewSolidBgPropStates creates a BackgroundProperty with per-state solid backgrounds.
	NewSolidBgPropStates = theme.NewSolidBgPropStates
	// NewSolidBgPropUniform creates a BackgroundProperty with the same solid background for all states.
	NewSolidBgPropUniform = theme.NewSolidBgPropUniform
	// NewFloatPropUniform creates a FloatProperty with the same value for all states.
	NewFloatPropUniform = theme.NewFloatPropUniform
	// NewFloatPropStates creates a FloatProperty with per-state values.
	NewFloatPropStates = theme.NewFloatPropStates
)

// Config is a generic component config with variant group support.
type Config[G any] = theme.Config[G]

// ShadowConfig describes a drop shadow for tooltip components.
type ShadowConfig = theme.ShadowConfig

// SpriteRef holds a resolved texture region from the theme atlas.
type SpriteRef = theme.SpriteRef

// Per-component Group types. Each Group holds the resolved visual properties
// (colors, sizes, corner radii, etc.) for one variant of a widget. Groups
// are loaded from the theme JSON and looked up at render time via Config[G].

// ButtonGroup holds theme properties for Button variants.
type ButtonGroup = theme.ButtonGroup

// LabelGroup holds theme properties for Label variants.
type LabelGroup = theme.LabelGroup

// BadgeGroup holds theme properties for Badge variants.
type BadgeGroup = theme.BadgeGroup

// ToggleGroup holds theme properties for Toggle variants.
type ToggleGroup = theme.ToggleGroup

// CheckboxGroup holds theme properties for Checkbox variants.
type CheckboxGroup = theme.CheckboxGroup

// RadioGroup holds theme properties for Radio variants.
type RadioGroup = theme.RadioGroup

// TextInputGroup holds theme properties for TextInput variants.
type TextInputGroup = theme.TextInputGroup

// MaskedInputGroup holds theme properties for MaskedInput variants.
type MaskedInputGroup = theme.MaskedInputGroup

// TextAreaGroup holds theme properties for TextArea variants.
type TextAreaGroup = theme.TextAreaGroup

// SliderGroup holds theme properties for Slider variants.
type SliderGroup = theme.SliderGroup

// ScrollBarGroup holds theme properties for ScrollBar variants.
type ScrollBarGroup = theme.ScrollBarGroup

// MeterBarGroup holds theme properties for MeterBar variants.
type MeterBarGroup = theme.MeterBarGroup

// PanelGroup holds theme properties for Panel variants.
type PanelGroup = theme.PanelGroup

// NavDrawerGroup holds theme properties for NavDrawer variants.
type NavDrawerGroup = theme.NavDrawerGroup

// WindowGroup holds theme properties for Window variants.
type WindowGroup = theme.WindowGroup

// TabsGroup holds theme properties for TabBar variants.
type TabsGroup = theme.TabsGroup

// ListGroup holds theme properties for List variants.
type ListGroup = theme.ListGroup

// TreeListGroup holds theme properties for TreeList variants.
type TreeListGroup = theme.TreeListGroup

// TileListGroup holds theme properties for TileList variants.
type TileListGroup = theme.TileListGroup

// RichTextGroup holds theme properties for RichText variants.
type RichTextGroup = theme.RichTextGroup

// OptionRotatorChevronGroup holds theme properties for OptionRotator chevron arrows.
type OptionRotatorChevronGroup = theme.OptionRotatorChevronGroup

// OptionRotatorGroup holds theme properties for OptionRotator variants.
type OptionRotatorGroup = theme.OptionRotatorGroup

// ToggleButtonBarGroup holds theme properties for ToggleButtonBar variants.
type ToggleButtonBarGroup = theme.ToggleButtonBarGroup

// TooltipGroup holds theme properties for Tooltip variants.
type TooltipGroup = theme.TooltipGroup

// MenuBarGroup holds theme properties for MenuBar variants.
type MenuBarGroup = theme.MenuBarGroup

// MenuPopupGroup holds theme properties for MenuPopup variants.
type MenuPopupGroup = theme.MenuPopupGroup

// SelectGroup holds theme properties for Select dropdown variants.
type SelectGroup = theme.SelectGroup

// DragHandleGroup holds theme properties for DragHandle variants.
type DragHandleGroup = theme.DragHandleGroup

// ImageGroup holds theme properties for Image variants.
type ImageGroup = theme.ImageGroup

// AnimatedImageGroup holds theme properties for AnimatedImage variants.
type AnimatedImageGroup = theme.AnimatedImageGroup

// ColorPickerGroup holds theme properties for ColorPicker variants.
type ColorPickerGroup = theme.ColorPickerGroup

// GradientEditorGroup holds theme properties for GradientEditor variants.
type GradientEditorGroup = theme.GradientEditorGroup

// ToastGroup holds theme properties for Toast variants.
type ToastGroup = theme.ToastGroup

// SortableListGroup holds theme properties for SortableList variants.
type SortableListGroup = theme.SortableListGroup

// SortableTreeListGroup holds theme properties for SortableTreeList variants.
type SortableTreeListGroup = theme.SortableTreeListGroup

// IconButtonGroup holds theme properties for IconButton variants.
type IconButtonGroup = theme.IconButtonGroup

// StatWebGroup holds theme properties for StatWeb variants.
type StatWebGroup = theme.StatWebGroup

// AccordionGroup holds theme properties for Accordion variants.
type AccordionGroup = theme.AccordionGroup

// TagGroup holds theme properties for Tag variants.
type TagGroup = theme.TagGroup

// TagBarGroup holds theme properties for TagBar variants.
type TagBarGroup = theme.TagBarGroup

// PopoverGroup holds theme properties for Popover variants.
type PopoverGroup = theme.PopoverGroup

// TreeTableGroup holds theme properties for TreeTable variants.
type TreeTableGroup = theme.TreeTableGroup

// DataTableGroup holds theme properties for DataTable variants.
type DataTableGroup = theme.DataTableGroup

// KeybindInputGroup holds theme properties for KeybindInput variants.
type KeybindInputGroup = theme.KeybindInputGroup

// TimePickerGroup holds theme properties for TimePicker variants.
type TimePickerGroup = theme.TimePickerGroup

// ImageCropperGroup holds theme properties for ImageCropper variants.
type ImageCropperGroup = theme.ImageCropperGroup

// ToolBarGroup holds theme properties for ToolBar variants.
type ToolBarGroup = theme.ToolBarGroup

// CalendarSelectorGroup holds theme properties for CalendarSelector variants.
type CalendarSelectorGroup = theme.CalendarSelectorGroup

// RichTextEditorGroup holds theme properties for RichTextEditor variants.
type RichTextEditorGroup = theme.RichTextEditorGroup

// PropertyInspectorGroup holds theme properties for PropertyInspector variants.
type PropertyInspectorGroup = theme.PropertyInspectorGroup

// UserGroup holds the parsed visual properties for a user-defined component variant.
type UserGroup = theme.UserGroup

// UserConfig holds a user-defined component configuration with variant support.
type UserConfig = theme.UserConfig

// Per-component Config types. Each Config is a Config[G] alias mapping
// variant names to their Group. Use SetTheme to load configs from JSON.

// ButtonConfig maps variant names to ButtonGroup.
type ButtonConfig = theme.ButtonConfig

// LabelConfig maps variant names to LabelGroup.
type LabelConfig = theme.LabelConfig

// BadgeConfig maps variant names to BadgeGroup.
type BadgeConfig = theme.BadgeConfig

// ToggleConfig maps variant names to ToggleGroup.
type ToggleConfig = theme.ToggleConfig

// CheckboxConfig maps variant names to CheckboxGroup.
type CheckboxConfig = theme.CheckboxConfig

// RadioConfig maps variant names to RadioGroup.
type RadioConfig = theme.RadioConfig

// TextInputConfig maps variant names to TextInputGroup.
type TextInputConfig = theme.TextInputConfig

// MaskedInputConfig maps variant names to MaskedInputGroup.
type MaskedInputConfig = theme.MaskedInputConfig

// TextAreaConfig maps variant names to TextAreaGroup.
type TextAreaConfig = theme.TextAreaConfig

// SliderConfig maps variant names to SliderGroup.
type SliderConfig = theme.SliderConfig

// ScrollBarConfig maps variant names to ScrollBarGroup.
type ScrollBarConfig = theme.ScrollBarConfig

// MeterBarConfig maps variant names to MeterBarGroup.
type MeterBarConfig = theme.MeterBarConfig

// PanelConfig maps variant names to PanelGroup.
type PanelConfig = theme.PanelConfig

// NavDrawerConfig maps variant names to NavDrawerGroup.
type NavDrawerConfig = theme.NavDrawerConfig

// WindowConfig maps variant names to WindowGroup.
type WindowConfig = theme.WindowConfig

// TabsConfig maps variant names to TabsGroup.
type TabsConfig = theme.TabsConfig

// ListConfig maps variant names to ListGroup.
type ListConfig = theme.ListConfig

// TreeListConfig maps variant names to TreeListGroup.
type TreeListConfig = theme.TreeListConfig

// TileListConfig maps variant names to TileListGroup.
type TileListConfig = theme.TileListConfig

// RichTextConfig maps variant names to RichTextGroup.
type RichTextConfig = theme.RichTextConfig

// OptionRotatorConfig maps variant names to OptionRotatorGroup.
type OptionRotatorConfig = theme.OptionRotatorConfig

// ToggleButtonBarConfig maps variant names to ToggleButtonBarGroup.
type ToggleButtonBarConfig = theme.ToggleButtonBarConfig

// TooltipConfig maps variant names to TooltipGroup.
type TooltipConfig = theme.TooltipConfig

// MenuBarConfig maps variant names to MenuBarGroup.
type MenuBarConfig = theme.MenuBarConfig

// MenuPopupConfig maps variant names to MenuPopupGroup.
type MenuPopupConfig = theme.MenuPopupConfig

// SelectConfig maps variant names to SelectGroup.
type SelectConfig = theme.SelectConfig

// DragHandleConfig maps variant names to DragHandleGroup.
type DragHandleConfig = theme.DragHandleConfig

// ImageConfig maps variant names to ImageGroup.
type ImageConfig = theme.ImageConfig

// AnimatedImageConfig maps variant names to AnimatedImageGroup.
type AnimatedImageConfig = theme.AnimatedImageConfig

// ColorPickerConfig maps variant names to ColorPickerGroup.
type ColorPickerConfig = theme.ColorPickerConfig

// GradientEditorConfig maps variant names to GradientEditorGroup.
type GradientEditorConfig = theme.GradientEditorConfig

// ToastConfig maps variant names to ToastGroup.
type ToastConfig = theme.ToastConfig

// SortableListConfig maps variant names to SortableListGroup.
type SortableListConfig = theme.SortableListConfig

// SortableTreeListConfig maps variant names to SortableTreeListGroup.
type SortableTreeListConfig = theme.SortableTreeListConfig

// IconButtonConfig maps variant names to IconButtonGroup.
type IconButtonConfig = theme.IconButtonConfig

// StatWebConfig maps variant names to StatWebGroup.
type StatWebConfig = theme.StatWebConfig

// AccordionConfig maps variant names to AccordionGroup.
type AccordionConfig = theme.AccordionConfig

// TagConfig maps variant names to TagGroup.
type TagConfig = theme.TagConfig

// TagBarConfig maps variant names to TagBarGroup.
type TagBarConfig = theme.TagBarConfig

// PopoverConfig maps variant names to PopoverGroup.
type PopoverConfig = theme.PopoverConfig

// TreeTableConfig maps variant names to TreeTableGroup.
type TreeTableConfig = theme.TreeTableConfig

// DataTableConfig maps variant names to DataTableGroup.
type DataTableConfig = theme.DataTableConfig

// KeybindInputConfig maps variant names to KeybindInputGroup.
type KeybindInputConfig = theme.KeybindInputConfig

// TimePickerConfig maps variant names to TimePickerGroup.
type TimePickerConfig = theme.TimePickerConfig

// ImageCropperConfig maps variant names to ImageCropperGroup.
type ImageCropperConfig = theme.ImageCropperConfig

// ToolBarConfig maps variant names to ToolBarGroup.
type ToolBarConfig = theme.ToolBarConfig

// CalendarSelectorConfig maps variant names to CalendarSelectorGroup.
type CalendarSelectorConfig = theme.CalendarSelectorConfig

// RichTextEditorConfig maps variant names to RichTextEditorGroup.
type RichTextEditorConfig = theme.RichTextEditorConfig

// PropertyInspectorConfig maps variant names to PropertyInspectorGroup.
type PropertyInspectorConfig = theme.PropertyInspectorConfig

// HeadingLevel represents a heading size level for rich text content.
type HeadingLevel = int

const (
	HeadingNone HeadingLevel = 0 // No heading style.
	Heading1    HeadingLevel = 1 // Largest heading.
	Heading2    HeadingLevel = 2 // Medium heading.
	Heading3    HeadingLevel = 3 // Smallest heading.
)

// Theme holds the complete visual configuration for all WillowUI components.
type Theme = theme.Theme

// DefaultTheme is the fallback theme used when no explicit theme is set.
// This is the canonical variable; theme.DefaultThemeRef is redirected here
// at init time so that widget.EffectiveTheme() always reads from this variable.
var DefaultTheme = theme.DefaultTheme

func init() {
	// Redirect the internal/theme indirection to point at this package's
	// DefaultTheme variable. This ensures that widget.EffectiveTheme()
	// returns the current value of willowui.DefaultTheme whenever it is
	// reassigned (e.g. in tests or user code).
	theme.DefaultThemeRef = &DefaultTheme

	// Inject the embedded glyph spritesheet into the widget package so
	// default icons (chevrons, close X, sort arrows, etc.) are decoded
	// from the pre-baked PNG rather than generated procedurally.
	widget.SetGlyphSheet(embeddedGlyphSheet)
}

// LoadTheme parses JSON theme data and produces a *Theme.
// Returns an error if validation fails (bad colors, missing required groups, etc.).
// Nine-slice images are rejected — use LoadThemeFromFile or LoadThemeFromFS.
var LoadTheme = theme.LoadTheme

// LoadThemeRelative loads a theme JSON file resolved relative to the caller's
// source file. This is convenient for examples and tests where the JSON file
// sits next to the Go source.
func LoadThemeRelative(filename string) (*Theme, error) {
	_, src, _, ok := runtime.Caller(1)
	if !ok {
		return nil, fmt.Errorf("LoadThemeRelative: unable to determine caller path")
	}
	return theme.LoadThemeFromFile(filepath.Join(filepath.Dir(src), filename))
}

// LoadThemeFromFile reads a JSON file and compiles the theme.
// Nine-slice image paths are resolved relative to the JSON file's directory.
var LoadThemeFromFile = theme.LoadThemeFromFile

// LoadThemeFromFS reads a JSON file from an fs.FS and compiles the theme.
// Nine-slice image paths are resolved within the FS.
var LoadThemeFromFS = func(fsys fs.FS, path string) (*Theme, error) {
	return theme.LoadThemeFromFS(fsys, path)
}

// ValidateTheme checks that the given theme defines configs for all the
// named component types. Returns an error listing any missing configs.
var ValidateTheme = theme.ValidateTheme

// CollectThemeImages extracts all nine-slice image paths from theme JSON
// without loading them. Use this for prebaked atlas tooling.
var CollectThemeImages = theme.CollectThemeImages

// LoadThemeBinary decodes a WUIT binary (.theme file) and compiles the theme.
// The atlas (if present) is decoded from the embedded PNG + JSON sections.
var LoadThemeBinary = theme.LoadThemeBinary

// EncodeThemeBinary encodes theme JSON, atlas JSON, and atlas PNG into
// the WUIT binary format. Use the themec CLI tool for full compilation.
var EncodeThemeBinary = theme.EncodeThemeBinary

// =============================================================================
// Component & Layout
// =============================================================================

// Insets holds top/right/bottom/left spacing values used for padding and margin.
type Insets = widget.Insets

// UIElement is implemented by all UI component types in this package.
type UIElement = widget.UIElement

// Component is the base type for all WillowUI widgets.
type Component = widget.Component

// NewComponent creates a new Component with sensible defaults.
var NewComponent = widget.NewComponent

// NewHBox creates a Component with LayoutHBox pre-configured.
var NewHBox = widget.NewHBox

// NewVBox creates a Component with LayoutVBox pre-configured.
var NewVBox = widget.NewVBox

// NewFlow creates a Component with LayoutFlow pre-configured.
var NewFlow = widget.NewFlow

// LayoutMode controls how a Component arranges its children.
type LayoutMode = widget.LayoutMode

const (
	LayoutNone   = widget.LayoutNone   // No automatic layout; children are manually positioned.
	LayoutVBox   = widget.LayoutVBox   // Vertical stack layout.
	LayoutHBox   = widget.LayoutHBox   // Horizontal row layout.
	LayoutGrid   = widget.LayoutGrid   // Grid layout with rows and columns.
	LayoutFlow   = widget.LayoutFlow   // Flowing wrap layout.
	LayoutAnchor = widget.LayoutAnchor // Anchor-based absolute positioning within parent.
)

// Alignment controls child positioning.
type Alignment = widget.Alignment

const (
	AlignStart        = widget.AlignStart        // Align children to the start (left or top).
	AlignCenter       = widget.AlignCenter       // Center children along the axis.
	AlignEnd          = widget.AlignEnd          // Align children to the end (right or bottom).
	AlignSpaceBetween = widget.AlignSpaceBetween // Distribute children with equal space between.
)

// FillMode controls how a component stretches to fill its parent's content area.
type FillMode = widget.FillMode

const (
	FillNone   = widget.FillNone   // No stretching.
	FillWidth  = widget.FillWidth  // Stretch to parent's content width.
	FillHeight = widget.FillHeight // Stretch to parent's content height.
	FillBoth   = widget.FillBoth   // Stretch to fill both dimensions.
)

// Orientation represents horizontal or vertical direction.
type Orientation = widget.Orientation

const (
	Horizontal = widget.Horizontal // Left-to-right orientation.
	Vertical   = widget.Vertical   // Top-to-bottom orientation.
)

// =============================================================================
// Focus & Input
// =============================================================================

// InputManager reads keyboard state once per frame, tracks consumed keys,
// and exposes availability queries and event-style listeners for game logic.
type InputManager = widget.InputManager

// Input is the package-level InputManager singleton. Game logic reads key
// state through this instead of ebiten directly.
var Input = widget.DefaultInputManager

// NewInputManager creates an isolated InputManager (primarily for tests).
var NewInputManager = widget.NewInputManager

// ListenerHandle identifies a registered key listener for later removal.
type ListenerHandle = widget.ListenerHandle

// FocusManager tracks which component has keyboard focus.
type FocusManager = widget.FocusManager

// FM is the package-level FocusManager singleton. UI widgets and the
// screen system use this for focus dispatch, hotkeys, and navigation.
var FM = widget.DefaultFocusManager

// DefaultFocusManager is an alias for FM (backwards compatibility).
var DefaultFocusManager = widget.DefaultFocusManager

// NewFocusManager creates an empty focus manager.
var NewFocusManager = widget.NewFocusManager

// ModifierMask is a bitmask of modifier keys for keybind registration.
type ModifierMask = widget.ModifierMask

const (
	ModNone  = widget.ModNone  // No modifier keys.
	ModCtrl  = widget.ModCtrl  // Ctrl (or Cmd on macOS).
	ModShift = widget.ModShift // Shift key.
	ModAlt   = widget.ModAlt   // Alt (or Option on macOS).
)

// KeyCombo pairs an ebiten key with a modifier mask.
type KeyCombo = widget.KeyCombo

// Key creates a KeyCombo from a key and modifier mask.
var Key = widget.Key

// BindHandle identifies a registered keybind for later removal.
type BindHandle = widget.BindHandle

// =============================================================================
// Stage API
// =============================================================================

// FXAAConfig holds tunable parameters for the FXAA post-process pass.
// Use DefaultFXAAConfig for sensible defaults.
type FXAAConfig = sg.FXAAConfig

// DefaultFXAAConfig returns an FXAAConfig with FXAA 3.11 quality-15 defaults.
var DefaultFXAAConfig = sg.DefaultFXAAConfig

//go:embed assets/gofont.fontbundle
var embeddedGoFont []byte

//go:embed assets/icons/default-glyphs.png
var embeddedGlyphSheet []byte

// DefaultFont is the default FontFamily. If not set by StageConfig.Font or
// directly, Setup auto-loads the embedded Go font bundle.
var DefaultFont *sg.FontFamily

// EmbeddedGoFont is the raw bytes of the embedded Go font bundle
// (Regular + Bold + Italic + BoldItalic). Pass to
// willow.NewFontFamilyFromFontBundle to create a FontFamily manually,
// or use MustLoadDefaultFont for convenience.
var EmbeddedGoFont = embeddedGoFont

// MustLoadDefaultFont loads the embedded Go font bundle into DefaultFont
// and returns it. If DefaultFont is already loaded, it returns the existing
// value. This is intended for use in examples that need a font before
// ui.Setup runs. Panics on load failure.
func MustLoadDefaultFont() *sg.FontFamily {
	if DefaultFont != nil {
		return DefaultFont
	}
	ff, err := sg.NewFontFamilyFromFontBundle(embeddedGoFont)
	if err != nil {
		panic("willowui: failed to load embedded font: " + err.Error())
	}
	DefaultFont = ff
	return DefaultFont
}

// DefaultSharpness is the recommended SDF sharpness value for labels.
const DefaultSharpness = 0.15

// FontSource is a type alias for *sg.FontFamily, kept for backward
// compatibility. New code should use *sg.FontFamily directly.
type FontSource = *sg.FontFamily

// RecommendedFXAAConfig returns the moderate FXAA config used by default in
// Setup — SubpixQuality=0.5, EdgeThreshold=0.275, EdgeThresholdMin=0.0505.
func RecommendedFXAAConfig() FXAAConfig {
	return FXAAConfig{
		SubpixQuality:    0.5,
		EdgeThreshold:    0.275,
		EdgeThresholdMin: 0.0505,
	}
}

// StageConfig holds the window and scene configuration for ui.Setup.
type StageConfig struct {
	Title      string
	Width      int
	Height     int
	ClearColor sg.Color
	// Font, when non-nil, is a *sg.FontFamily used as DefaultFont
	// before any controller OnCreate is called.
	Font *sg.FontFamily
	// FXAA overrides the FXAA post-process configuration. When nil, Setup uses
	// RecommendedFXAAConfig() automatically. Set DisableFXAA to opt out entirely.
	FXAA *FXAAConfig
	// DisableFXAA opts out of the default FXAA pass. Has no effect when FXAA
	// is set explicitly.
	DisableFXAA bool
}

// StageManager manages a stack of Screens. Use the package-level Stage
// singleton in application code.
type StageManager = widget.StageManager

// Stage is the package-level screen-stack singleton. Add, remove, and replace
// screens here; the scene is wired automatically by Setup.
var Stage = widget.DefaultStage

// NewStageManager creates an isolated StageManager. Intended for testing;
// production code uses ui.Stage.
var NewStageManager = widget.NewStageManager

// Setup configures the application window, creates the internal scene, and
// starts the game loop.
// Setup never returns; errors are printed to stderr and the process exits.
//
// When components are passed, Setup creates a Screen, adds them to it, and
// pushes it onto the Stage automatically — no manual Screen/Stage wiring needed.
func Setup(cfg StageConfig, components ...UIElement) {
	if len(components) > 0 {
		screen := widget.NewScreen()
		for _, c := range components {
			screen.Add(c)
		}
		Stage.Add(screen)
	}

	// Set DefaultFont: use config font, or fall back to embedded Go font bundle.
	if cfg.Font != nil && DefaultFont == nil {
		DefaultFont = cfg.Font
	}
	if DefaultFont == nil {
		ff, err := sg.NewFontFamilyFromFontBundle(embeddedGoFont)
		if err != nil {
			fmt.Fprintf(os.Stderr, "willowui: failed to load embedded font: %v\n", err)
			os.Exit(1)
		}
		DefaultFont = ff
	}

	// Default FXAA to RecommendedFXAAConfig unless explicitly overridden or disabled.
	fxaa := cfg.FXAA
	if fxaa == nil && !cfg.DisableFXAA {
		def := RecommendedFXAAConfig()
		fxaa = &def
	}

	scene := sg.NewScene()
	scene.ClearColor = cfg.ClearColor
	Stage.SetScene(scene)
	SetScene(scene)
	scene.SetUpdateFunc(func() error {
		DefaultScheduler.Flush()
		Stage.Update(1.0 / 60.0)
		return nil
	})
	if err := sg.Run(scene, sg.RunConfig{
		Title:  cfg.Title,
		Width:  cfg.Width,
		Height: cfg.Height,
		FXAA:   fxaa,
	}); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// =============================================================================
// Screen & Controller
// =============================================================================

// Controller is implemented by per-screen state owners.
type Controller = widget.Controller

// ScreenOption configures a Screen during construction.
type ScreenOption = widget.ScreenOption

// WithController attaches a Controller to the screen.
var WithController = widget.WithController

// WithScene sets the scene on a Screen explicitly.
// Intended for use in tests; in production the scene is set automatically by Stage.Add.
var WithScene = widget.WithScene

// Screen is the unit of UI in WillowUI. It owns a node subtree and an
// optional controller.
type Screen = widget.Screen

// NewScreen creates a new Screen. Options: WithController, WithScene.
var NewScreen = widget.NewScreen

// SetScene registers the active scene so UI components can read injected
// keyboard input from test runners. The most recently registered non-nil scene
// is retained as a fallback when active scene is nil.
func SetScene(s *sg.Scene) {
	widget.SetScene(s)
	core.SetScene(s)
}

// =============================================================================
// Widgets
// =============================================================================

// Button is an interactive component with a colored background and centered text label.
type Button = widget.Button

// IconButton is an icon-first button that renders a sprite as its primary content,
// with an optional text label beneath or beside it.
type IconButton = widget.IconButton

// IconLabelPosition controls where the text label appears relative to the icon.
type IconLabelPosition = widget.IconLabelPosition

const (
	// IconLabelBelow places the label beneath the icon (default).
	IconLabelBelow = widget.IconLabelBelow
	// IconLabelRight places the label to the right of the icon.
	IconLabelRight = widget.IconLabelRight
)

// NewButton creates a Button with the given name, text label, font, and display size.
var NewButton = widget.NewButton

// NewIconButton creates an icon-first button. Call SetIconImage or SetIconKey to
// provide an icon source.
var NewIconButton = widget.NewIconButton

// Label is a text display component.
type Label = widget.Label

// NewLabel creates a Label with the given name, initial text, font, and display size.
var NewLabel = widget.NewLabel

// NewSectionLabel creates a Label pre-styled as a muted section header using
// the standard dim blue-grey palette common across WillowUI example layouts.
func NewSectionLabel(name, text string, source *sg.FontFamily, size float64) *widget.Label {
	l := widget.NewLabel(name, text, source, size)
	l.SetColor(sg.RGBA(0.45, 0.55, 0.65, 1))
	return l
}

// Badge is a small pill or dot overlay for counts, labels, or status indicators.
type Badge = widget.Badge

// NewBadge creates a Badge with the given name, font, and display size.
var NewBadge = widget.NewBadge

// Tag is a compact pill widget used as a category marker, filter chip, or
// item classifier with optional remove and toggle modes.
type Tag = widget.Tag

// NewTag creates a Tag with the given name, font, and display size.
var NewTag = widget.NewTag

// TagBar is a tag-input widget where the user types text and presses Space
// to create Tag chips. Each tag shows a × to delete it.
type TagBar = widget.TagBar

// NewTagBar creates a TagBar with the given name, font, and display size.
var NewTagBar = widget.NewTagBar

// TextInput is a single-line text entry field.
type TextInput = widget.TextInput

// NewTextInput creates a single-line text input with the given font and display size.
var NewTextInput = widget.NewTextInput

// MaskedInput is a single-line text entry field constrained by a mask pattern.
type MaskedInput = widget.MaskedInput

// NewMaskedInput creates a single-line masked input with the given font and display size.
var NewMaskedInput = widget.NewMaskedInput

// InputField is a labeled text input that combines a Label, TextInput, and
// optional validation message into a single composable unit.
type InputField = widget.InputField

// LabelPosition controls where the field label is placed.
type LabelPosition = widget.LabelPosition

// ValidationState indicates the current validation status.
type ValidationState = widget.ValidationState

// NewInputField creates a new InputField with the given name, font, and display size.
var NewInputField = widget.NewInputField

const (
	LabelAbove = widget.LabelAbove // Label is rendered above the input.
	LabelLeft  = widget.LabelLeft  // Label is rendered to the left of the input.

	ValidationNone    = widget.ValidationNone    // No validation state shown.
	ValidationError   = widget.ValidationError   // Field has a validation error.
	ValidationWarning = widget.ValidationWarning // Field has a validation warning.
	ValidationSuccess = widget.ValidationSuccess // Field passed validation.
)

// SearchBox is a search-oriented single-line input with a magnifier icon,
// optional clear button, debounce, and reactive result population.
type SearchBox = widget.SearchBox

// NewSearchBox creates a SearchBox with the given name, font, and display size.
var NewSearchBox = widget.NewSearchBox

// SetSearchBoxFunc configures the SearchBox with an automatic search callback
// that returns a slice. Results replace the bound reactive Array on each search.
func SetSearchBoxFunc[T any](sb *SearchBox, results *Array[T], fn func(query string) []T) {
	widget.SetSearchBoxFunc(sb, results, fn)
}

// SetSearchBoxIntoFunc configures the SearchBox with an advanced search callback
// that directly mutates the bound reactive Array.
func SetSearchBoxIntoFunc[T any](sb *SearchBox, results *Array[T], fn func(query string, results *Array[T])) {
	widget.SetSearchBoxIntoFunc(sb, results, fn)
}

// TextArea is a multi-line text entry field.
type TextArea = widget.TextArea

// NewTextArea creates a multi-line text area with the given font and display size.
var NewTextArea = widget.NewTextArea

// VisualLine is the exported representation of a visual line for testing.
type VisualLine = widget.VisualLine

// RichText is a text display component supporting markup with bold, italic,
// colors, links, and other inline formatting.
type RichText = widget.RichText

// NewRichText creates a RichText component with the given font and display size.
var NewRichText = widget.NewRichText

// ListItem represents a single item in a List.
type ListItem = widget.ListItem

// List is a scrollable list of items.
type List = widget.List

// NewList creates a scrollable list with the given item height.
var NewList = widget.NewList

// TileList is a scrollable grid of tile items.
type TileList = widget.TileList

// NewTileList creates a scrollable tile list with the given tile dimensions.
var NewTileList = widget.NewTileList

// TreeNode represents a node in a tree hierarchy.
type TreeNode = widget.TreeNode

// ReactiveTreeNode is a tree node with a reactive Children array.
// Use with TreeList.BindRoots so mutations anywhere in the subtree
// automatically update the tree view.
type ReactiveTreeNode = widget.ReactiveTreeNode

// NewReactiveTreeNode creates a ReactiveTreeNode with an empty Children array.
var NewReactiveTreeNode = widget.NewReactiveTreeNode

// TreeList is a collapsible tree view.
type TreeList = widget.TreeList

// NewTreeList creates a TreeList with the given item height.
var NewTreeList = widget.NewTreeList

// NewTreeToggle creates a toggle button for expanding/collapsing a tree node.
var NewTreeToggle = widget.NewTreeToggle

// TreeToggleSize is the width and height (in pixels) of a tree toggle button.
const TreeToggleSize = widget.TreeToggleSize

// treeExpandGlyph returns the expand glyph image. Used for testing.
var treeExpandGlyph = widget.TreeExpandGlyph

// treeCollapseGlyph returns the collapse glyph image. Used for testing.
var treeCollapseGlyph = widget.TreeCollapseGlyph

// PasswordDotGlyph returns the procedural dot glyph used for password masking.
var PasswordDotGlyph = widget.PasswordDotGlyph

// TabBar is a tab navigation component.
type TabBar = widget.TabBar

// NewTabBar creates a TabBar with the given font and display size.
var NewTabBar = widget.NewTabBar

// TabOverflowMode controls what happens when tabs overflow the bar width.
type TabOverflowMode = widget.TabOverflowMode

const (
	TabOverflowClip   = widget.TabOverflowClip   // Clip tabs that overflow the bar width.
	TabOverflowScroll = widget.TabOverflowScroll // Enable horizontal scrolling for overflow tabs.
)

// Slider is a draggable range control for selecting a numeric value.
type Slider = widget.Slider

// NewSlider creates a horizontal slider with range [0, 1].
var NewSlider = widget.NewSlider

// NumberStepper is a numeric input combining a text field with decrement and
// increment step buttons.
type NumberStepper = widget.NumberStepper

// NewNumberStepper creates a NumberStepper with range (-∞, +∞), step 1, and
// zero decimal places.
var NewNumberStepper = widget.NewNumberStepper

// OptionRotator is a compact selection widget with left/right chevrons and a
// centered value label. Clicking the chevrons — or pressing Left/Right arrow
// keys when focused — cycles through a fixed list of string options.
type OptionRotator = widget.OptionRotator

// NewOptionRotator creates an OptionRotator with the given name, options list,
// font, and display size. Panics if options is empty.
var NewOptionRotator = widget.NewOptionRotator

// MeterBar displays a horizontal fill bar for values like HP, mana, or XP.
type MeterBar = widget.MeterBar

// ProgressBar is an alias for MeterBar.
type ProgressBar = widget.MeterBar

// NewMeterBar creates a MeterBar with range [0, 1].
var NewMeterBar = widget.NewMeterBar

// NewProgressBar creates a MeterBar with range [0, 1]. Alias for NewMeterBar.
var NewProgressBar = widget.NewProgressBar

// ScrollBar is a scrollbar with a draggable thumb.
type ScrollBar = widget.ScrollBar

// NewScrollBar creates a ScrollBar.
var NewScrollBar = widget.NewScrollBar

// ScrollPanel is a scrollable container panel.
type ScrollPanel = widget.ScrollPanel

// NewScrollPanel creates a ScrollPanel.
var NewScrollPanel = widget.NewScrollPanel

// Panel is a static container with optional background color, border, and layout.
type Panel = widget.Panel

// NewPanel creates a Panel with no background and no border.
var NewPanel = widget.NewPanel

// AnchorLayout is a container that pins children to edges or corners of the
// parent with pixel offsets. The primary tool for HUD composition.
type AnchorLayout = widget.AnchorLayout

// NewAnchorLayout creates an AnchorLayout container.
var NewAnchorLayout = widget.NewAnchorLayout

// TwoColumnLayout arranges children in labeled two-column rows. Each column is
// independently aligned. Primary use case is settings screens, stat sheets, and
// form dialogs.
type TwoColumnLayout = widget.TwoColumnLayout

// NewTwoColumnLayout creates a TwoColumnLayout container with sensible defaults:
// left column right-aligned, right column left-aligned, 12px gap, 8px row spacing.
var NewTwoColumnLayout = widget.NewTwoColumnLayout

// Anchor identifies a position within a parent container.
type Anchor = widget.Anchor

const (
	AnchorTopLeft      = widget.AnchorTopLeft      // Top-left corner.
	AnchorTopCenter    = widget.AnchorTopCenter    // Top edge, centered horizontally.
	AnchorTopRight     = widget.AnchorTopRight     // Top-right corner.
	AnchorMiddleLeft   = widget.AnchorMiddleLeft   // Left edge, centered vertically.
	AnchorCenter       = widget.AnchorCenter       // Centered in both axes.
	AnchorMiddleRight  = widget.AnchorMiddleRight  // Right edge, centered vertically.
	AnchorBottomLeft   = widget.AnchorBottomLeft   // Bottom-left corner.
	AnchorBottomCenter = widget.AnchorBottomCenter // Bottom edge, centered horizontally.
	AnchorBottomRight  = widget.AnchorBottomRight  // Bottom-right corner.
)

// Window is a draggable, resizable floating window component.
type Window = widget.Window

// WindowManager manages multiple floating windows.
type WindowManager = widget.WindowManager

// NavDrawer is a slide-out navigation panel anchored to an edge of the screen.
type NavDrawer = widget.NavDrawer

// NavDrawerAnchor specifies which edge the drawer slides from.
type NavDrawerAnchor = widget.NavDrawerAnchor

const (
	NavDrawerLeft  = widget.NavDrawerLeft  // Drawer slides from the left edge.
	NavDrawerRight = widget.NavDrawerRight // Drawer slides from the right edge.
)

// NewNavDrawer creates a NavDrawer anchored to the left edge by default.
var NewNavDrawer = widget.NewNavDrawer

// NewWindow creates a Window with the given title, font, and display size.
var NewWindow = widget.NewWindow

// NewWindowManager creates a new window manager.
var NewWindowManager = widget.NewWindowManager

// DefaultWindowManager is the package-level default window manager instance.
var DefaultWindowManager = widget.DefaultWindowManager

// DefaultWindowWidth is the default width for new windows.
const DefaultWindowWidth = widget.DefaultWindowWidth

// DefaultWindowHeight is the default height for new windows.
const DefaultWindowHeight = widget.DefaultWindowHeight

// WindowTitleBarHeight is the height of the title bar in pixels.
const WindowTitleBarHeight = widget.WindowTitleBarHeight

// Checkbox is a toggle control with a box and check mark.
type Checkbox = widget.Checkbox

// NewCheckbox creates a Checkbox with the given label text, font, and display size.
var NewCheckbox = widget.NewCheckbox

// Toggle is a binary on/off switch with an animated sliding thumb.
type Toggle = widget.Toggle

// NewToggle creates a Toggle switch with default dimensions.
var NewToggle = widget.NewToggle

// Radio manages a group of mutually exclusive radio buttons.
type Radio = widget.Radio

// RadioButton is a single option within a Radio widget.
type RadioButton = widget.RadioButton

// NewRadio creates a new empty Radio widget.
var NewRadio = widget.NewRadio

// ToggleButtonBar is a row of mutually exclusive toggle buttons.
type ToggleButtonBar = widget.ToggleButtonBar

// NewToggleButtonBar creates a ToggleButtonBar with the given font and display size.
var NewToggleButtonBar = widget.NewToggleButtonBar

// Tooltip is a floating overlay that appears after a hover delay. It embeds
// Component, so all layout, sizing, and child-management methods are available.
// Tooltips are never added directly to the scene — use SetTooltip on a trigger.
type Tooltip = widget.Tooltip

// NewTooltip creates a Tooltip with sensible defaults.
var NewTooltip = widget.NewTooltip

// TooltipAnchor controls where a tooltip is placed relative to its trigger.
type TooltipAnchor = widget.TooltipAnchor

// TooltipManager manages tooltip visibility for the scene.
type TooltipManager = widget.TooltipManager

// DefaultTooltipManager is the package-level tooltip manager.
// Set DefaultTooltipManager.Enabled = false to disable all tooltips globally.
var DefaultTooltipManager = widget.DefaultTooltipManager

// MenuItem is a single entry in a MenuPopup.
type MenuItem = widget.MenuItem

// MenuPopup is a floating list of items shown by MenuPopupManager.
type MenuPopup = widget.MenuPopup

// NewMenuPopup creates a MenuPopup that will display items using font at displaySize.
var NewMenuPopup = widget.NewMenuPopup

// MenuPopupManager manages the single active floating menu popup.
type MenuPopupManager = widget.MenuPopupManager

// DefaultMenuPopupManager is the package-level menu popup manager.
var DefaultMenuPopupManager = widget.DefaultMenuPopupManager

// SelectOption is a single choice in a Select widget.
type SelectOption = widget.SelectOption

// Select is a dropdown widget that opens a MenuPopup when clicked.
type Select = widget.Select

// NewSelect creates a Select with the given name, options list, font, and display size.
var NewSelect = widget.NewSelect

// ContextMenu is a list of items shown on right-click.
type ContextMenu = widget.ContextMenu

// NewContextMenu creates a ContextMenu with the given font and display size.
var NewContextMenu = widget.NewContextMenu

// MenuBarEntry defines one top-level menu in the bar.
type MenuBarEntry = widget.MenuBarEntry

// MenuBar is a horizontal bar of labeled menu buttons that open dropdown panels.
type MenuBar = widget.MenuBar

// NewMenuBar creates a new MenuBar with the given font and display size.
var NewMenuBar = widget.NewMenuBar

// DragHandle is a visible grip primitive that emits drag delta events and can
// optionally resize a target component directly.
type DragHandle = widget.DragHandle

// DragAxis specifies which axis a DragHandle operates on.
type DragAxis = widget.DragAxis

// DragGripStyle specifies the visual indicator rendered on a DragHandle.
type DragGripStyle = widget.DragGripStyle

// NewDragHandle creates a DragHandle with default dot grip style.
var NewDragHandle = widget.NewDragHandle

const (
	DragAxisX        = widget.DragAxisX        // Constrain drag to horizontal movement.
	DragAxisY        = widget.DragAxisY        // Constrain drag to vertical movement.
	DragAxisDiagonal = widget.DragAxisDiagonal // Allow free diagonal drag.

	DragGripDots  = widget.DragGripDots  // Render grip as a dot pattern.
	DragGripLines = widget.DragGripLines // Render grip as horizontal lines.
	DragGripNone  = widget.DragGripNone  // No visual grip indicator.
)

// Image is a display-only component that renders a sprite, texture region,
// or engine.Image with configurable fit/fill modes, tinting, and corner radius.
type Image = widget.Image

// ImageScaleMode controls how the image is laid out within the widget bounds.
type ImageScaleMode = widget.ImageScaleMode

// NewImage creates an Image widget with no source set.
var NewImage = widget.NewImage

const (
	ImageScaleStretch = widget.ImageScaleStretch // Stretch to fill bounds, ignoring aspect ratio.
	ImageScaleFit     = widget.ImageScaleFit     // Scale to fit within bounds, preserving aspect ratio.
	ImageScaleFill    = widget.ImageScaleFill    // Scale to cover bounds, cropping overflow.
	ImageScaleCenter  = widget.ImageScaleCenter  // Center at original size, no scaling.
	ImageScaleTile    = widget.ImageScaleTile    // Tile the image to fill bounds.
)

// AnimatedImage extends Image to play back a frame-strip sprite animation,
// cycling through regions at a configurable FPS.
type AnimatedImage = widget.AnimatedImage

// AnimPlayMode controls how an AnimatedImage loops.
type AnimPlayMode = widget.AnimPlayMode

// NewAnimatedImage creates an AnimatedImage widget with no frames set.
var NewAnimatedImage = widget.NewAnimatedImage

const (
	AnimPlayOnce     = widget.AnimPlayOnce     // Play frames once and stop on the last frame.
	AnimPlayLoop     = widget.AnimPlayLoop     // Loop from the beginning after the last frame.
	AnimPlayPingPong = widget.AnimPlayPingPong // Reverse direction at each end of the sequence.
)

// TooltipAnchor constants control tooltip placement relative to the trigger.
const (
	TooltipAuto              = widget.TooltipAuto              // Automatically choose the best side.
	TooltipAbove             = widget.TooltipAbove             // Place above the trigger.
	TooltipBelow             = widget.TooltipBelow             // Place below the trigger.
	TooltipLeft              = widget.TooltipLeft              // Place to the left of the trigger.
	TooltipRight             = widget.TooltipRight             // Place to the right of the trigger.
	TooltipCornerTopLeft     = widget.TooltipCornerTopLeft     // Anchor to the trigger's top-left corner.
	TooltipCornerTopRight    = widget.TooltipCornerTopRight    // Anchor to the trigger's top-right corner.
	TooltipCornerBottomLeft  = widget.TooltipCornerBottomLeft  // Anchor to the trigger's bottom-left corner.
	TooltipCornerBottomRight = widget.TooltipCornerBottomRight // Anchor to the trigger's bottom-right corner.
	TooltipFollowMouse       = widget.TooltipFollowMouse       // Follow the mouse cursor.
)

// =============================================================================
// ColorPicker
// =============================================================================

// ColorPicker is a swatch + label trigger that opens a floating picker popup.
type ColorPicker = widget.ColorPicker

// ColorMode selects which input mode the picker popup displays.
type ColorMode = widget.ColorMode

const (
	ColorModeHex   = widget.ColorModeHex   // Hexadecimal color input (#RRGGBB).
	ColorModeRGB   = widget.ColorModeRGB   // RGB sliders (0-255).
	ColorModeHSV   = widget.ColorModeHSV   // Hue/Saturation/Value sliders.
	ColorModeHSL   = widget.ColorModeHSL   // Hue/Saturation/Lightness sliders.
	ColorModeFloat = widget.ColorModeFloat // Floating-point RGB (0.0-1.0).
)

// NewColorPicker creates a ColorPicker trigger.
var NewColorPicker = widget.NewColorPicker

// ColorPickerManager manages the single active floating color picker popup.
type ColorPickerManager = widget.ColorPickerManager

// DefaultColorPickerManager is the package-level color picker manager.
var DefaultColorPickerManager = widget.DefaultColorPickerManager

// =============================================================================
// Color Utilities
// =============================================================================

var (
	// ParseHex parses a hex color string (#RGB, #RGBA, #RRGGBB, or #RRGGBBAA) into a color.Color.
	ParseHex = colorutil.ParseHex
	// FormatHex formats a color as a #RRGGBB hex string.
	FormatHex = colorutil.FormatHex
	// FormatHexA formats a color as a #RRGGBBAA hex string including alpha.
	FormatHexA = colorutil.FormatHexA
	// ToRGB255 converts a color.Color to 0-255 R, G, B, A components.
	ToRGB255 = colorutil.ToRGB255
	// FromRGB255 creates a color from 0-255 R, G, B, A components.
	FromRGB255 = colorutil.FromRGB255
	// ToHSL converts a color.Color to hue (0-360), saturation (0-1), lightness (0-1).
	ToHSL = colorutil.ToHSL
	// FromHSL creates a color from hue (0-360), saturation (0-1), lightness (0-1).
	FromHSL = colorutil.FromHSL
	// ToHSV converts a color.Color to hue (0-360), saturation (0-1), value (0-1).
	ToHSV = colorutil.ToHSV
	// FromHSV creates a color from hue (0-360), saturation (0-1), value (0-1).
	FromHSV = colorutil.FromHSV
	// NormalizeRGB clamps and converts a color to 0.0-1.0 float64 RGBA components.
	NormalizeRGB = colorutil.NormalizeRGB
)

// =============================================================================
// Toast
// =============================================================================

// ToastAnchor specifies which screen corner toasts stack at.
type ToastAnchor = widget.ToastAnchor

// ToastOption is a functional option for configuring a toast.
type ToastOption = widget.ToastOption

// ToastManager manages a stack of transient toast notifications.
type ToastManager = widget.ToastManager

// DefaultToastManager is the package-level singleton used by ShowToast.
var DefaultToastManager = widget.DefaultToastManager

// ShowToast shows a toast via DefaultToastManager with the given variant.
var ShowToast = widget.ShowToast

// WithDuration sets the auto-dismiss duration.
var WithDuration = widget.WithDuration

// WithDismissOnClick enables or disables click-to-dismiss (default true).
var WithDismissOnClick = widget.WithDismissOnClick

// WithProgress shows a shrinking remaining-time bar at the bottom of the toast.
var WithProgress = widget.WithProgress

// WithOnDismiss sets a callback invoked when the toast is dismissed.
var WithOnDismiss = widget.WithOnDismiss

const (
	ToastBottomRight = widget.ToastBottomRight // Stack toasts in the bottom-right corner.
	ToastBottomLeft  = widget.ToastBottomLeft  // Stack toasts in the bottom-left corner.
	ToastTopRight    = widget.ToastTopRight    // Stack toasts in the top-right corner.
	ToastTopLeft     = widget.ToastTopLeft     // Stack toasts in the top-left corner.
)

// =============================================================================
// Popover
// =============================================================================

// Popover is a floating rich-content panel anchored to a trigger component.
// It is dismissable, interactive, and designed for heavier content than Tooltip.
type Popover = widget.Popover

// PopoverSide controls which side of the trigger the popover prefers to appear on.
type PopoverSide = widget.PopoverSide

// PopoverManager manages the single active floating popover.
type PopoverManager = widget.PopoverManager

// DefaultPopoverManager is the package-level singleton used by all Popover instances.
var DefaultPopoverManager = widget.DefaultPopoverManager

// NewPopover creates a new Popover with the given name.
var NewPopover = widget.NewPopover

const (
	PopoverBelow = widget.PopoverBelow // Prefer placement below the trigger.
	PopoverAbove = widget.PopoverAbove // Prefer placement above the trigger.
	PopoverRight = widget.PopoverRight // Prefer placement to the right.
	PopoverLeft  = widget.PopoverLeft  // Prefer placement to the left.
)

// =============================================================================
// DataTable
// =============================================================================

// DataTable is a virtualized, sortable, filterable data grid widget.
type DataTable = widget.DataTable

// DataTableColumn defines a column in a DataTable.
type DataTableColumn = widget.DataTableColumn

// CellCoord identifies a cell by row and column index.
type CellCoord = widget.CellCoord

// SortType controls how a column's values are compared during sort.
type SortType = widget.SortType

// CellStyle holds styling overrides and hooks for DataTable cells and headers.
type CellStyle = widget.CellStyle

// LabelStyle is a deprecated alias for CellStyle.
type LabelStyle = widget.LabelStyle

// CellClipMode controls how cell content is clipped when it overflows.
type CellClipMode = widget.CellClipMode

// DataTableScrollMode controls how the DataTable scrolls.
// Note: named DataTableScrollMode to avoid conflict with existing types.
type DataTableScrollMode = widget.ScrollMode

// DataTableSelectionMode controls row selection behavior in a DataTable.
type DataTableSelectionMode = widget.SelectionMode

// SortDirection indicates the sort order for a column.
type SortDirection = widget.SortDirection

// SortKey identifies a column and its sort direction in a multi-sort stack.
type SortKey = widget.SortKey

// OnSortScroll controls scroll behavior after a sort operation.
type OnSortScroll = widget.OnSortScroll

// NewDataTable creates a DataTable with the given name and row height.
var NewDataTable = widget.NewDataTable

// LabelColumn creates a simple text-label column with the given key, header and accessor.
var LabelColumn = widget.LabelColumn

// MeterColumn creates a column that renders an inline MeterBar for each row.
// The accessor returns a float64 in [0, 1]. Use Cell.OnPostUpdate to
// customize the fill color dynamically.
var MeterColumn = widget.MeterColumn

// SelectionColumn creates a column that renders checkboxes (multi-select)
// or radio dots (single-select) per row. Visibility and mode are controlled
// by reactive Refs so external UI can toggle batch mode dynamically.
var SelectionColumn = widget.SelectionColumn

// SetRowClickSelects configures whether clicking anywhere on a row triggers
// the selection toggle for a SelectionColumn.
var SetRowClickSelects = widget.SetRowClickSelects

// EllipsisLabel creates a Label for use in cell rendering.
var EllipsisLabel = widget.EllipsisLabel

// UpdateEllipsisLabel updates the text of an EllipsisLabel component.
var UpdateEllipsisLabel = widget.UpdateEllipsisLabel

const (
	SortAlpha   = widget.SortAlpha   // Alphabetic string comparison.
	SortNumeric = widget.SortNumeric // Numeric comparison.
	SortCustom  = widget.SortCustom  // User-supplied comparison function.

	ClipEllipsis = widget.ClipEllipsis // Truncate overflow text with an ellipsis.
	ClipMask     = widget.ClipMask     // Mask overflow with a clipping rectangle.

	ScrollModeVirtual = widget.ScrollModeVirtual // Only render visible rows (virtualized).
	ScrollModeStatic  = widget.ScrollModeStatic  // Render all rows (no virtualization).

	SelectionModeNone   = widget.SelectionModeNone   // Row selection disabled.
	SelectionModeSingle = widget.SelectionModeSingle // Only one row may be selected.
	SelectionModeMulti  = widget.SelectionModeMulti  // Multiple rows may be selected.

	SortNone = widget.SortNone // Column is unsorted.
	SortAsc  = widget.SortAsc  // Ascending sort order.
	SortDesc = widget.SortDesc // Descending sort order.

	OnSortScrollNone        = widget.OnSortScrollNone        // No scroll adjustment after sort.
	OnSortScrollToSelection = widget.OnSortScrollToSelection // Scroll to keep selection visible after sort.
	OnSortScrollToTop       = widget.OnSortScrollToTop       // Scroll to top after sort.
)

// =============================================================================
// TreeTable
// =============================================================================

// TreeTable is a hybrid tree + column grid where rows can be expanded/collapsed
// while each row spans multiple data columns.
type TreeTable = widget.TreeTable

// TableColumn defines a column in a TreeTable.
type TableColumn = widget.TableColumn

// SortDir indicates the sort order for a TreeTable column.
type SortDir = widget.SortDir

const (
	SortDirAsc  = widget.SortDirAsc  // Ascending sort direction.
	SortDirDesc = widget.SortDirDesc // Descending sort direction.
)

// TreeTableRow represents a row in the TreeTable hierarchy.
type TreeTableRow = widget.TreeTableRow

// NewTreeTable creates a new TreeTable with the given name, font, and display size.
var NewTreeTable = widget.NewTreeTable

// =============================================================================
// KeybindInput
// =============================================================================

// KeybindInput is a settings control that captures a keyboard or gamepad
// binding. It displays the current binding as a styled key cap label and
// enters listening mode on click to capture a new binding.
type KeybindInput = widget.KeybindInput

// KeyBinding represents a keyboard or gamepad binding.
type KeyBinding = widget.KeyBinding

// NewKeybindInput creates a KeybindInput with the given name, font, and display size.
var NewKeybindInput = widget.NewKeybindInput

// =============================================================================
// SortableList
// =============================================================================

// SortableList is a vertical list widget specialized for ordered collections
// with drag-handle-based reordering and keyboard reorder commands.
type SortableList = widget.SortableList

// SortHandleSide specifies which side of each row the drag handle appears on.
type SortHandleSide = widget.SortHandleSide

// NewSortableList creates a new sortable list with fixed item height.
var NewSortableList = widget.NewSortableList

const (
	SortHandleLeft  = widget.SortHandleLeft  // Drag handle on the left side of each row.
	SortHandleRight = widget.SortHandleRight // Drag handle on the right side of each row.
)

// BindSortableListItems binds a reactive Array[T] to a SortableList.
func BindSortableListItems[T any](sl *SortableList, items *Array[T]) {
	widget.BindSortableListItems(sl, items)
}

// =============================================================================
// SortableTreeList
// =============================================================================

// SortableTreeList is a hierarchical list where nodes can be reordered by drag
// within their level and optionally reparented by dragging onto another node.
type SortableTreeList = widget.SortableTreeList

// SortableTreeItem represents a node in a sortable tree hierarchy.
type SortableTreeItem = widget.SortableTreeItem

// NewSortableTreeList creates a new sortable tree list.
var NewSortableTreeList = widget.NewSortableTreeList

// =============================================================================
// StatWeb
// =============================================================================

// StatWeb is an editable polygon stat display (spider/radar chart) with named
// axes and draggable handles for attribute editing.
type StatWeb = widget.StatWeb

// StatAxis defines a single spoke on a StatWeb.
type StatAxis = widget.StatAxis

// NewStatWeb creates a StatWeb with the given name, font, and font size.
var NewStatWeb = widget.NewStatWeb

// =============================================================================
// Accordion
// =============================================================================

// Accordion is a vertically stacked list of collapsible sections, each with
// a header row and an arbitrary content panel.
type Accordion = widget.Accordion

// AccordionSection defines a section to add to an Accordion.
type AccordionSection = widget.AccordionSection

// NewAccordion creates an Accordion with default settings.
var NewAccordion = widget.NewAccordion

// =============================================================================
// GradientEditor
// =============================================================================

// GradientMode selects which corners are independently editable in a GradientEditor.
type GradientMode = widget.GradientMode

const (
	// GradientModeH is a horizontal gradient (TL=BL, TR=BR).
	GradientModeH = widget.GradientModeH
	// GradientModeV is a vertical gradient (TL=TR, BL=BR).
	GradientModeV = widget.GradientModeV
	// GradientModeFourCorner is a 4-corner bilinear gradient (all corners independent).
	GradientModeFourCorner = widget.GradientModeFourCorner
)

// Gradient is the value type produced by GradientEditor.
type Gradient = widget.Gradient

// GradientEditor edits horizontal, vertical, or 4-corner gradients.
type GradientEditor = widget.GradientEditor

// NewGradientEditor creates a new GradientEditor with the given name, font, and display size.
var NewGradientEditor = widget.NewGradientEditor

// =============================================================================
// Gradient utilities
// =============================================================================

var (
	// SampleBilinear returns the bilinearly interpolated color at normalized (u, v).
	SampleBilinear = colorutil.SampleBilinear
	// FormatGradientString returns the theme-compatible JSON fill string for a Gradient.
	FormatGradientString = colorutil.FormatGradientString
	// DefaultGradient returns a horizontal black→white gradient.
	DefaultGradient = colorutil.DefaultGradient
)

// =============================================================================
// Template
// =============================================================================

// FactoryContext provides resources needed by component factories.
type FactoryContext = template.FactoryContext

// WidgetFactory creates a custom widget component by name for use in XML templates.
type WidgetFactory = template.WidgetFactory

// AttrSetter applies a named attribute value to a custom widget component.
type AttrSetter = template.AttrSetter

// TemplateRegistry stores compiled XML templates and instantiates them.
type TemplateRegistry = template.TemplateRegistry

// NewTemplateRegistry creates a new template registry.
var NewTemplateRegistry = template.NewTemplateRegistry

// NewTemplateRegistryWithFont creates a new template registry with a default
// font loaded from raw TTF data and a display font size.
var NewTemplateRegistryWithFont = template.NewTemplateRegistryWithFont

// DataProvider is implemented by controllers that support XML template data binding.
type DataProvider = template.DataProvider

// EvalContext provides the evaluation environment for expressions.
type EvalContext = template.EvalContext

// ExprNode is the interface for all expression AST nodes.
type ExprNode = template.ExprNode

// CompileXML parses XML template data and compiles it to an IR tree.
var CompileXML = template.CompileXML

// CompileXMLWithTypes compiles XML template data, accepting extra custom widget type names.
var CompileXMLWithTypes = template.CompileXMLWithTypes

// IRNode is the intermediate representation of a compiled XML template element.
type IRNode = template.IRNode

// IRAttribute represents a single attribute on an IR node.
type IRAttribute = template.IRAttribute

// IRDirective represents a structural directive attached to an IR node.
type IRDirective = template.IRDirective

// DirectiveType identifies a structural directive in a compiled template.
type DirectiveType = template.DirectiveType

const (
	DirectiveIf   DirectiveType = template.DirectiveIf   // Conditional rendering (if:expr).
	DirectiveFor  DirectiveType = template.DirectiveFor  // List rendering (for:item in collection).
	DirectiveShow DirectiveType = template.DirectiveShow // Visibility toggle (show:expr).
)

// ParseExpression parses an expression string into an AST.
var ParseExpression = template.ParseExpression

// EvalExpression evaluates an expression AST node in the given context.
var EvalExpression = template.EvalExpression

// ExprRef is a data-binding path reference (e.g. "user.name").
type ExprRef = template.ExprRef

// ExprLiteral is a constant value node.
type ExprLiteral = template.ExprLiteral

// BinOp identifies a binary operator.
type BinOp = template.BinOp

const (
	BinAdd BinOp = template.BinAdd // Addition (+).
	BinSub BinOp = template.BinSub // Subtraction (-).
	BinMul BinOp = template.BinMul // Multiplication (*).
	BinDiv BinOp = template.BinDiv // Division (/).
	BinMod BinOp = template.BinMod // Modulo (%).
	BinEq  BinOp = template.BinEq  // Equality (==).
	BinNeq BinOp = template.BinNeq // Inequality (!=).
	BinLt  BinOp = template.BinLt  // Less than (<).
	BinLte BinOp = template.BinLte // Less than or equal (<=).
	BinGt  BinOp = template.BinGt  // Greater than (>).
	BinGte BinOp = template.BinGte // Greater than or equal (>=).
	BinAnd BinOp = template.BinAnd // Logical AND (&&).
	BinOr  BinOp = template.BinOr  // Logical OR (||).
)

// ExprBinary is a binary operation node (e.g. a + b, x == y).
type ExprBinary = template.ExprBinary

// UnaryOp identifies a unary operator.
type UnaryOp = template.UnaryOp

const (
	UnaryNot UnaryOp = template.UnaryNot // Logical negation (!).
	UnaryNeg UnaryOp = template.UnaryNeg // Arithmetic negation (-).
)

// ExprUnary is a unary operation node (e.g. !visible, -offset).
type ExprUnary = template.ExprUnary

// ExprTernary is a ternary conditional node (cond ? then : else).
type ExprTernary = template.ExprTernary

// ExprConcat is a string-interpolation concat node.
type ExprConcat = template.ExprConcat

// DecodeIR decodes a .xmlui binary blob into an IR tree.
var DecodeIR = template.DecodeIR

// EncodeIR encodes an IR tree into a .xmlui binary blob.
var EncodeIR = template.EncodeIR

// =============================================================================
// Scene helpers
// =============================================================================

// SetUpdateFunc wraps fn and sets it as the scene's per-frame update function.
// The reactive scheduler is automatically flushed at the start of each frame
// before fn runs, so Ref.Set() calls made inside event callbacks (OnClick, etc.)
// propagate to watchers and bound widgets within the same frame — no manual
// DefaultScheduler.Flush() required.
func SetUpdateFunc(scene *sg.Scene, fn func() error) {
	scene.SetUpdateFunc(func() error {
		DefaultScheduler.Flush()
		return fn()
	})
}

// NewSpacer creates an invisible fixed-size gap for use in VBox/HBox layouts.
// It occupies layout space without rendering anything.
var NewSpacer = widget.NewSpacer

// NewDivider creates a horizontal rule sprite styled with the standard divider
// color. Position it with SetPosition and add it to a node with AddChild.
func NewDivider(name string, width float64) *sg.Node {
	d := sg.NewSprite(name, sg.TextureRegion{})
	d.SetScale(width, 1)
	d.SetColor(sg.RGBA(0.25, 0.3, 0.35, 1))
	return d
}

// =============================================================================
// Dev tools
// =============================================================================

// HotReloader watches XML template and JSON theme files, live-reloading
// when they change. Only available with the "hotreload" build tag.
type HotReloader = dev.HotReloader

// NewHotReloader creates a hot reloader that watches xmlPath for changes and
// recompiles the template, swapping the live component tree on the screen.
var NewHotReloader = dev.NewHotReloader

// NewHotReloaderDirect creates a HotReloader without starting a file watcher.
// Intended for unit tests that call Reload() directly.
var NewHotReloaderDirect = dev.NewHotReloaderDirect

// =============================================================================
// TimePicker
// =============================================================================

// TimePicker is a compact hour/minute/second picker with up/down stepper
// columns and an optional AM/PM toggle.
type TimePicker = widget.TimePicker

// TimeValue represents a time-of-day as hour (0-23), minute, and second.
type TimeValue = widget.TimeValue

// TimeFormat selects 12-hour or 24-hour time display.
type TimeFormat = widget.TimeFormat

// NewTimePicker creates a TimePicker with default 24h format, no seconds.
var NewTimePicker = widget.NewTimePicker

const (
	TimeFormat24h = widget.TimeFormat24h // 24-hour time display (00:00 - 23:59).
	TimeFormat12h = widget.TimeFormat12h // 12-hour time display with AM/PM.
)

// =============================================================================
// ImageCropper
// =============================================================================

// ImageCropper displays an image with a draggable crop rectangle.
type ImageCropper = widget.ImageCropper

// NewImageCropper creates an ImageCropper widget.
var NewImageCropper = widget.NewImageCropper

// =============================================================================
// ToolBar
// =============================================================================

// ToolBar is a horizontal or vertical command strip for housing actions,
// toggle groups, separators, and compact controls.
type ToolBar = widget.ToolBar

// ToolBarOverflowMode controls how items behave when they exceed the toolbar's bounds.
type ToolBarOverflowMode = widget.ToolBarOverflowMode

// NewToolBar creates a new ToolBar with horizontal orientation and clip overflow.
var NewToolBar = widget.NewToolBar

const (
	ToolBarClip   = widget.ToolBarClip   // Clip items that overflow the toolbar bounds.
	ToolBarScroll = widget.ToolBarScroll // Enable scrolling for overflow items.
	ToolBarWrap   = widget.ToolBarWrap   // Wrap overflow items to a new row/column.
)

// ToolGroup manages mutually-exclusive selection among a set of IconButtons.
type ToolGroup = widget.ToolGroup

// NewToolGroup creates a new ToolGroup for radio-style icon button selection.
var NewToolGroup = widget.NewToolGroup

// =============================================================================
// CalendarSelector
// =============================================================================

// CalendarSelector is a month-grid date picker with prev/next month navigation.
type CalendarSelector = widget.CalendarSelector

// NewCalendarSelector creates a CalendarSelector with today's date selected.
var NewCalendarSelector = widget.NewCalendarSelector
