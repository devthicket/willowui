package theme

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	_ "image/png" // register PNG decoder
	"io/fs"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strconv"
	"strings"

	"github.com/devthicket/willowui/internal/core"
	"github.com/devthicket/willowui/internal/engine"
	"github.com/devthicket/willowui/internal/markup"
	"github.com/devthicket/willowui/internal/render"
	"github.com/devthicket/willowui/internal/sg"
)

// ---------------------------------------------------------------------------
// Image loading
// ---------------------------------------------------------------------------

// imageLoader abstracts image and file loading for the compile pipeline.
type imageLoader interface {
	loadImage(path string) (image.Image, error)
	readFile(path string) ([]byte, error)
}

// nilImageLoader rejects all loads. Used by LoadTheme (raw bytes).
type nilImageLoader struct{}

func (nilImageLoader) loadImage(path string) (image.Image, error) {
	return nil, fmt.Errorf("nine-slice images require LoadThemeFromFile or LoadThemeFromFS (image: %s)", path)
}

func (nilImageLoader) readFile(path string) ([]byte, error) {
	return nil, fmt.Errorf("file reads require LoadThemeFromFile or LoadThemeFromFS (file: %s)", path)
}

// fileImageLoader loads images relative to a base directory.
type fileImageLoader struct{ baseDir string }

func (l fileImageLoader) loadImage(path string) (image.Image, error) {
	f, err := os.Open(filepath.Join(l.baseDir, path))
	if err != nil {
		return nil, err
	}
	defer f.Close()
	img, _, err := image.Decode(f)
	return img, err
}

func (l fileImageLoader) readFile(path string) ([]byte, error) {
	return os.ReadFile(filepath.Join(l.baseDir, path))
}

// fsImageLoader loads images from an fs.FS.
type fsImageLoader struct{ fsys fs.FS }

func (l fsImageLoader) loadImage(path string) (image.Image, error) {
	f, err := l.fsys.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	img, _, err := image.Decode(f)
	return img, err
}

func (l fsImageLoader) readFile(path string) ([]byte, error) {
	return fs.ReadFile(l.fsys, path)
}

// nineSliceEntry tracks a pending nine-slice that needs atlas resolution.
type nineSliceEntry struct {
	imagePath string
	slice     *render.NineSlice
}

// ---------------------------------------------------------------------------
// JSON key → ComponentState mapping
// ---------------------------------------------------------------------------

var stateNames = map[string]core.ComponentState{
	"default":       core.StateDefault,
	"hover":         core.StateHover,
	"active":        core.StateActive,
	"disabled":      core.StateDisabled,
	"focus":         core.StateFocus,
	"focusHover":    core.StateFocusHover,
	"focusActive":   core.StateFocusActive,
	"focusDisabled": core.StateFocusDisabled,
}

// lookupVariant checks built-in variant names first, then user-defined ones.
func lookupVariant(key string, userVariants map[string]Variant) (Variant, bool) {
	if v, ok := builtinVariantNames[key]; ok {
		return v, true
	}
	if userVariants != nil {
		if v, ok := userVariants[key]; ok {
			return v, true
		}
	}
	return 0, false
}

// ---------------------------------------------------------------------------
// Per-component property maps
// ---------------------------------------------------------------------------

// propertyKind describes the type of a property.
type propertyKind int

const (
	propColor      propertyKind = iota // ColorProperty
	propBackground                     // BackgroundProperty (BgSolid from color strings)
	propPadding                        // render.Insets
	propFloat                          // float64 scalar
	propStateFloat                     // FloatProperty (per-state float64)
	propGridRef                        // BackgroundProperty via nine-grid string key reference
	propBool                           // bool scalar
	propSpriteRef                      // SpriteRef via sprite name string reference
	propStr                            // plain string scalar
)

// nineGridDef holds a parsed nine-grid definition from the "nine-grids" section.
type nineGridDef struct {
	source      string
	region      *render.Rect // nil = use full image
	innerRegion render.Rect
	insets      render.Insets          // derived from slices or auto-slice
	centerFill  *render.GradientColors // optional: replaces center cell with gradient
}

type propInfo struct {
	goField string
	kind    propertyKind
}

// keyAliases maps common incorrect JSON key names to their correct forms.
// Used to emit warnings when a theme author uses a bare key instead of
// the *Color convention.
var keyAliases = map[string]string{
	"background": "backgroundColor",
	"border":     "borderColor",
}

