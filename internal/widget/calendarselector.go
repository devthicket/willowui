package widget

import (
	"fmt"
	"time"

	"github.com/devthicket/willowui/internal/sg"
)

// CalendarSelector is a month-grid date picker with prev/next month navigation.
// It supports inline mode (always visible) and popup mode (trigger button opens
// a floating calendar overlay via PopoverManager).
type CalendarSelector struct {
	Component
	source      *sg.FontFamily
	font        *sg.FontFamily
	displaySize float64

	selectedDate time.Time
	viewYear     int
	viewMonth    time.Month

	minDate   *time.Time
	maxDate   *time.Time
	popupMode bool
	popupOpen bool

	// Sub-components.
	prevBtn    *Button
	nextBtn    *Button
	monthLabel *sg.Node // text node for "March 2026"
	headerBg   *sg.Node // header background strip
	weekdaySep *sg.Node // separator line below weekday row

	weekdayLabels [7]*sg.Node
	dayCells      [42]*calendarDayCell // 6 rows x 7 cols
	hoveredCell   int                  // index of hovered cell, -1 = none

	// Inline mode container: holds the full calendar grid.
	calendarRoot *sg.Node

	// Popup mode: trigger button + popover.
	triggerBtn *Button
	popover    *Popover

	onDateSelected func(time.Time)

	// Reactive binding.
	dateRef   *Ref[time.Time]
	dateWatch WatchHandle
}

// calendarDayCell is a single day cell in the calendar grid.
type calendarDayCell struct {
	container  *sg.Node
	bg         *sg.Node
	todayRing  *sg.Node // border ring for today (when not selected)
	label      *sg.Node
	day        int       // 0 = empty cell
	date       time.Time // the actual date this cell represents
	inMonth    bool      // true if this day is in the currently viewed month
	disabled   bool      // true if outside min/max range
	isToday    bool
	isSelected bool
}

// Default CalendarSelector dimensions.
const (
	DefaultCalendarWidth  = 280.0
	DefaultCalendarHeight = 300.0
)

