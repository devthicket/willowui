package widget

import (
	"time"

	"github.com/devthicket/willowui/internal/engine"
	"github.com/devthicket/willowui/internal/sg"
)

// defaultSearchIcon returns the search magnifier glyph from the default spritesheet.
func defaultSearchIcon() engine.Image { return IconSearch() }

// defaultSearchClearIcon returns the close X glyph from the default spritesheet.
func defaultSearchClearIcon() engine.Image { return IconCloseX() }

// ---------------------------------------------------------------------------
// SearchBox widget
// ---------------------------------------------------------------------------

// SearchBox is a search-oriented single-line input with a magnifier icon,
// optional clear button, debounce, and automatic reactive result population.
// It wraps a TextInput internally and adds search-specific behavior.
type SearchBox struct {
	Component

	input *TextInput // embedded text input (handles all text editing)

	iconNode *sg.Node    // magnifier glyph sprite
	clearBtn *IconButton // × clear button

	// AutoHeight, when true, causes SetSize to ignore the height argument and
	// compute it automatically from font size and padding.
	AutoHeight bool

	// Visual options.
	showSearchIcon  bool
	showClearButton bool

	// TextInput-level callbacks (wired through to the inner TextInput).
	onChange func(string)
	onSubmit func(string)
	onBlur   func()

	// Debounce: counts down in real seconds.
	debounce          time.Duration
	debounceRemaining time.Duration
	debounceActive    bool

	// Query settings.
	minQueryLength int
	searchOnChange bool
	searchOnSubmit bool

	// Search lifecycle callbacks.
	onClear        func()
	onSearchStart  func(string)
	onSearchFinish func(string, int)
	onSearchEmpty  func(string)

	// Type-erased search runner (set by SetSearchBoxFunc / SetSearchBoxIntoFunc).
	searchRunner func(query string)
	resultsCount int
	searching    bool
}

// NewSearchBox creates a SearchBox with the given name, font source, and display size.
// By default the search icon is shown and the clear button is enabled.
func NewSearchBox(name string, source *sg.FontFamily, displaySize float64) *SearchBox {
	sb := &SearchBox{
		showSearchIcon:  true,
		showClearButton: true,
		searchOnChange:  true,
		debounce:        150 * time.Millisecond,
	}
	initComponent(&sb.Component, name)

	sb.initBackground(name)
	sb.initBorder(name)

	// Create the inner TextInput. We'll manage its node as a direct child
	// of our own node so it renders inside our background/border.
	sb.input = NewTextInput(name+"-input", source, displaySize)
	// Mark as embedded so it skips its own background/border/focus ring.
	sb.input.embedded = true
	sb.input.hideBackground()
	sb.input.hideBorder()
	sb.input.hideFocusRing()
	sb.node.AddChild(sb.input.Node())

	// Wire input callbacks to SearchBox-level handlers.
	sb.input.SetOnChange(func(v string) {
		if sb.onChange != nil {
			sb.onChange(v)
		}
		sb.updateClearButtonVisibility()
		sb.scheduleSearch()
	})
	sb.input.SetOnSubmit(func(v string) {
		if sb.onSubmit != nil {
			sb.onSubmit(v)
		}
		if sb.searchOnSubmit {
			sb.debounceActive = false
			sb.runSearch()
		}
	})
	sb.input.onFocusChange = func(focused bool) {
		if !focused {
			if sb.onBlur != nil {
				sb.onBlur()
			}
		}
		sb.UpdateVisuals()
	}

	// Search icon (raw sprite — glyph tinted via node color).
	sb.iconNode = sg.NewSprite(name+"-icon", sg.TextureRegion{})
	sb.iconNode.SetCustomImage(defaultSearchIcon())
	sb.iconNode.SetScale(1, 1)
	sb.iconNode.Interactable = false
	sb.iconNode.SetVisible(sb.showSearchIcon)
	sb.node.AddChild(sb.iconNode)

	// Clear button (IconButton component — interactive, gets pointer cursor).
	sb.clearBtn = NewIconButton(name + "-clear")
	sb.clearBtn.embedded = true
	sb.clearBtn.hideBackground()
	sb.clearBtn.hideBorder()
	sb.clearBtn.hideFocusRing()
	sb.clearBtn.SetIconImage(defaultSearchClearIcon())
	sb.clearBtn.SetIconSize(10, 10)
	sb.clearBtn.SetOnClick(func() { sb.Clear() })
	sb.clearBtn.Node().SetVisible(false)
	sb.node.AddChild(sb.clearBtn.Node())

	// Default size.
	defaultW := 200.0
	var resolvedFont *sg.FontFamily
	if source != nil {
		resolvedFont = source
	}
	defaultH := displayLineHeight(resolvedFont, displaySize) + 12
	sb.SetSize(defaultW, defaultH)

	// Pointer handling on the SearchBox container focuses the inner input.
	sb.node.OnPointerDown(func(ctx sg.PointerContext) {
		if !sb.enabled {
			return
		}
		sb.pressed = true
		sb.bubbleActivation()
		DefaultFocusManager.SetFocus(&sb.input.Component)
	})

	sb.onVisualStateChange = func() { sb.UpdateVisuals() }
	sb.onThemeChange = func() { sb.UpdateVisuals() }

	sb.SetCursorShape(engine.CursorShapeText)

	sb.UpdateVisuals()

	sb.node.OnUpdate = func(dt float64) {
		sb.update(dt)
	}

	return sb
}

