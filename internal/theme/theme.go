package theme

import (
	"math"

	"github.com/devthicket/willowui/internal/core"
	"github.com/devthicket/willowui/internal/engine"
	"github.com/devthicket/willowui/internal/render"
	"github.com/devthicket/willowui/internal/sg"
)

// ---------------------------------------------------------------------------
// Variant
// ---------------------------------------------------------------------------

// Variant selects a color group for a component (e.g. Primary, Danger).
type Variant uint8

const (
	Primary Variant = iota // default for all components
	Secondary
	Accent
	Neutral
	Danger
	Success
	Warning
	Info
	Custom1
	Custom2
	Custom3
	Custom4
	Custom5
	Custom6
	Custom7
	Custom8
	Custom9
	Custom10
	Custom11
	Custom12
	Custom13
	Custom14
	Custom15
	Custom16
	Custom17
	Custom18
	Custom19
	Custom20
	Custom21
	Custom22
	Custom23
	Custom24
	Custom25
	Custom26
	Custom27
	Custom28
	Custom29
	Custom30
	Custom31
	Custom32
	Custom33
	Custom34
	Custom35
	Custom36
	Custom37
	Custom38
	Custom39
	Custom40
	Custom41
	Custom42
	Custom43
	Custom44
	Custom45
	Custom46
	Custom47
	Custom48
	Custom49
	Custom50
	Custom51
	Custom52
	Custom53
	Custom54
	Custom55
	Custom56

	VariantCount = 64
)

// ---------------------------------------------------------------------------
// Property types
// ---------------------------------------------------------------------------

// SpriteRef holds a resolved sprite image from the theme.
// Set is false when the slot was not specified in the theme JSON —
// the widget falls back to its built-in procedural glyph.
type SpriteRef struct {
	Image engine.Image // sub-image extracted from the sprite sheet
	Set   bool
}

// ColorProperty holds a color value for each component state.
type ColorProperty [core.StateCount]sg.Color

// Resolve returns the color for the given state.
func (p *ColorProperty) Resolve(s core.ComponentState) sg.Color {
	return p[s]
}

// BackgroundProperty holds a background value for each component state.
type BackgroundProperty [core.StateCount]core.Background

// Resolve returns the background for the given state.
func (p *BackgroundProperty) Resolve(s core.ComponentState) core.Background {
	return p[s]
}

// FloatProperty holds a float64 value for each component state.
type FloatProperty [core.StateCount]float64

// Resolve returns the float for the given state.
func (p *FloatProperty) Resolve(s core.ComponentState) float64 {
	return p[s]
}

// ---------------------------------------------------------------------------
// Property construction helpers
// ---------------------------------------------------------------------------

// NewColorPropUniform creates a ColorProperty where all states have the same color.
func NewColorPropUniform(c sg.Color) ColorProperty {
	var p ColorProperty
	for i := range p {
		p[i] = c
	}
	return p
}

// NewColorPropStates creates a ColorProperty from explicit state values.
// Unset states (zero Color) are filled via the fallback chain.
func NewColorPropStates(m map[core.ComponentState]sg.Color) ColorProperty {
	var p ColorProperty
	for s, c := range m {
		p[s] = c
	}
	ResolveColorFallbacks(&p)
	return p
}

// ResolveColorFallbacks fills unset (zero) states using the fallback chain.
func ResolveColorFallbacks(p *ColorProperty) {
	zero := sg.Color{}
	for s := core.ComponentState(0); s < core.StateCount; s++ {
		if p[s] != zero {
			continue
		}
		for _, fb := range core.StateFallbacks[s] {
			if p[fb] != zero {
				p[s] = p[fb]
				break
			}
		}
	}
}

// NewSolidBgPropStates creates a BackgroundProperty from explicit state colors.
// Unset states are filled via the fallback chain.
func NewSolidBgPropStates(m map[core.ComponentState]sg.Color) BackgroundProperty {
	var p BackgroundProperty
	for s, c := range m {
		p[s] = core.SolidBackground(c)
	}
	ResolveBgFallbacks(&p)
	return p
}

// NewSolidBgPropUniform creates a BackgroundProperty where all states have the
// same solid color.
func NewSolidBgPropUniform(c sg.Color) BackgroundProperty {
	var p BackgroundProperty
	for i := range p {
		p[i] = core.SolidBackground(c)
	}
	return p
}

// ResolveBgFallbacks fills unset (BgNone) states using the fallback chain.
func ResolveBgFallbacks(p *BackgroundProperty) {
	for s := core.ComponentState(0); s < core.StateCount; s++ {
		if p[s].Type != core.BgNone {
			continue
		}
		for _, fb := range core.StateFallbacks[s] {
			if p[fb].Type != core.BgNone {
				p[s] = p[fb]
				break
			}
		}
	}
}

// NewFloatPropUniform creates a FloatProperty where all states have the same value.
func NewFloatPropUniform(v float64) FloatProperty {
	var p FloatProperty
	for i := range p {
		p[i] = v
	}
	return p
}

// NewFloatPropStates creates a FloatProperty from explicit state values.
// Unset states (NaN) are filled via the fallback chain; remaining NaN → 0.
func NewFloatPropStates(m map[core.ComponentState]float64) FloatProperty {
	var p FloatProperty
	for i := range p {
		p[i] = math.NaN()
	}
	for s, v := range m {
		p[s] = v
	}
	ResolveFloatFallbacks(&p)
	return p
}

// ResolveFloatFallbacks fills unset (NaN) states using the fallback chain.
// Any remaining NaN after fallback resolution is set to 0.
func ResolveFloatFallbacks(p *FloatProperty) {
	for s := core.ComponentState(0); s < core.StateCount; s++ {
		if !math.IsNaN(p[s]) {
			continue
		}
		resolved := false
		for _, fb := range core.StateFallbacks[s] {
			if !math.IsNaN(p[fb]) {
				p[s] = p[fb]
				resolved = true
				break
			}
		}
		if !resolved {
			p[s] = 0
		}
	}
}

// ---------------------------------------------------------------------------
// Generic Config type
// ---------------------------------------------------------------------------

// Config is a generic component config with variant group support.
// Primary is required (stored by value). All other variants are optional
// (stored as pointers) and fall back to Primary if nil.
type Config[G any] struct {
	Primary  G
	Variants [VariantCount - 1]*G
}

// Group returns the group for the given variant, falling back to Primary
// if the requested variant is not defined or out of range.
func (c *Config[G]) Group(v Variant) *G {
	if v > Primary && v < VariantCount && c.Variants[v-1] != nil {
		return c.Variants[v-1]
	}
	return &c.Primary
}

// ---------------------------------------------------------------------------
// Per-component Group types
// ---------------------------------------------------------------------------

// ButtonGroup defines the visual properties for Button components.
type ButtonGroup struct {
	Background   BackgroundProperty
	TextColor    ColorProperty
	Border       ColorProperty
	BorderWidth  float64
	Padding      render.Insets
	CornerRadius float64
	OffsetX      FloatProperty
	OffsetY      FloatProperty
	TextOffsetX  FloatProperty
	TextOffsetY  FloatProperty

	FocusColor     ColorProperty
	FocusRingWidth float64
}

// LabelGroup defines the visual properties for Label components.
type LabelGroup struct {
	TextColor ColorProperty
}

// BadgeGroup defines the visual properties for Badge components.
type BadgeGroup struct {
	Background   BackgroundProperty
	TextColor    ColorProperty
	CornerRadius float64 // -1 = pill
	Padding      render.Insets
	DotSize      float64 // diameter in dot mode, default 8
}

// TagGroup defines the visual properties for Tag components.
type TagGroup struct {
	Background         BackgroundProperty
	SelectedBackground BackgroundProperty
	TextColor          ColorProperty
	SelectedTextColor  ColorProperty
	BorderColor        ColorProperty
	BorderWidth        float64
	CornerRadius       float64 // -1 = pill
	Padding            render.Insets
	RemoveButtonSize   float64 // diameter of the × click area
	RemoveButtonColor  ColorProperty
	Gap                float64 // space between text and × button
}

// TagBarGroup defines the visual properties for TagBar components.
type TagBarGroup struct {
	Background     BackgroundProperty
	Border         ColorProperty
	BorderWidth    float64
	CornerRadius   float64
	Padding        render.Insets
	Spacing        float64 // gap between tags and between tags and input
	FocusColor     ColorProperty
	FocusRingWidth float64
}

// ToggleGroup defines the visual properties for Toggle components.
type ToggleGroup struct {
	TrackColor   ColorProperty
	ThumbColor   ColorProperty
	CornerRadius float64

	FocusColor     ColorProperty
	FocusRingWidth float64
}

// CheckboxGroup defines the visual properties for Checkbox components.
type CheckboxGroup struct {
	BoxColor   ColorProperty
	CheckColor ColorProperty

	CheckIcon SpriteRef // theme icon for the check mark

	FocusColor     ColorProperty
	FocusRingWidth float64
}

// RadioGroup defines the visual properties for RadioButton components.
type RadioGroup struct {
	CircleColor  ColorProperty
	DotColor     ColorProperty
	CornerRadius float64 // -1 = auto (50% of size = circle)

	DotIcon SpriteRef // theme icon for the radio dot

	FocusColor     ColorProperty
	FocusRingWidth float64
}

// TextInputGroup defines the visual properties for TextInput components.
type TextInputGroup struct {
	Background       BackgroundProperty
	TextColor        ColorProperty
	CursorColor      ColorProperty
	SelectionColor   ColorProperty
	Border           ColorProperty
	BorderWidth      float64
	CornerRadius     float64
	PlaceholderAlpha float64
	Padding          render.Insets

	FocusColor       ColorProperty
	FocusRingWidth   float64
	PasswordDotColor ColorProperty
}

// MaskedInputGroup defines the visual properties for MaskedInput components.
type MaskedInputGroup struct {
	Background       BackgroundProperty
	TextColor        ColorProperty
	CursorColor      ColorProperty
	SelectionColor   ColorProperty
	Border           ColorProperty
	BorderWidth      float64
	CornerRadius     float64
	PlaceholderAlpha float64
	Padding          render.Insets

	// LiteralColor is used to render fixed separator characters in the mask.
	LiteralColor ColorProperty
	// MaskPlaceholderColor is used to render the placeholder character in empty slots.
	MaskPlaceholderColor ColorProperty

	// SlotPadding is the horizontal padding added to each side of a slot cell,
	// creating visual spacing between adjacent characters. Default 3.
	SlotPadding float64

	FocusColor     ColorProperty
	FocusRingWidth float64
}

// InputFieldGroup defines the visual properties for InputField components.
type InputFieldGroup struct {
	LabelColor    ColorProperty
	RequiredColor ColorProperty
	ErrorColor    ColorProperty
	WarningColor  ColorProperty
	SuccessColor  ColorProperty
	LabelGap      float64 // space between label and input, default 4
	MessageGap    float64 // space between input and message, default 3
	LabelLeftGap  float64 // gap between label and input in left-label mode, default 8
}

// SearchBoxGroup defines the visual properties for SearchBox components.
// Defaults fall back to TextInput values when not set.
type SearchBoxGroup struct {
	Background       BackgroundProperty
	TextColor        ColorProperty
	CursorColor      ColorProperty
	SelectionColor   ColorProperty
	Border           ColorProperty
	BorderWidth      float64
	CornerRadius     float64
	PlaceholderAlpha float64
	Padding          render.Insets

	IconColor        ColorProperty
	ClearButtonColor ColorProperty
	ClearHoverColor  ColorProperty
	ClearActiveColor ColorProperty
	IconGap          float64

	FocusColor     ColorProperty
	FocusRingWidth float64
}

