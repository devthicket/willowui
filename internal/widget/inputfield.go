package widget

import (
	"github.com/devthicket/willowui/internal/sg"
)

// ---------------------------------------------------------------------------
// Supporting types
// ---------------------------------------------------------------------------

// LabelPosition controls where the field label is placed relative to the input.
type LabelPosition int

const (
	LabelAbove LabelPosition = iota // default: label above the input
	LabelLeft                       // label to the left of the input
)

// ValidationState indicates the current validation status of an InputField.
type ValidationState int

const (
	ValidationNone    ValidationState = iota // no validation state
	ValidationError                          // error state
	ValidationWarning                        // warning state
	ValidationSuccess                        // success state
)

// ---------------------------------------------------------------------------
// InputField widget
// ---------------------------------------------------------------------------

// InputField is a labeled text input widget that combines a Label, a TextInput,
// and an optional validation message into a single composable unit.
type InputField struct {
	Component

	input    *TextInput
	labelLbl *Label
	msgLbl   *Label

	font        *sg.FontFamily
	displaySize float64

	labelText      string
	labelPosition  LabelPosition
	required       bool
	requiredMarker string

	validationState ValidationState
	validationMsg   string
}

// NewInputField creates a new InputField with a label, text input, and validation message.
func NewInputField(name string, source *sg.FontFamily, displaySize float64) *InputField {
	var font *sg.FontFamily
	if source != nil {
		font = source
	}
	f := &InputField{
		font:           font,
		displaySize:    displaySize,
		requiredMarker: "*",
	}
	initComponent(&f.Component, name)

	// Label (hidden until text is set).
	f.labelLbl = NewLabel(name+"-label", "", source, displaySize)
	f.AddChild(&f.labelLbl.Component)
	f.labelLbl.Node().SetVisible(false)

	// TextInput.
	f.input = NewTextInput(name+"-input", source, displaySize)
	f.AddChild(&f.input.Component)

	// Validation message label (hidden by default).
	// Use a smaller font size for the message.
	msgSize := displaySize * 0.85
	if msgSize < 10 {
		msgSize = 10
	}
	f.msgLbl = NewLabel(name+"-msg", "", source, msgSize)
	f.AddChild(&f.msgLbl.Component)
	f.msgLbl.Node().SetVisible(false)

	f.onThemeChange = func() { f.updateLayout() }

	// Default size.
	f.SetSize(200, 32)

	return f
}

// ---------------------------------------------------------------------------
// Label API
// ---------------------------------------------------------------------------

// SetLabel sets the field label text.
func (f *InputField) SetLabel(text string) {
	f.labelText = text
	f.updateLabelText()
	f.updateLayout()
}

// SetLabelPosition sets whether the label is above or to the left of the input.
func (f *InputField) SetLabelPosition(pos LabelPosition) {
	f.labelPosition = pos
	f.updateLayout()
}

// SetRequired marks the field as required and appends the required marker.
func (f *InputField) SetRequired(v bool) {
	f.required = v
	f.updateLabelText()
	f.updateLayout()
}

// SetRequiredMarker sets the character(s) appended to the label when required. Default "*".
func (f *InputField) SetRequiredMarker(s string) {
	f.requiredMarker = s
	f.updateLabelText()
	f.updateLayout()
}

func (f *InputField) updateLabelText() {
	text := f.labelText
	if f.required && f.requiredMarker != "" {
		text += " " + f.requiredMarker
	}
	hasLabel := text != ""
	f.labelLbl.SetText(text)
	f.labelLbl.Node().SetVisible(hasLabel)

	// Apply colors from theme.
	group := f.effectiveInputFieldGroup()
	labelColor := group.LabelColor.Resolve(StateDefault)
	if labelColor != (sg.Color{}) {
		f.labelLbl.SetColor(labelColor)
	}
	if f.required {
		reqColor := group.RequiredColor.Resolve(StateDefault)
		if reqColor != (sg.Color{}) {
			// For simplicity, color the whole label. A richer approach
			// would color only the marker, but that requires markup.
			f.labelLbl.SetColor(reqColor)
		}
	}
}

// ---------------------------------------------------------------------------
// Validation API
// ---------------------------------------------------------------------------

// SetValidationState sets the validation state and updates the input border color.
func (f *InputField) SetValidationState(state ValidationState) {
	f.validationState = state
	f.applyValidationVisuals()
}

// SetValidationMessage sets the validation message text.
func (f *InputField) SetValidationMessage(msg string) {
	f.validationMsg = msg
	hasMsg := msg != ""
	f.msgLbl.SetText(msg)
	f.msgLbl.Node().SetVisible(hasMsg)
	f.applyValidationVisuals()
	f.updateLayout()
}

// ClearValidation resets validation state and hides the message.
func (f *InputField) ClearValidation() {
	f.validationState = ValidationNone
	f.validationMsg = ""
	f.msgLbl.SetText("")
	f.msgLbl.Node().SetVisible(false)
	f.applyValidationVisuals()
	f.updateLayout()
}