// ---------------------------------------------------------------------------
// Core input API (delegates to inner TextInput)
// ---------------------------------------------------------------------------

// Value returns the current query text.
func (sb *SearchBox) Value() string { return sb.input.Value() }

// SetValue sets the query text.
func (sb *SearchBox) SetValue(v string) {
	sb.input.SetValue(v)
	sb.updateClearButtonVisibility()
}

// BindValue binds the search box to a reactive Ref[string].
func (sb *SearchBox) BindValue(ref *Ref[string]) {
	sb.input.BindValue(ref)
	sb.updateClearButtonVisibility()
}

// SetPlaceholder sets the placeholder text.
func (sb *SearchBox) SetPlaceholder(v string) {
	sb.input.SetPlaceholder(v)
}

// SetSize sets the SearchBox dimensions.
func (sb *SearchBox) SetSize(w, h float64) {
	if sb.AutoHeight {
		pad := sb.effectivePadding()
		h = displayLineHeight(sb.input.font, sb.input.displaySize) + pad.Top + pad.Bottom
	}
	sb.Width = w
	sb.Height = h
	sb.resizeBackground(w, h)
	sb.resizeBorder(w, h)
	sb.node.HitShape = sg.HitRect{X: 0, Y: 0, Width: w, Height: h}
	sb.repositionInternals()
	sb.MarkLayoutDirty()
}

// SetWidth sets only the width; height is computed automatically.
func (sb *SearchBox) SetWidth(w float64) {
	pad := sb.effectivePadding()
	h := displayLineHeight(sb.input.font, sb.input.displaySize) + pad.Top + pad.Bottom
	sb.SetSize(w, h)
}

// SetShowSearchIcon toggles the magnifier icon.
func (sb *SearchBox) SetShowSearchIcon(v bool) {
	sb.showSearchIcon = v
	sb.iconNode.SetVisible(v)
	sb.repositionInternals()
}

// SetShowClearButton toggles the clear button.
func (sb *SearchBox) SetShowClearButton(v bool) {
	sb.showClearButton = v
	sb.updateClearButtonVisibility()
	sb.repositionInternals()
}

// Clear empties the query, hides the clear button, clears results, and fires OnClear.
func (sb *SearchBox) Clear() {
	sb.SetValue("")
	if sb.searchRunner != nil {
		sb.resultsCount = 0
		sb.searchRunner("")
	}
	if sb.onClear != nil {
		sb.onClear()
	}
	sb.debounceActive = false
}

// SetOnChange sets the callback fired when the query text changes.
func (sb *SearchBox) SetOnChange(fn func(query string)) { sb.onChange = fn }

// SetOnSubmit sets the callback fired on Enter.
func (sb *SearchBox) SetOnSubmit(fn func(query string)) { sb.onSubmit = fn }

// SetOnClear sets the callback fired when the clear button is pressed.
func (sb *SearchBox) SetOnClear(fn func()) { sb.onClear = fn }

// SetOnBlur sets the callback fired when the field loses focus.
func (sb *SearchBox) SetOnBlur(fn func()) { sb.onBlur = fn }