// TextAreaGroup defines the visual properties for TextArea components.
type TextAreaGroup struct {
	Background       BackgroundProperty
	TextColor        ColorProperty
	CursorColor      ColorProperty
	SelectionColor   ColorProperty
	Border           ColorProperty
	BorderWidth      float64
	CornerRadius     float64
	PlaceholderAlpha float64
	Padding          render.Insets

	FocusColor     ColorProperty
	FocusRingWidth float64
}

// SliderGroup defines the visual properties for Slider components.
type SliderGroup struct {
	Background        BackgroundProperty
	Border            ColorProperty
	BorderWidth       float64
	CornerRadius      float64
	ThumbBackground   BackgroundProperty
	ThumbBorder       ColorProperty
	ThumbBorderWidth  float64
	ThumbCornerRadius float64
	// ThumbSize overrides the cross-axis thumb dimension (height for horizontal
	// sliders, width for vertical). When > 0 the thumb is centered on the track
	// and can overflow it. Combine with ThumbCornerRadius -1 for a circle.
	ThumbSize float64
	// ThumbLength overrides the along-track thumb dimension (width for horizontal,
	// height for vertical). When 0, ThumbSize is used for both axes (square/circle).
	// Set ThumbLength < ThumbSize for a tall thin pill thumb (macOS style).
	ThumbLength float64

	FocusColor     ColorProperty
	FocusRingWidth float64
}

// ScrollBarGroup defines the visual properties for ScrollBar components.
type ScrollBarGroup struct {
	Background        BackgroundProperty
	Border            ColorProperty
	BorderWidth       float64
	CornerRadius      float64
	ThumbBackground   BackgroundProperty
	ThumbBorder       ColorProperty
	ThumbBorderWidth  float64
	ThumbCornerRadius float64

	ArrowUpIcon   SpriteRef // theme icon for scroll-up arrow
	ArrowDownIcon SpriteRef // theme icon for scroll-down arrow
}

// MeterBarGroup defines the visual properties for MeterBar components.
type MeterBarGroup struct {
	Background       BackgroundProperty
	Border           ColorProperty
	BorderWidth      float64
	CornerRadius     float64
	FillBackground   BackgroundProperty
	FillBorder       ColorProperty
	FillBorderWidth  float64
	FillCornerRadius float64
	TextColor        ColorProperty
}

// PanelGroup defines the visual properties for Panel components.
type PanelGroup struct {
	Background   BackgroundProperty
	Border       ColorProperty
	BorderWidth  float64
	CornerRadius float64
	Padding      render.Insets // default content padding for VBox/HBox child layout
}

// NavDrawerGroup defines the visual properties for NavDrawer components.
type NavDrawerGroup struct {
	Background        BackgroundProperty
	BorderColor       ColorProperty
	BorderWidth       float64
	Padding           render.Insets
	BackdropColor     ColorProperty // semi-transparent overlay color
	AnimationDuration float64       // seconds, default 0.25
}

// WindowGroup defines the visual properties for Window components.
type WindowGroup struct {
	Background               BackgroundProperty
	TitleBackground          BackgroundProperty
	TitleTextColor           ColorProperty
	ResizeHandleColor        ColorProperty
	Border                   ColorProperty
	BorderWidth              float64
	CornerRadius             float64
	ContentPaneUnderTitleBar bool // body background extends behind the title bar

	CloseIcon  SpriteRef // theme icon for the close button
	ResizeIcon SpriteRef // theme icon for the resize handle
}

// TabsGroup defines the visual properties for TabBar components.
type TabsGroup struct {
	BarBackground           BackgroundProperty
	SelectedTabColor        ColorProperty
	UnselectedTabColor      ColorProperty
	SelectedTabBackground   BackgroundProperty
	UnselectedTabBackground BackgroundProperty
	ScrollArrowBackground   BackgroundProperty
	ScrollArrowColor        ColorProperty
	ScrollArrowWidth        float64
}

// ListGroup defines the visual properties for List components.
type ListGroup struct {
	Background     BackgroundProperty
	ItemBackground BackgroundProperty
	Border         ColorProperty

	ItemPadding render.Insets // padding applied around each rendered item (left keeps text off the edge)

	FocusColor     ColorProperty
	FocusRingWidth float64
}

// TreeListGroup defines the visual properties for TreeList components.
type TreeListGroup struct {
	Background     BackgroundProperty
	ItemBackground BackgroundProperty

	ExpandIcon   SpriteRef // theme icon for collapsed tree node toggle
	CollapseIcon SpriteRef // theme icon for expanded tree node toggle

	FocusColor     ColorProperty
	FocusRingWidth float64
}

// TileListGroup defines the visual properties for TileList components.
type TileListGroup struct {
	Background     BackgroundProperty
	ItemBackground BackgroundProperty

	FocusColor     ColorProperty
	FocusRingWidth float64
}

// RichTextGroup defines the visual properties for RichText components.
type RichTextGroup struct {
	TextColor ColorProperty
}

// ShadowConfig describes a drop shadow for tooltip components.
type ShadowConfig struct {
	OffsetX float64
	OffsetY float64
	Blur    float64
	Color   sg.Color
}

// MenuBarGroup defines the visual properties for MenuBar components.
type MenuBarGroup struct {
	Background      BackgroundProperty // bar background
	EntryTextColor  ColorProperty      // entry label color (default, hover, active, disabled)
	EntryBackground BackgroundProperty // entry background on hover/active
	EntryPadding    render.Insets      // padding inside each entry label
	Spacing         float64            // gap between entries
	Height          float64            // bar height
	BorderColor     ColorProperty      // bottom border of the bar
	BorderWidth     float64            // bottom border width
}

// MenuPopupGroup defines the visual properties for MenuPopup components.
type MenuPopupGroup struct {
	Background     BackgroundProperty
	ItemBackground BackgroundProperty // hover/highlighted item background
	TextColor      ColorProperty
	DisabledColor  ColorProperty
	SeparatorColor ColorProperty
	Border         ColorProperty
	BorderWidth    float64
	CornerRadius   float64
	Padding        render.Insets
	ItemPadding    render.Insets
	ItemHeight     float64
	MaxHeight      float64            // max visible height before scroll; 0 = default (280)
	SelectedColor  BackgroundProperty // background of the currently-selected item
}

// DragHandleGroup defines the visual properties for DragHandle components.
type DragHandleGroup struct {
	Background      BackgroundProperty
	GripColor       ColorProperty
	GripHoverColor  ColorProperty
	GripActiveColor ColorProperty
	GripDotSize     float64
	GripSpacing     float64
	GripCount       int
}

// SelectGroup defines the visual properties for Select components.
type SelectGroup struct {
	Background     BackgroundProperty
	TextColor      ColorProperty
	Border         ColorProperty
	BorderWidth    float64
	CornerRadius   float64
	Padding        render.Insets
	ChevronColor   ColorProperty
	FocusColor     ColorProperty
	FocusRingWidth float64
}

// TooltipGroup defines the visual properties for Tooltip components.
type TooltipGroup struct {
	Background   BackgroundProperty
	BorderColor  ColorProperty
	BorderWidth  float64
	CornerRadius float64
	Padding      render.Insets
	MaxWidth     float64
	Shadow       ShadowConfig
}

// ToastGroup defines the visual properties for Toast notification components.
type ToastGroup struct {
	Background   BackgroundProperty
	TextColor    ColorProperty
	BorderColor  ColorProperty
	BorderWidth  float64
	CornerRadius float64
	Padding      render.Insets
	IconColor    ColorProperty

	ProgressBarColor ColorProperty
	MinWidth         float64 // default 200
	MaxWidth         float64 // default 360
	ItemSpacing      float64 // gap between stacked toasts, default 6
	AnimDuration     float64 // tween duration in seconds, default 0.2
}

// OptionRotatorChevronGroup defines the visual properties for the chevron
// buttons inside an OptionRotator.
type OptionRotatorChevronGroup struct {
	Background   BackgroundProperty
	Border       ColorProperty
	BorderWidth  float64
	CornerRadius float64
	Width        float64 // fixed pixel width of each chevron hit area
	IconColor    ColorProperty
	IconSize     float64 // scale multiplier for the procedural glyph
}

// OptionRotatorGroup defines the visual properties for OptionRotator components.
type OptionRotatorGroup struct {
	Background   BackgroundProperty
	Border       ColorProperty
	BorderWidth  float64
	CornerRadius float64
	Padding      render.Insets
	TextColor    ColorProperty
	Chevron      OptionRotatorChevronGroup

	ChevronLeftIcon  SpriteRef // theme icon for left chevron
	ChevronRightIcon SpriteRef // theme icon for right chevron

	FocusColor     ColorProperty
	FocusRingWidth float64
}

// ToggleButtonBarGroup defines the visual properties for ToggleButtonBar components.
type ToggleButtonBarGroup struct {
	Background             BackgroundProperty
	Border                 ColorProperty
	BorderWidth            float64
	CornerRadius           float64
	Padding                render.Insets
	Spacing                float64
	SelectedBackground     BackgroundProperty
	SelectedTextColor      ColorProperty
	SelectedBorder         ColorProperty
	SelectedBorderWidth    float64
	SelectedCornerRadius   float64
	UnselectedBackground   BackgroundProperty
	UnselectedTextColor    ColorProperty
	UnselectedBorder       ColorProperty
	UnselectedBorderWidth  float64
	UnselectedCornerRadius float64

	FocusColor     ColorProperty
	FocusRingWidth float64
}

// ImageGroup defines the visual properties for Image components.
type ImageGroup struct {
	Background   BackgroundProperty
	CornerRadius float64
}

// AnimatedImageGroup defines the visual properties for AnimatedImage components.
type AnimatedImageGroup struct {
	Background   BackgroundProperty
	CornerRadius float64
}

// SortableListGroup defines the visual properties for SortableList components.
type SortableListGroup struct {
	Background           BackgroundProperty
	ItemBackground       BackgroundProperty
	ItemBorderColor      ColorProperty
	ItemBorderWidth      float64
	ItemCornerRadius     float64
	ItemPadding          render.Insets
	SelectionColor       ColorProperty
	BorderColor          ColorProperty
	BorderWidth          float64
	HandleColor          ColorProperty
	HandleHoverColor     ColorProperty
	HandleActiveColor    ColorProperty
	HandleWidth          float64
	HandleGap            float64
	InsertIndicatorColor ColorProperty
	InsertIndicatorWidth float64
	FocusColor           ColorProperty
	FocusRingWidth       float64
}

// SortableTreeListGroup defines the visual properties for SortableTreeList components.
type SortableTreeListGroup struct {
	Background  BackgroundProperty
	BorderColor ColorProperty
	BorderWidth float64

	RowBackground BackgroundProperty
	RowHoverBg    BackgroundProperty
	RowSelectedBg BackgroundProperty
	RowHeight     float64
	RowPadding    render.Insets
	IndentWidth   float64

	LabelColor ColorProperty
	IconSize   float64
	IconGap    float64

	ChevronColor ColorProperty
	ChevronSize  float64

	DropLineColor         ColorProperty
	DropLineWidth         float64
	DropTargetBg          BackgroundProperty
	DropTargetBorderColor ColorProperty

	FocusColor     ColorProperty
	FocusRingWidth float64
}

