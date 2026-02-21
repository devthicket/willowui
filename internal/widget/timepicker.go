package widget

import (
	"fmt"

	"github.com/devthicket/willowui/internal/sg"
)

// TimeValue represents a time-of-day as hour (0-23), minute, and second.
// It is comparable and suitable for use with Ref[TimeValue].
type TimeValue struct {
	Hour, Minute, Second int
}

// TimeFormat selects 12-hour or 24-hour time display.
type TimeFormat int

const (
	TimeFormat24h TimeFormat = iota
	TimeFormat12h
)

// TimePicker is a compact hour/minute/second picker using up/down stepper
// columns separated by ":" dividers, with an optional AM/PM toggle in 12h mode.
type TimePicker struct {
	Component
	source      *sg.FontFamily
	font        *sg.FontFamily
	displaySize float64

	hour, minute, second int
	showSeconds          bool
	format               TimeFormat
	isPM                 bool // only used in 12h mode

	// Sub-components: columns + separator labels.
	hourCol   *timeColumn
	minuteCol *timeColumn
	secondCol *timeColumn
	sep1      *sg.Node // ":" between hour and minute
	sep2      *sg.Node // ":" between minute and second
	ampmBtn   *Button

	onTimeChanged func(h, m, s int)

	// Reactive binding.
	timeRef   *Ref[TimeValue]
	timeWatch WatchHandle
}

// timeColumn is an internal up/down stepper column for a single time field.
type timeColumn struct {
	container *sg.Node
	upBtn     *Button
	downBtn   *Button
	label     *sg.Node
	labelText *sg.Node
}

// Default TimePicker dimensions.
const (
	DefaultTimePickerWidth  = 180.0
	DefaultTimePickerHeight = 80.0
)

// NewTimePicker creates a TimePicker with default 24h format, no seconds.
func NewTimePicker(name string, source *sg.FontFamily, displaySize float64) *TimePicker {
	var font *sg.FontFamily
	if source != nil {
		font = source
	}
	tp := &TimePicker{
		source:      source,
		font:        font,
		displaySize: displaySize,
	}
	initComponent(&tp.Component, name)
	tp.node.Interactable = true

	// Create hour column.
	tp.hourCol = tp.newTimeColumn(name+"-hour", 0, 23, func() int { return tp.displayHour() }, func(v int) {
		if tp.format == TimeFormat12h {
			tp.setDisplayHour(v)
		} else {
			tp.hour = v
		}
		tp.fireChanged()
	})
	tp.node.AddChild(tp.hourCol.container)

	// Separator ":".
	tp.sep1 = sg.NewText(name+"-sep1", ":", font)
	tp.sep1.TextBlock.FontSize = displaySize * 1.2
	tp.node.AddChild(tp.sep1)

	// Create minute column.
	tp.minuteCol = tp.newTimeColumn(name+"-minute", 0, 59, func() int { return tp.minute }, func(v int) {
		tp.minute = v
		tp.fireChanged()
	})
	tp.node.AddChild(tp.minuteCol.container)

	// Separator ":" (between minute and second).
	tp.sep2 = sg.NewText(name+"-sep2", ":", font)
	tp.sep2.TextBlock.FontSize = displaySize * 1.2
	tp.sep2.SetVisible(false)
	tp.node.AddChild(tp.sep2)

	// Create second column.
	tp.secondCol = tp.newTimeColumn(name+"-second", 0, 59, func() int { return tp.second }, func(v int) {
		tp.second = v
		tp.fireChanged()
	})
	tp.secondCol.container.SetVisible(false)
	tp.node.AddChild(tp.secondCol.container)

	// AM/PM button (hidden by default in 24h mode).
	tp.ampmBtn = NewButton(name+"-ampm", "AM", source, displaySize*0.9)
	tp.ampmBtn.SetVariant(Secondary)
	tp.ampmBtn.SetOnClick(func() {
		tp.toggleAmPm()
		tp.syncHourFrom12h()
		tp.fireChanged()
	})
	tp.ampmBtn.Node().SetVisible(false)
	tp.node.AddChild(tp.ampmBtn.Node())

	tp.onVisualStateChange = func() { tp.UpdateVisuals() }
	tp.onThemeChange = func() { tp.UpdateVisuals() }

	tp.SetSize(DefaultTimePickerWidth, DefaultTimePickerHeight)
	tp.syncAllLabels()
	tp.UpdateVisuals()
	return tp
}