// Input returns the inner TextInput for advanced configuration.
func (sb *SearchBox) Input() *TextInput { return sb.input }

// ---------------------------------------------------------------------------
// Debounce API
// ---------------------------------------------------------------------------

// SetDebounce sets the debounce duration. Default is 150ms.
func (sb *SearchBox) SetDebounce(d time.Duration) { sb.debounce = d }

// Debounce returns the current debounce duration.
func (sb *SearchBox) Debounce() time.Duration { return sb.debounce }

// TriggerSearchNow bypasses the debounce and runs the search immediately.
func (sb *SearchBox) TriggerSearchNow() {
	sb.debounceActive = false
	sb.runSearch()
}

// CancelPendingSearch cancels any pending debounced search without running it.
// Useful after programmatically setting the value (e.g. accepting an autocomplete
// suggestion) to prevent the dropdown from reopening.
func (sb *SearchBox) CancelPendingSearch() {
	sb.debounceActive = false
}

// ---------------------------------------------------------------------------
// Query settings
// ---------------------------------------------------------------------------

// SetMinQueryLength sets the minimum character count before automatic search.
func (sb *SearchBox) SetMinQueryLength(n int) { sb.minQueryLength = n }

// MinQueryLength returns the minimum query length.
func (sb *SearchBox) MinQueryLength() int { return sb.minQueryLength }

// SetSearchOnChange controls whether typing triggers automatic search.
func (sb *SearchBox) SetSearchOnChange(v bool) { sb.searchOnChange = v }

// SetSearchOnSubmit controls whether Enter triggers automatic search.
func (sb *SearchBox) SetSearchOnSubmit(v bool) { sb.searchOnSubmit = v }

// ---------------------------------------------------------------------------
// Lifecycle callbacks
// ---------------------------------------------------------------------------

// SetOnSearchStart sets the callback fired before each search execution.
func (sb *SearchBox) SetOnSearchStart(fn func(query string)) { sb.onSearchStart = fn }

// SetOnSearchFinish sets the callback fired after each search with result count.
func (sb *SearchBox) SetOnSearchFinish(fn func(query string, count int)) { sb.onSearchFinish = fn }

// SetOnSearchEmpty sets the callback fired when search returns zero results.
func (sb *SearchBox) SetOnSearchEmpty(fn func(query string)) { sb.onSearchEmpty = fn }

// ---------------------------------------------------------------------------
// State helpers
// ---------------------------------------------------------------------------

// ResultsCount returns the number of results from the last search.
func (sb *SearchBox) ResultsCount() int { return sb.resultsCount }

// IsSearching returns true if a search is currently running.
func (sb *SearchBox) IsSearching() bool { return sb.searching }

// setSearchRunner sets the type-erased search runner. Called by package-level
// generic helpers SetSearchBoxFunc / SetSearchBoxIntoFunc.
func (sb *SearchBox) setSearchRunner(fn func(query string)) {
	sb.searchRunner = fn
}

// ---------------------------------------------------------------------------
// Internal: search execution
// ---------------------------------------------------------------------------

func (sb *SearchBox) runSearch() {
	if sb.searchRunner == nil {
		return
	}
	q := sb.input.Value()
	if sb.minQueryLength > 0 && len([]rune(q)) < sb.minQueryLength {
		sb.searchRunner("")
		sb.resultsCount = 0
		return
	}
	sb.searching = true
	if sb.onSearchStart != nil {
		sb.onSearchStart(q)
	}
	sb.searchRunner(q)
	sb.searching = false
	if sb.onSearchFinish != nil {
		sb.onSearchFinish(q, sb.resultsCount)
	}
	if sb.resultsCount == 0 && sb.onSearchEmpty != nil {
		sb.onSearchEmpty(q)
	}
}

func (sb *SearchBox) scheduleSearch() {
	if !sb.searchOnChange || sb.searchRunner == nil {
		return
	}
	if sb.debounce <= 0 {
		sb.runSearch()
		return
	}
	sb.debounceRemaining = sb.debounce
	sb.debounceActive = true
}

// ---------------------------------------------------------------------------
// Delegation helpers removed — text editing handled by inner TextInput
// ---------------------------------------------------------------------------

// HasSelection returns true when text is selected.
func (sb *SearchBox) HasSelection() bool { return sb.input.HasSelection() }

