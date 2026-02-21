// binding-gallery demonstrates every widget Bind* method in WillowUI.
//
// Patterns shown:
//   - Fan-out: one Ref[float64] drives Slider.BindValue, ProgressBar.BindValue
//     and MeterBar (via WatchEffect) simultaneously
//   - Derived text: TextInput.BindValue → Ref → Computed → WatchEffect → Label
//   - BindEnabled / BindVisible: Checkbox + Toggle gate widget groups
//   - Select two-way pattern (no BindSelected): SetOnChange + WatchValue
//   - OptionRotator.BindSelected and BindValue (index vs string refs)
//   - TabBar.BindSelected + ToggleButtonBar.BindSelected sharing one Ref
//   - Ref.Update: +/- buttons mutate a ref without reading it first
package main

import (
	"fmt"

	"github.com/devthicket/willow"
	ui "github.com/devthicket/willowui"
)

const (
	screenW  = 900
	screenH  = 680
	colL     = 40.0
	colR     = 480.0
	fontSize = 15.0
)

func main() {
	font := ui.MustLoadDefaultFont()

	screen := ui.NewScreen()

	title := willow.NewText("title", "Reactive: Binding Gallery", font)
	title.TextBlock.FontSize = 20
	title.TextBlock.Color = willow.RGBA(1, 1, 1, 1)
	title.SetPosition(24, 12)
	screen.AddNode(title)
	screen.AddNode(divLine("div0", screenW, 44))

	y := 56.0

	// ── Section 1: Fan-out pipeline ──────────────────────────────────────────
	// One Ref[float64] (0–1) drives Slider.BindValue and ProgressBar.BindValue
	// directly. MeterBar (which has no BindValue) is driven via WatchEffect.

	addSection(screen, font, "FAN-OUT:  Slider.BindValue + ProgressBar.BindValue + MeterBar via WatchEffect", y)
	y += 22

	levelRef := ui.NewRef(0.6) // 0-1

	// Slider (0-1 range)
	slider := ui.NewSlider("lvl-slider")
	slider.SetSize(220, 22)
	slider.SetPosition(colL, y+2)
	slider.BindValue(levelRef) // two-way: slider ↔ ref
	screen.Add(slider)

	minusBtn := ui.NewButton("lvl-minus", "−", font, fontSize)
	minusBtn.SetSize(28, 28)
	minusBtn.SetPosition(colL+232, y)
	minusBtn.SetOnClick(func() {
		levelRef.Update(func(v float64) float64 {
			if v >= 0.05 {
				return v - 0.05
			}
			return 0
		})
	})
	screen.Add(minusBtn)

	plusBtn := ui.NewButton("lvl-plus", "+", font, fontSize)
	plusBtn.SetSize(28, 28)
	plusBtn.SetPosition(colL+268, y)
	plusBtn.SetOnClick(func() {
		levelRef.Update(func(v float64) float64 {
			if v <= 0.95 {
				return v + 0.05
			}
			return 1
		})
	})
	screen.Add(plusBtn)

	// ProgressBar: direct BindValue (both 0-1)
	pb := ui.NewProgressBar("lvl-pb")
	pb.SetSize(320, 16)
	pb.SetPosition(colR, y)
	pb.BindValue(levelRef)
	screen.Add(pb)

	// MeterBar: no BindValue → use WatchEffect
	mb := ui.NewMeterBar("lvl-mb")
	mb.SetRange(0, 100)
	mb.SetSize(320, 16)
	mb.SetPosition(colR, y+28)
	screen.Add(mb)
	ui.WatchEffect(func() {
		mb.SetValue(levelRef.Get() * 100)
	})

	valLbl := ui.NewLabel("lvl-val", "", font, fontSize)
	valLbl.SetColor(willow.RGBA(0.65, 0.8, 1, 1))
	valLbl.SetPosition(colR, y+54)
	screen.Add(valLbl)
	ui.WatchValue(levelRef, func(_, v float64) {
		valLbl.SetText(fmt.Sprintf("levelRef = %.2f  (%.0f%%)", v, v*100))
	})

	y += 84
	screen.AddNode(divLine("d1", screenW-80, y))
	y += 12

	// ── Section 2: Derived text ───────────────────────────────────────────────
	// TextInput.BindValue → Ref[string] → Computed[string] → WatchEffect → Label

	addSection(screen, font, "DERIVED TEXT:  TextInput.BindValue -> Computed -> WatchEffect -> Label", y)
	y += 22

	nameRef := ui.NewRef("")

	ti := ui.NewTextInput("name-input", font, fontSize)
	ti.SetPlaceholder("Type your name...")
	ti.SetWidth(240)
	ti.SetPosition(colL, y)
	ti.BindValue(nameRef) // two-way: typing updates nameRef
	screen.Add(ti)

	greeting := ui.NewComputed(func() string {
		n := nameRef.Get()
		if n == "" {
			return "Hello, stranger!"
		}
		return fmt.Sprintf("Hello, %s!  (%d chars)", n, len(n))
	})

	greetLbl := ui.NewLabel("greet-lbl", "", font, fontSize)
	greetLbl.SetColor(willow.RGBA(0.65, 0.9, 0.65, 1))
	greetLbl.SetPosition(colR, y+4)
	screen.Add(greetLbl)
	ui.WatchEffect(func() { greetLbl.SetText(greeting.Get()) })

	y += 50
	screen.AddNode(divLine("d2", screenW-80, y))
	y += 12

	// ── Section 3: BindEnabled + BindVisible ─────────────────────────────────

	addSection(screen, font, "BindEnabled + BindVisible:  Checkbox / Toggle gate widget groups", y)
	y += 22

	enabledRef := ui.NewRef(true)
	visibleRef := ui.NewRef(true)

	cb := ui.NewCheckbox("cb-enable", "Enable stepper + button", font, fontSize)
	cb.SetChecked(true)
	cb.SetPosition(colL, y)
	cb.BindValue(enabledRef)
	screen.Add(cb)

	tgl := ui.NewToggle("tgl-visible")
	tgl.SetValue(true)
	tgl.SetPosition(colL, y+36)
	tgl.BindValue(visibleRef)
	screen.Add(tgl)
	tglLbl := ui.NewLabel("tgl-lbl", "  Show detail panel", font, fontSize)
	tglLbl.SetPosition(colL+54, y+40)
	screen.Add(tglLbl)

	stepper := ui.NewNumberStepper("gated-stepper", font, fontSize)
	stepper.SetMin(0)
	stepper.SetMax(50)
	stepper.SetSize(160, 32)
	stepper.SetPosition(colR, y)
	stepper.BindEnabled(enabledRef)
	screen.Add(stepper)

	gatedBtn := ui.NewButton("gated-btn", "Gated Action", font, fontSize)
	gatedBtn.SetSize(160, 32)
	gatedBtn.SetPosition(colR+170, y)
	gatedBtn.BindEnabled(enabledRef)
	screen.Add(gatedBtn)

	detailPanel := ui.NewPanel("detail-panel")
	detailPanel.SetBackground(willow.RGBA(0.10, 0.20, 0.30, 1))
	detailPanel.SetSize(340, 34)
	detailPanel.SetPosition(colR, y+44)
	detailPanel.BindVisible(visibleRef)
	screen.Add(detailPanel)

	detailLbl := ui.NewLabel("detail-lbl", "  Detail panel  (bind:visible)", font, fontSize)
	detailLbl.SetColor(willow.RGBA(0.7, 0.85, 1, 1))
	detailPanel.AddChild(detailLbl)

	y += 96
	screen.AddNode(divLine("d3", screenW-80, y))
	y += 12

	// ── Section 4: Selection widgets ─────────────────────────────────────────
	// Select: no BindSelected — demonstrate manual two-way pattern.
	// OptionRotator: BindSelected (index) and BindValue (string).
	// TabBar + ToggleButtonBar share one Ref[int].

	addSection(screen, font, "SELECTION BINDING:  Select (manual 2-way) / OptionRotator / TabBar+ToggleButtonBar sync", y)
	y += 22

	// Select: manual two-way via SetOnChange + WatchValue
	classNames := []string{"Warrior", "Mage", "Rogue", "Paladin"}
	classRef := ui.NewRef(0)

	sel := ui.NewSelect("class-sel", toSelectOpts(classNames), font, fontSize)
	sel.SetSize(180, 32)
	sel.SetPosition(colL, y)
	sel.SetOnChange(func(idx int, _ ui.SelectOption) { classRef.Set(idx) }) // widget → ref
	ui.WatchValue(classRef, func(_, idx int) { sel.SetSelected(idx) })      // ref → widget
	screen.Add(sel)

	classLbl := ui.NewLabel("class-lbl", "", font, fontSize)
	classLbl.SetColor(willow.RGBA(1, 0.85, 0.5, 1))
	classLbl.SetPosition(colR, y+8)
	screen.Add(classLbl)
	ui.WatchValue(classRef, func(_, idx int) {
		if idx >= 0 && idx < len(classNames) {
			classLbl.SetText("classRef = " + classNames[idx])
		}
	})

	y += 46

	// OptionRotator.BindSelected (index) and BindValue (string) on the same widget
	diffOpts := []string{"Easy", "Normal", "Hard", "Nightmare"}
	diffIdxRef := ui.NewRef(1)
	diffValRef := ui.NewRef("Normal")

	or := ui.NewOptionRotator("diff-or", diffOpts, font, fontSize)
	or.SetSize(220, 32)
	or.SetPosition(colL, y)
	or.BindSelected(diffIdxRef) // keeps index ref in sync
	or.BindValue(diffValRef)    // also keeps string ref in sync
	screen.Add(or)

	diffLbl := ui.NewLabel("diff-lbl", "", font, fontSize)
	diffLbl.SetColor(willow.RGBA(1, 0.7, 0.7, 1))
	diffLbl.SetPosition(colR, y+8)
	screen.Add(diffLbl)
	ui.WatchEffect(func() {
		diffLbl.SetText(fmt.Sprintf("idx=%d  val=%q", diffIdxRef.Get(), diffValRef.Get()))
	})

	y += 46

	// TabBar + ToggleButtonBar sharing one Ref[int]
	tabNames := []string{"Info", "Stats", "Skills"}
	tabRef := ui.NewRef(0)

	tb := ui.NewTabBar("sync-tabs", font, fontSize)
	for _, n := range tabNames {
		tb.AddTabPage(n, ui.LayoutNone, 0, ui.Insets{})
	}
	tb.SetSize(260, 36)
	tb.SetPosition(colL, y)
	tb.BindSelected(tabRef) // two-way
	screen.Add(tb)

	tbb := ui.NewToggleButtonBar("sync-tbb", font, fontSize)
	for _, n := range tabNames {
		tbb.AddButton(n)
	}
	tbb.SetSize(260, 36)
	tbb.SetPosition(colR, y)
	tbb.BindSelected(tabRef) // same ref → clicking either syncs both
	screen.Add(tbb)

	syncLbl := ui.NewLabel("sync-lbl", "", font, fontSize)
	syncLbl.SetColor(willow.RGBA(0.6, 0.6, 0.6, 1))
	syncLbl.SetPosition(colL, y+44)
	screen.Add(syncLbl)
	ui.WatchValue(tabRef, func(_, idx int) {
		if idx >= 0 && idx < len(tabNames) {
			syncLbl.SetText(fmt.Sprintf("tabRef = %d (%s)  -- click either widget to sync", idx, tabNames[idx]))
		}
	})

	ui.Stage.Add(screen)
	ui.Setup(ui.StageConfig{
		Title:      "WillowUI — Binding Gallery",
		Width:      screenW,
		Height:     screenH,
		ClearColor: willow.RGBA(0.08, 0.08, 0.10, 1),
	})
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func addSection(s *ui.Screen, font *willow.FontFamily, text string, y float64) {
	lbl := ui.NewLabel("sec-"+text[:12], text, font, 12)
	lbl.SetColor(willow.RGBA(0.45, 0.55, 0.68, 1))
	lbl.SetPosition(colL, y)
	s.Add(lbl)
}

func divLine(name string, w, y float64) *willow.Node {
	d := willow.NewSprite(name, willow.TextureRegion{})
	d.SetPosition(40, y)
	d.SetScale(w, 1)
	d.SetColor(willow.RGBA(0.18, 0.23, 0.28, 1))
	return d
}

func toSelectOpts(names []string) []ui.SelectOption {
	opts := make([]ui.SelectOption, len(names))
	for i, n := range names {
		opts[i] = ui.SelectOption{Label: n}
	}
	return opts
}