// newTimeColumn creates an up/down stepper column for a time field.
func (tp *TimePicker) newTimeColumn(name string, min, max int, getter func() int, setter func(int)) *timeColumn {
	tc := &timeColumn{}

	tc.container = sg.NewContainer(name)
	tc.container.Interactable = true

	// Up button.
	tc.upBtn = NewButton(name+"-up", "+", tp.source, tp.displaySize*0.8)
	tc.upBtn.SetVariant(Neutral)
	tc.upBtn.SetOnClick(func() {
		v := getter()
		v++
		if v > max {
			v = min
		}
		setter(v)
		tp.syncAllLabels()
	})
	tc.container.AddChild(tc.upBtn.Node())

	// Value label (sprite background + text).
	tc.label = sg.NewSprite(name+"-bg", sg.TextureRegion{})
	tc.container.AddChild(tc.label)

	tc.labelText = sg.NewText(name+"-val", "00", tp.font)
	tc.labelText.TextBlock.FontSize = tp.displaySize * 1.25
	tc.container.AddChild(tc.labelText)

	// Down button.
	tc.downBtn = NewButton(name+"-down", "-", tp.source, tp.displaySize*0.8)
	tc.downBtn.SetVariant(Neutral)
	tc.downBtn.SetOnClick(func() {
		v := getter()
		v--
		if v < min {
			v = max
		}
		setter(v)
		tp.syncAllLabels()
	})
	tc.container.AddChild(tc.downBtn.Node())

	return tc
}

// SetTime sets the time. h is in 24h range (0-23).
func (tp *TimePicker) SetTime(h, m, s int) {
	tp.hour = clampInt(h, 0, 23)
	tp.minute = clampInt(m, 0, 59)
	tp.second = clampInt(s, 0, 59)
	if tp.format == TimeFormat12h {
		tp.isPM = tp.hour >= 12
		if tp.isPM {
			tp.ampmBtn.SetText("PM")
		} else {
			tp.ampmBtn.SetText("AM")
		}
	}
	tp.syncAllLabels()
	tp.updateColumnRanges()
}

// Hour returns the current hour in 24h format (0-23).
func (tp *TimePicker) Hour() int { return tp.hour }

// Minute returns the current minute (0-59).
func (tp *TimePicker) Minute() int { return tp.minute }

// Second returns the current second (0-59).
func (tp *TimePicker) Second() int { return tp.second }

// SetShowSeconds shows or hides the seconds column.
func (tp *TimePicker) SetShowSeconds(v bool) {
	tp.showSeconds = v
	tp.secondCol.container.SetVisible(v)
	tp.sep2.SetVisible(v)
	tp.relayout()
}

// SetFormat sets 12h or 24h display format.
func (tp *TimePicker) SetFormat(format TimeFormat) {
	tp.format = format
	if format == TimeFormat12h {
		tp.isPM = tp.hour >= 12
		if tp.isPM {
			tp.ampmBtn.SetText("PM")
		} else {
			tp.ampmBtn.SetText("AM")
		}
		tp.ampmBtn.Node().SetVisible(true)
	} else {
		tp.ampmBtn.Node().SetVisible(false)
	}
	tp.updateColumnRanges()
	tp.syncAllLabels()
	tp.relayout()
}

// SetOnTimeChanged sets the callback invoked when any time field changes.
func (tp *TimePicker) SetOnTimeChanged(fn func(h, m, s int)) {
	tp.onTimeChanged = fn
}

// BindTime binds the time value to a reactive Ref. Changes to the Ref update
// the picker, and user interactions update the Ref.
func (tp *TimePicker) BindTime(ref *Ref[TimeValue]) {
	tp.timeWatch.Stop()
	tp.timeRef = ref
	v := ref.Peek()
	tp.SetTime(v.Hour, v.Minute, v.Second)
	tp.timeWatch = WatchValue(ref, func(_, newVal TimeValue) {
		tp.SetTime(newVal.Hour, newVal.Minute, newVal.Second)
	})
}

// SetSize resizes the picker and repositions sub-components.
func (tp *TimePicker) SetSize(w, h float64) {
	tp.Width = w
	tp.Height = h
	tp.node.HitShape = sg.HitRect{X: 0, Y: 0, Width: w, Height: h}
	tp.MarkLayoutDirty()
	tp.relayout()
}

// SetEnabled enables or disables the picker and all sub-components.
func (tp *TimePicker) SetEnabled(v bool) {
	tp.Component.SetEnabled(v)
	tp.hourCol.setEnabled(v)
	tp.minuteCol.setEnabled(v)
	tp.secondCol.setEnabled(v)
	tp.ampmBtn.SetEnabled(v)
}