// NewCalendarSelector creates a CalendarSelector with today's date selected.
func NewCalendarSelector(name string, source *sg.FontFamily, displaySize float64) *CalendarSelector {
	var font *sg.FontFamily
	if source != nil {
		font = source
	}
	now := time.Now()
	cs := &CalendarSelector{
		source:       source,
		font:         font,
		displaySize:  displaySize,
		selectedDate: time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local),
		viewYear:     now.Year(),
		viewMonth:    now.Month(),
		hoveredCell:  -1,
	}
	initComponent(&cs.Component, name)
	cs.node.Interactable = true
	cs.initBackground(name)
	cs.initBorder(name)

	// Build the calendar content tree.
	cs.calendarRoot = sg.NewContainer(name + "-cal-root")
	cs.calendarRoot.Interactable = true
	cs.node.AddChild(cs.calendarRoot)

	// Header background.
	cs.headerBg = sg.NewSprite(name+"-header-bg", sg.TextureRegion{})
	cs.calendarRoot.AddChild(cs.headerBg)

	// Header: prev button, month label, next button.
	cs.prevBtn = NewButton(name+"-prev", "<", source, displaySize)
	cs.prevBtn.SetVariant(Neutral)
	cs.prevBtn.SetOnClick(func() {
		cs.prevMonth()
	})
	cs.calendarRoot.AddChild(cs.prevBtn.Node())

	cs.monthLabel = sg.NewText(name+"-month", cs.monthYearString(), font)
	cs.monthLabel.TextBlock.FontSize = displaySize * 1.1
	cs.monthLabel.TextBlock.Color = sg.RGBA(1, 1, 1, 1)
	cs.calendarRoot.AddChild(cs.monthLabel)

	cs.nextBtn = NewButton(name+"-next", ">", source, displaySize)
	cs.nextBtn.SetVariant(Neutral)
	cs.nextBtn.SetOnClick(func() {
		cs.nextMonth()
	})
	cs.calendarRoot.AddChild(cs.nextBtn.Node())

	// Weekday header labels.
	weekdays := [7]string{"Su", "Mo", "Tu", "We", "Th", "Fr", "Sa"}
	for i := 0; i < 7; i++ {
		lbl := sg.NewText(fmt.Sprintf("%s-wd-%d", name, i), weekdays[i], font)
		lbl.TextBlock.FontSize = displaySize * 0.85
		cs.weekdayLabels[i] = lbl
		cs.calendarRoot.AddChild(lbl)
	}

	// Separator line below weekday row.
	cs.weekdaySep = sg.NewSprite(name+"-wd-sep", sg.TextureRegion{})
	cs.calendarRoot.AddChild(cs.weekdaySep)

	// Day cells: 6 rows x 7 columns.
	for i := 0; i < 42; i++ {
		cell := &calendarDayCell{}
		cell.container = sg.NewContainer(fmt.Sprintf("%s-day-%d", name, i))
		cell.container.Interactable = true

		cell.bg = sg.NewSprite(fmt.Sprintf("%s-day-%d-bg", name, i), sg.TextureRegion{})
		cell.container.AddChild(cell.bg)

		// Today ring: 4 border sprites forming an outline.
		cell.todayRing = sg.NewContainer(fmt.Sprintf("%s-day-%d-ring", name, i))
		cell.todayRing.SetVisible(false)
		cell.container.AddChild(cell.todayRing)

		cell.label = sg.NewText(fmt.Sprintf("%s-day-%d-lbl", name, i), "", font)
		cell.label.TextBlock.FontSize = displaySize
		cell.label.SetZIndex(1)
		cell.container.AddChild(cell.label)

		idx := i
		cell.container.OnClick(func(_ sg.ClickContext) {
			cs.onDayCellClick(idx)
		})
		cell.container.OnPointerEnter(func(_ sg.PointerContext) {
			if cs.hoveredCell != idx {
				cs.hoveredCell = idx
				cs.UpdateVisuals()
			}
		})
		cell.container.OnPointerLeave(func(_ sg.PointerContext) {
			if cs.hoveredCell == idx {
				cs.hoveredCell = -1
				cs.UpdateVisuals()
			}
		})

		cs.dayCells[i] = cell
		cs.calendarRoot.AddChild(cell.container)
	}

	// Popup mode trigger button (hidden by default).
	cs.triggerBtn = NewButton(name+"-trigger", cs.formatDate(cs.selectedDate), source, displaySize)
	cs.triggerBtn.SetVariant(Neutral)
	cs.triggerBtn.Node().SetVisible(false)
	cs.triggerBtn.SetOnClick(func() {
		if cs.popupOpen {
			cs.ClosePopup()
		} else {
			cs.OpenPopup()
		}
	})
	cs.node.AddChild(cs.triggerBtn.Node())

	// Create popover for popup mode.
	cs.popover = NewPopover(name + "-popover")
	cs.popover.SetPreferredSide(PopoverBelow)
	cs.popover.SetOnClose(func() {
		cs.popupOpen = false
	})

	cs.onVisualStateChange = func() { cs.UpdateVisuals() }
	cs.onThemeChange = func() { cs.UpdateVisuals() }

	cs.SetSize(DefaultCalendarWidth, DefaultCalendarHeight)
	cs.rebuildGrid()
	cs.UpdateVisuals()
	return cs
}

// SetDate sets the selected date and navigates to its month.
func (cs *CalendarSelector) SetDate(t time.Time) {
	cs.selectedDate = time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.Local)
	cs.viewYear = t.Year()
	cs.viewMonth = t.Month()
	cs.rebuildGrid()
	cs.UpdateVisuals()
	if cs.popupMode {
		cs.triggerBtn.SetText(cs.formatDate(cs.selectedDate))
	}
}

// Date returns the currently selected date.
func (cs *CalendarSelector) Date() time.Time {
	return cs.selectedDate
}

// SetMinDate sets the minimum selectable date.
func (cs *CalendarSelector) SetMinDate(t time.Time) {
	d := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.Local)
	cs.minDate = &d
	cs.rebuildGrid()
	cs.UpdateVisuals()
}

// SetMaxDate sets the maximum selectable date.
func (cs *CalendarSelector) SetMaxDate(t time.Time) {
	d := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.Local)
	cs.maxDate = &d
	cs.rebuildGrid()
	cs.UpdateVisuals()
}

// ClearMinDate removes the minimum date constraint.
func (cs *CalendarSelector) ClearMinDate() {
	cs.minDate = nil
	cs.rebuildGrid()
	cs.UpdateVisuals()
}

// ClearMaxDate removes the maximum date constraint.
func (cs *CalendarSelector) ClearMaxDate() {
	cs.maxDate = nil
	cs.rebuildGrid()
	cs.UpdateVisuals()
}

// SetPopupMode switches between inline and popup display.
func (cs *CalendarSelector) SetPopupMode(v bool) {
	cs.popupMode = v
	cs.calendarRoot.SetVisible(!v)
	cs.triggerBtn.Node().SetVisible(v)
	if v {
		cs.triggerBtn.SetText(cs.formatDate(cs.selectedDate))
	}
}