// IconButtonGroup defines the visual properties for IconButton components.
type IconButtonGroup struct {
	Background   BackgroundProperty
	BorderColor  ColorProperty
	BorderWidth  float64
	CornerRadius float64

	IconColor   ColorProperty      // tints the icon sprite per state
	ActiveColor BackgroundProperty // background when Active == true
	LabelColor  ColorProperty

	IconSize float64 // default icon width = height
	Padding  render.Insets
	LabelGap float64 // gap between icon and label, default 4

	FocusColor     ColorProperty
	FocusRingWidth float64
}

// AccordionGroup defines the visual properties for Accordion components.
type AccordionGroup struct {
	Background  BackgroundProperty
	BorderColor ColorProperty
	BorderWidth float64

	HeaderBackground BackgroundProperty
	HeaderHoverBg    BackgroundProperty
	HeaderTextColor  ColorProperty
	HeaderHeight     float64
	HeaderPadding    render.Insets
	HeaderIconSize   float64
	HeaderIconGap    float64

	ChevronColor ColorProperty
	ChevronSize  float64
	ExpandIcon   SpriteRef // theme icon for collapsed section (default: built-in chevron)
	CollapseIcon SpriteRef // theme icon for expanded section (default: built-in chevron)

	ContentBackground BackgroundProperty
	ContentPadding    render.Insets

	DividerColor  ColorProperty
	DividerHeight float64

	AnimationDuration float64 // seconds, default 0.2
	CornerRadius      float64
}

// StatWebGroup defines the visual properties for StatWeb (spider/radar chart) components.
type StatWebGroup struct {
	Background BackgroundProperty

	PolygonFill        ColorProperty // inner area color (semi-transparent)
	PolygonStroke      ColorProperty // outline color
	PolygonStrokeWidth float64

	SpokeColor ColorProperty
	SpokeWidth float64

	GridColor  ColorProperty // concentric polygon grid lines
	GridLevels int           // number of concentric rings, default 4

	HandleColor      ColorProperty
	HandleHoverColor ColorProperty
	HandleRadius     float64

	LabelColor    ColorProperty
	LabelFontSize float64
	LabelOffset   float64 // distance from spoke tip to label center
}

// GradientEditorGroup defines the visual properties for GradientEditor components.
type GradientEditorGroup struct {
	Background   BackgroundProperty
	BorderColor  ColorProperty
	BorderWidth  float64
	CornerRadius float64
	Padding      render.Insets

	PreviewHeight float64 // height of the preview bar in H/V mode, default 40
	PreviewSize   float64 // side length of the preview square in 4-corner mode, default 140
}

// PopoverGroup defines the visual properties for Popover components.
type PopoverGroup struct {
	Background   BackgroundProperty
	BorderColor  ColorProperty
	BorderWidth  float64
	CornerRadius float64
	Padding      render.Insets
	TitleColor   ColorProperty
}

// TreeTableGroup defines the visual properties for TreeTable components.
type TreeTableGroup struct {
	Background     BackgroundProperty
	Border         ColorProperty
	BorderWidth    float64
	HeaderBg       BackgroundProperty
	HeaderText     ColorProperty
	RowBg          BackgroundProperty
	RowAltBg       BackgroundProperty
	RowSelectedBg  ColorProperty
	CellText       ColorProperty
	ChevronColor   ColorProperty
	DividerColor   ColorProperty
	SortIndicator  ColorProperty
	CornerRadius   float64
}

// DataTableGroup defines the visual properties for DataTable components.
type DataTableGroup struct {
	Background        BackgroundProperty
	HeaderBackground  BackgroundProperty
	HeaderText        ColorProperty
	HeaderHoverColor  ColorProperty
	HeaderBorderColor ColorProperty
	HeaderBorderWidth float64

	RowBackground    BackgroundProperty
	RowBackgroundAlt BackgroundProperty
	RowHoverColor    ColorProperty
	SelectionColor   ColorProperty

	CellText    ColorProperty
	CellPadding float64

	DividerColor ColorProperty
	DividerWidth float64

	SortGlyphColor    ColorProperty
	SortGlyphInactive ColorProperty

	BorderColor  ColorProperty
	BorderWidth  float64
	CornerRadius float64
}

// ColorPickerGroup defines the visual properties for ColorPicker components.
type ColorPickerGroup struct {
	Background         BackgroundProperty
	BorderColor        ColorProperty
	BorderWidth        float64
	CornerRadius       float64
	Padding            render.Insets
	SwatchBorderColor  ColorProperty
	SwatchBorderWidth  float64
	SwatchCornerRadius float64
	PopupWidth         float64 // default 280
	SVFieldSize        float64 // default 200 (square)
	HueBarHeight       float64 // default 14
	AlphaBarHeight     float64 // default 14
}

// KeybindInputGroup defines the visual properties for KeybindInput components.
type KeybindInputGroup struct {
	Background   BackgroundProperty
	BorderColor  ColorProperty
	BorderWidth  float64
	CornerRadius float64
	Padding      render.Insets

	KeyCapBackground   BackgroundProperty
	KeyCapTextColor    ColorProperty
	KeyCapBorderColor  ColorProperty
	KeyCapBorderWidth  float64
	KeyCapCornerRadius float64
	KeyCapPadding      render.Insets

	ListeningBackground  BackgroundProperty
	ListeningTextColor   ColorProperty
	ListeningBorderColor ColorProperty

	ClearButtonColor ColorProperty
	ClearButtonSize  float64

	UnsetTextColor ColorProperty
	UnsetText      string // default "---"
	ListeningText  string // default "Press any key..."
}

// ImageCropperGroup defines the visual properties for ImageCropper components.
type ImageCropperGroup struct {
	Background   BackgroundProperty
	CornerRadius float64

	CropBorderColor ColorProperty
	CropBorderWidth float64

	HandleBackground   BackgroundProperty
	HandleSize         float64
	HandleCornerRadius float64

	DimColor ColorProperty

	GridColor     ColorProperty
	GridLineWidth float64
}

// TimePickerGroup defines the visual properties for TimePicker components.
type TimePickerGroup struct {
	Background   BackgroundProperty
	BorderColor  ColorProperty
	BorderWidth  float64
	CornerRadius float64
	Padding      render.Insets

	ColumnBackground BackgroundProperty
	ColumnWidth      float64
	ColumnHeight     float64

	ValueTextColor ColorProperty
	ValueFontSize  float64

	ArrowColor      ColorProperty
	ArrowHoverColor ColorProperty
	ArrowSize       float64

	SeparatorColor ColorProperty
	SeparatorWidth float64

	AmPmBackground   BackgroundProperty
	AmPmActiveColor  ColorProperty
	AmPmTextColor    ColorProperty
	AmPmCornerRadius float64
}

// ToolBarGroup defines the visual properties for ToolBar components.
type ToolBarGroup struct {
	Background         BackgroundProperty
	BorderColor        ColorProperty
	BorderWidth        float64
	CornerRadius       float64
	Padding            render.Insets
	Spacing            float64       // gap between items, default 4
	SeparatorColor     ColorProperty // divider line color
	SeparatorThickness float64       // divider line width, default 1
	SeparatorHeight    float64       // fraction of toolbar height, default 0.6
}

// CalendarSelectorGroup defines the visual properties for CalendarSelector components.
type CalendarSelectorGroup struct {
	Background   BackgroundProperty
	BorderColor  ColorProperty
	BorderWidth  float64
	CornerRadius float64
	Padding      render.Insets

	HeaderBackground    BackgroundProperty
	HeaderTextColor     ColorProperty
	NavButtonColor      ColorProperty
	NavButtonHoverColor ColorProperty

	WeekdayTextColor  ColorProperty
	WeekdayBackground BackgroundProperty

	DayTextColor     ColorProperty
	DayBackground    BackgroundProperty
	DayHoverBg       BackgroundProperty
	DaySelectedBg    BackgroundProperty
	DaySelectedColor ColorProperty
	DayTodayBg       BackgroundProperty
	DayTodayColor    ColorProperty
	DayMutedColor    ColorProperty
	DaySize          float64
	DayCornerRadius  float64

	TriggerBackground  BackgroundProperty
	TriggerTextColor   ColorProperty
	TriggerBorderColor ColorProperty
}

// RichTextEditorGroup defines the visual properties for RichTextEditor components.
type RichTextEditorGroup struct {
	Background   BackgroundProperty
	BorderColor  ColorProperty
	BorderWidth  float64
	CornerRadius float64
	Padding      render.Insets

	ToolbarBackground  BackgroundProperty
	ToolbarHeight      float64
	ToolbarPadding     render.Insets
	ToolbarButtonSize  float64
	ToolbarButtonGap   float64
	ToolbarActiveColor ColorProperty

	ContentBackground BackgroundProperty
	ContentPadding    render.Insets

	CursorColor    ColorProperty
	SelectionColor ColorProperty
}

// PropertyInspectorGroup defines the visual properties for PropertyInspector.
type PropertyInspectorGroup struct {
	Background   BackgroundProperty
	BorderColor  ColorProperty
	BorderWidth  float64
	CornerRadius float64

	SearchBarHeight float64
	SearchBarGap    float64

	GroupHeaderBackground BackgroundProperty
	GroupHeaderTextColor  ColorProperty
	GroupHeaderHeight     float64

	RowBackground      BackgroundProperty
	RowAltBackground   BackgroundProperty
	RowHoverBackground BackgroundProperty
	RowHeight          float64

	LabelColor   ColorProperty
	LabelWidth   float64
	DividerColor ColorProperty
}

// ---------------------------------------------------------------------------
// Config type aliases
// ---------------------------------------------------------------------------

type ButtonConfig = Config[ButtonGroup]
type LabelConfig = Config[LabelGroup]
type BadgeConfig = Config[BadgeGroup]
type ToggleConfig = Config[ToggleGroup]
type CheckboxConfig = Config[CheckboxGroup]
type RadioConfig = Config[RadioGroup]
type TextInputConfig = Config[TextInputGroup]
type MaskedInputConfig = Config[MaskedInputGroup]
type InputFieldConfig = Config[InputFieldGroup]
type SearchBoxConfig = Config[SearchBoxGroup]
type TextAreaConfig = Config[TextAreaGroup]
type SliderConfig = Config[SliderGroup]
type ScrollBarConfig = Config[ScrollBarGroup]
type MeterBarConfig = Config[MeterBarGroup]
type PanelConfig = Config[PanelGroup]
type NavDrawerConfig = Config[NavDrawerGroup]
type WindowConfig = Config[WindowGroup]
type TabsConfig = Config[TabsGroup]
type ListConfig = Config[ListGroup]
type TreeListConfig = Config[TreeListGroup]
type TileListConfig = Config[TileListGroup]
type RichTextConfig = Config[RichTextGroup]
type OptionRotatorConfig = Config[OptionRotatorGroup]
type ToggleButtonBarConfig = Config[ToggleButtonBarGroup]
type MenuBarConfig = Config[MenuBarGroup]
type MenuPopupConfig = Config[MenuPopupGroup]
type DragHandleConfig = Config[DragHandleGroup]
type ImageConfig = Config[ImageGroup]
type AnimatedImageConfig = Config[AnimatedImageGroup]
type GradientEditorConfig = Config[GradientEditorGroup]
type ColorPickerConfig = Config[ColorPickerGroup]
type SelectConfig = Config[SelectGroup]
type TooltipConfig = Config[TooltipGroup]
type ToastConfig = Config[ToastGroup]
type SortableListConfig = Config[SortableListGroup]
type SortableTreeListConfig = Config[SortableTreeListGroup]
type IconButtonConfig = Config[IconButtonGroup]
type StatWebConfig = Config[StatWebGroup]
type AccordionConfig = Config[AccordionGroup]
type TagConfig = Config[TagGroup]
type TagBarConfig = Config[TagBarGroup]
type PopoverConfig = Config[PopoverGroup]
type TreeTableConfig = Config[TreeTableGroup]
type DataTableConfig = Config[DataTableGroup]
type KeybindInputConfig = Config[KeybindInputGroup]
type TimePickerConfig = Config[TimePickerGroup]
type ImageCropperConfig = Config[ImageCropperGroup]
type ToolBarConfig = Config[ToolBarGroup]
type CalendarSelectorConfig = Config[CalendarSelectorGroup]
type RichTextEditorConfig = Config[RichTextEditorGroup]
type PropertyInspectorConfig = Config[PropertyInspectorGroup]