var componentPropertyMaps = map[string]map[string]propInfo{
	"button": {
		"backgroundColor": {"Background", propBackground},
		"backgroundGrid":  {"Background", propGridRef},
		"textColor":       {"TextColor", propColor},
		"borderColor":     {"Border", propColor},
		"borderWidth":     {"BorderWidth", propFloat},
		"padding":         {"Padding", propPadding},
		"cornerRadius":    {"CornerRadius", propFloat},
		"offsetX":         {"OffsetX", propStateFloat},
		"offsetY":         {"OffsetY", propStateFloat},
		"textOffsetX":     {"TextOffsetX", propStateFloat},
		"textOffsetY":     {"TextOffsetY", propStateFloat},
		"focusColor":      {"FocusColor", propColor},
		"focusRingWidth":  {"FocusRingWidth", propFloat},
	},
	"label": {
		"textColor": {"TextColor", propColor},
	},
	"badge": {
		"backgroundColor": {"Background", propBackground},
		"backgroundGrid":  {"Background", propGridRef},
		"textColor":       {"TextColor", propColor},
		"cornerRadius":    {"CornerRadius", propFloat},
		"padding":         {"Padding", propPadding},
		"dotSize":         {"DotSize", propFloat},
	},
	"toggle": {
		"trackColor":     {"TrackColor", propColor},
		"thumbColor":     {"ThumbColor", propColor},
		"cornerRadius":   {"CornerRadius", propFloat},
		"focusColor":     {"FocusColor", propColor},
		"focusRingWidth": {"FocusRingWidth", propFloat},
	},
	"checkbox": {
		"boxColor":       {"BoxColor", propColor},
		"checkColor":     {"CheckColor", propColor},
		"checkIcon":      {"CheckIcon", propSpriteRef},
		"focusColor":     {"FocusColor", propColor},
		"focusRingWidth": {"FocusRingWidth", propFloat},
	},
	"radio": {
		"circleColor":    {"CircleColor", propColor},
		"dotColor":       {"DotColor", propColor},
		"cornerRadius":   {"CornerRadius", propFloat},
		"dotIcon":        {"DotIcon", propSpriteRef},
		"focusColor":     {"FocusColor", propColor},
		"focusRingWidth": {"FocusRingWidth", propFloat},
	},
	"textInput": {
		"backgroundColor":  {"Background", propBackground},
		"backgroundGrid":   {"Background", propGridRef},
		"textColor":        {"TextColor", propColor},
		"cursorColor":      {"CursorColor", propColor},
		"selectionColor":   {"SelectionColor", propColor},
		"borderColor":      {"Border", propColor},
		"borderWidth":      {"BorderWidth", propFloat},
		"cornerRadius":     {"CornerRadius", propFloat},
		"placeholderAlpha": {"PlaceholderAlpha", propFloat},
		"padding":          {"Padding", propPadding},
		"focusColor":       {"FocusColor", propColor},
		"focusRingWidth":   {"FocusRingWidth", propFloat},
	},
	"textArea": {
		"backgroundColor":  {"Background", propBackground},
		"backgroundGrid":   {"Background", propGridRef},
		"textColor":        {"TextColor", propColor},
		"cursorColor":      {"CursorColor", propColor},
		"selectionColor":   {"SelectionColor", propColor},
		"borderColor":      {"Border", propColor},
		"borderWidth":      {"BorderWidth", propFloat},
		"cornerRadius":     {"CornerRadius", propFloat},
		"placeholderAlpha": {"PlaceholderAlpha", propFloat},
		"padding":          {"Padding", propPadding},
		"focusColor":       {"FocusColor", propColor},
		"focusRingWidth":   {"FocusRingWidth", propFloat},
	},
	"slider": {
		"backgroundColor":      {"Background", propBackground},
		"backgroundGrid":       {"Background", propGridRef},
		"borderColor":          {"Border", propColor},
		"borderWidth":          {"BorderWidth", propFloat},
		"cornerRadius":         {"CornerRadius", propFloat},
		"thumbBackgroundColor": {"ThumbBackground", propBackground},
		"thumbBackgroundGrid":  {"ThumbBackground", propGridRef},
		"thumbBorderColor":     {"ThumbBorder", propColor},
		"thumbBorderWidth":     {"ThumbBorderWidth", propFloat},
		"thumbCornerRadius":    {"ThumbCornerRadius", propFloat},
		"thumbSize":            {"ThumbSize", propFloat},
		"thumbLength":          {"ThumbLength", propFloat},
		"focusColor":           {"FocusColor", propColor},
		"focusRingWidth":       {"FocusRingWidth", propFloat},
	},
	"scrollBar": {
		"backgroundColor":      {"Background", propBackground},
		"backgroundGrid":       {"Background", propGridRef},
		"borderColor":          {"Border", propColor},
		"borderWidth":          {"BorderWidth", propFloat},
		"cornerRadius":         {"CornerRadius", propFloat},
		"thumbBackgroundColor": {"ThumbBackground", propBackground},
		"thumbBackgroundGrid":  {"ThumbBackground", propGridRef},
		"thumbBorderColor":     {"ThumbBorder", propColor},
		"thumbBorderWidth":     {"ThumbBorderWidth", propFloat},
		"thumbCornerRadius":    {"ThumbCornerRadius", propFloat},
		"arrowUpIcon":          {"ArrowUpIcon", propSpriteRef},
		"arrowDownIcon":        {"ArrowDownIcon", propSpriteRef},
	},
	"meterBar": {
		"backgroundColor":     {"Background", propBackground},
		"backgroundGrid":      {"Background", propGridRef},
		"borderColor":         {"Border", propColor},
		"borderWidth":         {"BorderWidth", propFloat},
		"cornerRadius":        {"CornerRadius", propFloat},
		"fillBackgroundColor": {"FillBackground", propBackground},
		"fillBackgroundGrid":  {"FillBackground", propGridRef},
		"fillBorderColor":     {"FillBorder", propColor},
		"fillBorderWidth":     {"FillBorderWidth", propFloat},
		"fillCornerRadius":    {"FillCornerRadius", propFloat},
		"textColor":           {"TextColor", propColor},
	},
	"panel": {
		"backgroundColor": {"Background", propBackground},
		"backgroundGrid":  {"Background", propGridRef},
		"borderColor":     {"Border", propColor},
		"borderWidth":     {"BorderWidth", propFloat},
		"cornerRadius":    {"CornerRadius", propFloat},
		"padding":         {"Padding", propPadding},
	},
	"window": {
		"backgroundColor":          {"Background", propBackground},
		"backgroundGrid":           {"Background", propGridRef},
		"titleBackgroundColor":     {"TitleBackground", propBackground},
		"titleBackgroundGrid":      {"TitleBackground", propGridRef},
		"titleTextColor":           {"TitleTextColor", propColor},
		"resizeHandleColor":        {"ResizeHandleColor", propColor},
		"borderColor":              {"Border", propColor},
		"borderWidth":              {"BorderWidth", propFloat},
		"cornerRadius":             {"CornerRadius", propFloat},
		"contentPaneUnderTitleBar": {"ContentPaneUnderTitleBar", propBool},
		"closeIcon":                {"CloseIcon", propSpriteRef},
		"resizeIcon":               {"ResizeIcon", propSpriteRef},
	},
	"tabs": {
		"barBackgroundColor":           {"BarBackground", propBackground},
		"barBackgroundGrid":            {"BarBackground", propGridRef},
		"selectedTabColor":             {"SelectedTabColor", propColor},
		"unselectedTabColor":           {"UnselectedTabColor", propColor},
		"selectedTabBackgroundColor":   {"SelectedTabBackground", propBackground},
		"unselectedTabBackgroundColor": {"UnselectedTabBackground", propBackground},
	},
	"list": {
		"backgroundColor":     {"Background", propBackground},
		"backgroundGrid":      {"Background", propGridRef},
		"itemBackgroundColor": {"ItemBackground", propBackground},
		"itemBackgroundGrid":  {"ItemBackground", propGridRef},
		"borderColor":         {"Border", propColor},
		"focusColor":          {"FocusColor", propColor},
		"focusRingWidth":      {"FocusRingWidth", propFloat},
	},
	"treeList": {
		"backgroundColor":     {"Background", propBackground},
		"backgroundGrid":      {"Background", propGridRef},
		"itemBackgroundColor": {"ItemBackground", propBackground},
		"itemBackgroundGrid":  {"ItemBackground", propGridRef},
		"expandIcon":          {"ExpandIcon", propSpriteRef},
		"collapseIcon":        {"CollapseIcon", propSpriteRef},
		"focusColor":          {"FocusColor", propColor},
		"focusRingWidth":      {"FocusRingWidth", propFloat},
	},
	"tileList": {
		"backgroundColor":     {"Background", propBackground},
		"backgroundGrid":      {"Background", propGridRef},
		"itemBackgroundColor": {"ItemBackground", propBackground},
		"itemBackgroundGrid":  {"ItemBackground", propGridRef},
		"focusColor":          {"FocusColor", propColor},
		"focusRingWidth":      {"FocusRingWidth", propFloat},
	},
	"inventory": {
		"backgroundColor": {"Background", propBackground},
		"backgroundGrid":  {"Background", propGridRef},
		"cellColor":       {"CellColor", propColor},
	},
	"richText": {
		"textColor": {"TextColor", propColor},
	},
	"optionRotator": {
		"backgroundColor":     {"Background", propBackground},
		"borderColor":         {"Border", propColor},
		"borderWidth":         {"BorderWidth", propFloat},
		"cornerRadius":        {"CornerRadius", propFloat},
		"padding":             {"Padding", propPadding},
		"textColor":           {"TextColor", propColor},
		"chevronLeftIcon":     {"ChevronLeftIcon", propSpriteRef},
		"chevronRightIcon":    {"ChevronRightIcon", propSpriteRef},
		"chevronBackground":   {"ChevronBackground", propBackground},
		"chevronBorderColor":  {"ChevronBorder", propColor},
		"chevronBorderWidth":  {"ChevronBorderWidth", propFloat},
		"chevronCornerRadius": {"ChevronCornerRadius", propFloat},
		"chevronWidth":        {"ChevronWidth", propFloat},
		"chevronIconColor":    {"ChevronIconColor", propColor},
		"chevronIconSize":     {"ChevronIconSize", propFloat},
		"focusColor":          {"FocusColor", propColor},
		"focusRingWidth":      {"FocusRingWidth", propFloat},
	},
	"toggleButtonBar": {
		"backgroundColor":           {"Background", propBackground},
		"backgroundGrid":            {"Background", propGridRef},
		"borderColor":               {"Border", propColor},
		"borderWidth":               {"BorderWidth", propFloat},
		"cornerRadius":              {"CornerRadius", propFloat},
		"padding":                   {"Padding", propPadding},
		"spacing":                   {"Spacing", propFloat},
		"selectedBackgroundColor":   {"SelectedBackground", propBackground},
		"selectedBackgroundGrid":    {"SelectedBackground", propGridRef},
		"selectedTextColor":         {"SelectedTextColor", propColor},
		"selectedBorderColor":       {"SelectedBorder", propColor},
		"selectedBorderWidth":       {"SelectedBorderWidth", propFloat},
		"selectedCornerRadius":      {"SelectedCornerRadius", propFloat},
		"unselectedBackgroundColor": {"UnselectedBackground", propBackground},
		"unselectedBackgroundGrid":  {"UnselectedBackground", propGridRef},
		"unselectedTextColor":       {"UnselectedTextColor", propColor},
		"unselectedBorderColor":     {"UnselectedBorder", propColor},
		"unselectedBorderWidth":     {"UnselectedBorderWidth", propFloat},
		"unselectedCornerRadius":    {"UnselectedCornerRadius", propFloat},
		"focusColor":                {"FocusColor", propColor},
		"focusRingWidth":            {"FocusRingWidth", propFloat},
	},
	"tooltip": {
		"backgroundColor": {"Background", propBackground},
		"backgroundGrid":  {"Background", propGridRef},
		"borderColor":     {"BorderColor", propColor},
		"borderWidth":     {"BorderWidth", propFloat},
		"cornerRadius":    {"CornerRadius", propFloat},
		"padding":         {"Padding", propPadding},
		"maxWidth":        {"MaxWidth", propFloat},
	},
	"popover": {
		"backgroundColor": {"Background", propBackground},
		"backgroundGrid":  {"Background", propGridRef},
		"borderColor":     {"BorderColor", propColor},
		"borderWidth":     {"BorderWidth", propFloat},
		"cornerRadius":    {"CornerRadius", propFloat},
		"padding":         {"Padding", propPadding},
		"titleColor":      {"TitleColor", propColor},
	},
	"menuPopup": {
		"backgroundColor":     {"Background", propBackground},
		"backgroundGrid":      {"Background", propGridRef},
		"itemBackgroundColor": {"ItemBackground", propBackground},
		"itemBackgroundGrid":  {"ItemBackground", propGridRef},
		"textColor":           {"TextColor", propColor},
		"disabledColor":       {"DisabledColor", propColor},
		"separatorColor":      {"SeparatorColor", propColor},
		"borderColor":         {"Border", propColor},
		"borderWidth":         {"BorderWidth", propFloat},
		"cornerRadius":        {"CornerRadius", propFloat},
		"padding":             {"Padding", propPadding},
		"itemHeight":          {"ItemHeight", propFloat},
	},
	"select": {
		"backgroundColor": {"Background", propBackground},
		"backgroundGrid":  {"Background", propGridRef},
		"textColor":       {"TextColor", propColor},
		"borderColor":     {"Border", propColor},
		"borderWidth":     {"BorderWidth", propFloat},
		"cornerRadius":    {"CornerRadius", propFloat},
		"padding":         {"Padding", propPadding},
		"chevronColor":    {"ChevronColor", propColor},
		"focusColor":      {"FocusColor", propColor},
		"focusRingWidth":  {"FocusRingWidth", propFloat},
	},
	"dragHandle": {
		"backgroundColor": {"Background", propBackground},
		"backgroundGrid":  {"Background", propGridRef},
		"gripColor":       {"GripColor", propColor},
		"gripHoverColor":  {"GripHoverColor", propColor},
		"gripActiveColor": {"GripActiveColor", propColor},
		"gripDotSize":     {"GripDotSize", propFloat},
		"gripSpacing":     {"GripSpacing", propFloat},
		"gripCount":       {"GripCount", propFloat},
	},
	"sortableList": {
		"backgroundColor":      {"Background", propBackground},
		"backgroundGrid":       {"Background", propGridRef},
		"itemBackgroundColor":  {"ItemBackground", propBackground},
		"itemBackgroundGrid":   {"ItemBackground", propGridRef},
		"itemBorderColor":      {"ItemBorderColor", propColor},
		"itemBorderWidth":      {"ItemBorderWidth", propFloat},
		"itemCornerRadius":     {"ItemCornerRadius", propFloat},
		"itemPadding":          {"ItemPadding", propPadding},
		"selectionColor":       {"SelectionColor", propColor},
		"borderColor":          {"BorderColor", propColor},
		"borderWidth":          {"BorderWidth", propFloat},
		"handleColor":          {"HandleColor", propColor},
		"handleHoverColor":     {"HandleHoverColor", propColor},
		"handleActiveColor":    {"HandleActiveColor", propColor},
		"handleWidth":          {"HandleWidth", propFloat},
		"handleGap":            {"HandleGap", propFloat},
		"insertIndicatorColor": {"InsertIndicatorColor", propColor},
		"insertIndicatorWidth": {"InsertIndicatorWidth", propFloat},
		"focusColor":           {"FocusColor", propColor},
		"focusRingWidth":       {"FocusRingWidth", propFloat},
	},
	"image": {
		"backgroundColor": {"Background", propBackground},
		"backgroundGrid":  {"Background", propGridRef},
		"cornerRadius":    {"CornerRadius", propFloat},
	},
	"iconButton": {
		"backgroundColor": {"Background", propBackground},
		"backgroundGrid":  {"Background", propGridRef},
		"borderColor":     {"BorderColor", propColor},
		"borderWidth":     {"BorderWidth", propFloat},
		"cornerRadius":    {"CornerRadius", propFloat},
		"iconColor":       {"IconColor", propColor},
		"activeColor":     {"ActiveColor", propBackground},
		"labelColor":      {"LabelColor", propColor},
		"iconSize":        {"IconSize", propFloat},
		"padding":         {"Padding", propPadding},
		"labelGap":        {"LabelGap", propFloat},
		"focusColor":      {"FocusColor", propColor},
		"focusRingWidth":  {"FocusRingWidth", propFloat},
	},
	"statWeb": {
		"backgroundColor":    {"Background", propBackground},
		"polygonFill":        {"PolygonFill", propColor},
		"polygonStroke":      {"PolygonStroke", propColor},
		"polygonStrokeWidth": {"PolygonStrokeWidth", propFloat},
		"spokeColor":         {"SpokeColor", propColor},
		"spokeWidth":         {"SpokeWidth", propFloat},
		"gridColor":          {"GridColor", propColor},
		"gridLevels":         {"GridLevels", propFloat},
		"handleColor":        {"HandleColor", propColor},
		"handleHoverColor":   {"HandleHoverColor", propColor},
		"handleRadius":       {"HandleRadius", propFloat},
		"labelColor":         {"LabelColor", propColor},
		"labelFontSize":      {"LabelFontSize", propFloat},
		"labelOffset":        {"LabelOffset", propFloat},
	},
	"accordion": {
		"backgroundColor":   {"Background", propBackground},
		"borderColor":       {"BorderColor", propColor},
		"borderWidth":       {"BorderWidth", propFloat},
		"headerBackground":  {"HeaderBackground", propBackground},
		"headerHoverBg":     {"HeaderHoverBg", propBackground},
		"headerTextColor":   {"HeaderTextColor", propColor},
		"headerHeight":      {"HeaderHeight", propFloat},
		"headerIconSize":    {"HeaderIconSize", propFloat},
		"headerIconGap":     {"HeaderIconGap", propFloat},
		"chevronColor":      {"ChevronColor", propColor},
		"chevronSize":       {"ChevronSize", propFloat},
		"expandIcon":        {"ExpandIcon", propSpriteRef},
		"collapseIcon":      {"CollapseIcon", propSpriteRef},
		"contentBackground": {"ContentBackground", propBackground},
		"dividerColor":      {"DividerColor", propColor},
		"dividerHeight":     {"DividerHeight", propFloat},
		"animationDuration": {"AnimationDuration", propFloat},
		"cornerRadius":      {"CornerRadius", propFloat},
	},
	"dataTable": {
		"backgroundColor":   {"Background", propBackground},
		"headerBackground":  {"HeaderBackground", propBackground},
		"headerText":        {"HeaderText", propColor},
		"headerHoverColor":  {"HeaderHoverColor", propColor},
		"headerBorderColor": {"HeaderBorderColor", propColor},
		"headerBorderWidth": {"HeaderBorderWidth", propFloat},
		"rowBackground":     {"RowBackground", propBackground},
		"rowBackgroundAlt":  {"RowBackgroundAlt", propBackground},
		"rowHoverColor":     {"RowHoverColor", propColor},
		"selectionColor":    {"SelectionColor", propColor},
		"cellText":          {"CellText", propColor},
		"cellPadding":       {"CellPadding", propFloat},
		"dividerColor":      {"DividerColor", propColor},
		"dividerWidth":      {"DividerWidth", propFloat},
		"sortGlyphColor":    {"SortGlyphColor", propColor},
		"sortGlyphInactive": {"SortGlyphInactive", propColor},
		"sortGlyphAsc":      {"SortGlyphAsc", propStr},
		"sortGlyphDesc":     {"SortGlyphDesc", propStr},
		"sortGlyphNone":     {"SortGlyphNone", propStr},
		"borderColor":       {"BorderColor", propColor},
		"borderWidth":       {"BorderWidth", propFloat},
		"cornerRadius":      {"CornerRadius", propFloat},
	},
	"timePicker": {
		"backgroundColor":  {"Background", propBackground},
		"borderColor":      {"BorderColor", propColor},
		"borderWidth":      {"BorderWidth", propFloat},
		"cornerRadius":     {"CornerRadius", propFloat},
		"padding":          {"Padding", propPadding},
		"columnBackground": {"ColumnBackground", propBackground},
		"columnWidth":      {"ColumnWidth", propFloat},
		"columnHeight":     {"ColumnHeight", propFloat},
		"valueTextColor":   {"ValueTextColor", propColor},
		"valueFontSize":    {"ValueFontSize", propFloat},
		"arrowColor":       {"ArrowColor", propColor},
		"arrowHoverColor":  {"ArrowHoverColor", propColor},
		"arrowSize":        {"ArrowSize", propFloat},
		"separatorColor":   {"SeparatorColor", propColor},
		"separatorWidth":   {"SeparatorWidth", propFloat},
		"amPmBackground":   {"AmPmBackground", propBackground},
		"amPmActiveColor":  {"AmPmActiveColor", propColor},
		"amPmTextColor":    {"AmPmTextColor", propColor},
		"amPmCornerRadius": {"AmPmCornerRadius", propFloat},
	},
	"imageCropper": {
		"backgroundColor":    {"Background", propBackground},
		"cornerRadius":       {"CornerRadius", propFloat},
		"cropBorderColor":    {"CropBorderColor", propColor},
		"cropBorderWidth":    {"CropBorderWidth", propFloat},
		"handleBackground":   {"HandleBackground", propBackground},
		"handleSize":         {"HandleSize", propFloat},
		"handleCornerRadius": {"HandleCornerRadius", propFloat},
		"dimColor":           {"DimColor", propColor},
		"gridColor":          {"GridColor", propColor},
		"gridLineWidth":      {"GridLineWidth", propFloat},
	},
	"colorPicker": {
		"backgroundColor":    {"Background", propBackground},
		"borderColor":        {"BorderColor", propColor},
		"borderWidth":        {"BorderWidth", propFloat},
		"cornerRadius":       {"CornerRadius", propFloat},
		"padding":            {"Padding", propPadding},
		"swatchBorderColor":  {"SwatchBorderColor", propColor},
		"swatchBorderWidth":  {"SwatchBorderWidth", propFloat},
		"swatchCornerRadius": {"SwatchCornerRadius", propFloat},
		"popupWidth":         {"PopupWidth", propFloat},
		"svFieldSize":        {"SVFieldSize", propFloat},
		"hueBarHeight":       {"HueBarHeight", propFloat},
		"alphaBarHeight":     {"AlphaBarHeight", propFloat},
	},
	"toolBar": {
		"backgroundColor":    {"Background", propBackground},
		"borderColor":        {"BorderColor", propColor},
		"borderWidth":        {"BorderWidth", propFloat},
		"cornerRadius":       {"CornerRadius", propFloat},
		"padding":            {"Padding", propPadding},
		"spacing":            {"Spacing", propFloat},
		"separatorColor":     {"SeparatorColor", propColor},
		"separatorThickness": {"SeparatorThickness", propFloat},
		"separatorHeight":    {"SeparatorHeight", propFloat},
	},
	"calendarSelector": {
		"backgroundColor":     {"Background", propBackground},
		"borderColor":         {"BorderColor", propColor},
		"borderWidth":         {"BorderWidth", propFloat},
		"cornerRadius":        {"CornerRadius", propFloat},
		"padding":             {"Padding", propPadding},
		"headerBackground":    {"HeaderBackground", propBackground},
		"headerTextColor":     {"HeaderTextColor", propColor},
		"navButtonColor":      {"NavButtonColor", propColor},
		"navButtonHoverColor": {"NavButtonHoverColor", propColor},
		"weekdayTextColor":    {"WeekdayTextColor", propColor},
		"weekdayBackground":   {"WeekdayBackground", propBackground},
		"dayTextColor":        {"DayTextColor", propColor},
		"dayBackground":       {"DayBackground", propBackground},
		"dayHoverBg":          {"DayHoverBg", propBackground},
		"daySelectedBg":       {"DaySelectedBg", propBackground},
		"daySelectedColor":    {"DaySelectedColor", propColor},
		"dayTodayBg":          {"DayTodayBg", propBackground},
		"dayTodayColor":       {"DayTodayColor", propColor},
		"dayMutedColor":       {"DayMutedColor", propColor},
		"daySize":             {"DaySize", propFloat},
		"dayCornerRadius":     {"DayCornerRadius", propFloat},
		"triggerBackground":   {"TriggerBackground", propBackground},
		"triggerTextColor":    {"TriggerTextColor", propColor},
		"triggerBorderColor":  {"TriggerBorderColor", propColor},
	},
	"richTextEditor": {
		"backgroundColor":    {"Background", propBackground},
		"backgroundGrid":     {"Background", propGridRef},
		"borderColor":        {"BorderColor", propColor},
		"borderWidth":        {"BorderWidth", propFloat},
		"cornerRadius":       {"CornerRadius", propFloat},
		"padding":            {"Padding", propPadding},
		"toolbarBackground":  {"ToolbarBackground", propBackground},
		"toolbarHeight":      {"ToolbarHeight", propFloat},
		"toolbarButtonSize":  {"ToolbarButtonSize", propFloat},
		"toolbarButtonGap":   {"ToolbarButtonGap", propFloat},
		"toolbarActiveColor": {"ToolbarActiveColor", propColor},
		"contentBackground":  {"ContentBackground", propBackground},
		"cursorColor":        {"CursorColor", propColor},
		"selectionColor":     {"SelectionColor", propColor},
	},
	"propertyInspector": {
		"backgroundColor":       {"Background", propBackground},
		"backgroundGrid":        {"Background", propGridRef},
		"borderColor":           {"BorderColor", propColor},
		"borderWidth":           {"BorderWidth", propFloat},
		"cornerRadius":          {"CornerRadius", propFloat},
		"searchBarHeight":       {"SearchBarHeight", propFloat},
		"searchBarGap":          {"SearchBarGap", propFloat},
		"groupHeaderBackground": {"GroupHeaderBackground", propBackground},
		"groupHeaderTextColor":  {"GroupHeaderTextColor", propColor},
		"groupHeaderHeight":     {"GroupHeaderHeight", propFloat},
		"rowBackground":         {"RowBackground", propBackground},
		"rowAltBackground":      {"RowAltBackground", propBackground},
		"rowHoverBackground":    {"RowHoverBackground", propBackground},
		"rowHeight":             {"RowHeight", propFloat},
		"labelColor":            {"LabelColor", propColor},
		"labelWidth":            {"LabelWidth", propFloat},
		"dividerColor":          {"DividerColor", propColor},
	},
	"animatedImage": {
		"backgroundColor": {"Background", propBackground},
		"cornerRadius":    {"CornerRadius", propFloat},
	},
	"gradientEditor": {
		"backgroundColor": {"Background", propBackground},
		"borderColor":     {"BorderColor", propColor},
		"borderWidth":     {"BorderWidth", propFloat},
		"cornerRadius":    {"CornerRadius", propFloat},
		"padding":         {"Padding", propPadding},
		"previewHeight":   {"PreviewHeight", propFloat},
		"previewSize":     {"PreviewSize", propFloat},
	},
	"inputField": {
		"labelColor":    {"LabelColor", propColor},
		"requiredColor": {"RequiredColor", propColor},
		"errorColor":    {"ErrorColor", propColor},
		"warningColor":  {"WarningColor", propColor},
		"successColor":  {"SuccessColor", propColor},
		"labelGap":      {"LabelGap", propFloat},
		"messageGap":    {"MessageGap", propFloat},
		"labelLeftGap":  {"LabelLeftGap", propFloat},
	},
	"keybindInput": {
		"backgroundColor":          {"Background", propBackground},
		"borderColor":              {"BorderColor", propColor},
		"borderWidth":              {"BorderWidth", propFloat},
		"cornerRadius":             {"CornerRadius", propFloat},
		"padding":                  {"Padding", propPadding},
		"keyCapBackgroundColor":    {"KeyCapBackground", propBackground},
		"keyCapTextColor":          {"KeyCapTextColor", propColor},
		"keyCapBorderColor":        {"KeyCapBorderColor", propColor},
		"keyCapBorderWidth":        {"KeyCapBorderWidth", propFloat},
		"keyCapCornerRadius":       {"KeyCapCornerRadius", propFloat},
		"keyCapPadding":            {"KeyCapPadding", propPadding},
		"listeningBackgroundColor": {"ListeningBackground", propBackground},
		"listeningTextColor":       {"ListeningTextColor", propColor},
		"listeningBorderColor":     {"ListeningBorderColor", propColor},
		"clearButtonColor":         {"ClearButtonColor", propColor},
		"clearButtonSize":          {"ClearButtonSize", propFloat},
		"unsetTextColor":           {"UnsetTextColor", propColor},
		"unsetText":                {"UnsetText", propStr},
		"listeningText":            {"ListeningText", propStr},
	},
	"maskedInput": {
		"backgroundColor":     {"Background", propBackground},
		"textColor":           {"TextColor", propColor},
		"cursorColor":         {"CursorColor", propColor},
		"selectionColor":      {"SelectionColor", propColor},
		"borderColor":         {"Border", propColor},
		"borderWidth":         {"BorderWidth", propFloat},
		"cornerRadius":        {"CornerRadius", propFloat},
		"placeholderAlpha":    {"PlaceholderAlpha", propFloat},
		"padding":             {"Padding", propPadding},
		"literalColor":        {"LiteralColor", propColor},
		"maskPlaceholderColor": {"MaskPlaceholderColor", propColor},
		"focusColor":          {"FocusColor", propColor},
		"focusRingWidth":      {"FocusRingWidth", propFloat},
	},
	"menuBar": {
		"backgroundColor":      {"Background", propBackground},
		"entryTextColor":       {"EntryTextColor", propColor},
		"entryBackgroundColor": {"EntryBackground", propBackground},
		"entryPadding":         {"EntryPadding", propPadding},
		"spacing":              {"Spacing", propFloat},
		"height":               {"Height", propFloat},
		"borderColor":          {"BorderColor", propColor},
		"borderWidth":          {"BorderWidth", propFloat},
	},
	"navDrawer": {
		"backgroundColor":   {"Background", propBackground},
		"borderColor":       {"BorderColor", propColor},
		"borderWidth":       {"BorderWidth", propFloat},
		"padding":           {"Padding", propPadding},
		"backdropColor":     {"BackdropColor", propColor},
		"animationDuration": {"AnimationDuration", propFloat},
	},
	"searchBox": {
		"backgroundColor":  {"Background", propBackground},
		"textColor":        {"TextColor", propColor},
		"cursorColor":      {"CursorColor", propColor},
		"selectionColor":   {"SelectionColor", propColor},
		"borderColor":      {"Border", propColor},
		"borderWidth":      {"BorderWidth", propFloat},
		"cornerRadius":     {"CornerRadius", propFloat},
		"placeholderAlpha": {"PlaceholderAlpha", propFloat},
		"padding":          {"Padding", propPadding},
		"iconColor":        {"IconColor", propColor},
		"clearButtonColor": {"ClearButtonColor", propColor},
		"clearHoverColor":  {"ClearHoverColor", propColor},
		"clearActiveColor": {"ClearActiveColor", propColor},
		"iconGap":          {"IconGap", propFloat},
		"focusColor":       {"FocusColor", propColor},
		"focusRingWidth":   {"FocusRingWidth", propFloat},
	},
	"sortableTreeList": {
		"backgroundColor":      {"Background", propBackground},
		"borderColor":          {"BorderColor", propColor},
		"borderWidth":          {"BorderWidth", propFloat},
		"rowBackgroundColor":   {"RowBackground", propBackground},
		"rowHoverBgColor":      {"RowHoverBg", propBackground},
		"rowSelectedBgColor":   {"RowSelectedBg", propBackground},
		"rowHeight":            {"RowHeight", propFloat},
		"rowPadding":           {"RowPadding", propPadding},
		"indentWidth":          {"IndentWidth", propFloat},
		"labelColor":           {"LabelColor", propColor},
		"iconSize":             {"IconSize", propFloat},
		"iconGap":              {"IconGap", propFloat},
		"chevronColor":         {"ChevronColor", propColor},
		"chevronSize":          {"ChevronSize", propFloat},
		"dropLineColor":        {"DropLineColor", propColor},
		"dropLineWidth":        {"DropLineWidth", propFloat},
		"dropTargetBgColor":    {"DropTargetBg", propBackground},
		"dropTargetBorderColor": {"DropTargetBorderColor", propColor},
		"focusColor":           {"FocusColor", propColor},
		"focusRingWidth":       {"FocusRingWidth", propFloat},
	},
	"tag": {
		"backgroundColor":         {"Background", propBackground},
		"selectedBackgroundColor": {"SelectedBackground", propBackground},
		"textColor":               {"TextColor", propColor},
		"selectedTextColor":       {"SelectedTextColor", propColor},
		"borderColor":             {"BorderColor", propColor},
		"borderWidth":             {"BorderWidth", propFloat},
		"cornerRadius":            {"CornerRadius", propFloat},
		"padding":                 {"Padding", propPadding},
		"removeButtonSize":        {"RemoveButtonSize", propFloat},
		"removeButtonColor":       {"RemoveButtonColor", propColor},
		"gap":                     {"Gap", propFloat},
	},
	"tagBar": {
		"backgroundColor": {"Background", propBackground},
		"borderColor":     {"Border", propColor},
		"borderWidth":     {"BorderWidth", propFloat},
		"cornerRadius":    {"CornerRadius", propFloat},
		"padding":         {"Padding", propPadding},
		"spacing":         {"Spacing", propFloat},
		"focusColor":      {"FocusColor", propColor},
		"focusRingWidth":  {"FocusRingWidth", propFloat},
	},
	"toast": {
		"backgroundColor":  {"Background", propBackground},
		"textColor":        {"TextColor", propColor},
		"borderColor":      {"BorderColor", propColor},
		"borderWidth":      {"BorderWidth", propFloat},
		"cornerRadius":     {"CornerRadius", propFloat},
		"padding":          {"Padding", propPadding},
		"iconColor":        {"IconColor", propColor},
		"progressBarColor": {"ProgressBarColor", propColor},
		"minWidth":         {"MinWidth", propFloat},
		"maxWidth":         {"MaxWidth", propFloat},
		"itemSpacing":      {"ItemSpacing", propFloat},
		"animDuration":     {"AnimDuration", propFloat},
	},
}