// UpdateVisuals applies theme colors to all sub-components.
func (tp *TimePicker) UpdateVisuals() {
	group := tp.EffectiveTheme().TimePicker.Group(tp.Variant())

	// Separator colors — text nodes use TextBlock.Color.
	sepColor := group.SeparatorColor.Resolve(tp.state)
	tp.sep1.SetTextColor(sepColor)
	tp.sep2.SetTextColor(sepColor)

	// Column backgrounds and value text.
	colBg := group.ColumnBackground.Resolve(tp.state)
	valColor := group.ValueTextColor.Resolve(tp.state)

	for _, col := range []*timeColumn{tp.hourCol, tp.minuteCol, tp.secondCol} {
		col.label.SetColor(colBg.Color)
		col.labelText.SetTextColor(valColor)
	}

	tp.MarkDrawDirty()
}

// displayHour returns the hour value for display based on format.
func (tp *TimePicker) displayHour() int {
	if tp.format == TimeFormat12h {
		h := tp.hour % 12
		if h == 0 {
			h = 12
		}
		return h
	}
	return tp.hour
}

// setDisplayHour converts a 12h display hour back to 24h and stores it.
func (tp *TimePicker) setDisplayHour(dispH int) {
	if tp.format == TimeFormat12h {
		if dispH > 12 {
			dispH = 1
		}
		if dispH < 1 {
			dispH = 12
		}
		if tp.isPM {
			if dispH == 12 {
				tp.hour = 12
			} else {
				tp.hour = dispH + 12
			}
		} else {
			if dispH == 12 {
				tp.hour = 0
			} else {
				tp.hour = dispH
			}
		}
	} else {
		tp.hour = dispH
	}
}

// syncHourFrom12h recalculates the 24h hour from the current display hour and AM/PM state.
func (tp *TimePicker) syncHourFrom12h() {
	if tp.format != TimeFormat12h {
		return
	}
	dispH := tp.hour % 12
	if dispH == 0 {
		dispH = 12
	}
	tp.setDisplayHour(dispH)
}

// updateColumnRanges updates the up/down button callbacks with correct wrap ranges.
func (tp *TimePicker) updateColumnRanges() {
	var hourMin, hourMax int
	if tp.format == TimeFormat12h {
		hourMin = 1
		hourMax = 12
	} else {
		hourMin = 0
		hourMax = 23
	}

	tp.hourCol.upBtn.SetOnClick(func() {
		old := tp.displayHour()
		v := old + 1
		if v > hourMax {
			v = hourMin
		}
		// 12h: crossing 11→12 toggles AM/PM (11 AM→12 PM, 11 PM→12 AM).
		if tp.format == TimeFormat12h && old == 11 && v == 12 {
			tp.toggleAmPm()
		}
		tp.setDisplayHour(v)
		tp.syncAllLabels()
		tp.fireChanged()
	})
	tp.hourCol.downBtn.SetOnClick(func() {
		old := tp.displayHour()
		v := old - 1
		if v < hourMin {
			v = hourMax
		}
		// 12h: crossing 12→11 toggles AM/PM (12 PM→11 AM, 12 AM→11 PM).
		if tp.format == TimeFormat12h && old == 12 && v == 11 {
			tp.toggleAmPm()
		}
		tp.setDisplayHour(v)
		tp.syncAllLabels()
		tp.fireChanged()
	})
}

// syncAllLabels updates all column display text.
func (tp *TimePicker) syncAllLabels() {
	setTextContent(tp.hourCol.labelText, fmt.Sprintf("%02d", tp.displayHour()))
	setTextContent(tp.minuteCol.labelText, fmt.Sprintf("%02d", tp.minute))
	setTextContent(tp.secondCol.labelText, fmt.Sprintf("%02d", tp.second))
}

// setTextContent updates the text content of a willow text node.
func setTextContent(node *sg.Node, text string) {
	node.SetContent(text)
}