// SelectAll selects the entire text.
func (sb *SearchBox) SelectAll() { sb.input.SelectAll() }

// InsertText inserts text at the cursor.
func (sb *SearchBox) InsertText(s string) { sb.input.InsertText(s) }

// DeleteBack deletes the character before the cursor.
func (sb *SearchBox) DeleteBack() { sb.input.DeleteBack() }

// DeleteForward deletes the character after the cursor.
func (sb *SearchBox) DeleteForward() { sb.input.DeleteForward() }

// Submit fires OnSubmit and optionally triggers search.
func (sb *SearchBox) Submit() { sb.input.Submit() }

// SetEnabled overrides Component.SetEnabled to also update visuals.
func (sb *SearchBox) SetEnabled(v bool) {
	sb.Component.SetEnabled(v)
	sb.input.SetEnabled(v)
	sb.UpdateVisuals()
}

// ---------------------------------------------------------------------------
// Layout helpers
// ---------------------------------------------------------------------------

const sbIconSize = 16.0
const sbClearBtnAreaSize = 20.0

func (sb *SearchBox) effectiveSearchBoxGroup() *SearchBoxGroup {
	t := sb.EffectiveTheme()
	sbGroup := t.SearchBox.Group(sb.Variant())
	zero := ColorProperty{}
	if sbGroup.Background[StateDefault].Type == BgNone && sbGroup.TextColor == zero {
		ti := t.TextInput.Group(sb.Variant())
		merged := *sbGroup
		merged.Background = ti.Background
		merged.TextColor = ti.TextColor
		merged.CursorColor = ti.CursorColor
		merged.SelectionColor = ti.SelectionColor
		merged.Border = ti.Border
		merged.BorderWidth = ti.BorderWidth
		merged.CornerRadius = ti.CornerRadius
		merged.PlaceholderAlpha = ti.PlaceholderAlpha
		merged.Padding = ti.Padding
		merged.FocusColor = ti.FocusColor
		merged.FocusRingWidth = ti.FocusRingWidth
		// Derive icon/clear colors from the text color when not explicitly themed.
		if merged.IconColor == zero {
			merged.IconColor = ti.TextColor
		}
		if merged.ClearButtonColor == zero {
			merged.ClearButtonColor = ti.TextColor
		}
		return &merged
	}
	return sbGroup
}

func (sb *SearchBox) effectivePadding() Insets {
	group := sb.effectiveSearchBoxGroup()
	return resolveAutoInsets(group.Padding, defaultTextInputPadding)
}

func (sb *SearchBox) sbIconGap() float64 {
	group := sb.effectiveSearchBoxGroup()
	if group.IconGap > 0 {
		return group.IconGap
	}
	return 6.0
}

func (sb *SearchBox) contentLeftOffset() float64 {
	pad := sb.effectivePadding()
	if sb.showSearchIcon {
		gap := sb.sbIconGap()
		return pad.Left + gap + sbIconSize + gap
	}
	return pad.Left
}

func (sb *SearchBox) contentRightInset() float64 {
	pad := sb.effectivePadding()
	if sb.showClearButton {
		return pad.Right + sbClearBtnAreaSize
	}
	return pad.Right
}

func (sb *SearchBox) repositionInternals() {
	pad := sb.effectivePadding()
	w, h := sb.Width, sb.Height
	lineH := displayLineHeight(sb.input.font, sb.input.displaySize)

	// Position search icon (Image component).
	if sb.showSearchIcon {
		gap := sb.sbIconGap()
		iconX := pad.Left + gap/2
		iconY := (h - sbIconSize) / 2
		sb.iconNode.SetSize(sbIconSize, sbIconSize)
		sb.iconNode.SetPosition(iconX, iconY)
	}

	// Position the inner TextInput to fill the content area.
	leftOff := sb.contentLeftOffset()
	rightInset := sb.contentRightInset()
	innerW := w - leftOff - rightInset
	if innerW < 0 {
		innerW = 0
	}

	// The inner TextInput has its own background/border hidden.
	// Position it so its content area aligns with our content area.
	sb.input.Node().SetPosition(leftOff, 0)
	// Resize the input to fit the available area without its own padding.
	sb.input.SetSize(innerW, h)

	// Override the inner input's content position to remove its own padding
	// since we're managing padding at the SearchBox level.
	vCenter := (h - lineH) / 2
	sb.input.content.SetPosition(0, vCenter)
	sb.input.updateContentMask(innerW, lineH)

	// Position clear button (IconButton component).
	clearX := w - pad.Right - sbClearBtnAreaSize
	clearY := (h - sbClearBtnAreaSize) / 2
	sb.clearBtn.applySize(sbClearBtnAreaSize, sbClearBtnAreaSize)
	sb.clearBtn.Node().SetPosition(clearX, clearY)

	sb.MarkLayoutDirty()
}