// ---------------------------------------------------------------------------
// LoadTheme
// ---------------------------------------------------------------------------

// LoadTheme parses JSON theme data and produces a *Theme.
// Returns an error if validation fails (bad colors, missing required groups, etc.).
// Errors are collected so the author sees all problems at once.
// Nine-slice images are rejected — use LoadThemeFromFile or LoadThemeFromFS.
func LoadTheme(data []byte) (*Theme, error) {
	return compileThemeInternal(data, nilImageLoader{})
}

// LoadThemeRelative loads a theme JSON file resolved relative to the caller's
// source file. This is convenient for examples and tests where the JSON file
// sits next to the Go source.
func LoadThemeRelative(filename string) (*Theme, error) {
	_, src, _, ok := runtime.Caller(1)
	if !ok {
		return nil, fmt.Errorf("LoadThemeRelative: unable to determine caller path")
	}
	return LoadThemeFromFile(filepath.Join(filepath.Dir(src), filename))
}

// LoadThemeFromFile reads a JSON file and compiles the theme.
// Nine-slice image paths are resolved relative to the JSON file's directory.
func LoadThemeFromFile(path string) (*Theme, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("theme compile: %w", err)
	}
	dir := filepath.Dir(path)
	return compileThemeInternal(data, fileImageLoader{baseDir: dir})
}

// LoadThemeFromFS reads a JSON file from an fs.FS and compiles the theme.
// Nine-slice image paths are resolved within the FS.
func LoadThemeFromFS(fsys fs.FS, path string) (*Theme, error) {
	data, err := fs.ReadFile(fsys, path)
	if err != nil {
		return nil, fmt.Errorf("theme compile: %w", err)
	}
	return compileThemeInternal(data, fsImageLoader{fsys: fsys})
}