func (f *InputField) applyValidationVisuals() {
	group := f.effectiveInputFieldGroup()

	// Apply message color based on validation state.
	var msgColor sg.Color
	switch f.validationState {
	case ValidationError:
		msgColor = group.ErrorColor.Resolve(StateDefault)
	case ValidationWarning:
		msgColor = group.WarningColor.Resolve(StateDefault)
	case ValidationSuccess:
		msgColor = group.SuccessColor.Resolve(StateDefault)
	default:
		msgColor = group.LabelColor.Resolve(StateDefault)
	}
	if msgColor != (sg.Color{}) {
		f.msgLbl.SetColor(msgColor)
	}

	// Override the inner TextInput's border color when in a validation state.
	// We do this by setting the input's variant to match the state color.
	// For now, directly manipulate the border via the input's border nodes.
	if f.validationState != ValidationNone && msgColor != (sg.Color{}) {
		f.input.applyBorderColor(msgColor)
	} else {
		// Reset to normal theme border.
		f.input.UpdateVisuals()
	}
}

// ---------------------------------------------------------------------------
// TextInput proxy methods
// ---------------------------------------------------------------------------

// Input returns the inner TextInput for advanced configuration.
func (f *InputField) Input() *TextInput { return f.input }

// SetValue sets the text content.
func (f *InputField) SetValue(v string) { f.input.SetValue(v) }

// Value returns the current text.
func (f *InputField) Value() string { return f.input.Value() }

// SetPlaceholder sets the placeholder text.
func (f *InputField) SetPlaceholder(v string) { f.input.SetPlaceholder(v) }

// BindValue binds the input to a reactive Ref[string].
func (f *InputField) BindValue(ref *Ref[string]) { f.input.BindValue(ref) }

// SetOnChange sets the callback fired when text changes.
func (f *InputField) SetOnChange(fn func(string)) { f.input.SetOnChange(fn) }

// SetOnSubmit sets the callback fired on Enter.
func (f *InputField) SetOnSubmit(fn func(string)) { f.input.SetOnSubmit(fn) }

// SetOnBlur sets the callback fired when the input loses focus.
func (f *InputField) SetOnBlur(fn func()) { f.input.SetOnBlur(fn) }

// SetMaxLength limits the number of characters.
func (f *InputField) SetMaxLength(n int) { f.input.SetMaxLength(n) }

// SetReadOnly sets the input to read-only mode.
func (f *InputField) SetReadOnly(v bool) { f.input.SetEnabled(!v) }

// SetEnabled enables or disables the entire field.
func (f *InputField) SetEnabled(v bool) {
	f.Component.SetEnabled(v)
	f.input.SetEnabled(v)
}

// ---------------------------------------------------------------------------
// Size
// ---------------------------------------------------------------------------

// SetSize sets the input width and height. The label and message auto-size.
func (f *InputField) SetSize(w, h float64) {
	f.Width = w
	f.Height = h
	f.input.SetSize(w, h)
	f.updateLayout()
}

// SetWidth sets only the width; height is computed from font size.
func (f *InputField) SetWidth(w float64) {
	f.input.SetWidth(w)
	f.Width = w
	f.Height = f.input.Height
	f.updateLayout()
}

// ---------------------------------------------------------------------------
// Layout
// ---------------------------------------------------------------------------

func (f *InputField) effectiveInputFieldGroup() *InputFieldGroup {
	return f.EffectiveTheme().InputField.Group(f.Variant())
}

func (f *InputField) updateLayout() {
	group := f.effectiveInputFieldGroup()
	labelGap := group.LabelGap
	if labelGap <= 0 {
		labelGap = 4
	}
	msgGap := group.MessageGap
	if msgGap <= 0 {
		msgGap = 3
	}
	labelLeftGap := group.LabelLeftGap
	if labelLeftGap <= 0 {
		labelLeftGap = 8
	}

	hasLabel := f.labelText != ""
	hasMsg := f.validationMsg != ""

	inputW := f.input.Width
	inputH := f.input.Height

	if f.labelPosition == LabelLeft && hasLabel {
		// Label to the left, input to the right.
		labelW := f.labelLbl.Width
		labelH := f.labelLbl.Height

		f.labelLbl.Component.OffsetX = 0
		f.labelLbl.Component.OffsetY = (inputH - labelH) / 2

		f.input.Component.OffsetX = labelW + labelLeftGap
		f.input.Component.OffsetY = 0

		totalW := labelW + labelLeftGap + inputW
		totalH := inputH

		if hasMsg {
			f.msgLbl.Component.OffsetX = labelW + labelLeftGap
			f.msgLbl.Component.OffsetY = inputH + msgGap
			totalH = inputH + msgGap + f.msgLbl.Height
		}

		f.Width = totalW
		f.Height = totalH
	} else {
		// Label above (default).
		y := 0.0

		if hasLabel {
			f.labelLbl.Component.OffsetX = 0
			f.labelLbl.Component.OffsetY = y
			y += f.labelLbl.Height + labelGap
		}

		f.input.Component.OffsetX = 0
		f.input.Component.OffsetY = y
		y += inputH

		if hasMsg {
			y += msgGap
			f.msgLbl.Component.OffsetX = 0
			f.msgLbl.Component.OffsetY = y
			y += f.msgLbl.Height
		}

		f.Height = y
	}

	f.MarkLayoutDirty()
}

// Dispose cleans up the input field and its children.
func (f *InputField) Dispose() {
	f.input.Dispose()
	f.labelLbl.Dispose()
	f.msgLbl.Dispose()
	f.Component.Dispose()
}