func (sb *SearchBox) updateClearButtonVisibility() {
	hasText := sb.input.Value() != ""
	sb.clearBtn.Node().SetVisible(sb.showClearButton && hasText)
}

// ---------------------------------------------------------------------------
// Visuals
// ---------------------------------------------------------------------------

// UpdateVisuals applies theme colors based on state.
func (sb *SearchBox) UpdateVisuals() {
	// Use the inner input's focus state for our visual state.
	sb.state = computeState(sb.enabled, sb.input.focused, sb.hovered, sb.pressed)
	group := sb.effectiveSearchBoxGroup()
	sb.applyCornerRadius(group.CornerRadius)
	bg := group.Background.Resolve(sb.state)
	sb.applyBackground(bg)
	sb.applyBorder(group.Border.Resolve(sb.state), group.BorderWidth, bg)

	// Icon tint.
	sb.iconNode.SetColor(group.IconColor.Resolve(sb.state))

	// Clear button icon tint.
	sb.clearBtn.icon.SetColor(group.ClearButtonColor.Resolve(sb.state))

	sb.repositionInternals()
	sb.applyFocusRing(group.FocusColor.Resolve(sb.state), group.FocusRingWidth)
}

// update ticks the debounce timer each frame.
func (sb *SearchBox) update(dt float64) {
	if sb.debounceActive && sb.debounce > 0 {
		sb.debounceRemaining -= time.Duration(float64(time.Second) * dt)
		if sb.debounceRemaining <= 0 {
			sb.debounceActive = false
			sb.runSearch()
		}
	}

	sb.UpdateVisuals()
}

// Dispose stops reactive watches and disposes the component tree.
func (sb *SearchBox) Dispose() {
	sb.input.Dispose()
	sb.Component.Dispose()
}

// ---------------------------------------------------------------------------
// Test helpers
// ---------------------------------------------------------------------------

// GetPlaceholder returns the placeholder string. Used for testing.
func (sb *SearchBox) GetPlaceholder() string { return sb.input.GetPlaceholder() }

// GetCursorPos returns the current cursor rune position. Used for testing.
func (sb *SearchBox) GetCursorPos() int { return sb.input.GetCursorPos() }

// TextNode returns the willow text node. Used for testing.
func (sb *SearchBox) TextNode() *sg.Node { return sb.input.TextNode() }

// ClearVisible returns whether the clear button is visible. Used for testing.
func (sb *SearchBox) ClearVisible() bool {
	if sb.clearBtn == nil {
		return false
	}
	return sb.clearBtn.Node().Visible()
}

// ---------------------------------------------------------------------------
// Package-level generic helpers (Go doesn't allow generic methods)
// ---------------------------------------------------------------------------

// SetSearchBoxFunc sets the automatic search callback that returns a result slice.
func SetSearchBoxFunc[T any](sb *SearchBox, results *Array[T], fn func(query string) []T) {
	sb.setSearchRunner(func(query string) {
		if query == "" {
			results.Clear()
			sb.resultsCount = 0
			return
		}
		out := fn(query)
		results.Batch(func() {
			results.Clear()
			for _, v := range out {
				results.Push(v)
			}
		})
		sb.resultsCount = results.Len()
	})
}

// SetSearchBoxIntoFunc sets the advanced search callback that directly mutates the result array.
func SetSearchBoxIntoFunc[T any](sb *SearchBox, results *Array[T], fn func(query string, results *Array[T])) {
	sb.setSearchRunner(func(query string) {
		if query == "" {
			results.Clear()
			sb.resultsCount = 0
			return
		}
		results.Batch(func() {
			fn(query, results)
		})
		sb.resultsCount = results.Len()
	})
}