func compileThemeInternal(data []byte, loader imageLoader) (*Theme, error) {
	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("theme compile: invalid JSON: %w", err)
	}

	var errs []error
	t := &Theme{}

	// Check for prebaked atlas.
	var prebakedAtlas *sg.Atlas
	if atlasObj, ok := raw["atlas"].(map[string]any); ok {
		a, err := loadPrebakedAtlas(atlasObj, loader)
		if err != nil {
			return nil, err
		}
		prebakedAtlas = a
		t.Atlas = a
	}

	// Parse colors section for $reference resolution.
	colorMap := make(map[string]sg.Color)
	if colorsObj, ok := raw["colors"].(map[string]any); ok {
		for name, val := range colorsObj {
			colorStr, ok := val.(string)
			if !ok {
				errs = append(errs, fmt.Errorf("theme compile: colors.%s: expected color string", name))
				continue
			}
			c, err := markup.ParseColor(colorStr)
			if err != nil {
				errs = append(errs, fmt.Errorf("theme compile: colors.%s: %w", name, err))
				continue
			}
			colorMap[name] = c
		}
	}
	if len(errs) > 0 {
		return nil, errors.Join(errs...)
	}

	// Determine component source: new format uses "components" wrapper,
	// old format has component sections at the top level.
	componentSource := raw
	if comps, ok := raw["components"].(map[string]any); ok {
		componentSource = comps
	}

	// Parse nine-grids section.
	gridDefs, gridErrs := parseNineGrids(raw, loader, colorMap)
	errs = append(errs, gridErrs...)
	if gridDefs == nil {
		gridDefs = make(map[string]*nineGridDef)
	}
	if len(errs) > 0 {
		return nil, errors.Join(errs...)
	}

	// Parse sprites section.
	resolvedSprites, spriteErrs := parseSprites(raw, loader)
	errs = append(errs, spriteErrs...)
	if resolvedSprites == nil {
		resolvedSprites = make(map[string]SpriteRef)
	}
	if len(errs) > 0 {
		return nil, errors.Join(errs...)
	}

	// Unknown top-level keys are silently accepted for forward compatibility.

	// Parse user-defined variant names. These map friendly names to Custom1..Custom56 slots.
	userVariants := make(map[string]Variant)
	nextCustom := Custom1
	if variantsArr, ok := raw["variants"].([]any); ok {
		for _, v := range variantsArr {
			name, ok := v.(string)
			if !ok {
				errs = append(errs, fmt.Errorf("theme compile: variants: expected string entries"))
				continue
			}
			if _, builtin := builtinVariantNames[name]; builtin {
				errs = append(errs, fmt.Errorf("theme compile: variants: %q shadows built-in variant name", name))
				continue
			}
			if nextCustom >= VariantCount {
				errs = append(errs, fmt.Errorf("theme compile: variants: too many custom variants (max %d)", Custom56-Custom1+1))
				break
			}
			userVariants[name] = nextCustom
			nextCustom++
		}
	}
	if len(userVariants) > 0 {
		t.CustomVariants = userVariants
	}
	if len(resolvedSprites) > 0 {
		t.Sprites = resolvedSprites
	}

	// Parse fonts section.
	if fontsObj, ok := raw["fonts"].(map[string]any); ok {
		t.Fonts = make(map[string]string, len(fontsObj))
		for role, val := range fontsObj {
			name, ok := val.(string)
			if !ok {
				errs = append(errs, fmt.Errorf("theme compile: fonts.%s: expected string", role))
				continue
			}
			t.Fonts[role] = name
		}
	}

	if len(errs) > 0 {
		return nil, errors.Join(errs...)
	}

	// Parse global "_" defaults from component source. These cascade into
	// every component that declares a matching property.
	var globalUnderscoreRaw map[string]any
	if underscoreObj, ok := componentSource["_"].(map[string]any); ok {
		globalUnderscoreRaw = underscoreObj
	}

	// Build per-component top-level defaults by matching global "_" keys
	// against each component's property map. Also checks bare top-level keys
	// (old format) as fallback.
	topLevelByComp := make(map[string]*parsedGroup, len(componentPropertyMaps))
	for compName, propMap := range componentPropertyMaps {
		// Start with bare top-level keys (old format backward compat).
		pg, pgErrs := parseProps(componentSource, propMap, "", loader, nil, nil, colorMap, gridDefs, resolvedSprites)
		errs = append(errs, pgErrs...)
		// Merge global "_" defaults (these take priority for new format).
		if globalUnderscoreRaw != nil {
			uPg, uErrs := parseProps(globalUnderscoreRaw, propMap, "_.", loader, nil, nil, colorMap, gridDefs, resolvedSprites)
			errs = append(errs, uErrs...)
			// underscore values override bare top-level values
			for field, val := range uPg.colors {
				pg.colors[field] = val
			}
			for field, val := range uPg.bgs {
				pg.bgs[field] = val
			}
			for field, val := range uPg.floats {
				pg.floats[field] = val
			}
			for field, val := range uPg.stateFloats {
				pg.stateFloats[field] = val
			}
			if uPg.padding != nil {
				pg.padding = uPg.padding
			}
			for field := range uPg.unset {
				pg.unset[field] = true
			}
		}
		topLevelByComp[compName] = pg
	}

	// Parse each component section (pass 1: parse + load images).
	var pending []nineSliceEntry
	loadedImages := make(map[string]image.Image)

	for compName, propMap := range componentPropertyMaps {
		compData, ok := componentSource[compName]
		if !ok {
			continue
		}
		compObj, ok := compData.(map[string]any)
		if !ok {
			errs = append(errs, fmt.Errorf("theme compile: %s: expected object", compName))
			continue
		}
		compErrs := compileComponent(t, compName, propMap, compObj, topLevelByComp[compName], loader, loadedImages, &pending, colorMap, gridDefs, userVariants, resolvedSprites)
		errs = append(errs, compErrs...)
	}

	// Parse user-defined component sections.
	for compName, compData := range componentSource {
		if compName == "_" {
			continue
		}
		if _, isBuiltin := componentPropertyMaps[compName]; isBuiltin {
			continue
		}
		compObj, ok := compData.(map[string]any)
		if !ok {
			continue
		}
		config, compErrs := compileUserComponent(compName, compObj, loader, loadedImages, &pending, colorMap, gridDefs, userVariants)
		errs = append(errs, compErrs...)
		if config != nil {
			if t.UserComponents == nil {
				t.UserComponents = make(map[string]*UserConfig)
			}
			t.UserComponents[compName] = config
		}
	}

	if len(errs) > 0 {
		return nil, errors.Join(errs...)
	}

	// Pass 2: Resolve nine-slice regions.
	if len(pending) > 0 {
		if prebakedAtlas != nil {
			// Use prebaked atlas for region lookups.
			for i := range pending {
				region := prebakedAtlas.Region(pending[i].imagePath)
				pending[i].slice.Region = region
			}
		} else {
			// Pack atlas at runtime.
			atlas := sg.NewBatchAtlas()
			for imgPath, img := range loadedImages {
				if err := atlas.Stage(imgPath, engine.NewImageFromImage(img)); err != nil {
					return nil, fmt.Errorf("theme compile: atlas stage %q: %w", imgPath, err)
				}
			}
			if err := atlas.Pack(); err != nil {
				return nil, fmt.Errorf("theme compile: atlas pack: %w", err)
			}
			t.Atlas = atlas

			for i := range pending {
				region := atlas.Region(pending[i].imagePath)
				pending[i].slice.Region = region
			}
		}
	}

	return t, nil
}

// loadPrebakedAtlas loads a prebaked atlas from the "atlas" top-level key.
// Expected format: {"image": "atlas.png", "data": "atlas.json"}
// or multi-page: {"images": ["atlas-0.png", "atlas-1.png"], "data": "atlas.json"}
func loadPrebakedAtlas(obj map[string]any, loader imageLoader) (*sg.Atlas, error) {
	dataPath, _ := obj["data"].(string)
	if dataPath == "" {
		return nil, fmt.Errorf("theme compile: atlas: missing \"data\" field")
	}

	// Collect page image paths.
	var imagePaths []string
	if single, ok := obj["image"].(string); ok && single != "" {
		imagePaths = []string{single}
	} else if arr, ok := obj["images"].([]any); ok {
		for _, v := range arr {
			s, ok := v.(string)
			if !ok || s == "" {
				return nil, fmt.Errorf("theme compile: atlas.images: expected array of strings")
			}
			imagePaths = append(imagePaths, s)
		}
	}
	if len(imagePaths) == 0 {
		return nil, fmt.Errorf("theme compile: atlas: missing \"image\" or \"images\" field")
	}

	// Read atlas JSON descriptor.
	jsonData, err := loader.readFile(dataPath)
	if err != nil {
		return nil, fmt.Errorf("theme compile: atlas.data: %w", err)
	}

	// Load page images.
	pages := make([]engine.Image, len(imagePaths))
	for i, imgPath := range imagePaths {
		img, err := loader.loadImage(imgPath)
		if err != nil {
			return nil, fmt.Errorf("theme compile: atlas image %q: %w", imgPath, err)
		}
		pages[i] = engine.NewImageFromImage(img)
	}

	// Parse atlas using willow's LoadAtlas.
	atlas, err := sg.LoadAtlas(jsonData, pages)
	if err != nil {
		return nil, fmt.Errorf("theme compile: load atlas: %w", err)
	}

	return atlas, nil
}

// ---------------------------------------------------------------------------
// ValidateTheme
// ---------------------------------------------------------------------------