// IsPopupMode returns true if the calendar is in popup mode.
func (cs *CalendarSelector) IsPopupMode() bool {
	return cs.popupMode
}

// OpenPopup opens the popup calendar (popup mode only).
func (cs *CalendarSelector) OpenPopup() {
	if !cs.popupMode || cs.popupOpen {
		return
	}
	cs.popupOpen = true

	calW := 280.0
	calH := 280.0

	wrapper := NewComponent(cs.node.Name + "-popup-cal")
	wrapper.Width = calW
	wrapper.Height = calH
	wrapper.node.Interactable = true
	wrapper.node.HitShape = sg.HitRect{X: 0, Y: 0, Width: calW, Height: calH}

	cs.node.RemoveChild(cs.calendarRoot)
	wrapper.node.AddChild(cs.calendarRoot)
	cs.calendarRoot.SetVisible(true)

	cs.layoutCalendar(calW, calH)
	cs.UpdateVisuals()

	cs.popover.SetContentSize(calW, calH)
	cs.popover.SetContent(wrapper)
	cs.popover.Open(&cs.triggerBtn.Component)
}

// ClosePopup closes the popup calendar (popup mode only).
func (cs *CalendarSelector) ClosePopup() {
	if !cs.popupMode || !cs.popupOpen {
		return
	}
	cs.popover.Close()
	cs.popupOpen = false

	if cs.calendarRoot.Parent != nil {
		cs.calendarRoot.Parent.RemoveChild(cs.calendarRoot)
	}
	cs.node.AddChild(cs.calendarRoot)
	cs.calendarRoot.SetVisible(!cs.popupMode)
}

// SetMonth navigates the calendar view to a specific year and month.
func (cs *CalendarSelector) SetMonth(year, month int) {
	cs.viewYear = year
	cs.viewMonth = time.Month(month)
	cs.rebuildGrid()
	cs.UpdateVisuals()
}

// SetOnDateSelected sets the callback fired when a day is clicked.
func (cs *CalendarSelector) SetOnDateSelected(fn func(time.Time)) {
	cs.onDateSelected = fn
}

// BindDate binds the selected date to a reactive Ref. Changes to the Ref
// update the calendar, and user selections update the Ref.
func (cs *CalendarSelector) BindDate(ref *Ref[time.Time]) {
	cs.dateRef = ref
	bindRef(&cs.dateWatch, ref, cs.SetDate)
}

// SetSize resizes the calendar and repositions sub-components.
func (cs *CalendarSelector) SetSize(w, h float64) {
	cs.Width = w
	cs.Height = h
	cs.resizeBackground(w, h)
	cs.resizeBorder(w, h)
	cs.node.HitShape = sg.HitRect{X: 0, Y: 0, Width: w, Height: h}
	cs.MarkLayoutDirty()

	if cs.popupMode {
		cs.triggerBtn.SetSize(w, h)
		cs.triggerBtn.SetPosition(0, 0)
	} else {
		cs.layoutCalendar(w, h)
	}
}

// SetEnabled enables or disables the calendar.
func (cs *CalendarSelector) SetEnabled(v bool) {
	cs.Component.SetEnabled(v)
	cs.prevBtn.SetEnabled(v)
	cs.nextBtn.SetEnabled(v)
	cs.triggerBtn.SetEnabled(v)
}

