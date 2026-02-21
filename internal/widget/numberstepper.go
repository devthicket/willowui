package widget

import (
	"fmt"
	"math"
	"strconv"

	"github.com/devthicket/willowui/internal/engine"
	"github.com/devthicket/willowui/internal/sg"
)

// NumberStepper is a numeric input that pairs a text field with decrement and
// increment buttons. Clicking the buttons adjusts the value by the configured
// step; the text field accepts direct entry. Min, max, and decimal-place
// display are all configurable.
type NumberStepper struct {
	Component
	decrementBtn *Button
	input        *TextInput
	incrementBtn *Button

	value    float64
	min, max float64
	step     float64
	pageStep float64 // 0 = auto (step × 10)
	decimals int

	ignoreWatch bool
	onChange    func(float64)
	boundRef    *Ref[float64]
	watch       WatchHandle
}

// Default NumberStepper dimensions.
const (
	DefaultNumberStepperWidth  = 120.0
	DefaultNumberStepperHeight = 28.0
)

// NewNumberStepper creates a NumberStepper with range (-∞, +∞), step 1, and
// zero decimal places.
func NewNumberStepper(name string, source *sg.FontFamily, displaySize float64) *NumberStepper {
	ns := &NumberStepper{
		step: 1,
		min:  math.Inf(-1),
		max:  math.Inf(1),
	}
	initComponent(&ns.Component, name)

	// Decrement button.
	ns.decrementBtn = NewButton(name+"-dec", "-", source, displaySize)
	ns.decrementBtn.SetVariant(Neutral)
	ns.decrementBtn.SetOnClick(func() {
		ns.SetValue(ns.value - ns.step)
	})
	ns.node.AddChild(ns.decrementBtn.Node())

	// Text input (center).
	ns.input = NewTextInput(name+"-input", source, displaySize)
	ns.input.SetOnChange(func(s string) {
		v, err := strconv.ParseFloat(s, 64)
		if err != nil || math.IsNaN(v) || math.IsInf(v, 0) {
			return
		}
		v = ns.clamp(v)
		ns.value = v
		DefaultScheduler.Flush()
		ns.fire(v)
	})
	ns.input.SetOnBlur(func() {
		// Reformat the displayed text to the proper decimal count and clamped value.
		ns.syncInputText()
	})
	ns.input.SetOnSubmit(func(s string) {
		v, err := strconv.ParseFloat(s, 64)
		if err != nil || math.IsNaN(v) || math.IsInf(v, 0) {
			ns.syncInputText() // revert to last valid value
			return
		}
		ns.SetValue(v)
	})
	ns.node.AddChild(ns.input.Node())

	// Increment button.
	ns.incrementBtn = NewButton(name+"-inc", "+", source, displaySize)
	ns.incrementBtn.SetVariant(Neutral)
	ns.incrementBtn.SetOnClick(func() {
		ns.SetValue(ns.value + ns.step)
	})
	ns.node.AddChild(ns.incrementBtn.Node())

	// Intercept Home/End in the text field so we can use them for min/max jumps.
	ns.input.SetKeyFilter(func(key engine.Key) bool {
		return key == engine.KeyHome || key == engine.KeyEnd
	})

	// Override the TextInput's HandleKey so Up/Down are intercepted by the
	// stepper rather than escaping to spatial nav. At boundary (min/max),
	// return false to allow spatial nav escape.
	ns.input.SetHandleKey(func(key engine.Key) bool {
		switch key {
		case engine.KeyUp:
			return ns.value < ns.max || math.IsInf(ns.max, 1)
		case engine.KeyDown:
			return ns.value > ns.min || math.IsInf(ns.min, -1)
		case engine.KeyLeft:
			return ns.input.GetCursorPos() > 0
		case engine.KeyRight:
			runes := []rune(ns.input.Value())
			return ns.input.GetCursorPos() < len(runes)
		}
		return false
	})

	// Keyboard navigation + select-all on focus + mouse wheel.
	var wasFocused bool
	ns.node.OnUpdate = func(_ float64) {
		focused := ns.input.IsFocused()

		// Select all text when the input first gains focus so the user can
		// immediately type a replacement value.
		if focused && !wasFocused {
			ns.input.SelectAll()
		}
		wasFocused = focused

		// Mouse wheel steps the value when hovering over any part of the stepper.
		if ns.enabled && ns.containsCursor() {
			_, wy := engine.Wheel()
			if wy > 0 {
				ns.SetValue(ns.value + ns.step)
			} else if wy < 0 {
				ns.SetValue(ns.value - ns.step)
			}
		}

		if !focused {
			return
		}
		im := DefaultInputManager
		pg := ns.pageStep
		if pg == 0 {
			pg = ns.step * 10
		}
		switch {
		case im.IsKeyJustAvailable(engine.KeyUp):
			ns.SetValue(ns.value + ns.step)
			im.Consume(engine.KeyUp)
		case im.IsKeyJustAvailable(engine.KeyDown):
			ns.SetValue(ns.value - ns.step)
			im.Consume(engine.KeyDown)
		case im.IsKeyJustAvailable(engine.KeyPageUp):
			ns.SetValue(ns.value + pg)
			im.Consume(engine.KeyPageUp)
		case im.IsKeyJustAvailable(engine.KeyPageDown):
			ns.SetValue(ns.value - pg)
			im.Consume(engine.KeyPageDown)
		case im.IsKeyJustAvailable(engine.KeyHome) && !math.IsInf(ns.max, 1):
			ns.SetValue(ns.max) // Home = upper limit (max)
			im.Consume(engine.KeyHome)
		case im.IsKeyJustAvailable(engine.KeyEnd) && !math.IsInf(ns.min, -1):
			ns.SetValue(ns.min) // End = lower limit (min)
			im.Consume(engine.KeyEnd)
		}
	}

	ns.SetSize(DefaultNumberStepperWidth, DefaultNumberStepperHeight)
	ns.syncInputText()
	return ns
}