// ValidateTheme checks that the given theme defines configs for all the
// named component types. Returns an error listing any missing configs.
// Use at boot to catch incomplete themes early.
func ValidateTheme(t *Theme, components ...string) error {
	var missing []string
	for _, name := range components {
		if !themeHasComponent(t, name) {
			missing = append(missing, name)
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("theme: missing configs: %s", strings.Join(missing, ", "))
	}
	return nil
}

// themeHasComponent returns true if the component's primary group has any
// non-zero values (meaning it was explicitly set, not left as zero-value).
func themeHasComponent(t *Theme, name string) bool {
	zero := sg.Color{}
	switch name {
	case "button":
		return t.Button.Primary.Background[core.StateDefault].Type != core.BgNone ||
			t.Button.Primary.TextColor[core.StateDefault] != zero
	case "label":
		return t.Label.Primary.TextColor[core.StateDefault] != zero
	case "badge":
		return t.Badge.Primary.Background[core.StateDefault].Type != core.BgNone ||
			t.Badge.Primary.TextColor[core.StateDefault] != zero
	case "toggle":
		return t.Toggle.Primary.TrackColor[core.StateDefault] != zero ||
			t.Toggle.Primary.ThumbColor[core.StateDefault] != zero
	case "checkbox":
		return t.Checkbox.Primary.BoxColor[core.StateDefault] != zero ||
			t.Checkbox.Primary.CheckColor[core.StateDefault] != zero
	case "radio":
		return t.Radio.Primary.CircleColor[core.StateDefault] != zero ||
			t.Radio.Primary.DotColor[core.StateDefault] != zero
	case "textInput":
		return t.TextInput.Primary.Background[core.StateDefault].Type != core.BgNone ||
			t.TextInput.Primary.TextColor[core.StateDefault] != zero
	case "maskedInput":
		return t.MaskedInput.Primary.Background[core.StateDefault].Type != core.BgNone ||
			t.MaskedInput.Primary.TextColor[core.StateDefault] != zero
	case "inputField":
		return t.InputField.Primary.LabelColor[core.StateDefault] != zero
	case "searchBox":
		return t.SearchBox.Primary.Background[core.StateDefault].Type != core.BgNone ||
			t.SearchBox.Primary.TextColor[core.StateDefault] != zero
	case "textArea":
		return t.TextArea.Primary.Background[core.StateDefault].Type != core.BgNone ||
			t.TextArea.Primary.TextColor[core.StateDefault] != zero
	case "slider":
		return t.Slider.Primary.Background[core.StateDefault].Type != core.BgNone
	case "scrollBar":
		return t.ScrollBar.Primary.Background[core.StateDefault].Type != core.BgNone
	case "meterBar":
		return t.MeterBar.Primary.Background[core.StateDefault].Type != core.BgNone
	case "panel":
		return t.Panel.Primary.Background[core.StateDefault].Type != core.BgNone
	case "navDrawer":
		return t.NavDrawer.Primary.Background[core.StateDefault].Type != core.BgNone
	case "window":
		return t.Window.Primary.Background[core.StateDefault].Type != core.BgNone ||
			t.Window.Primary.TitleBackground[core.StateDefault].Type != core.BgNone
	case "tabs":
		return t.Tabs.Primary.BarBackground[core.StateDefault].Type != core.BgNone
	case "list":
		return t.List.Primary.Background[core.StateDefault].Type != core.BgNone ||
			t.List.Primary.ItemBackground[core.StateDefault].Type != core.BgNone
	case "treeList":
		return t.TreeList.Primary.Background[core.StateDefault].Type != core.BgNone
	case "tileList":
		return t.TileList.Primary.Background[core.StateDefault].Type != core.BgNone ||
			t.TileList.Primary.ItemBackground[core.StateDefault].Type != core.BgNone
case "richText":
		return t.RichText.Primary.TextColor[core.StateDefault] != zero
	case "optionRotator":
		return t.OptionRotator.Primary.TextColor[core.StateDefault] != zero
	case "toggleButtonBar":
		return t.ToggleButtonBar.Primary.Background[core.StateDefault].Type != core.BgNone ||
			t.ToggleButtonBar.Primary.SelectedBackground[core.StateDefault].Type != core.BgNone
	case "tooltip":
		return t.Tooltip.Primary.Background[core.StateDefault].Type != core.BgNone ||
			t.Tooltip.Primary.BorderColor[core.StateDefault] != zero
	case "dragHandle":
		return t.DragHandle.Primary.GripColor[core.StateDefault] != zero
	case "sortableList":
		return t.SortableList.Primary.Background[core.StateDefault].Type != core.BgNone ||
			t.SortableList.Primary.HandleColor[core.StateDefault] != zero
	case "image":
		return true // image group is always valid (transparent bg is intentional)
	case "animatedImage":
		return true // animated image group is always valid (transparent bg is intentional)
	case "colorPicker":
		return t.ColorPicker.Primary.Background[core.StateDefault].Type != core.BgNone ||
			t.ColorPicker.Primary.BorderColor[core.StateDefault] != zero
	case "toast":
		return t.Toast.Primary.Background[core.StateDefault].Type != core.BgNone ||
			t.Toast.Primary.TextColor[core.StateDefault] != zero
	case "popover":
		return t.Popover.Primary.Background[core.StateDefault].Type != core.BgNone ||
			t.Popover.Primary.BorderColor[core.StateDefault] != zero
	case "iconButton":
		return t.IconButton.Primary.Background[core.StateDefault].Type != core.BgNone ||
			t.IconButton.Primary.IconColor[core.StateDefault] != zero
	case "statWeb":
		return t.StatWeb.Primary.PolygonFill[core.StateDefault] != zero ||
			t.StatWeb.Primary.SpokeColor[core.StateDefault] != zero
	case "gradientEditor":
		return t.GradientEditor.Primary.Background[core.StateDefault].Type != core.BgNone ||
			t.GradientEditor.Primary.BorderColor[core.StateDefault] != zero
	case "accordion":
		return t.Accordion.Primary.Background[core.StateDefault].Type != core.BgNone ||
			t.Accordion.Primary.HeaderBackground[core.StateDefault].Type != core.BgNone
	case "timePicker":
		return t.TimePicker.Primary.Background[core.StateDefault].Type != core.BgNone ||
			t.TimePicker.Primary.ValueTextColor[core.StateDefault] != zero
	case "imageCropper":
		return t.ImageCropper.Primary.Background[core.StateDefault].Type != core.BgNone ||
			t.ImageCropper.Primary.CropBorderColor[core.StateDefault] != zero
	case "toolBar":
		return t.ToolBar.Primary.Background[core.StateDefault].Type != core.BgNone ||
			t.ToolBar.Primary.BorderColor[core.StateDefault] != zero
	case "calendarSelector":
		return t.CalendarSelector.Primary.Background[core.StateDefault].Type != core.BgNone ||
			t.CalendarSelector.Primary.HeaderTextColor[core.StateDefault] != zero
	case "richTextEditor":
		return t.RichTextEditor.Primary.Background[core.StateDefault].Type != core.BgNone ||
			t.RichTextEditor.Primary.BorderColor[core.StateDefault] != zero
	case "propertyInspector":
		return t.PropertyInspector.Primary.Background[core.StateDefault].Type != core.BgNone ||
			t.PropertyInspector.Primary.BorderColor[core.StateDefault] != zero
	default:
		return false
	}
}

// ---------------------------------------------------------------------------
// Component compilation
// ---------------------------------------------------------------------------

// parseProps parses raw key-value pairs against a property map into a
// parsedGroup. Keys not found in propMap are silently skipped.
// colorMap provides $reference resolution for color values.
// gridDefs provides nine-grid definitions for propGridRef resolution.
func parseProps(raw map[string]any, propMap map[string]propInfo, pathPrefix string, loader imageLoader, loadedImages map[string]image.Image, pending *[]nineSliceEntry, colorMap map[string]sg.Color, gridDefs map[string]*nineGridDef, sprites map[string]SpriteRef) (*parsedGroup, []error) {
	var errs []error
	pg := &parsedGroup{
		colors:      make(map[string]ColorProperty),
		bgs:         make(map[string]BackgroundProperty),
		floats:      make(map[string]float64),
		bools:       make(map[string]bool),
		stateFloats: make(map[string]FloatProperty),
		spriteRefs:  make(map[string]SpriteRef),
		strs:        make(map[string]string),
		unset:       make(map[string]bool),
	}

	// Track grid ref keys for second-pass processing (grid wins over color).
	var gridRefKeys []string

	for key, val := range raw {
		info, ok := propMap[key]
		if !ok {
			if suggestion, hasSuggestion := keyAliases[key]; hasSuggestion {
				if _, valid := propMap[suggestion]; valid {
					fmt.Fprintf(os.Stderr, "theme warning: %s%q is not a valid key (did you mean %q?)\n", pathPrefix, key, suggestion)
				}
			}
			continue
		}
		path := pathPrefix + key

		// Check for unset values (null, "nil", "none", "").
		if isUnsetValue(val) {
			pg.unset[info.goField] = true
			continue
		}

		switch info.kind {
		case propColor:
			cp, propErrs := compileColorProperty(path, val, colorMap)
			errs = append(errs, propErrs...)
			if len(propErrs) == 0 {
				pg.colors[info.goField] = cp
			}
		case propBackground:
			bp, propErrs := compileBackgroundProperty(path, val, loader, loadedImages, pending, colorMap)
			errs = append(errs, propErrs...)
			if len(propErrs) == 0 {
				pg.bgs[info.goField] = bp
			}
		case propGridRef:
			// Defer grid refs to second pass so they override backgroundColor.
			gridRefKeys = append(gridRefKeys, key)
		case propPadding:
			insets, err := compilePadding(path, val)
			if err != nil {
				errs = append(errs, err)
			} else {
				pg.padding = &insets
			}
		case propFloat:
			if !isNumeric(val) {
				errs = append(errs, fmt.Errorf("theme compile: %s: expected number", path))
			} else {
				pg.floats[info.goField] = toFloat(val)
			}
		case propBool:
			if b, ok := val.(bool); ok {
				pg.bools[info.goField] = b
			} else {
				errs = append(errs, fmt.Errorf("theme compile: %s: expected boolean", path))
			}
		case propStateFloat:
			fp, propErrs := compileFloatProperty(path, val)
			errs = append(errs, propErrs...)
			if len(propErrs) == 0 {
				pg.stateFloats[info.goField] = fp
			}
		case propSpriteRef:
			name, ok := val.(string)
			if !ok {
				errs = append(errs, fmt.Errorf("theme compile: %s: expected sprite name string", path))
			} else if name != "" {
				if sr, found := sprites[name]; found {
					pg.spriteRefs[info.goField] = sr
				} else {
					errs = append(errs, fmt.Errorf("theme compile: %s: unknown sprite %q", path, name))
				}
			}
		case propStr:
			s, ok := val.(string)
			if !ok {
				errs = append(errs, fmt.Errorf("theme compile: %s: expected string", path))
			} else {
				pg.strs[info.goField] = s
			}
		}
	}

	// Second pass: resolve grid references (these override any solid bg).
	for _, key := range gridRefKeys {
		info := propMap[key]
		path := pathPrefix + key
		bp, propErrs := compileGridRefProperty(path, raw[key], gridDefs, loader, loadedImages, pending)
		errs = append(errs, propErrs...)
		if len(propErrs) == 0 {
			pg.bgs[info.goField] = bp
		}
	}

	return pg, errs
}

// mergeDefaults fills in missing values from a lower-priority parsedGroup.
// Fields that were explicitly unset (via null/"nil"/"none"/"") are not inherited.
func (pg *parsedGroup) mergeDefaults(defaults *parsedGroup) {
	if defaults == nil {
		return
	}
	for field, val := range defaults.colors {
		if _, exists := pg.colors[field]; !exists && !pg.unset[field] {
			pg.colors[field] = val
		}
	}
	for field, val := range defaults.bgs {
		if _, exists := pg.bgs[field]; !exists && !pg.unset[field] {
			pg.bgs[field] = val
		}
	}
	for field, val := range defaults.floats {
		if _, exists := pg.floats[field]; !exists && !pg.unset[field] {
			pg.floats[field] = val
		}
	}
	for field, val := range defaults.bools {
		if _, exists := pg.bools[field]; !exists && !pg.unset[field] {
			pg.bools[field] = val
		}
	}
	for field, val := range defaults.stateFloats {
		if _, exists := pg.stateFloats[field]; !exists && !pg.unset[field] {
			pg.stateFloats[field] = val
		}
	}
	for field, val := range defaults.spriteRefs {
		if _, exists := pg.spriteRefs[field]; !exists && !pg.unset[field] {
			pg.spriteRefs[field] = val
		}
	}
	for field, val := range defaults.strs {
		if _, exists := pg.strs[field]; !exists && !pg.unset[field] {
			pg.strs[field] = val
		}
	}
	if pg.padding == nil && defaults.padding != nil && !pg.unset["Padding"] {
		pg.padding = defaults.padding
	}
}

func compileComponent(t *Theme, compName string, propMap map[string]propInfo, compObj map[string]any, topLevel *parsedGroup, loader imageLoader, loadedImages map[string]image.Image, pending *[]nineSliceEntry, colorMap map[string]sg.Color, gridDefs map[string]*nineGridDef, userVariants map[string]Variant, sprites map[string]SpriteRef) []error {
	var errs []error

	// Check for primary group.
	if _, ok := compObj["primary"]; !ok {
		errs = append(errs, fmt.Errorf("theme compile: %s: missing required group \"primary\"", compName))
		return errs
	}

	// Parse component-level properties. These come from:
	// 1. Bare (non-variant, non-underscore) keys at the component level
	// 2. The "_" underscore defaults block (overrides bare keys)
	configRaw := make(map[string]any)
	for key, val := range compObj {
		if _, isVariant := lookupVariant(key, userVariants); !isVariant && key != "_" {
			configRaw[key] = val
		}
	}
	if underscoreObj, ok := compObj["_"].(map[string]any); ok {
		for key, val := range underscoreObj {
			configRaw[key] = val
		}
	}
	configGroup, configErrs := parseProps(configRaw, propMap, compName+".", loader, loadedImages, pending, colorMap, gridDefs, sprites)
	errs = append(errs, configErrs...)
	// Merge top-level defaults into config-level.
	configGroup.mergeDefaults(topLevel)

	// Parse each variant group.
	groups := make(map[Variant]*parsedGroup)

	for groupKey, groupData := range compObj {
		v, ok := lookupVariant(groupKey, userVariants)
		if !ok {
			continue
		}
		groupObj, ok := groupData.(map[string]any)
		if !ok {
			errs = append(errs, fmt.Errorf("theme compile: %s.%s: expected object", compName, groupKey))
			continue
		}
		pg, pgErrs := parseProps(groupObj, propMap, compName+"."+groupKey+".", loader, loadedImages, pending, colorMap, gridDefs, sprites)
		errs = append(errs, pgErrs...)
		// Merge: variant ← component ← top-level.
		pg.mergeDefaults(configGroup)
		groups[v] = pg
	}

	if len(errs) > 0 {
		return errs
	}

	// Build groups and assign to theme.
	assignComponent(t, compName, groups)
	return nil
}

// compileUserComponent parses a user-defined component section, inferring
// property types from JSON values. Returns nil on fatal error (missing primary).
func compileUserComponent(compName string, compObj map[string]any, loader imageLoader, loadedImages map[string]image.Image, pending *[]nineSliceEntry, colorMap map[string]sg.Color, gridDefs map[string]*nineGridDef, userVariants map[string]Variant) (*UserConfig, []error) {
	var errs []error

	if _, ok := compObj["primary"]; !ok {
		return nil, []error{fmt.Errorf("theme compile: %s: missing required group \"primary\"", compName)}
	}

	config := &UserConfig{}

	for groupKey, groupData := range compObj {
		v, ok := lookupVariant(groupKey, userVariants)
		if !ok {
			continue
		}
		groupObj, ok := groupData.(map[string]any)
		if !ok {
			errs = append(errs, fmt.Errorf("theme compile: %s.%s: expected object", compName, groupKey))
			continue
		}
		group, groupErrs := parseUserGroup(compName+"."+groupKey, groupObj, loader, loadedImages, pending, colorMap, gridDefs)
		errs = append(errs, groupErrs...)
		if v == Primary {
			config.Primary = *group
		} else {
			config.Variants[v-1] = group
		}
	}

	return config, errs
}

// parseUserGroup parses a single variant group for a user-defined component,
// inferring property types from JSON values.
func parseUserGroup(path string, obj map[string]any, loader imageLoader, loadedImages map[string]image.Image, pending *[]nineSliceEntry, colorMap map[string]sg.Color, gridDefs map[string]*nineGridDef) (*UserGroup, []error) {
	var errs []error
	g := &UserGroup{}

	for key, val := range obj {
		propPath := path + "." + key

		// Insets: object with top/right/bottom/left numeric keys.
		if m, ok := val.(map[string]any); ok {
			if _, hasTop := m["top"]; hasTop {
				insets, err := compilePadding(propPath, val)
				if err != nil {
					errs = append(errs, err)
				} else {
					if g.paddings == nil {
						g.paddings = make(map[string]render.Insets)
					}
					g.paddings[key] = insets
				}
				continue
			}

			// Object with state keys: color, background, or per-state float.
			// Peek at values to determine kind.
			isBackground := false
			isStateFloat := false
			for _, sv := range m {
				if s, ok := sv.(string); ok {
					t := strings.TrimSpace(s)
					if strings.HasPrefix(t, "gradient(") || strings.HasPrefix(t, "gradientV(") || strings.HasPrefix(t, "gradientH(") {
						isBackground = true
						break
					}
				} else if isNumeric(sv) {
					isStateFloat = true
				}
			}

			if isBackground {
				bp, propErrs := compileBackgroundProperty(propPath, val, loader, loadedImages, pending, colorMap)
				errs = append(errs, propErrs...)
				if len(propErrs) == 0 {
					if g.backgrounds == nil {
						g.backgrounds = make(map[string]BackgroundProperty)
					}
					g.backgrounds[key] = bp
				}
			} else if isStateFloat {
				fp, propErrs := compileFloatProperty(propPath, val)
				errs = append(errs, propErrs...)
				if len(propErrs) == 0 {
					if g.stateFloats == nil {
						g.stateFloats = make(map[string]FloatProperty)
					}
					g.stateFloats[key] = fp
				}
			} else {
				cp, propErrs := compileColorProperty(propPath, val, colorMap)
				errs = append(errs, propErrs...)
				if len(propErrs) == 0 {
					if g.colors == nil {
						g.colors = make(map[string]ColorProperty)
					}
					g.colors[key] = cp
				}
			}
			continue
		}

		// Scalar number.
		if isNumeric(val) {
			if g.floats == nil {
				g.floats = make(map[string]float64)
			}
			g.floats[key] = toFloat(val)
			continue
		}
	}

	return g, errs
}

// ---------------------------------------------------------------------------
// Property compilation
// ---------------------------------------------------------------------------

func compileColorProperty(path string, data any, colorMap map[string]sg.Color) (ColorProperty, []error) {
	stateMap, ok := data.(map[string]any)
	if !ok {
		return ColorProperty{}, []error{fmt.Errorf("theme compile: %s: expected object with state keys", path)}
	}

	var errs []error
	var prop ColorProperty

	for stateKey, colorVal := range stateMap {
		state, ok := stateNames[stateKey]
		if !ok {
			continue // unknown state — skip
		}
		colorStr, ok := colorVal.(string)
		if !ok {
			errs = append(errs, fmt.Errorf("theme compile: %s.%s: expected color string", path, stateKey))
			continue
		}
		// Reject gradient values in color properties.
		trimmed := strings.TrimSpace(colorStr)
		if strings.HasPrefix(trimmed, "gradient(") || strings.HasPrefix(trimmed, "gradientV(") || strings.HasPrefix(trimmed, "gradientH(") {
			errs = append(errs, fmt.Errorf("theme compile: %s.%s: gradients are not supported for color properties, use backgroundColor", path, stateKey))
			continue
		}
		c, err := resolveColorString(colorStr, colorMap)
		if err != nil {
			errs = append(errs, fmt.Errorf("theme compile: %s.%s: %w", path, stateKey, err))
			continue
		}
		prop[state] = c
	}

	if len(errs) == 0 {
		ResolveColorFallbacks(&prop)
	}
	return prop, errs
}

// CompileFloatProperty is exported for testing from the root package.
func CompileFloatProperty(path string, data any) (FloatProperty, []error) {
	return compileFloatProperty(path, data)
}

func compileFloatProperty(path string, data any) (FloatProperty, []error) {
	// Bare number → uniform (all states same value).
	if isNumeric(data) {
		return NewFloatPropUniform(toFloat(data)), nil
	}

	stateMap, ok := data.(map[string]any)
	if !ok {
		return FloatProperty{}, []error{fmt.Errorf("theme compile: %s: expected number or object with state keys", path)}
	}

	var errs []error
	var prop FloatProperty
	for i := range prop {
		prop[i] = math.NaN()
	}
	for stateKey, val := range stateMap {
		state, ok := stateNames[stateKey]
		if !ok {
			continue
		}
		if !isNumeric(val) {
			errs = append(errs, fmt.Errorf("theme compile: %s.%s: expected number", path, stateKey))
			continue
		}
		prop[state] = toFloat(val)
	}

	if len(errs) == 0 {
		ResolveFloatFallbacks(&prop)
	}
	return prop, errs
}

func compileBackgroundProperty(path string, data any, loader imageLoader, loadedImages map[string]image.Image, pending *[]nineSliceEntry, colorMap map[string]sg.Color) (BackgroundProperty, []error) {
	stateMap, ok := data.(map[string]any)
	if !ok {
		return BackgroundProperty{}, []error{fmt.Errorf("theme compile: %s: expected object with state keys", path)}
	}

	var errs []error
	var prop BackgroundProperty

	for stateKey, bgVal := range stateMap {
		state, ok := stateNames[stateKey]
		if !ok {
			continue // unknown state — skip
		}

		colorStr, ok := bgVal.(string)
		if !ok {
			errs = append(errs, fmt.Errorf("theme compile: %s.%s: expected color string", path, stateKey))
			continue
		}

		// Check for gradient values before trying color parse.
		trimmed := strings.TrimSpace(colorStr)
		if strings.HasPrefix(trimmed, "gradient(") || strings.HasPrefix(trimmed, "gradientV(") || strings.HasPrefix(trimmed, "gradientH(") {
			g, err := parseGradient(trimmed, colorMap)
			if err != nil {
				errs = append(errs, fmt.Errorf("theme compile: %s.%s: %w", path, stateKey, err))
				continue
			}
			prop[state] = core.GradientBackground(g)
			continue
		}

		c, err := resolveColorString(colorStr, colorMap)
		if err != nil {
			errs = append(errs, fmt.Errorf("theme compile: %s.%s: %w", path, stateKey, err))
			continue
		}
		prop[state] = core.SolidBackground(c)
	}

	if len(errs) == 0 {
		ResolveBgFallbacks(&prop)
	}
	return prop, errs
}

// compileGridRefProperty compiles a backgroundGrid property.
// Values are state maps where each value is a string key referencing a nine-grid.
func compileGridRefProperty(path string, data any, gridDefs map[string]*nineGridDef, loader imageLoader, loadedImages map[string]image.Image, pending *[]nineSliceEntry) (BackgroundProperty, []error) {
	stateMap, ok := data.(map[string]any)
	if !ok {
		return BackgroundProperty{}, []error{fmt.Errorf("theme compile: %s: expected object with state keys", path)}
	}

	var errs []error
	var prop BackgroundProperty

	for stateKey, gridVal := range stateMap {
		state, ok := stateNames[stateKey]
		if !ok {
			continue
		}
		gridKey, ok := gridVal.(string)
		if !ok {
			errs = append(errs, fmt.Errorf("theme compile: %s.%s: expected nine-grid key string", path, stateKey))
			continue
		}
		def, ok := gridDefs[gridKey]
		if !ok {
			errs = append(errs, fmt.Errorf("theme compile: %s.%s: unknown nine-grid key %q", path, stateKey, gridKey))
			continue
		}

		// Load source image (deduplicating). When a region is specified,
		// crop to that sub-rect and stage under a unique key so the atlas
		// maps each nine-grid to its own texture region.
		imageKey := def.source
		if _, loaded := loadedImages[def.source]; !loaded {
			img, loadErr := loader.loadImage(def.source)
			if loadErr != nil {
				errs = append(errs, fmt.Errorf("theme compile: %s.%s: load image %q: %w", path, stateKey, def.source, loadErr))
				continue
			}
			loadedImages[def.source] = img
		}
		if def.region != nil {
			imageKey = fmt.Sprintf("%s#%s", def.source, gridKey)
			if _, loaded := loadedImages[imageKey]; !loaded {
				src := loadedImages[def.source]
				r := image.Rect(
					int(def.region.X), int(def.region.Y),
					int(def.region.X+def.region.Width), int(def.region.Y+def.region.Height),
				)
				cropped := image.NewRGBA(image.Rect(0, 0, r.Dx(), r.Dy()))
				for cy := r.Min.Y; cy < r.Max.Y; cy++ {
					for cx := r.Min.X; cx < r.Max.X; cx++ {
						cropped.Set(cx-r.Min.X, cy-r.Min.Y, src.At(cx, cy))
					}
				}
				loadedImages[imageKey] = cropped
			}
		}

		ns := &render.NineSlice{
			Insets:      def.insets,
			InnerRegion: def.innerRegion,
			CenterFill:  def.centerFill,
		}
		prop[state] = core.SliceBackground(ns)
		*pending = append(*pending, nineSliceEntry{imagePath: imageKey, slice: ns})
	}

	if len(errs) == 0 {
		ResolveBgFallbacks(&prop)
	}
	return prop, errs
}

// parseNineGrids parses the "nine-grids" top-level section into a map of
// nineGridDef entries. Source image dimensions are used for auto-slice when
// region is omitted.
func parseNineGrids(raw map[string]any, loader imageLoader, colorMap map[string]sg.Color) (map[string]*nineGridDef, []error) {
	gridsObj, ok := raw["nine-grids"].(map[string]any)
	if !ok {
		return nil, nil
	}

	var errs []error
	defs := make(map[string]*nineGridDef)

	for name, val := range gridsObj {
		entry, ok := val.(map[string]any)
		if !ok {
			errs = append(errs, fmt.Errorf("theme compile: nine-grids.%s: expected object", name))
			continue
		}

		// Parse source path (required).
		source, _ := entry["source"].(string)
		if source == "" {
			errs = append(errs, fmt.Errorf("theme compile: nine-grids.%s: missing or empty \"source\"", name))
			continue
		}

		def := &nineGridDef{source: source}

		// Parse optional region.
		if regionObj, ok := entry["region"].(map[string]any); ok {
			r := &render.Rect{}
			if v, ok := regionObj["x"]; ok && isNumeric(v) {
				r.X = toFloat(v)
			}
			if v, ok := regionObj["y"]; ok && isNumeric(v) {
				r.Y = toFloat(v)
			}
			if v, ok := regionObj["width"]; ok && isNumeric(v) {
				r.Width = toFloat(v)
			}
			if v, ok := regionObj["height"]; ok && isNumeric(v) {
				r.Height = toFloat(v)
			}
			def.region = r
		}

		// Parse innerRegion (required). Can be "auto" with auto-slice.
		innerVal := entry["innerRegion"]
		isAutoInner := false
		if innerStr, ok := innerVal.(string); ok && innerStr == "auto" {
			isAutoInner = true
		} else if innerObj, ok := innerVal.(map[string]any); ok {
			if v, ok := innerObj["x"]; ok && isNumeric(v) {
				def.innerRegion.X = toFloat(v)
			}
			if v, ok := innerObj["y"]; ok && isNumeric(v) {
				def.innerRegion.Y = toFloat(v)
			}
			if v, ok := innerObj["width"]; ok && isNumeric(v) {
				def.innerRegion.Width = toFloat(v)
			}
			if v, ok := innerObj["height"]; ok && isNumeric(v) {
				def.innerRegion.Height = toFloat(v)
			}
		} else {
			errs = append(errs, fmt.Errorf("theme compile: nine-grids.%s: missing \"innerRegion\"", name))
			continue
		}

		// Parse slices or auto-slice (mutually exclusive, one required).
		hasSlices := false
		hasAutoSlice := false

		if slicesObj, ok := entry["slices"].(map[string]any); ok {
			hasSlices = true
			// Derive insets from corner dimensions.
			if tl, ok := slicesObj["topLeftCorner"].(map[string]any); ok {
				if v, ok := tl["width"]; ok && isNumeric(v) {
					def.insets.Left = toFloat(v)
				}
				if v, ok := tl["height"]; ok && isNumeric(v) {
					def.insets.Top = toFloat(v)
				}
			}
			if br, ok := slicesObj["bottomRightCorner"].(map[string]any); ok {
				if v, ok := br["width"]; ok && isNumeric(v) {
					def.insets.Right = toFloat(v)
				}
				if v, ok := br["height"]; ok && isNumeric(v) {
					def.insets.Bottom = toFloat(v)
				}
			}
		}

		if autoObj, ok := entry["auto-slice"].(map[string]any); ok {
			hasAutoSlice = true
			if v, ok := autoObj["top"]; ok && isNumeric(v) {
				def.insets.Top = toFloat(v)
			}
			if v, ok := autoObj["right"]; ok && isNumeric(v) {
				def.insets.Right = toFloat(v)
			}
			if v, ok := autoObj["bottom"]; ok && isNumeric(v) {
				def.insets.Bottom = toFloat(v)
			}
			if v, ok := autoObj["left"]; ok && isNumeric(v) {
				def.insets.Left = toFloat(v)
			}
		}

		if !hasSlices && !hasAutoSlice {
			errs = append(errs, fmt.Errorf("theme compile: nine-grids.%s: requires either \"slices\" or \"auto-slice\"", name))
			continue
		}
		if hasSlices && hasAutoSlice {
			errs = append(errs, fmt.Errorf("theme compile: nine-grids.%s: cannot have both \"slices\" and \"auto-slice\"", name))
			continue
		}

		// Resolve auto innerRegion from auto-slice insets.
		if isAutoInner {
			if !hasAutoSlice {
				errs = append(errs, fmt.Errorf("theme compile: nine-grids.%s: innerRegion \"auto\" requires \"auto-slice\"", name))
				continue
			}
			// Need image dimensions to compute inner region.
			// For now, store placeholder; resolved after image loading.
			def.innerRegion = render.Rect{
				X:      def.insets.Left,
				Y:      def.insets.Top,
				Width:  -1, // sentinel: needs image dimensions
				Height: -1,
			}
		}

		if def.insets.Top <= 0 && def.insets.Right <= 0 && def.insets.Bottom <= 0 && def.insets.Left <= 0 {
			errs = append(errs, fmt.Errorf("theme compile: nine-grids.%s: insets must have at least one positive value", name))
			continue
		}

		// Parse optional centerFill gradient.
		if cfStr, ok := entry["centerFill"].(string); ok && cfStr != "" {
			trimmed := strings.TrimSpace(cfStr)
			if strings.HasPrefix(trimmed, "gradient(") || strings.HasPrefix(trimmed, "gradientV(") || strings.HasPrefix(trimmed, "gradientH(") {
				g, err := parseGradient(trimmed, colorMap)
				if err != nil {
					errs = append(errs, fmt.Errorf("theme compile: nine-grids.%s.centerFill: %w", name, err))
				} else {
					def.centerFill = g
				}
			} else {
				errs = append(errs, fmt.Errorf("theme compile: nine-grids.%s.centerFill: expected gradient value (gradient/gradientV/gradientH)", name))
			}
		}

		defs[name] = def
	}

	return defs, errs
}

// parseSprites parses the "sprites" section of the theme JSON.
// Each entry maps a sprite name to a source image path and pixel rectangle.
// Source images are loaded and deduplicated. The returned map contains fully
// resolved SpriteRef values with TextureRegions pointing into the loaded images.
func parseSprites(raw map[string]any, loader imageLoader) (map[string]SpriteRef, []error) {
	spritesObj, ok := raw["sprites"].(map[string]any)
	if !ok {
		return nil, nil
	}

	var errs []error
	result := make(map[string]SpriteRef)
	// Cache source images so multiple sprites from the same file are loaded once.
	srcImages := make(map[string]image.Image)

	for name, val := range spritesObj {
		obj, ok := val.(map[string]any)
		if !ok {
			errs = append(errs, fmt.Errorf("theme compile: sprites.%s: expected object", name))
			continue
		}

		src, _ := obj["src"].(string)
		if src == "" {
			errs = append(errs, fmt.Errorf("theme compile: sprites.%s: missing \"src\" field", name))
			continue
		}

		x, xOk := obj["x"]
		y, yOk := obj["y"]
		w, wOk := obj["w"]
		h, hOk := obj["h"]
		if !xOk || !yOk || !wOk || !hOk {
			errs = append(errs, fmt.Errorf("theme compile: sprites.%s: missing x/y/w/h fields", name))
			continue
		}
		if !isNumeric(x) || !isNumeric(y) || !isNumeric(w) || !isNumeric(h) {
			errs = append(errs, fmt.Errorf("theme compile: sprites.%s: x/y/w/h must be numbers", name))
			continue
		}

		sx := int(toFloat(x))
		sy := int(toFloat(y))
		sw := int(toFloat(w))
		sh := int(toFloat(h))

		// Load or reuse the source image.
		goSrcImg, exists := srcImages[src]
		if !exists {
			img, err := loader.loadImage(src)
			if err != nil {
				errs = append(errs, fmt.Errorf("theme compile: sprites.%s: %w", name, err))
				continue
			}
			goSrcImg = img
			srcImages[src] = goSrcImg
		}

		// Extract the sub-image for this sprite.
		subRect := image.Rect(sx, sy, sx+sw, sy+sh)
		type subImager interface {
			SubImage(r image.Rectangle) image.Image
		}
		var subImg image.Image
		if si, ok := goSrcImg.(subImager); ok {
			subImg = si.SubImage(subRect)
		} else {
			// Fallback: copy pixels manually.
			dst := image.NewNRGBA(image.Rect(0, 0, sw, sh))
			for dy := 0; dy < sh; dy++ {
				for dx := 0; dx < sw; dx++ {
					dst.Set(dx, dy, goSrcImg.At(sx+dx, sy+dy))
				}
			}
			subImg = dst
		}

		result[name] = SpriteRef{
			Image: engine.NewImageFromImage(subImg),
			Set:   true,
		}
	}

	return result, errs
}

func compilePadding(path string, data any) (render.Insets, error) {
	obj, ok := data.(map[string]any)
	if !ok {
		return render.Insets{}, fmt.Errorf("theme compile: %s: expected object with top/right/bottom/left", path)
	}
	var insets render.Insets
	if v, ok := obj["top"]; ok {
		if !isNumeric(v) {
			return render.Insets{}, fmt.Errorf("theme compile: %s.top: expected number", path)
		}
		insets.Top = toFloat(v)
	}
	if v, ok := obj["right"]; ok {
		if !isNumeric(v) {
			return render.Insets{}, fmt.Errorf("theme compile: %s.right: expected number", path)
		}
		insets.Right = toFloat(v)
	}
	if v, ok := obj["bottom"]; ok {
		if !isNumeric(v) {
			return render.Insets{}, fmt.Errorf("theme compile: %s.bottom: expected number", path)
		}
		insets.Bottom = toFloat(v)
	}
	if v, ok := obj["left"]; ok {
		if !isNumeric(v) {
			return render.Insets{}, fmt.Errorf("theme compile: %s.left: expected number", path)
		}
		insets.Left = toFloat(v)
	}
	return insets, nil
}

// ---------------------------------------------------------------------------
// Assign compiled groups to theme
// ---------------------------------------------------------------------------

type parsedGroup struct {
	colors      map[string]ColorProperty
	bgs         map[string]BackgroundProperty
	padding     *render.Insets
	floats      map[string]float64
	bools       map[string]bool
	stateFloats map[string]FloatProperty
	spriteRefs  map[string]SpriteRef
	strs        map[string]string
	unset       map[string]bool
}

func assignComponent(t *Theme, compName string, groups map[Variant]*parsedGroup) {
	type entry struct {
		cfg      any // pointer to Config[G]
		defaults any // G with non-zero defaults
	}
	ap := core.AutoPadding
	e, ok := map[string]entry{
		"button":            {&t.Button, ButtonGroup{Padding: ap}},
		"label":             {&t.Label, LabelGroup{}},
		"badge":             {&t.Badge, BadgeGroup{CornerRadius: -1, Padding: render.Insets{Top: 2, Right: 6, Bottom: 2, Left: 6}, DotSize: 8}},
		"toggle":            {&t.Toggle, ToggleGroup{CornerRadius: -1}},
		"checkbox":          {&t.Checkbox, CheckboxGroup{}},
		"radio":             {&t.Radio, RadioGroup{}},
		"textInput":         {&t.TextInput, TextInputGroup{Padding: ap, PlaceholderAlpha: 0.4}},
		"maskedInput":       {&t.MaskedInput, MaskedInputGroup{Padding: ap, PlaceholderAlpha: 0.4}},
		"inputField":        {&t.InputField, InputFieldGroup{LabelGap: 4, MessageGap: 3, LabelLeftGap: 8}},
		"searchBox":         {&t.SearchBox, SearchBoxGroup{Padding: ap, PlaceholderAlpha: 0.4}},
		"textArea":          {&t.TextArea, TextAreaGroup{Padding: ap}},
		"slider":            {&t.Slider, SliderGroup{}},
		"scrollBar":         {&t.ScrollBar, ScrollBarGroup{}},
		"meterBar":          {&t.MeterBar, MeterBarGroup{}},
		"panel":             {&t.Panel, PanelGroup{}},
		"navDrawer":         {&t.NavDrawer, NavDrawerGroup{AnimationDuration: 0.25}},
		"window":            {&t.Window, WindowGroup{}},
		"tabs":              {&t.Tabs, TabsGroup{ScrollArrowWidth: 24}},
		"list":              {&t.List, ListGroup{ItemPadding: ap}},
		"treeList":          {&t.TreeList, TreeListGroup{}},
		"tileList":          {&t.TileList, TileListGroup{}},
		"richText":          {&t.RichText, RichTextGroup{}},
		"optionRotator":     {&t.OptionRotator, OptionRotatorGroup{CornerRadius: -1, Padding: ap, Chevron: OptionRotatorChevronGroup{Width: 20, IconSize: 1.0}}},
		"toggleButtonBar":   {&t.ToggleButtonBar, ToggleButtonBarGroup{Padding: ap}},
		"tooltip":           {&t.Tooltip, TooltipGroup{Padding: ap}},
		"popover":           {&t.Popover, PopoverGroup{Padding: ap}},
		"menuBar":           {&t.MenuBar, MenuBarGroup{EntryPadding: render.Insets{Top: 4, Right: 10, Bottom: 4, Left: 10}, Height: 28, BorderWidth: 1}},
		"menuPopup":         {&t.MenuPopup, MenuPopupGroup{Padding: ap, ItemPadding: render.Insets{Top: 6, Right: 12, Bottom: 6, Left: 12}}},
		"select":            {&t.Select, SelectGroup{Padding: ap}},
		"dragHandle":        {&t.DragHandle, DragHandleGroup{}},
		"toast":             {&t.Toast, ToastGroup{Padding: ap}},
		"image":             {&t.Image, ImageGroup{}},
		"animatedImage":     {&t.AnimatedImage, AnimatedImageGroup{}},
		"colorPicker":       {&t.ColorPicker, ColorPickerGroup{Padding: ap}},
		"sortableList":      {&t.SortableList, SortableListGroup{}},
		"iconButton":        {&t.IconButton, IconButtonGroup{Padding: ap, LabelGap: 4}},
		"gradientEditor":    {&t.GradientEditor, GradientEditorGroup{Padding: ap, PreviewHeight: 40, PreviewSize: 140}},
		"statWeb":           {&t.StatWeb, StatWebGroup{PolygonStrokeWidth: 2, SpokeWidth: 1, GridLevels: 4, HandleRadius: 6, LabelOffset: 16}},
		"tag":               {&t.Tag, TagGroup{CornerRadius: -1, Padding: render.Insets{Top: 2, Right: 8, Bottom: 2, Left: 8}, RemoveButtonSize: 16, Gap: 4}},
		"tagBar":            {&t.TagBar, TagBarGroup{BorderWidth: 1, CornerRadius: 4, Padding: render.Insets{Top: 4, Right: 4, Bottom: 4, Left: 4}, Spacing: 4}},
		"accordion":         {&t.Accordion, AccordionGroup{HeaderHeight: 36, HeaderPadding: render.Insets{Left: 8, Right: 8}, HeaderIconSize: 16, HeaderIconGap: 6, ChevronSize: 12, ContentPadding: render.Insets{Left: 8, Right: 8, Top: 8, Bottom: 8}, DividerHeight: 1, AnimationDuration: 0.2}},
		"dataTable":         {&t.DataTable, DataTableGroup{HeaderBorderWidth: 1, CellPadding: 6, DividerWidth: 1, BorderWidth: 1, CornerRadius: 4}},
		"timePicker":        {&t.TimePicker, TimePickerGroup{BorderWidth: 1, CornerRadius: 4, ColumnWidth: 40, ColumnHeight: 60, ValueFontSize: 16, ArrowSize: 12, SeparatorWidth: 2, AmPmCornerRadius: 4}},
		"keybindInput":      {&t.KeybindInput, KeybindInputGroup{BorderWidth: 1, CornerRadius: 4, Padding: render.Insets{Left: 6, Right: 6, Top: 4, Bottom: 4}, KeyCapBorderWidth: 1, KeyCapCornerRadius: 3, KeyCapPadding: render.Insets{Left: 6, Right: 6, Top: 2, Bottom: 2}, ClearButtonSize: 16, UnsetText: "---", ListeningText: "Press any key..."}},
		"imageCropper":      {&t.ImageCropper, ImageCropperGroup{CropBorderWidth: 2, HandleSize: 10, HandleCornerRadius: 5, GridLineWidth: 1}},
		"toolBar":           {&t.ToolBar, ToolBarGroup{BorderWidth: 1, Padding: render.Insets{Top: 4, Right: 8, Bottom: 4, Left: 8}, Spacing: 4, SeparatorThickness: 1, SeparatorHeight: 0.6}},
		"calendarSelector":  {&t.CalendarSelector, CalendarSelectorGroup{BorderWidth: 1, CornerRadius: 4, DaySize: 30, DayCornerRadius: 4}},
		"sortableTreeList":  {&t.SortableTreeList, SortableTreeListGroup{}},
		"richTextEditor":    {&t.RichTextEditor, RichTextEditorGroup{BorderWidth: 1, CornerRadius: 4, ToolbarHeight: 36, ToolbarPadding: render.Insets{Top: 4, Right: 4, Bottom: 4, Left: 4}, ToolbarButtonSize: 28, ToolbarButtonGap: 4, ContentPadding: render.Insets{Top: 4, Right: 4, Bottom: 4, Left: 4}}},
		"propertyInspector": {&t.PropertyInspector, PropertyInspectorGroup{BorderWidth: 1, SearchBarHeight: 28, SearchBarGap: 4, GroupHeaderHeight: 28, RowHeight: 30, LabelWidth: 120}},
	}[compName]
	if ok {
		assignConfigAny(e.cfg, groups, e.defaults)
	}
}

// Helpers for building groups from parsed data and handling group fallbacks.

func getColor(pg *parsedGroup, field string, fallback ColorProperty) ColorProperty {
	if pg != nil {
		if cp, ok := pg.colors[field]; ok {
			return cp
		}
	}
	return fallback
}

func getBg(pg *parsedGroup, field string, fallback BackgroundProperty) BackgroundProperty {
	if pg != nil {
		if bp, ok := pg.bgs[field]; ok {
			return bp
		}
	}
	return fallback
}

func getPadding(pg *parsedGroup, fallback render.Insets) render.Insets {
	if pg != nil {
		if pg.padding != nil {
			return *pg.padding
		}
		// Explicitly unset via null — return zero, not the auto fallback.
		if pg.unset["Padding"] {
			return render.Insets{}
		}
	}
	return fallback
}

func getFloat(pg *parsedGroup, field string, fallback float64) float64 {
	if pg != nil {
		if f, ok := pg.floats[field]; ok {
			return f
		}
	}
	return fallback
}

func getBool(pg *parsedGroup, field string, fallback bool) bool {
	if pg != nil {
		if b, ok := pg.bools[field]; ok {
			return b
		}
	}
	return fallback
}

func getStr(pg *parsedGroup, field string, fallback string) string {
	if pg != nil {
		if s, ok := pg.strs[field]; ok {
			return s
		}
	}
	return fallback
}

func getStateFloat(pg *parsedGroup, field string, fallback FloatProperty) FloatProperty {
	if pg != nil {
		if fp, ok := pg.stateFloats[field]; ok {
			return fp
		}
	}
	return fallback
}

func getSpriteRef(pg *parsedGroup, field string, fallback SpriteRef) SpriteRef {
	if pg != nil {
		if sr, ok := pg.spriteRefs[field]; ok {
			return sr
		}
	}
	return fallback
}

// --- Reflection-based config assignment ---
//
// Type tokens used to dispatch field assignment to the correct getter.
var (
	typeColorProperty = reflect.TypeOf(ColorProperty{})
	typeBgProperty    = reflect.TypeOf(BackgroundProperty{})
	typeFloatProperty = reflect.TypeOf(FloatProperty{})
	typeSpriteRef     = reflect.TypeOf(SpriteRef{})
	typeInsets        = reflect.TypeOf(render.Insets{})
	typeShadowConfig  = reflect.TypeOf(ShadowConfig{})
)

// buildGroup populates a Group struct from a parsedGroup using reflection.
// For each exported field it dispatches to the appropriate getter based on type.
// The fallback value comes from the corresponding field in defaults.
// prefix is prepended to the field name when looking up keys in the parsedGroup
// (used for nested sub-structs like OptionRotatorChevronGroup).
func buildGroup(dst, defaults reflect.Value, pg *parsedGroup, prefix string) {
	t := dst.Type()
	for i := 0; i < t.NumField(); i++ {
		sf := t.Field(i)
		if !sf.IsExported() {
			continue
		}
		df := dst.Field(i)
		fb := defaults.Field(i)
		key := prefix + sf.Name

		switch sf.Type {
		case typeColorProperty:
			df.Set(reflect.ValueOf(getColor(pg, key, fb.Interface().(ColorProperty))))
		case typeBgProperty:
			df.Set(reflect.ValueOf(getBg(pg, key, fb.Interface().(BackgroundProperty))))
		case typeFloatProperty:
			df.Set(reflect.ValueOf(getStateFloat(pg, key, fb.Interface().(FloatProperty))))
		case typeSpriteRef:
			df.Set(reflect.ValueOf(getSpriteRef(pg, key, fb.Interface().(SpriteRef))))
		case typeInsets:
			df.Set(reflect.ValueOf(getPadding(pg, fb.Interface().(render.Insets))))
		case typeShadowConfig:
			// ShadowConfig is not parsed from theme JSON; keep the default.
			df.Set(fb)
		default:
			switch sf.Type.Kind() {
			case reflect.Float64:
				df.SetFloat(getFloat(pg, key, fb.Float()))
			case reflect.Bool:
				df.SetBool(getBool(pg, key, fb.Bool()))
			case reflect.String:
				df.SetString(getStr(pg, key, fb.String()))
			case reflect.Int:
				df.SetInt(int64(getFloat(pg, key, float64(fb.Int()))))
			case reflect.Struct:
				// Nested sub-struct (e.g. OptionRotatorChevronGroup).
				buildGroup(df, fb, pg, key[len(prefix):])
			}
		}
	}
}

// assignConfigAny populates a Config[G] using reflection, where cfgPtr is
// a pointer to a Config[G] value and defaults is the zero/default G value.
func assignConfigAny(cfgPtr any, groups map[Variant]*parsedGroup, defaults any) {
	cv := reflect.ValueOf(cfgPtr).Elem() // Config[G]
	priField := cv.FieldByName("Primary")
	varField := cv.FieldByName("Variants")
	dv := reflect.ValueOf(defaults)

	pri := groups[Primary]
	buildGroup(priField, dv, pri, "")

	for v := Variant(1); v < VariantCount; v++ {
		pg := groups[v]
		if pg == nil {
			continue
		}
		gp := reflect.New(priField.Type())
		buildGroup(gp.Elem(), priField, pg, "")
		varField.Index(int(v - 1)).Set(gp)
	}
}

// ---------------------------------------------------------------------------
// Utilities
// ---------------------------------------------------------------------------

// isUnsetValue returns true if the value represents an explicit unset/clear.
// Valid unset values: null (JSON nil), "nil", "none", "".
func isUnsetValue(v any) bool {
	if v == nil {
		return true
	}
	s, ok := v.(string)
	if !ok {
		return false
	}
	return s == "" || s == "nil" || s == "none"
}

// resolveColorString resolves a color string, handling $references via colorMap.
func resolveColorString(s string, colorMap map[string]sg.Color) (sg.Color, error) {
	if strings.HasPrefix(s, "$") {
		name := s[1:]
		c, ok := colorMap[name]
		if !ok {
			return sg.Color{}, fmt.Errorf("undefined color reference %q", s)
		}
		return c, nil
	}
	return markup.ParseColor(s)
}

// ---------------------------------------------------------------------------
// Gradient parsing
// ---------------------------------------------------------------------------

// splitTopLevelCommas splits a string on commas that are NOT inside
// parentheses. This allows gradient arguments like rgba(1,2,3,0.5) to
// be treated as a single token.
func splitTopLevelCommas(s string) []string {
	var parts []string
	depth := 0
	start := 0
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '(':
			depth++
		case ')':
			if depth > 0 {
				depth--
			}
		case ',':
			if depth == 0 {
				parts = append(parts, s[start:i])
				start = i + 1
			}
		}
	}
	parts = append(parts, s[start:])
	return parts
}