// UpdateVisuals applies theme colors to all sub-components.
func (cs *CalendarSelector) UpdateVisuals() {
	group := cs.EffectiveTheme().CalendarSelector.Group(cs.Variant())
	state := cs.state

	// Container background and border.
	cr := resolveCornerRadius(group.CornerRadius, cs.Height)
	cs.applyCornerRadius(cr)
	bg := group.Background.Resolve(state)
	cs.applyBackground(bg)
	cs.applyBorder(group.BorderColor.Resolve(state), group.BorderWidth, bg)

	// Header background.
	headerBgColor := group.HeaderBackground.Resolve(state)
	cs.headerBg.SetColor(headerBgColor.Color)

	// Month label color.
	headerTextColor := group.HeaderTextColor.Resolve(state)
	cs.monthLabel.SetTextColor(headerTextColor)

	// Weekday separator color (uses border color).
	cs.weekdaySep.SetColor(group.BorderColor.Resolve(state))

	// Weekday label colors.
	wdColor := group.WeekdayTextColor.Resolve(state)
	for _, lbl := range cs.weekdayLabels {
		lbl.SetTextColor(wdColor)
	}

	// Day cells.
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)

	for i, cell := range cs.dayCells {
		if cell.day == 0 {
			cell.container.SetVisible(false)
			continue
		}
		cell.container.SetVisible(true)

		isHovered := cs.hoveredCell == i && !cell.disabled
		var bgColor, textColor sg.Color
		showTodayRing := false

		switch {
		case cell.isSelected:
			bgColor = group.DaySelectedBg.Resolve(state).Color
			textColor = group.DaySelectedColor.Resolve(state)
			// If today is also selected, no ring needed.
		case cell.date.Equal(today):
			bgColor = group.DayTodayBg.Resolve(state).Color
			textColor = group.DayTodayColor.Resolve(state)
			showTodayRing = true
		case isHovered:
			bgColor = group.DayHoverBg.Resolve(state).Color
			textColor = group.DayTextColor.Resolve(state)
		case cell.disabled:
			bgColor = group.DayBackground.Resolve(state).Color
			textColor = group.DayMutedColor.Resolve(state)
		default:
			bgColor = group.DayBackground.Resolve(state).Color
			textColor = group.DayTextColor.Resolve(state)
		}

		cell.bg.SetColor(bgColor)
		cell.todayRing.SetVisible(showTodayRing)
		cell.label.SetTextColor(textColor)

		_ = cell.disabled // cursor shape handled by container callbacks
	}

	cs.MarkDrawDirty()
}

// Dispose cleans up the calendar.
func (cs *CalendarSelector) Dispose() {
	cs.dateWatch.Stop()
	cs.Component.Dispose()
}

// PrevButton returns the previous-month navigation button.
func (cs *CalendarSelector) PrevButton() *Button { return cs.prevBtn }

// NextButton returns the next-month navigation button.
func (cs *CalendarSelector) NextButton() *Button { return cs.nextBtn }

// TriggerButton returns the popup mode trigger button.
func (cs *CalendarSelector) TriggerButton() *Button { return cs.triggerBtn }

// ---------------------------------------------------------------------------
// Internal
// ---------------------------------------------------------------------------

func (cs *CalendarSelector) monthYearString() string {
	return fmt.Sprintf("%s %d", cs.viewMonth.String(), cs.viewYear)
}

func (cs *CalendarSelector) formatDate(t time.Time) string {
	return fmt.Sprintf("%d-%02d-%02d", t.Year(), t.Month(), t.Day())
}

func (cs *CalendarSelector) prevMonth() {
	cs.viewMonth--
	if cs.viewMonth < time.January {
		cs.viewMonth = time.December
		cs.viewYear--
	}
	cs.rebuildGrid()
	cs.UpdateVisuals()
}

func (cs *CalendarSelector) nextMonth() {
	cs.viewMonth++
	if cs.viewMonth > time.December {
		cs.viewMonth = time.January
		cs.viewYear++
	}
	cs.rebuildGrid()
	cs.UpdateVisuals()
}

func (cs *CalendarSelector) rebuildGrid() {
	setTextContent(cs.monthLabel, cs.monthYearString())

	firstOfMonth := time.Date(cs.viewYear, cs.viewMonth, 1, 0, 0, 0, 0, time.Local)
	startWeekday := int(firstOfMonth.Weekday()) // Sunday = 0
	dim := daysIn(cs.viewYear, cs.viewMonth)

	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)

	for i := 0; i < 42; i++ {
		cell := cs.dayCells[i]
		dayNum := i - startWeekday + 1

		if dayNum < 1 || dayNum > dim {
			cell.day = 0
			cell.inMonth = false
			cell.isToday = false
			cell.isSelected = false
			cell.disabled = true
			setTextContent(cell.label, "")
		} else {
			cell.day = dayNum
			cell.date = time.Date(cs.viewYear, cs.viewMonth, dayNum, 0, 0, 0, 0, time.Local)
			cell.inMonth = true
			cell.isToday = cell.date.Equal(today)
			cell.isSelected = cell.date.Equal(cs.selectedDate)
			cell.disabled = cs.isDateDisabled(cell.date)
			setTextContent(cell.label, fmt.Sprintf("%d", dayNum))
		}
	}

	// Recompute layout so text positions match the new content.
	if cs.Width > 0 && cs.Height > 0 {
		cs.layoutCalendar(cs.Width, cs.Height)
	}
}

func (cs *CalendarSelector) isDateDisabled(d time.Time) bool {
	if cs.minDate != nil && d.Before(*cs.minDate) {
		return true
	}
	if cs.maxDate != nil && d.After(*cs.maxDate) {
		return true
	}
	return false
}