// ---------------------------------------------------------------------------
// Theme
// ---------------------------------------------------------------------------

// Theme holds the complete visual configuration for all WillowUI components.
type Theme struct {
	// Per-component configs.
	Button            ButtonConfig
	Label             LabelConfig
	Badge             BadgeConfig
	Toggle            ToggleConfig
	Checkbox          CheckboxConfig
	Radio             RadioConfig
	TextInput         TextInputConfig
	MaskedInput       MaskedInputConfig
	InputField        InputFieldConfig
	SearchBox         SearchBoxConfig
	TextArea          TextAreaConfig
	Slider            SliderConfig
	ScrollBar         ScrollBarConfig
	MeterBar          MeterBarConfig
	Panel             PanelConfig
	NavDrawer         NavDrawerConfig
	Window            WindowConfig
	Tabs              TabsConfig
	List              ListConfig
	TreeList          TreeListConfig
	TileList          TileListConfig
	RichText          RichTextConfig
	OptionRotator     OptionRotatorConfig
	ToggleButtonBar   ToggleButtonBarConfig
	Tooltip           TooltipConfig
	MenuBar           MenuBarConfig
	MenuPopup         MenuPopupConfig
	Select            SelectConfig
	DragHandle        DragHandleConfig
	Image             ImageConfig
	AnimatedImage     AnimatedImageConfig
	ColorPicker       ColorPickerConfig
	GradientEditor    GradientEditorConfig
	Toast             ToastConfig
	SortableList      SortableListConfig
	SortableTreeList  SortableTreeListConfig
	IconButton        IconButtonConfig
	StatWeb           StatWebConfig
	Accordion         AccordionConfig
	Tag               TagConfig
	TagBar            TagBarConfig
	Popover           PopoverConfig
	TreeTable         TreeTableConfig
	DataTable         DataTableConfig
	KeybindInput      KeybindInputConfig
	TimePicker        TimePickerConfig
	ImageCropper      ImageCropperConfig
	ToolBar           ToolBarConfig
	CalendarSelector  CalendarSelectorConfig
	RichTextEditor    RichTextEditorConfig
	PropertyInspector PropertyInspectorConfig

	// Atlas holds the packed texture atlas for nine-slice images.
	// nil if no nine-slice images are used.
	Atlas *sg.Atlas

	// Sprites maps sprite key names (from the theme JSON "sprites" section)
	// to their resolved SpriteRef values. Use GetSprite to look up by key.
	// Nil if no sprites were declared.
	Sprites map[string]SpriteRef

	// CustomVariants maps user-defined variant names (declared in the JSON
	// "variants" array) to their assigned Variant slot (Custom1..Custom56).
	// Nil if no custom names were declared.
	CustomVariants map[string]Variant

	// UserComponents holds configurations for user-defined component types.
	// Keys are the component names as they appear in the JSON "components" map.
	// Nil if no user-defined components were declared.
	UserComponents map[string]*UserConfig

	// Fonts maps theme font role names (e.g. "body", "heading") to registered
	// font family names. These are string-only references resolved at runtime
	// via RegisterFontFamily — no font data is stored in the theme.
	// Nil if no fonts were declared.
	Fonts map[string]string
}

// GetSprite returns the SpriteRef for the given key name, or an empty
// SpriteRef with Set=false if the key is not found.
func (t *Theme) GetSprite(key string) SpriteRef {
	if t.Sprites != nil {
		if sr, ok := t.Sprites[key]; ok {
			return sr
		}
	}
	return SpriteRef{}
}

// Variant looks up a variant by name. Supports both built-in names
// ("primary", "secondary", "accent", etc.) and user-defined names
// declared in the JSON "variants" array. Returns Primary if not found.
func (t *Theme) Variant(name string) Variant {
	if v, ok := builtinVariantNames[name]; ok {
		return v
	}
	if t.CustomVariants != nil {
		if v, ok := t.CustomVariants[name]; ok {
			return v
		}
	}
	return Primary
}

// FontName returns the registered font family name for the given role,
// or empty string if the role is not defined.
func (t *Theme) FontName(role string) string {
	if t.Fonts != nil {
		return t.Fonts[role]
	}
	return ""
}

// UserComponent returns the Config for a user-defined component type, or an
// empty Config (all variants fall back to zero) if the name is not found.
func (t *Theme) UserComponent(name string) *UserConfig {
	if t.UserComponents != nil {
		if c, ok := t.UserComponents[name]; ok {
			return c
		}
	}
	return &UserConfig{}
}

// ---------------------------------------------------------------------------
// User-defined component types
// ---------------------------------------------------------------------------

// UserGroup holds the parsed visual properties for a user-defined component
// variant. Property types are inferred from JSON values during compilation.
type UserGroup struct {
	colors      map[string]ColorProperty
	backgrounds map[string]BackgroundProperty
	floats      map[string]float64
	stateFloats map[string]FloatProperty
	paddings    map[string]render.Insets
}

// Color returns the ColorProperty for the given key, or a zero value if not found.
func (g *UserGroup) Color(key string) ColorProperty {
	if g != nil {
		if p, ok := g.colors[key]; ok {
			return p
		}
	}
	return ColorProperty{}
}

// Background returns the BackgroundProperty for the given key, or a zero value if not found.
func (g *UserGroup) Background(key string) BackgroundProperty {
	if g != nil {
		if p, ok := g.backgrounds[key]; ok {
			return p
		}
	}
	return BackgroundProperty{}
}

// Float returns the scalar float for the given key, or 0 if not found.
func (g *UserGroup) Float(key string) float64 {
	if g != nil {
		if v, ok := g.floats[key]; ok {
			return v
		}
	}
	return 0
}

// StateFloat returns the per-state FloatProperty for the given key, or a zero value if not found.
func (g *UserGroup) StateFloat(key string) FloatProperty {
	if g != nil {
		if p, ok := g.stateFloats[key]; ok {
			return p
		}
	}
	return FloatProperty{}
}

// Padding returns the Insets for the given key, or a zero value if not found.
func (g *UserGroup) Padding(key string) render.Insets {
	if g != nil {
		if p, ok := g.paddings[key]; ok {
			return p
		}
	}
	return render.Insets{}
}

// UserConfig holds a user-defined component configuration with variant support.
type UserConfig = Config[UserGroup]

// ---------------------------------------------------------------------------
// Built-in variant name table (shared between theme and themecompile)
// ---------------------------------------------------------------------------

// builtinVariantNames maps JSON variant key strings to their Variant constants.
// Used during JSON compilation and by Theme.Variant().
var builtinVariantNames = map[string]Variant{
	"primary":   Primary,
	"secondary": Secondary,
	"accent":    Accent,
	"neutral":   Neutral,
	"danger":    Danger,
	"success":   Success,
	"warning":   Warning,
	"info":      Info,
	"custom1":   Custom1,
	"custom2":   Custom2,
	"custom3":   Custom3,
	"custom4":   Custom4,
	"custom5":   Custom5,
	"custom6":   Custom6,
	"custom7":   Custom7,
	"custom8":   Custom8,
	"custom9":   Custom9,
	"custom10":  Custom10,
	"custom11":  Custom11,
	"custom12":  Custom12,
	"custom13":  Custom13,
	"custom14":  Custom14,
	"custom15":  Custom15,
	"custom16":  Custom16,
	"custom17":  Custom17,
	"custom18":  Custom18,
	"custom19":  Custom19,
	"custom20":  Custom20,
	"custom21":  Custom21,
	"custom22":  Custom22,
	"custom23":  Custom23,
	"custom24":  Custom24,
	"custom25":  Custom25,
	"custom26":  Custom26,
	"custom27":  Custom27,
	"custom28":  Custom28,
	"custom29":  Custom29,
	"custom30":  Custom30,
	"custom31":  Custom31,
	"custom32":  Custom32,
	"custom33":  Custom33,
	"custom34":  Custom34,
	"custom35":  Custom35,
	"custom36":  Custom36,
	"custom37":  Custom37,
	"custom38":  Custom38,
	"custom39":  Custom39,
	"custom40":  Custom40,
	"custom41":  Custom41,
	"custom42":  Custom42,
	"custom43":  Custom43,
	"custom44":  Custom44,
	"custom45":  Custom45,
	"custom46":  Custom46,
	"custom47":  Custom47,
	"custom48":  Custom48,
	"custom49":  Custom49,
	"custom50":  Custom50,
	"custom51":  Custom51,
	"custom52":  Custom52,
	"custom53":  Custom53,
	"custom54":  Custom54,
	"custom55":  Custom55,
	"custom56":  Custom56,
}

// ---------------------------------------------------------------------------
// Default colors (used to build the default theme)
// ---------------------------------------------------------------------------

var (
	colorPrimary             = sg.RGBA(0.26, 0.52, 0.96, 1)
	colorSecondary           = sg.RGBA(0.40, 0.40, 0.45, 1)
	colorBackground          = sg.RGBA(0.15, 0.15, 0.17, 1)
	colorSurface             = sg.RGBA(0.18, 0.18, 0.21, 1)
	colorText                = sg.RGBA(0.93, 0.93, 0.93, 1)
	colorBorder              = sg.RGBA(0.30, 0.30, 0.33, 1)
	colorDisabled            = sg.RGBA(0.45, 0.45, 0.48, 0.6)
	colorHover               = sg.RGBA(0.30, 0.58, 1.00, 1)
	colorPressed             = sg.RGBA(0.20, 0.42, 0.80, 1)
	colorFocused             = sg.RGBA(0.35, 0.65, 1.00, 1)
	colorInputFocusBg        = sg.RGBA(0.13, 0.14, 0.17, 1)
	colorSelection           = sg.RGBA(0.26, 0.52, 0.96, 0.35)
	colorWindowTitleActive   = sg.RGBA(0.26, 0.52, 0.96, 1)
	colorWindowTitleInactive = sg.RGBA(0.30, 0.30, 0.33, 1)
	colorTextDisabled        = sg.RGBA(colorText.R(), colorText.G(), colorText.B(), 0.4)
)