// parseGradient parses a gradient string in one of three formats:
//   - gradient(topLeft, topRight, bottomRight, bottomLeft) — 4 colors
//   - gradientV(top, bottom) — vertical: TL=TR=top, BL=BR=bottom
//   - gradientH(left, right) — horizontal: TL=BL=left, TR=BR=right
//
// Each color argument is resolved through resolveColorString (supports
// hex, rgb(), rgba(), $reference).
// ParseGradient is exported for testing from the root package.
func ParseGradient(s string, colorMap map[string]sg.Color) (*render.GradientColors, error) {
	return parseGradient(s, colorMap)
}

func parseGradient(s string, colorMap map[string]sg.Color) (*render.GradientColors, error) {
	s = strings.TrimSpace(s)

	var prefix string
	var expectedArgs int

	switch {
	case strings.HasPrefix(s, "gradient(") && strings.HasSuffix(s, ")"):
		prefix = "gradient("
		expectedArgs = 4
	case strings.HasPrefix(s, "gradientV(") && strings.HasSuffix(s, ")"):
		prefix = "gradientV("
		expectedArgs = 2
	case strings.HasPrefix(s, "gradientH(") && strings.HasSuffix(s, ")"):
		prefix = "gradientH("
		expectedArgs = 2
	default:
		return nil, fmt.Errorf("invalid gradient format %q", s)
	}

	inner := s[len(prefix) : len(s)-1]
	args := splitTopLevelCommas(inner)
	if len(args) != expectedArgs {
		return nil, fmt.Errorf("invalid gradient format %q: expected %d color arguments, got %d", s, expectedArgs, len(args))
	}

	colors := make([]sg.Color, len(args))
	for i, arg := range args {
		c, err := resolveColorString(strings.TrimSpace(arg), colorMap)
		if err != nil {
			return nil, fmt.Errorf("gradient argument %d: %w", i+1, err)
		}
		colors[i] = c
	}

	g := &render.GradientColors{}
	switch prefix {
	case "gradient(":
		g.TopLeft = colors[0]
		g.TopRight = colors[1]
		g.BottomRight = colors[2]
		g.BottomLeft = colors[3]
	case "gradientV(":
		g.TopLeft = colors[0]
		g.TopRight = colors[0]
		g.BottomLeft = colors[1]
		g.BottomRight = colors[1]
	case "gradientH(":
		g.TopLeft = colors[0]
		g.BottomLeft = colors[0]
		g.TopRight = colors[1]
		g.BottomRight = colors[1]
	}

	return g, nil
}