// Value returns the current numeric value.
func (ns *NumberStepper) Value() float64 {
	return ns.value
}

// SetValue sets the value, clamping it to [min, max], updates the displayed
// text, and fires onChange and any bound Ref.
func (ns *NumberStepper) SetValue(v float64) {
	v = ns.clamp(v)
	ns.value = v
	ns.syncInputText()
	DefaultScheduler.Flush()
	ns.fire(v)
}

// SetMin sets the minimum allowed value. The current value is re-clamped.
func (ns *NumberStepper) SetMin(v float64) {
	ns.min = v
	ns.value = ns.clamp(ns.value)
	ns.syncInputText()
}

// SetMax sets the maximum allowed value. The current value is re-clamped.
func (ns *NumberStepper) SetMax(v float64) {
	ns.max = v
	ns.value = ns.clamp(ns.value)
	ns.syncInputText()
}

// SetStep sets the increment/decrement step size (default 1).
func (ns *NumberStepper) SetStep(v float64) {
	ns.step = v
}

// SetPageStep sets the step size used for Page Up / Page Down (default 0,
// which auto-computes as step × 10).
func (ns *NumberStepper) SetPageStep(v float64) {
	ns.pageStep = v
}

// SetDecimals sets how many decimal places to display and accept (default 0).
// Changing this also reformats the current value.
func (ns *NumberStepper) SetDecimals(n int) {
	if n < 0 {
		n = 0
	}
	ns.decimals = n
	ns.syncInputText()
}

// SetOnChange sets the callback invoked whenever the value changes.
func (ns *NumberStepper) SetOnChange(fn func(float64)) {
	ns.onChange = fn
}

// BindValue binds the stepper to a reactive Ref[float64] for two-way sync.
// External changes to the ref update the stepper; stepper changes update the ref.
func (ns *NumberStepper) BindValue(ref *Ref[float64]) {
	ns.watch.Stop()
	ns.boundRef = ref
	ns.SetValue(ref.Peek())
	ns.watch = WatchValue(ref, func(_, newVal float64) {
		if ns.ignoreWatch {
			return
		}
		ns.SetValue(newVal)
	})
}

// SetSize resizes the stepper and repositions its sub-components.
// The step buttons are square (height × height); the text field fills the rest.
func (ns *NumberStepper) SetSize(w, h float64) {
	ns.Width = w
	ns.Height = h
	ns.node.HitShape = sg.HitRect{X: 0, Y: 0, Width: w, Height: h}
	ns.MarkLayoutDirty()

	btnW := h
	inputW := w - 2*btnW
	if inputW < 0 {
		inputW = 0
	}

	ns.decrementBtn.SetSize(btnW, h)
	ns.decrementBtn.SetPosition(0, 0)

	ns.input.SetSize(inputW, h)
	ns.input.SetPosition(btnW, 0)

	ns.incrementBtn.SetSize(btnW, h)
	ns.incrementBtn.SetPosition(w-btnW, 0)
}

// SetEnabled enables or disables the stepper and all sub-components.
func (ns *NumberStepper) SetEnabled(v bool) {
	ns.Component.SetEnabled(v)
	ns.decrementBtn.SetEnabled(v)
	ns.input.SetEnabled(v)
	ns.incrementBtn.SetEnabled(v)
}

// Dispose stops reactive watches and disposes the component tree.
func (ns *NumberStepper) Dispose() {
	ns.watch.Stop()
	ns.Component.Dispose()
}

// DecrementButton returns the "-" button. Useful for styling.
func (ns *NumberStepper) DecrementButton() *Button { return ns.decrementBtn }

// IncrementButton returns the "+" button. Useful for styling.
func (ns *NumberStepper) IncrementButton() *Button { return ns.incrementBtn }

// InputField returns the text input. Useful for styling.
func (ns *NumberStepper) InputField() *TextInput { return ns.input }

// clamp restricts v to [min, max].
func (ns *NumberStepper) clamp(v float64) float64 {
	if v < ns.min {
		return ns.min
	}
	if v > ns.max {
		return ns.max
	}
	return v
}

// syncInputText formats the current value and writes it to the text field.
// Does not trigger onChange.
func (ns *NumberStepper) syncInputText() {
	ns.input.SetValue(fmt.Sprintf("%.*f", ns.decimals, ns.value))
}

// fire notifies the bound Ref and the onChange callback.
func (ns *NumberStepper) fire(v float64) {
	if ns.boundRef != nil {
		ns.ignoreWatch = true
		ns.boundRef.Set(v)
		DefaultScheduler.Flush()
		ns.ignoreWatch = false
	}
	if ns.onChange != nil {
		ns.onChange(v)
	}
}