// defaultTheme is the backing value for DefaultTheme.
var defaultTheme = Theme{
	Button: ButtonConfig{
		Primary: ButtonGroup{
			Background: NewSolidBgPropStates(map[core.ComponentState]sg.Color{
				core.StateDefault:  colorPrimary,
				core.StateHover:    colorHover,
				core.StateActive:   colorPressed,
				core.StateDisabled: colorDisabled,
			}),
			TextColor: NewColorPropStates(map[core.ComponentState]sg.Color{
				core.StateDefault:  colorText,
				core.StateDisabled: colorTextDisabled,
			}),
			BorderWidth:    1,
			Padding:        render.Insets{Top: 8, Right: 16, Bottom: 8, Left: 16},
			FocusColor:     NewColorPropStates(map[core.ComponentState]sg.Color{core.StateFocus: colorPrimary}),
			FocusRingWidth: 2,
		},
		Variants: func() [VariantCount - 1]*ButtonGroup {
			var v [VariantCount - 1]*ButtonGroup
			v[Neutral-1] = &ButtonGroup{
				Background: NewSolidBgPropStates(map[core.ComponentState]sg.Color{
					core.StateDefault:  colorSecondary,
					core.StateHover:    sg.RGBA(0.48, 0.48, 0.53, 1),
					core.StateActive:   sg.RGBA(0.35, 0.35, 0.40, 1),
					core.StateDisabled: colorDisabled,
				}),
				TextColor: NewColorPropStates(map[core.ComponentState]sg.Color{
					core.StateDefault:  colorText,
					core.StateDisabled: colorTextDisabled,
				}),
				BorderWidth:    1,
				Padding:        render.Insets{Top: 8, Right: 16, Bottom: 8, Left: 16},
				FocusColor:     NewColorPropStates(map[core.ComponentState]sg.Color{core.StateFocus: colorPrimary}),
				FocusRingWidth: 2,
			}
			return v
		}(),
	},

	Label: LabelConfig{
		Primary: LabelGroup{
			TextColor: NewColorPropUniform(colorText),
		},
	},

	Badge: BadgeConfig{
		Primary: BadgeGroup{
			Background:   NewSolidBgPropUniform(colorPrimary),
			TextColor:    NewColorPropUniform(colorText),
			CornerRadius: -1,
			Padding:      render.Insets{Top: 2, Right: 6, Bottom: 2, Left: 6},
			DotSize:      8,
		},
		Variants: [VariantCount - 1]*BadgeGroup{
			Success - 1: {
				Background:   NewSolidBgPropUniform(sg.RGBA(0.18, 0.70, 0.35, 1)),
				TextColor:    NewColorPropUniform(colorText),
				CornerRadius: -1,
				Padding:      render.Insets{Top: 2, Right: 6, Bottom: 2, Left: 6},
				DotSize:      8,
			},
			Warning - 1: {
				Background:   NewSolidBgPropUniform(sg.RGBA(0.90, 0.65, 0.10, 1)),
				TextColor:    NewColorPropUniform(colorText),
				CornerRadius: -1,
				Padding:      render.Insets{Top: 2, Right: 6, Bottom: 2, Left: 6},
				DotSize:      8,
			},
			Danger - 1: {
				Background:   NewSolidBgPropUniform(sg.RGBA(0.85, 0.20, 0.20, 1)),
				TextColor:    NewColorPropUniform(colorText),
				CornerRadius: -1,
				Padding:      render.Insets{Top: 2, Right: 6, Bottom: 2, Left: 6},
				DotSize:      8,
			},
			Neutral - 1: {
				Background:   NewSolidBgPropUniform(sg.RGBA(0.50, 0.50, 0.55, 1)),
				TextColor:    NewColorPropUniform(colorText),
				CornerRadius: -1,
				Padding:      render.Insets{Top: 2, Right: 6, Bottom: 2, Left: 6},
				DotSize:      8,
			},
		},
	},

	Toggle: ToggleConfig{
		Primary: ToggleGroup{
			TrackColor: NewColorPropStates(map[core.ComponentState]sg.Color{
				core.StateDefault:     colorSecondary,
				core.StateActive:      colorPrimary,
				core.StateFocus:       colorSecondary,
				core.StateFocusActive: colorPrimary,
				core.StateDisabled:    colorDisabled,
			}),
			ThumbColor: NewColorPropStates(map[core.ComponentState]sg.Color{
				core.StateDefault:  colorText,
				core.StateDisabled: colorTextDisabled,
			}),
			CornerRadius:   -1,
			FocusColor:     NewColorPropStates(map[core.ComponentState]sg.Color{core.StateFocus: colorFocused, core.StateFocusActive: colorFocused}),
			FocusRingWidth: 2,
		},
	},

	Checkbox: CheckboxConfig{
		Primary: CheckboxGroup{
			BoxColor: NewColorPropStates(map[core.ComponentState]sg.Color{
				core.StateDefault:     colorBorder,
				core.StateActive:      colorPrimary,
				core.StateFocus:       colorBorder,
				core.StateFocusActive: colorPrimary,
				core.StateDisabled:    colorDisabled,
			}),
			CheckColor: NewColorPropStates(map[core.ComponentState]sg.Color{
				core.StateDefault:  colorPrimary,
				core.StateDisabled: colorDisabled,
			}),
			FocusColor:     NewColorPropStates(map[core.ComponentState]sg.Color{core.StateFocus: colorFocused, core.StateFocusActive: colorFocused}),
			FocusRingWidth: 2,
		},
	},

	Radio: RadioConfig{
		Primary: RadioGroup{
			CircleColor: NewColorPropStates(map[core.ComponentState]sg.Color{
				core.StateDefault:     colorBorder,
				core.StateActive:      colorPrimary,
				core.StateFocus:       colorBorder,
				core.StateFocusActive: colorPrimary,
			}),
			DotColor:       NewColorPropUniform(colorPrimary),
			CornerRadius:   -1,
			FocusColor:     NewColorPropStates(map[core.ComponentState]sg.Color{core.StateFocus: colorFocused, core.StateFocusActive: colorFocused}),
			FocusRingWidth: 2,
		},
	},

	TextInput: TextInputConfig{
		Primary: TextInputGroup{
			Background: NewSolidBgPropStates(map[core.ComponentState]sg.Color{
				core.StateDefault:  colorBackground,
				core.StateFocus:    colorInputFocusBg,
				core.StateDisabled: colorDisabled,
			}),
			TextColor:      NewColorPropUniform(colorText),
			CursorColor:    NewColorPropUniform(colorText),
			SelectionColor: NewColorPropUniform(colorSelection),
			Border: NewColorPropStates(map[core.ComponentState]sg.Color{
				core.StateDefault:  colorBorder,
				core.StateFocus:    colorPrimary,
				core.StateDisabled: colorDisabled,
			}),
			BorderWidth:      1,
			PlaceholderAlpha: 0.4,
			Padding:          render.Insets{Top: 8, Right: 8, Bottom: 8, Left: 8},
			FocusColor:       NewColorPropStates(map[core.ComponentState]sg.Color{core.StateFocus: colorPrimary}),
			FocusRingWidth:   2,
		},
	},

	MaskedInput: MaskedInputConfig{
		Primary: MaskedInputGroup{
			Background: NewSolidBgPropStates(map[core.ComponentState]sg.Color{
				core.StateDefault:  colorBackground,
				core.StateFocus:    colorInputFocusBg,
				core.StateDisabled: colorDisabled,
			}),
			TextColor:      NewColorPropUniform(colorText),
			CursorColor:    NewColorPropUniform(colorText),
			SelectionColor: NewColorPropUniform(colorSelection),
			Border: NewColorPropStates(map[core.ComponentState]sg.Color{
				core.StateDefault:  colorBorder,
				core.StateFocus:    colorPrimary,
				core.StateDisabled: colorDisabled,
			}),
			BorderWidth:          1,
			PlaceholderAlpha:     0.4,
			Padding:              render.Insets{Top: 8, Right: 8, Bottom: 8, Left: 8},
			LiteralColor:         NewColorPropUniform(sg.RGBA(colorText.R(), colorText.G(), colorText.B(), 0.70)),
			MaskPlaceholderColor: NewColorPropUniform(colorBorder),
			SlotPadding:          3,
			FocusColor:           NewColorPropStates(map[core.ComponentState]sg.Color{core.StateFocus: colorPrimary}),
			FocusRingWidth:       2,
		},
	},

	TextArea: TextAreaConfig{
		Primary: TextAreaGroup{
			Background: NewSolidBgPropStates(map[core.ComponentState]sg.Color{
				core.StateDefault:  colorBackground,
				core.StateDisabled: colorDisabled,
			}),
			TextColor:      NewColorPropUniform(colorText),
			CursorColor:    NewColorPropUniform(colorText),
			SelectionColor: NewColorPropUniform(colorSelection),
			Border: NewColorPropStates(map[core.ComponentState]sg.Color{
				core.StateDefault:  colorBorder,
				core.StateFocus:    colorPrimary,
				core.StateDisabled: colorDisabled,
			}),
			BorderWidth:    1,
			Padding:        render.Insets{Top: 8, Right: 8, Bottom: 8, Left: 8},
			FocusColor:     NewColorPropStates(map[core.ComponentState]sg.Color{core.StateFocus: colorPrimary}),
			FocusRingWidth: 2,
		},
	},

	Slider: SliderConfig{
		Primary: SliderGroup{
			Background: NewSolidBgPropStates(map[core.ComponentState]sg.Color{
				core.StateDefault:  colorSecondary,
				core.StateDisabled: colorDisabled,
			}),
			ThumbBackground: NewSolidBgPropStates(map[core.ComponentState]sg.Color{
				core.StateDefault:  colorText,
				core.StateDisabled: colorTextDisabled,
			}),
			FocusColor:     NewColorPropStates(map[core.ComponentState]sg.Color{core.StateFocus: colorPrimary}),
			FocusRingWidth: 2,
		},
	},

	ScrollBar: ScrollBarConfig{
		Primary: ScrollBarGroup{
			Background:      NewSolidBgPropUniform(colorBackground),
			ThumbBackground: NewSolidBgPropUniform(colorSecondary),
		},
	},

	MeterBar: MeterBarConfig{
		Primary: MeterBarGroup{
			Background:     NewSolidBgPropUniform(colorSecondary),
			FillBackground: NewSolidBgPropUniform(colorPrimary),
			TextColor:      NewColorPropUniform(colorText),
		},
	},

	Panel: PanelConfig{
		Primary: PanelGroup{
			Background:  NewSolidBgPropUniform(sg.Color{}),
			BorderWidth: 1,
			Padding:     render.Insets{Left: 8, Right: 8, Top: 8, Bottom: 8},
		},
	},

	NavDrawer: NavDrawerConfig{
		Primary: NavDrawerGroup{
			Background:        NewSolidBgPropUniform(colorSurface),
			BorderColor:       NewColorPropUniform(colorBorder),
			BorderWidth:       1,
			Padding:           render.Insets{Left: 8, Right: 8, Top: 8, Bottom: 8},
			BackdropColor:     NewColorPropUniform(sg.RGBA(0, 0, 0, 0.5)),
			AnimationDuration: 0.25,
		},
	},

	Window: WindowConfig{
		Primary: WindowGroup{
			Background: NewSolidBgPropUniform(colorBackground),
			TitleBackground: NewSolidBgPropStates(map[core.ComponentState]sg.Color{
				core.StateDefault: colorWindowTitleInactive,
				core.StateFocus:   colorWindowTitleActive,
			}),
			TitleTextColor:    NewColorPropUniform(colorText),
			ResizeHandleColor: NewColorPropUniform(colorSecondary),
		},
	},

	Tabs: TabsConfig{
		Primary: TabsGroup{
			BarBackground:         NewSolidBgPropUniform(colorBackground),
			SelectedTabColor:      NewColorPropUniform(colorPrimary),
			UnselectedTabColor:    NewColorPropUniform(colorSecondary),
			ScrollArrowBackground: NewSolidBgPropUniform(colorBackground),
			ScrollArrowColor:      NewColorPropUniform(colorSecondary),
			ScrollArrowWidth:      24,
		},
	},

	List: ListConfig{
		Primary: ListGroup{
			Background:     NewSolidBgPropUniform(colorBackground),
			ItemBackground: NewSolidBgPropUniform(colorSelection),
			ItemPadding:    render.Insets{Left: 8},
			FocusColor:     NewColorPropStates(map[core.ComponentState]sg.Color{core.StateFocus: colorPrimary}),
			FocusRingWidth: 2,
		},
	},

	TreeList: TreeListConfig{
		Primary: TreeListGroup{
			Background:     NewSolidBgPropUniform(colorBackground),
			ItemBackground: NewSolidBgPropUniform(colorSelection),
			FocusColor:     NewColorPropStates(map[core.ComponentState]sg.Color{core.StateFocus: colorPrimary}),
			FocusRingWidth: 2,
		},
	},

	TileList: TileListConfig{
		Primary: TileListGroup{
			Background:     NewSolidBgPropUniform(colorBackground),
			ItemBackground: NewSolidBgPropUniform(colorSelection),
			FocusColor:     NewColorPropStates(map[core.ComponentState]sg.Color{core.StateFocus: colorPrimary}),
			FocusRingWidth: 2,
		},
	},

	RichText: RichTextConfig{
		Primary: RichTextGroup{
			TextColor: NewColorPropUniform(colorText),
		},
	},

	OptionRotator: OptionRotatorConfig{
		Primary: OptionRotatorGroup{
			Background: NewSolidBgPropStates(map[core.ComponentState]sg.Color{
				core.StateFocus: colorBackground,
			}),
			Border: NewColorPropStates(map[core.ComponentState]sg.Color{
				core.StateFocus: colorPrimary,
			}),
			BorderWidth:  1,
			CornerRadius: -1,
			Padding:      render.Insets{Top: 0, Right: 4, Bottom: 0, Left: 4},
			TextColor: NewColorPropStates(map[core.ComponentState]sg.Color{
				core.StateDefault:  colorText,
				core.StateDisabled: colorTextDisabled,
			}),
			FocusColor:     NewColorPropStates(map[core.ComponentState]sg.Color{core.StateFocus: colorPrimary}),
			FocusRingWidth: 2,
			Chevron: OptionRotatorChevronGroup{
				Width: 20,
				IconColor: NewColorPropStates(map[core.ComponentState]sg.Color{
					core.StateDefault:  sg.RGBA(colorText.R(), colorText.G(), colorText.B(), 0.55),
					core.StateHover:    colorText,
					core.StateFocus:    colorText,
					core.StateDisabled: colorTextDisabled,
				}),
				IconSize: 1.0,
			},
		},
	},

	ToggleButtonBar: ToggleButtonBarConfig{
		Primary: ToggleButtonBarGroup{
			Background:             NewSolidBgPropUniform(colorBackground),
			Border:                 NewColorPropStates(map[core.ComponentState]sg.Color{core.StateFocus: colorPrimary}),
			BorderWidth:            1,
			CornerRadius:           4,
			Padding:                render.Insets{Top: 4, Right: 4, Bottom: 4, Left: 4},
			Spacing:                2,
			SelectedBackground:     NewSolidBgPropUniform(colorPrimary),
			SelectedCornerRadius:   4,
			SelectedTextColor:      NewColorPropUniform(colorText),
			UnselectedBackground:   NewSolidBgPropUniform(colorSecondary),
			UnselectedCornerRadius: 4,
			UnselectedTextColor:    NewColorPropUniform(colorText),
		},
	},

	Tooltip: TooltipConfig{
		Primary: TooltipGroup{
			Background:   NewSolidBgPropUniform(sg.RGBA(0.20, 0.20, 0.27, 0.97)),
			BorderColor:  NewColorPropUniform(sg.RGBA(0.55, 0.55, 0.68, 1)),
			BorderWidth:  1,
			CornerRadius: 6,
			Padding:      render.Insets{Top: 8, Right: 12, Bottom: 8, Left: 12},
		},
	},

	MenuBar: MenuBarConfig{
		Primary: MenuBarGroup{
			Background: NewSolidBgPropUniform(sg.RGBA(0.15, 0.15, 0.19, 1)),
			EntryTextColor: NewColorPropStates(map[core.ComponentState]sg.Color{
				core.StateDefault:  colorText,
				core.StateHover:    sg.RGBA(0.95, 0.95, 0.97, 1),
				core.StateActive:   sg.RGBA(1, 1, 1, 1),
				core.StateDisabled: colorTextDisabled,
			}),
			EntryBackground: NewSolidBgPropStates(map[core.ComponentState]sg.Color{
				core.StateDefault: sg.RGBA(0, 0, 0, 0),
				core.StateHover:   sg.RGBA(0.22, 0.22, 0.28, 1),
				core.StateActive:  sg.RGBA(0.26, 0.26, 0.32, 1),
			}),
			EntryPadding: render.Insets{Top: 4, Right: 10, Bottom: 4, Left: 10},
			Spacing:      0,
			Height:       28,
			BorderColor:  NewColorPropUniform(colorBorder),
			BorderWidth:  1,
		},
	},

	MenuPopup: MenuPopupConfig{
		Primary: MenuPopupGroup{
			Background: NewSolidBgPropUniform(sg.RGBA(0.18, 0.18, 0.22, 0.98)),
			ItemBackground: NewSolidBgPropStates(map[core.ComponentState]sg.Color{
				core.StateDefault: sg.RGBA(0, 0, 0, 0),
				core.StateHover:   sg.RGBA(0.26, 0.52, 0.96, 0.3),
			}),
			TextColor:      NewColorPropUniform(colorText),
			DisabledColor:  NewColorPropUniform(colorTextDisabled),
			SeparatorColor: NewColorPropUniform(colorBorder),
			Border:         NewColorPropUniform(colorBorder),
			BorderWidth:    1,
			CornerRadius:   6,
			Padding:        render.Insets{Top: 4, Right: 0, Bottom: 4, Left: 0},
			ItemPadding:    render.Insets{Top: 6, Right: 12, Bottom: 6, Left: 12},
			ItemHeight:     28,
			SelectedColor:  NewSolidBgPropUniform(sg.RGBA(0.26, 0.52, 0.96, 0.18)),
		},
	},

	Select: SelectConfig{
		Primary: SelectGroup{
			Background: NewSolidBgPropStates(map[core.ComponentState]sg.Color{
				core.StateDefault:  colorBackground,
				core.StateHover:    sg.RGBA(0.18, 0.18, 0.22, 1),
				core.StateDisabled: colorDisabled,
			}),
			TextColor: NewColorPropStates(map[core.ComponentState]sg.Color{
				core.StateDefault:  colorText,
				core.StateDisabled: colorTextDisabled,
			}),
			Border: NewColorPropStates(map[core.ComponentState]sg.Color{
				core.StateDefault:  colorBorder,
				core.StateFocus:    colorPrimary,
				core.StateDisabled: colorDisabled,
			}),
			BorderWidth:    1,
			CornerRadius:   4,
			Padding:        render.Insets{Top: 6, Right: 8, Bottom: 6, Left: 8},
			ChevronColor:   NewColorPropUniform(colorText),
			FocusColor:     NewColorPropStates(map[core.ComponentState]sg.Color{core.StateFocus: colorPrimary}),
			FocusRingWidth: 2,
		},
	},

	DragHandle: DragHandleConfig{
		Primary: DragHandleGroup{
			GripColor:       NewColorPropUniform(sg.RGBA(0.42, 0.47, 0.52, 1)),
			GripHoverColor:  NewColorPropUniform(sg.RGBA(0.60, 0.65, 0.70, 1)),
			GripActiveColor: NewColorPropUniform(sg.RGBA(0.78, 0.82, 0.85, 1)),
			GripDotSize:     3,
			GripSpacing:     4,
			GripCount:       3,
		},
	},

	Image: ImageConfig{
		Primary: ImageGroup{
			Background:   NewSolidBgPropUniform(sg.RGBA(0, 0, 0, 0)),
			CornerRadius: 0,
		},
	},

	AnimatedImage: AnimatedImageConfig{
		Primary: AnimatedImageGroup{
			Background:   NewSolidBgPropUniform(sg.RGBA(0, 0, 0, 0)),
			CornerRadius: 0,
		},
	},

	ColorPicker: ColorPickerConfig{
		Primary: ColorPickerGroup{
			Background: NewSolidBgPropStates(map[core.ComponentState]sg.Color{
				core.StateDefault: sg.RGBA(0.118, 0.118, 0.165, 1),
				core.StateHover:   sg.RGBA(0.165, 0.165, 0.22, 1),
			}),
			BorderColor: NewColorPropStates(map[core.ComponentState]sg.Color{
				core.StateDefault: sg.RGBA(0.227, 0.227, 0.314, 1),
				core.StateFocus:   sg.RGBA(0.376, 0.376, 0.667, 1),
			}),
			BorderWidth:  1,
			CornerRadius: 6,
			Padding:      render.Insets{Top: 10, Right: 10, Bottom: 10, Left: 10},
			SwatchBorderColor: NewColorPropStates(map[core.ComponentState]sg.Color{
				core.StateDefault: sg.RGBA(0.227, 0.227, 0.314, 1),
				core.StateHover:   sg.RGBA(0.376, 0.376, 0.667, 1),
			}),
			SwatchBorderWidth:  1,
			SwatchCornerRadius: 3,
			PopupWidth:         440,
			SVFieldSize:        200,
			HueBarHeight:       14,
			AlphaBarHeight:     14,
		},
	},

	Toast: ToastConfig{
		Primary: ToastGroup{
			Background:       NewSolidBgPropUniform(sg.RGBA(0.145, 0.165, 0.22, 1)),
			TextColor:        NewColorPropUniform(colorText),
			BorderColor:      NewColorPropUniform(colorBorder),
			BorderWidth:      1,
			CornerRadius:     6,
			Padding:          render.Insets{Top: 10, Right: 14, Bottom: 10, Left: 14},
			IconColor:        NewColorPropUniform(colorPrimary),
			ProgressBarColor: NewColorPropUniform(sg.RGBA(0.29, 0.34, 0.50, 1)),
			MinWidth:         200,
			MaxWidth:         360,
			ItemSpacing:      6,
			AnimDuration:     0.2,
		},
		Variants: [VariantCount - 1]*ToastGroup{
			Info - 1: {
				Background:       NewSolidBgPropUniform(sg.RGBA(0.145, 0.165, 0.22, 1)),
				TextColor:        NewColorPropUniform(colorText),
				BorderColor:      NewColorPropUniform(colorBorder),
				BorderWidth:      1,
				CornerRadius:     6,
				Padding:          render.Insets{Top: 10, Right: 14, Bottom: 10, Left: 14},
				IconColor:        NewColorPropUniform(colorPrimary),
				ProgressBarColor: NewColorPropUniform(sg.RGBA(0.29, 0.34, 0.50, 1)),
				MinWidth:         200,
				MaxWidth:         360,
				ItemSpacing:      6,
				AnimDuration:     0.2,
			},
			Success - 1: {
				Background:       NewSolidBgPropUniform(sg.RGBA(0.075, 0.22, 0.10, 1)),
				TextColor:        NewColorPropUniform(colorText),
				BorderColor:      NewColorPropUniform(colorBorder),
				BorderWidth:      1,
				CornerRadius:     6,
				Padding:          render.Insets{Top: 10, Right: 14, Bottom: 10, Left: 14},
				IconColor:        NewColorPropUniform(sg.RGBA(0.29, 0.855, 0.50, 1)),
				ProgressBarColor: NewColorPropUniform(sg.RGBA(0.29, 0.34, 0.50, 1)),
				MinWidth:         200,
				MaxWidth:         360,
				ItemSpacing:      6,
				AnimDuration:     0.2,
			},
			Warning - 1: {
				Background:       NewSolidBgPropUniform(sg.RGBA(0.22, 0.18, 0.06, 1)),
				TextColor:        NewColorPropUniform(colorText),
				BorderColor:      NewColorPropUniform(colorBorder),
				BorderWidth:      1,
				CornerRadius:     6,
				Padding:          render.Insets{Top: 10, Right: 14, Bottom: 10, Left: 14},
				IconColor:        NewColorPropUniform(sg.RGBA(0.98, 0.69, 0.25, 1)),
				ProgressBarColor: NewColorPropUniform(sg.RGBA(0.29, 0.34, 0.50, 1)),
				MinWidth:         200,
				MaxWidth:         360,
				ItemSpacing:      6,
				AnimDuration:     0.2,
			},
			Danger - 1: {
				Background:       NewSolidBgPropUniform(sg.RGBA(0.22, 0.075, 0.075, 1)),
				TextColor:        NewColorPropUniform(colorText),
				BorderColor:      NewColorPropUniform(colorBorder),
				BorderWidth:      1,
				CornerRadius:     6,
				Padding:          render.Insets{Top: 10, Right: 14, Bottom: 10, Left: 14},
				IconColor:        NewColorPropUniform(sg.RGBA(0.97, 0.44, 0.44, 1)),
				ProgressBarColor: NewColorPropUniform(sg.RGBA(0.29, 0.34, 0.50, 1)),
				MinWidth:         200,
				MaxWidth:         360,
				ItemSpacing:      6,
				AnimDuration:     0.2,
			},
		},
	},

	SortableList: SortableListConfig{
		Primary: SortableListGroup{
			Background:           NewSolidBgPropUniform(colorBackground),
			ItemBackground:       NewSolidBgPropUniform(sg.Color{}),
			ItemPadding:          render.Insets{Top: 0, Right: 4, Bottom: 0, Left: 8},
			SelectionColor:       NewColorPropUniform(colorSelection),
			BorderColor:          NewColorPropUniform(colorBorder),
			BorderWidth:          1,
			HandleColor:          NewColorPropUniform(sg.RGBA(0.42, 0.47, 0.52, 1)),
			HandleHoverColor:     NewColorPropUniform(sg.RGBA(0.60, 0.65, 0.70, 1)),
			HandleActiveColor:    NewColorPropUniform(sg.RGBA(0.78, 0.82, 0.85, 1)),
			HandleWidth:          24,
			HandleGap:            4,
			InsertIndicatorColor: NewColorPropUniform(colorPrimary),
			InsertIndicatorWidth: 2,
			FocusColor:           NewColorPropStates(map[core.ComponentState]sg.Color{core.StateFocus: colorPrimary}),
			FocusRingWidth:       2,
		},
	},

	SortableTreeList: SortableTreeListConfig{
		Primary: SortableTreeListGroup{
			Background:            NewSolidBgPropUniform(colorBackground),
			BorderColor:           NewColorPropUniform(colorBorder),
			BorderWidth:           1,
			RowBackground:         NewSolidBgPropUniform(sg.Color{}),
			RowHoverBg:            NewSolidBgPropUniform(sg.RGBA(0.25, 0.27, 0.35, 1)),
			RowSelectedBg:         NewSolidBgPropUniform(colorSelection),
			RowHeight:             30,
			RowPadding:            render.Insets{Top: 0, Right: 4, Bottom: 0, Left: 4},
			IndentWidth:           20,
			LabelColor:            NewColorPropUniform(colorText),
			IconSize:              16,
			IconGap:               4,
			ChevronColor:          NewColorPropUniform(sg.RGBA(0.60, 0.65, 0.70, 1)),
			ChevronSize:           12,
			DropLineColor:         NewColorPropUniform(colorPrimary),
			DropLineWidth:         2,
			DropTargetBg:          NewSolidBgPropUniform(sg.RGBA(0.25, 0.35, 0.55, 0.5)),
			DropTargetBorderColor: NewColorPropUniform(colorPrimary),
			FocusColor:            NewColorPropStates(map[core.ComponentState]sg.Color{core.StateFocus: colorPrimary}),
			FocusRingWidth:        2,
		},
	},

	IconButton: IconButtonConfig{
		Primary: IconButtonGroup{
			Background: NewSolidBgPropStates(map[core.ComponentState]sg.Color{
				core.StateDefault:  colorSecondary,
				core.StateHover:    sg.RGBA(0.48, 0.48, 0.53, 1),
				core.StateActive:   colorPressed,
				core.StateDisabled: colorDisabled,
				// Pin focus states to default so they don't fall back to StateActive (blue).
				core.StateFocus:         colorSecondary,
				core.StateFocusHover:    sg.RGBA(0.48, 0.48, 0.53, 1),
				core.StateFocusActive:   colorPressed,
				core.StateFocusDisabled: colorDisabled,
			}),
			BorderColor: NewColorPropStates(map[core.ComponentState]sg.Color{
				core.StateDefault:       sg.RGBA(0.30, 0.30, 0.35, 1),
				core.StateHover:         sg.RGBA(0.50, 0.50, 0.55, 1),
				core.StateActive:        sg.RGBA(0.20, 0.42, 0.80, 1),
				core.StateDisabled:      sg.RGBA(0.30, 0.30, 0.35, 0.5),
				core.StateFocus:         sg.RGBA(0.30, 0.30, 0.35, 1),
				core.StateFocusHover:    sg.RGBA(0.50, 0.50, 0.55, 1),
				core.StateFocusActive:   sg.RGBA(0.20, 0.42, 0.80, 1),
				core.StateFocusDisabled: sg.RGBA(0.30, 0.30, 0.35, 0.5),
			}),
			BorderWidth:  1,
			CornerRadius: 6,
			IconColor: NewColorPropStates(map[core.ComponentState]sg.Color{
				core.StateDefault:  colorText,
				core.StateHover:    sg.RGBA(1, 1, 1, 1),
				core.StateActive:   sg.RGBA(1, 1, 1, 1),
				core.StateDisabled: colorTextDisabled,
			}),
			ActiveColor: NewSolidBgPropUniform(colorPrimary),
			LabelColor: NewColorPropStates(map[core.ComponentState]sg.Color{
				core.StateDefault:  sg.RGBA(0.75, 0.75, 0.80, 1),
				core.StateHover:    colorText,
				core.StateDisabled: colorTextDisabled,
			}),
			IconSize:       20,
			Padding:        render.Insets{Top: 6, Right: 6, Bottom: 6, Left: 6},
			LabelGap:       4,
			FocusColor:     NewColorPropStates(map[core.ComponentState]sg.Color{core.StateFocus: colorPrimary}),
			FocusRingWidth: 2,
		},
	},

	Tag: TagConfig{
		Primary: TagGroup{
			Background:         NewSolidBgPropUniform(colorSecondary),
			SelectedBackground: NewSolidBgPropUniform(colorPrimary),
			TextColor:          NewColorPropUniform(colorText),
			SelectedTextColor:  NewColorPropUniform(colorText),
			CornerRadius:       -1,
			Padding:            render.Insets{Top: 2, Right: 8, Bottom: 2, Left: 8},
			RemoveButtonSize:   16,
			RemoveButtonColor:  NewColorPropUniform(sg.RGBA(0.7, 0.7, 0.75, 1)),
			Gap:                4,
		},
		Variants: [VariantCount - 1]*TagGroup{
			Success - 1: {
				Background:         NewSolidBgPropUniform(sg.RGBA(0.18, 0.70, 0.35, 1)),
				SelectedBackground: NewSolidBgPropUniform(sg.RGBA(0.18, 0.70, 0.35, 1)),
				TextColor:          NewColorPropUniform(colorText),
				SelectedTextColor:  NewColorPropUniform(colorText),
				CornerRadius:       -1,
				Padding:            render.Insets{Top: 2, Right: 8, Bottom: 2, Left: 8},
				RemoveButtonSize:   16,
				RemoveButtonColor:  NewColorPropUniform(sg.RGBA(0.7, 0.7, 0.75, 1)),
				Gap:                4,
			},
			Warning - 1: {
				Background:         NewSolidBgPropUniform(sg.RGBA(0.90, 0.65, 0.10, 1)),
				SelectedBackground: NewSolidBgPropUniform(sg.RGBA(0.90, 0.65, 0.10, 1)),
				TextColor:          NewColorPropUniform(colorText),
				SelectedTextColor:  NewColorPropUniform(colorText),
				CornerRadius:       -1,
				Padding:            render.Insets{Top: 2, Right: 8, Bottom: 2, Left: 8},
				RemoveButtonSize:   16,
				RemoveButtonColor:  NewColorPropUniform(sg.RGBA(0.7, 0.7, 0.75, 1)),
				Gap:                4,
			},
			Danger - 1: {
				Background:         NewSolidBgPropUniform(sg.RGBA(0.85, 0.20, 0.20, 1)),
				SelectedBackground: NewSolidBgPropUniform(sg.RGBA(0.85, 0.20, 0.20, 1)),
				TextColor:          NewColorPropUniform(colorText),
				SelectedTextColor:  NewColorPropUniform(colorText),
				CornerRadius:       -1,
				Padding:            render.Insets{Top: 2, Right: 8, Bottom: 2, Left: 8},
				RemoveButtonSize:   16,
				RemoveButtonColor:  NewColorPropUniform(sg.RGBA(0.7, 0.7, 0.75, 1)),
				Gap:                4,
			},
			Neutral - 1: {
				Background:         NewSolidBgPropUniform(sg.RGBA(0.50, 0.50, 0.55, 1)),
				SelectedBackground: NewSolidBgPropUniform(sg.RGBA(0.50, 0.50, 0.55, 1)),
				TextColor:          NewColorPropUniform(colorText),
				SelectedTextColor:  NewColorPropUniform(colorText),
				CornerRadius:       -1,
				Padding:            render.Insets{Top: 2, Right: 8, Bottom: 2, Left: 8},
				RemoveButtonSize:   16,
				RemoveButtonColor:  NewColorPropUniform(sg.RGBA(0.7, 0.7, 0.75, 1)),
				Gap:                4,
			},
		},
	},

	TagBar: TagBarConfig{
		Primary: TagBarGroup{
			Background: NewSolidBgPropStates(map[core.ComponentState]sg.Color{
				core.StateDefault: colorSurface,
				core.StateHover:   colorSurface,
			}),
			Border: NewColorPropStates(map[core.ComponentState]sg.Color{
				core.StateDefault: colorBorder,
				core.StateFocus:   colorPrimary,
			}),
			BorderWidth:    1,
			CornerRadius:   4,
			Padding:        render.Insets{Top: 4, Right: 4, Bottom: 4, Left: 4},
			Spacing:        4,
			FocusColor:     NewColorPropStates(map[core.ComponentState]sg.Color{core.StateFocus: colorPrimary}),
			FocusRingWidth: 2,
		},
	},
	Accordion: AccordionConfig{
		Primary: AccordionGroup{
			HeaderBackground: NewSolidBgPropStates(map[core.ComponentState]sg.Color{
				core.StateDefault: sg.RGBA(0.12, 0.12, 0.14, 1),
			}),
			HeaderTextColor:   NewColorPropUniform(sg.RGBA(0.9, 0.9, 0.9, 1)),
			HeaderHeight:      36,
			HeaderPadding:     render.Insets{Left: 8, Right: 8},
			ChevronColor:      NewColorPropUniform(sg.RGBA(0.7, 0.7, 0.7, 1)),
			ChevronSize:       12,
			ContentPadding:    render.Insets{Top: 8, Bottom: 8, Left: 8, Right: 8},
			DividerColor:      NewColorPropUniform(sg.RGBA(0.2, 0.2, 0.22, 1)),
			DividerHeight:     1,
			AnimationDuration: 0.2,
		},
	},

	Popover: PopoverConfig{
		Primary: PopoverGroup{
			Background:   NewSolidBgPropUniform(sg.RGBA(0.18, 0.18, 0.22, 0.98)),
			BorderColor:  NewColorPropUniform(colorBorder),
			BorderWidth:  1,
			CornerRadius: 6,
			Padding:      render.Insets{Top: 8, Right: 8, Bottom: 8, Left: 8},
			TitleColor:   NewColorPropUniform(colorText),
		},
	},

	TreeTable: TreeTableConfig{
		Primary: TreeTableGroup{
			Background:    NewSolidBgPropUniform(colorBackground),
			Border:        NewColorPropUniform(colorBorder),
			BorderWidth:   1,
			HeaderBg:      NewSolidBgPropUniform(sg.RGBA(0.20, 0.20, 0.23, 1)),
			HeaderText:    NewColorPropUniform(colorText),
			RowBg:         NewSolidBgPropUniform(colorBackground),
			RowAltBg:      NewSolidBgPropUniform(sg.RGBA(0.17, 0.17, 0.20, 1)),
			RowSelectedBg: NewColorPropUniform(colorSelection),
			CellText:      NewColorPropUniform(colorText),
			ChevronColor:  NewColorPropUniform(sg.RGBA(0.60, 0.65, 0.70, 1)),
			DividerColor:  NewColorPropUniform(colorBorder),
			SortIndicator: NewColorPropUniform(colorText),
			CornerRadius:  0,
		},
	},

	DataTable: DataTableConfig{
		Primary: DataTableGroup{
			Background:        NewSolidBgPropUniform(colorBackground),
			HeaderBackground:  NewSolidBgPropUniform(sg.RGBA(0.20, 0.20, 0.23, 1)),
			HeaderText:        NewColorPropUniform(colorText),
			HeaderHoverColor:  NewColorPropUniform(sg.RGBA(0.30, 0.30, 0.34, 1)),
			HeaderBorderColor: NewColorPropUniform(colorBorder),
			HeaderBorderWidth: 1,
			RowBackground:     NewSolidBgPropUniform(colorBackground),
			RowBackgroundAlt:  NewSolidBgPropUniform(sg.RGBA(0.17, 0.17, 0.20, 1)),
			RowHoverColor:     NewColorPropUniform(sg.RGBA(0.25, 0.25, 0.29, 1)),
			SelectionColor:    NewColorPropUniform(colorSelection),
			CellText:          NewColorPropUniform(colorText),
			CellPadding:       6,
			DividerColor:      NewColorPropUniform(colorBorder),
			DividerWidth:      1,
			SortGlyphColor:    NewColorPropUniform(colorText),
			SortGlyphInactive: NewColorPropUniform(colorDisabled),
			BorderColor:       NewColorPropUniform(colorBorder),
			BorderWidth:       1,
			CornerRadius:      4,
		},
	},

	TimePicker: TimePickerConfig{
		Primary: TimePickerGroup{
			Background:       NewSolidBgPropUniform(colorBackground),
			BorderColor:      NewColorPropUniform(colorBorder),
			BorderWidth:      1,
			CornerRadius:     4,
			Padding:          render.Insets{Top: 4, Right: 4, Bottom: 4, Left: 4},
			ColumnBackground: NewSolidBgPropUniform(sg.RGBA(0.15, 0.15, 0.18, 1)),
			ColumnWidth:      40,
			ColumnHeight:     60,
			ValueTextColor:   NewColorPropUniform(colorText),
			ValueFontSize:    16,
			ArrowColor:       NewColorPropUniform(colorText),
			ArrowHoverColor:  NewColorPropUniform(colorPrimary),
			ArrowSize:        12,
			SeparatorColor:   NewColorPropUniform(colorText),
			SeparatorWidth:   2,
			AmPmBackground:   NewSolidBgPropUniform(colorSecondary),
			AmPmActiveColor:  NewColorPropUniform(colorPrimary),
			AmPmTextColor:    NewColorPropUniform(colorText),
			AmPmCornerRadius: 4,
		},
	},

	ImageCropper: ImageCropperConfig{
		Primary: ImageCropperGroup{
			Background:         NewSolidBgPropUniform(sg.RGBA(0, 0, 0, 0)),
			CornerRadius:       0,
			CropBorderColor:    NewColorPropUniform(sg.RGBA(1, 1, 1, 1)),
			CropBorderWidth:    2,
			HandleBackground:   NewSolidBgPropUniform(sg.RGBA(1, 1, 1, 1)),
			HandleSize:         12,
			HandleCornerRadius: 6,
			DimColor:           NewColorPropUniform(sg.RGBA(0, 0, 0, 0.6)),
			GridColor:          NewColorPropUniform(sg.RGBA(1, 1, 1, 0.3)),
			GridLineWidth:      1,
		},
	},

	ToolBar: ToolBarConfig{
		Primary: ToolBarGroup{
			Background:         NewSolidBgPropUniform(sg.RGBA(0.12, 0.13, 0.18, 1)),
			BorderColor:        NewColorPropUniform(colorBorder),
			BorderWidth:        1,
			CornerRadius:       0,
			Padding:            render.Insets{Top: 4, Right: 8, Bottom: 4, Left: 8},
			Spacing:            4,
			SeparatorColor:     NewColorPropUniform(sg.RGBA(0.35, 0.38, 0.48, 1)),
			SeparatorThickness: 2,
			SeparatorHeight:    0.6,
		},
	},

	CalendarSelector: CalendarSelectorConfig{
		Primary: CalendarSelectorGroup{
			Background:   NewSolidBgPropUniform(sg.RGBA(0.11, 0.12, 0.15, 1)),
			BorderColor:  NewColorPropUniform(sg.RGBA(0.28, 0.30, 0.38, 1)),
			BorderWidth:  1,
			CornerRadius: 6,
			Padding:      render.Insets{Top: 0, Right: 0, Bottom: 4, Left: 0},

			HeaderBackground:    NewSolidBgPropUniform(sg.RGBA(0.15, 0.16, 0.21, 1)),
			HeaderTextColor:     NewColorPropUniform(sg.RGBA(0.95, 0.96, 0.98, 1)),
			NavButtonColor:      NewColorPropUniform(colorText),
			NavButtonHoverColor: NewColorPropUniform(colorPrimary),

			WeekdayTextColor:  NewColorPropUniform(sg.RGBA(0.45, 0.50, 0.58, 1)),
			WeekdayBackground: NewSolidBgPropUniform(sg.RGBA(0.11, 0.12, 0.15, 1)),

			DayTextColor:     NewColorPropUniform(sg.RGBA(0.82, 0.84, 0.88, 1)),
			DayBackground:    NewSolidBgPropUniform(sg.RGBA(0.14, 0.15, 0.19, 1)),
			DayHoverBg:       NewSolidBgPropUniform(sg.RGBA(0.22, 0.24, 0.32, 1)),
			DaySelectedBg:    NewSolidBgPropUniform(colorPrimary),
			DaySelectedColor: NewColorPropUniform(sg.RGBA(1, 1, 1, 1)),
			DayTodayBg:       NewSolidBgPropUniform(sg.RGBA(0.18, 0.20, 0.28, 1)),
			DayTodayColor:    NewColorPropUniform(colorPrimary),
			DayMutedColor:    NewColorPropUniform(sg.RGBA(0.28, 0.30, 0.35, 1)),
			DaySize:          32,
			DayCornerRadius:  4,

			TriggerBackground:  NewSolidBgPropUniform(sg.RGBA(0.14, 0.15, 0.19, 1)),
			TriggerTextColor:   NewColorPropUniform(colorText),
			TriggerBorderColor: NewColorPropUniform(sg.RGBA(0.28, 0.30, 0.38, 1)),
		},
	},

	RichTextEditor: RichTextEditorConfig{
		Primary: RichTextEditorGroup{
			Background: NewSolidBgPropStates(map[core.ComponentState]sg.Color{
				core.StateDefault:  colorBackground,
				core.StateDisabled: colorDisabled,
			}),
			BorderColor: NewColorPropStates(map[core.ComponentState]sg.Color{
				core.StateDefault:  colorBorder,
				core.StateFocus:    colorPrimary,
				core.StateDisabled: colorDisabled,
			}),
			BorderWidth:  1,
			CornerRadius: 4,
			Padding:      render.Insets{Top: 0, Right: 0, Bottom: 0, Left: 0},

			ToolbarBackground:  NewSolidBgPropUniform(sg.RGBA(0.15, 0.16, 0.21, 1)),
			ToolbarHeight:      36,
			ToolbarPadding:     render.Insets{Top: 4, Right: 4, Bottom: 4, Left: 4},
			ToolbarButtonSize:  28,
			ToolbarButtonGap:   4,
			ToolbarActiveColor: NewColorPropUniform(colorPrimary),

			ContentBackground: NewSolidBgPropUniform(colorBackground),
			ContentPadding:    render.Insets{Top: 4, Right: 4, Bottom: 4, Left: 4},

			CursorColor:    NewColorPropUniform(colorText),
			SelectionColor: NewColorPropUniform(colorSelection),
		},
	},
}

// DefaultTheme is the fallback theme used when no explicit theme is set.
var DefaultTheme = &defaultTheme

// DefaultThemeRef is a pointer to the canonical DefaultTheme variable.
// It is initially set to &DefaultTheme (pointing at this package's variable),
// but the root willowui package overwrites it with &willowui.DefaultTheme at
// init time so that the widget package always reads the root's variable when
// the user reassigns willowui.DefaultTheme.
var DefaultThemeRef = &DefaultTheme