// isNumeric returns true if the value is a JSON number (float64 from json.Unmarshal).
func isNumeric(v any) bool {
	_, ok := v.(float64)
	return ok
}

// toFloat converts a JSON-unmarshaled value to float64.
func toFloat(v any) float64 {
	if v == nil {
		return 0
	}
	switch val := v.(type) {
	case float64:
		return val
	case int:
		return float64(val)
	case bool:
		if val {
			return 1
		}
		return 0
	case string:
		f, _ := strconv.ParseFloat(val, 64)
		return f
	}
	return 0
}

// CollectThemeImages extracts all nine-slice image paths from theme JSON
// without loading them. Use this for prebaked atlas tooling.
func CollectThemeImages(data []byte) ([]string, error) {
	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("theme compile: invalid JSON: %w", err)
	}

	seen := make(map[string]bool)
	var paths []string

	gridsRaw, ok := raw["nine-grids"].(map[string]any)
	if ok {
		for name, entry := range gridsRaw {
			obj, ok := entry.(map[string]any)
			if !ok {
				continue
			}
			src, _ := obj["source"].(string)
			if src == "" {
				return nil, fmt.Errorf("theme compile: nine-grids[%q]: missing source", name)
			}
			if !seen[src] {
				seen[src] = true
				paths = append(paths, src)
			}
		}
	}

	// Collect sprite source images.
	if spritesRaw, ok := raw["sprites"].(map[string]any); ok {
		for _, entry := range spritesRaw {
			obj, ok := entry.(map[string]any)
			if !ok {
				continue
			}
			src, _ := obj["src"].(string)
			if src != "" && !seen[src] {
				seen[src] = true
				paths = append(paths, src)
			}
		}
	}

	return paths, nil
}