func (cs *CalendarSelector) onDayCellClick(idx int) {
	if !cs.enabled {
		return
	}
	cell := cs.dayCells[idx]
	if cell.day == 0 || cell.disabled {
		return
	}

	cs.selectedDate = cell.date
	cs.rebuildGrid()
	cs.UpdateVisuals()

	if cs.popupMode {
		cs.triggerBtn.SetText(cs.formatDate(cs.selectedDate))
		cs.ClosePopup()
	}

	if cs.dateRef != nil {
		cs.dateRef.Set(cs.selectedDate)
	}
	if cs.onDateSelected != nil {
		cs.onDateSelected(cs.selectedDate)
	}
}

func (cs *CalendarSelector) layoutCalendar(w, h float64) {
	headerH := cs.displaySize * 2.4
	weekdayH := cs.displaySize * 1.8
	pad := 4.0 // small internal padding

	// Header background.
	cs.headerBg.SetScale(w, headerH)
	cs.headerBg.SetPosition(0, 0)

	// Prev button.
	btnW := headerH * 0.85
	cs.prevBtn.SetSize(btnW, headerH)
	cs.prevBtn.SetPosition(pad, 0)

	// Next button.
	cs.nextBtn.SetSize(btnW, headerH)
	cs.nextBtn.SetPosition(w-btnW-pad, 0)

	// Month label centered.
	labelW, labelH := measureDisplay(cs.font, cs.monthYearString(), cs.displaySize*1.1)
	cs.monthLabel.SetPosition((w-labelW)/2, (headerH-labelH)/2)

	// Weekday labels.
	cellW := w / 7
	wdY := headerH
	for i, lbl := range cs.weekdayLabels {
		tw, th := measureDisplay(cs.font, lbl.TextBlock.Content, cs.displaySize*0.85)
		lbl.SetPosition(float64(i)*cellW+(cellW-tw)/2, wdY+(weekdayH-th)/2)
	}

	// Separator line.
	sepY := headerH + weekdayH - 1
	cs.weekdaySep.SetScale(w, 1)
	cs.weekdaySep.SetPosition(0, sepY)

	// Day grid.
	gridTop := headerH + weekdayH
	gridH := h - gridTop
	cellH := gridH / 6

	for i, cell := range cs.dayCells {
		col := i % 7
		row := i / 7
		cx := float64(col) * cellW
		cy := gridTop + float64(row)*cellH

		cell.container.SetPosition(cx, cy)
		cell.container.HitShape = sg.HitRect{X: 0, Y: 0, Width: cellW, Height: cellH}

		// Background: inset slightly for visual separation.
		inset := 1.0
		bgW := cellW - inset*2
		bgH := cellH - inset*2
		cell.bg.SetScale(bgW, bgH)
		cell.bg.SetPosition(inset, inset)

		// Today ring: 4 border sprites forming a colored outline.
		cs.buildTodayRing(cell, bgW, bgH, inset)

		// Center text in cell.
		if cell.day > 0 {
			tw, th := measureDisplay(cs.font, cell.label.TextBlock.Content, cs.displaySize)
			cell.label.SetPosition((cellW-tw)/2, (cellH-th)/2)
		}
	}
}

// buildTodayRing creates/updates the 4 border sprites that outline "today".
func (cs *CalendarSelector) buildTodayRing(cell *calendarDayCell, bgW, bgH, inset float64) {
	group := cs.EffectiveTheme().CalendarSelector.Group(cs.Variant())
	ringColor := group.DayTodayColor.Resolve(cs.state)
	ringW := 2.0

	// Lazy-create the 4 border sprites on first call.
	children := cell.todayRing.Children()
	if len(children) < 4 {
		for _, tag := range []string{"-t", "-b", "-l", "-r"} {
			s := sg.NewSprite(cell.todayRing.Name+tag, sg.TextureRegion{})
			cell.todayRing.AddChild(s)
		}
		children = cell.todayRing.Children()
	}

	top, bot, left, right := children[0], children[1], children[2], children[3]

	top.SetScale(bgW, ringW)
	top.SetPosition(inset, inset)
	top.SetColor(ringColor)

	bot.SetScale(bgW, ringW)
	bot.SetPosition(inset, inset+bgH-ringW)
	bot.SetColor(ringColor)

	left.SetScale(ringW, bgH)
	left.SetPosition(inset, inset)
	left.SetColor(ringColor)

	right.SetScale(ringW, bgH)
	right.SetPosition(inset+bgW-ringW, inset)
	right.SetColor(ringColor)
}

// daysIn returns the number of days in the given month.
func daysIn(year int, month time.Month) int {
	return time.Date(year, month+1, 0, 0, 0, 0, 0, time.Local).Day()
}