// relayout positions all sub-components within the current bounds.
func (tp *TimePicker) relayout() {
	w, h := tp.Width, tp.Height

	// Count visible columns.
	numCols := 2 // hour + minute always visible
	if tp.showSeconds {
		numCols = 3
	}
	hasAmPm := tp.format == TimeFormat12h

	// Measure ":" text for precise centering (matches the 1.2x font size).
	sepTextW, sepTextH := measureDisplay(tp.font, ":", tp.displaySize*1.2)
	sepWidth := sepTextW + tp.displaySize*0.5 // padding around ":"

	ampmWidth := 0.0
	if hasAmPm {
		ampmWidth = tp.displaySize * 2.8
	}

	totalSepWidth := sepWidth * float64(numCols-1)
	availWidth := w - totalSepWidth - ampmWidth
	colWidth := availWidth / float64(numCols)
	if colWidth < 0 {
		colWidth = 0
	}

	gap := 2.0
	btnHeight := h * 0.24
	valHeight := h - 2*btnHeight - 2*gap

	// Measure "00" to center value text properly (matches the 1.25x font size).
	valTextW, valTextH := measureDisplay(tp.font, "00", tp.displaySize*1.25)

	x := 0.0

	// Helper to layout a column.
	layoutCol := func(col *timeColumn, xPos float64) float64 {
		col.container.SetPosition(xPos, 0)

		col.upBtn.SetSize(colWidth, btnHeight)
		col.upBtn.SetPosition(0, 0)

		valTop := btnHeight + gap
		col.label.SetScale(colWidth, valHeight)
		col.label.SetPosition(0, valTop)

		// Center text in value area.
		tx := (colWidth - valTextW) / 2
		ty := valTop + (valHeight-valTextH)/2
		col.labelText.SetPosition(tx, ty)

		col.downBtn.SetSize(colWidth, btnHeight)
		col.downBtn.SetPosition(0, valTop+valHeight+gap)

		return xPos + colWidth
	}

	// Hour column.
	x = layoutCol(tp.hourCol, x)

	// Separator 1: center ":" vertically and horizontally in the gap.
	tp.sep1.SetPosition(x+(sepWidth-sepTextW)/2, (h-sepTextH)/2)
	x += sepWidth

	// Minute column.
	x = layoutCol(tp.minuteCol, x)

	// Separator 2 and second column.
	if tp.showSeconds {
		tp.sep2.SetPosition(x+(sepWidth-sepTextW)/2, (h-sepTextH)/2)
		x += sepWidth
		x = layoutCol(tp.secondCol, x)
	}

	// AM/PM button.
	if hasAmPm {
		ampmH := h * 0.35
		tp.ampmBtn.SetSize(ampmWidth, ampmH)
		tp.ampmBtn.SetPosition(x+4, (h-ampmH)/2)
	}
}

// toggleAmPm flips the AM/PM state and updates the button label.
func (tp *TimePicker) toggleAmPm() {
	tp.isPM = !tp.isPM
	if tp.isPM {
		tp.ampmBtn.SetText("PM")
	} else {
		tp.ampmBtn.SetText("AM")
	}
}

// fireChanged notifies the callback with 24h hour values.
func (tp *TimePicker) fireChanged() {
	if tp.timeRef != nil {
		tp.timeRef.Set(TimeValue{Hour: tp.hour, Minute: tp.minute, Second: tp.second})
	}
	if tp.onTimeChanged != nil {
		tp.onTimeChanged(tp.hour, tp.minute, tp.second)
	}
}

// clampInt clamps v to [lo, hi].
func clampInt(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

// setEnabled enables/disables a time column's buttons.
func (tc *timeColumn) setEnabled(v bool) {
	tc.upBtn.SetEnabled(v)
	tc.downBtn.SetEnabled(v)
}

// Dispose cleans up the picker.
func (tp *TimePicker) Dispose() {
	tp.timeWatch.Stop()
	tp.Component.Dispose()
}

// AmPmButton returns the AM/PM toggle button. Useful for styling.
func (tp *TimePicker) AmPmButton() *Button { return tp.ampmBtn }

// HourUpButton returns the hour column's up button. Useful for testing.
func (tp *TimePicker) HourUpButton() *Button { return tp.hourCol.upBtn }

// HourDownButton returns the hour column's down button.
func (tp *TimePicker) HourDownButton() *Button { return tp.hourCol.downBtn }

// MinuteUpButton returns the minute column's up button.
func (tp *TimePicker) MinuteUpButton() *Button { return tp.minuteCol.upBtn }

// MinuteDownButton returns the minute column's down button.
func (tp *TimePicker) MinuteDownButton() *Button { return tp.minuteCol.downBtn }

// SecondUpButton returns the second column's up button.
func (tp *TimePicker) SecondUpButton() *Button { return tp.secondCol.upBtn }

// SecondDownButton returns the second column's down button.
func (tp *TimePicker) SecondDownButton() *Button { return tp.secondCol.downBtn }
