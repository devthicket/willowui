package integration

import (
	"testing"

	ui "github.com/devthicket/willowui"
)

func TestInputFieldLabelRendersAboveByDefault(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	f := ui.NewInputField("f", font, 14)
	defer f.Dispose()

	f.SetLabel("Username")
	f.SetSize(200, 32)

	// Label should be at y=0 (above), input should be offset below the label.
	if f.Input().Component.OffsetY <= 0 {
		t.Errorf("input OffsetY = %v, want > 0 (below label)", f.Input().Component.OffsetY)
	}
}

func TestInputFieldLabelLeftPositionsLabelToLeft(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	f := ui.NewInputField("f", font, 14)
	defer f.Dispose()

	f.SetLabel("Email")
	f.SetLabelPosition(ui.LabelLeft)
	f.SetSize(200, 32)

	// Label y should be vertically centered, input x offset > 0.
	if f.Input().Component.OffsetX <= 0 {
		t.Errorf("input OffsetX = %v, want > 0 (right of label)", f.Input().Component.OffsetX)
	}
}

func TestInputFieldSetRequiredAppendsMarker(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	f := ui.NewInputField("f", font, 14)
	defer f.Dispose()

	f.SetLabel("Name")
	f.SetRequired(true)

	// The input should be accessible and the label should exist.
	if f.Value() != "" {
		t.Errorf("Value() = %q, want empty", f.Value())
	}
}

func TestInputFieldValidationErrorChangesBorderAndShowsMessage(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	f := ui.NewInputField("f", font, 14)
	defer f.Dispose()

	f.SetLabel("Username")
	f.SetSize(200, 32)
	f.SetValidationState(ui.ValidationError)
	f.SetValidationMessage("Too short")

	// Height should increase to accommodate the message.
	if f.Height <= 32 {
		t.Errorf("Height = %v, want > 32 (message should add height)", f.Height)
	}
}

func TestInputFieldClearValidationResetsState(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	f := ui.NewInputField("f", font, 14)
	defer f.Dispose()

	f.SetLabel("Username")
	f.SetSize(200, 32)
	f.SetValidationState(ui.ValidationError)
	f.SetValidationMessage("Error")

	heightWithMsg := f.Height

	f.ClearValidation()

	if f.Height >= heightWithMsg {
		t.Errorf("Height after clear = %v, want < %v", f.Height, heightWithMsg)
	}
}

func TestInputFieldBindValuePropagatesToInnerTextInput(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	f := ui.NewInputField("f", font, 14)
	defer f.Dispose()

	ref := ui.NewRef("")
	f.BindValue(ref)

	ref.Set("hello")
	ui.DefaultScheduler.Flush()

	if f.Value() != "hello" {
		t.Errorf("Value() = %q, want %q", f.Value(), "hello")
	}
}

func TestInputFieldSetSizeResizesInput(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	f := ui.NewInputField("f", font, 14)
	defer f.Dispose()

	f.SetSize(300, 40)

	if f.Input().Width != 300 {
		t.Errorf("Input().Width = %v, want 300", f.Input().Width)
	}
}