// ---------------------------------------------------------------------------
// Binary theme loading
// ---------------------------------------------------------------------------

// LoadThemeBinary decodes a WUIT binary and compiles the theme.
// The atlas (if present) is decoded from the embedded PNG + JSON sections.
func LoadThemeBinary(data []byte) (*Theme, error) {
	decoded, err := DecodeThemeBinary(data)
	if err != nil {
		return nil, err
	}

	if len(decoded.AtlasPNG) > 0 && len(decoded.AtlasJSON) > 0 {
		return compileThemeWithBinaryAtlas(decoded.ThemeJSON, decoded.AtlasJSON, decoded.AtlasPNG)
	}

	// No atlas — compile from JSON alone.
	return compileThemeInternal(decoded.ThemeJSON, nilImageLoader{})
}

// compileThemeWithBinaryAtlas compiles a theme from JSON and uses the provided
// atlas PNG + JSON to create a prebaked atlas.
func compileThemeWithBinaryAtlas(themeJSON, atlasJSON, atlasPNG []byte) (*Theme, error) {
	img, _, err := image.Decode(bytes.NewReader(atlasPNG))
	if err != nil {
		return nil, fmt.Errorf("theme binary: decode atlas PNG: %w", err)
	}
	page := engine.NewImageFromImage(img)

	atlas, err := sg.LoadAtlas(atlasJSON, []engine.Image{page})
	if err != nil {
		return nil, fmt.Errorf("theme binary: load atlas: %w", err)
	}

	return compileThemeWithAtlas(themeJSON, atlas)
}

// compileThemeWithAtlas is like compileThemeInternal but injects a prebaked atlas.
func compileThemeWithAtlas(data []byte, prebakedAtlas *sg.Atlas) (*Theme, error) {
	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("theme compile: invalid JSON: %w", err)
	}

	var errs []error
	t := &Theme{}
	t.Atlas = prebakedAtlas

	// Parse colors section.
	colorMap := make(map[string]sg.Color)
	if colorsObj, ok := raw["colors"].(map[string]any); ok {
		for name, val := range colorsObj {
			colorStr, ok := val.(string)
			if !ok {
				errs = append(errs, fmt.Errorf("theme compile: colors.%s: expected color string", name))
				continue
			}
			c, err := markup.ParseColor(colorStr)
			if err != nil {
				errs = append(errs, fmt.Errorf("theme compile: colors.%s: %w", name, err))
				continue
			}
			colorMap[name] = c
		}
	}
	if len(errs) > 0 {
		return nil, errors.Join(errs...)
	}

	componentSource := raw
	if comps, ok := raw["components"].(map[string]any); ok {
		componentSource = comps
	}

	// Parse nine-grids — images are in the atlas, not on disk.
	gridDefs, gridErrs := parseNineGrids(raw, nilImageLoader{}, colorMap)
	errs = append(errs, gridErrs...)
	if gridDefs == nil {
		gridDefs = make(map[string]*nineGridDef)
	}

	// Build resolved sprites from the atlas.
	resolvedSprites := make(map[string]SpriteRef)
	if spritesObj, ok := raw["sprites"].(map[string]any); ok {
		for name := range spritesObj {
			region := prebakedAtlas.Region(name)
			if int(region.Page) < len(prebakedAtlas.Pages) {
				pg := prebakedAtlas.Pages[region.Page]
				sub := pg.SubImage(image.Rect(
					int(region.X), int(region.Y),
					int(region.X)+int(region.Width), int(region.Y)+int(region.Height),
				)).(engine.Image)
				resolvedSprites[name] = SpriteRef{Image: sub, Set: true}
			}
		}
	}

	// Parse variants.
	userVariants := make(map[string]Variant)
	nextCustom := Custom1
	if variantsArr, ok := raw["variants"].([]any); ok {
		for _, v := range variantsArr {
			name, ok := v.(string)
			if !ok {
				continue
			}
			if _, builtin := builtinVariantNames[name]; builtin {
				continue
			}
			if nextCustom >= VariantCount {
				break
			}
			userVariants[name] = nextCustom
			nextCustom++
		}
	}
	if len(userVariants) > 0 {
		t.CustomVariants = userVariants
	}
	if len(resolvedSprites) > 0 {
		t.Sprites = resolvedSprites
	}
	if len(errs) > 0 {
		return nil, errors.Join(errs...)
	}

	// Parse fonts section.
	if fontsObj, ok := raw["fonts"].(map[string]any); ok {
		t.Fonts = make(map[string]string, len(fontsObj))
		for role, val := range fontsObj {
			name, ok := val.(string)
			if !ok {
				errs = append(errs, fmt.Errorf("theme compile: fonts.%s: expected string", role))
				continue
			}
			t.Fonts[role] = name
		}
	}

	// Parse global underscore defaults.
	var globalUnderscoreRaw map[string]any
	if underscoreObj, ok := componentSource["_"].(map[string]any); ok {
		globalUnderscoreRaw = underscoreObj
	}

	topLevelByComp := make(map[string]*parsedGroup, len(componentPropertyMaps))
	for compName, propMap := range componentPropertyMaps {
		pg, pgErrs := parseProps(componentSource, propMap, "", nilImageLoader{}, nil, nil, colorMap, gridDefs, resolvedSprites)
		errs = append(errs, pgErrs...)
		if globalUnderscoreRaw != nil {
			uPg, uErrs := parseProps(globalUnderscoreRaw, propMap, "_.", nilImageLoader{}, nil, nil, colorMap, gridDefs, resolvedSprites)
			errs = append(errs, uErrs...)
			for field, val := range uPg.colors {
				pg.colors[field] = val
			}
			for field, val := range uPg.bgs {
				pg.bgs[field] = val
			}
			for field, val := range uPg.floats {
				pg.floats[field] = val
			}
			for field, val := range uPg.stateFloats {
				pg.stateFloats[field] = val
			}
			if uPg.padding != nil {
				pg.padding = uPg.padding
			}
			for field := range uPg.unset {
				pg.unset[field] = true
			}
		}
		topLevelByComp[compName] = pg
	}

	// Parse each component.
	var pending []nineSliceEntry
	loadedImages := make(map[string]image.Image)

	for compName, propMap := range componentPropertyMaps {
		compData, ok := componentSource[compName]
		if !ok {
			continue
		}
		compObj, ok := compData.(map[string]any)
		if !ok {
			errs = append(errs, fmt.Errorf("theme compile: %s: expected object", compName))
			continue
		}
		compErrs := compileComponent(t, compName, propMap, compObj, topLevelByComp[compName], nilImageLoader{}, loadedImages, &pending, colorMap, gridDefs, userVariants, resolvedSprites)
		errs = append(errs, compErrs...)
	}

	// Parse user-defined component sections.
	for compName, compData := range componentSource {
		if compName == "_" {
			continue
		}
		if _, isBuiltin := componentPropertyMaps[compName]; isBuiltin {
			continue
		}
		compObj, ok := compData.(map[string]any)
		if !ok {
			continue
		}
		config, compErrs := compileUserComponent(compName, compObj, nilImageLoader{}, loadedImages, &pending, colorMap, gridDefs, userVariants)
		errs = append(errs, compErrs...)
		if config != nil {
			if t.UserComponents == nil {
				t.UserComponents = make(map[string]*UserConfig)
			}
			t.UserComponents[compName] = config
		}
	}

	if len(errs) > 0 {
		return nil, errors.Join(errs...)
	}

	// Resolve nine-slice regions from the prebaked atlas.
	for i := range pending {
		region := prebakedAtlas.Region(pending[i].imagePath)
		pending[i].slice.Region = region
	}

	return t, nil
}
