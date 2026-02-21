// Watch Lifecycle - reactive example.
// A live match spectator HUD. Three WatchEffect handles drive stat labels
// (score, stamina, coins) while spectating. Tabbing out stops all three
// display watchers so the HUD freezes — tabbing back in re-subscribes and
// snaps to current values instantly. A milestone watcher on the coin count
// stays active regardless of tab state, so you never miss a reward crossing
// a threshold. Ref.Update drives the frame-simulated stat changes.
package main

import (
	"fmt"

	"github.com/devthicket/willow"
	ui "github.com/devthicket/willowui"
)

const (
	screenW  = 800
	screenH  = 420
	colLeft  = 40.0
	colRight = 460.0
)

type simulationController struct {
	scoreRef      *ui.Ref[int]
	staminaRef    *ui.Ref[int]
	coinsRef      *ui.Ref[int]
	milestoneLbl  *ui.Label
	frame         int
	nextMilestone *int
	milestones    []int
	spectatingRef *ui.Ref[bool]
}

func (c *simulationController) OnCreate(s *ui.Screen) {}

func (c *simulationController) OnUpdate(dt float64) {
	c.frame++
	// Score: +50 every 30 frames.
	if c.frame%30 == 0 {
		c.scoreRef.Update(func(v int) int { return v + 50 })
	}
	// Stamina: triangle wave 20–100.
	phase := c.frame % 180
	var stamina int
	if phase < 90 {
		stamina = 20 + phase
	} else {
		stamina = 110 - (phase - 90)
	}
	c.staminaRef.Set(stamina)
	// Coins: +1 every 80 frames.
	if c.frame%80 == 0 {
		c.coinsRef.Update(func(v int) int { return v + 1 })
	}
}

func (c *simulationController) OnDestroy() {}

func main() {
	font := ui.MustLoadDefaultFont()

	const (
		sizeLarge  = 22.0
		sizeMedium = 16.0
		sizeSmall  = 16.0
	)

	// ── Reactive state ────────────────────────────────────────────────────────
	scoreRef := ui.NewRef(0)
	staminaRef := ui.NewRef(100)
	coinsRef := ui.NewRef(0)
	spectatingRef := ui.NewRef(true)
	nextMilestone := 0
	milestones := []int{10, 25, 50, 100}

	simCtrl := &simulationController{
		scoreRef:      scoreRef,
		staminaRef:    staminaRef,
		coinsRef:      coinsRef,
		frame:         0,
		nextMilestone: &nextMilestone,
		milestones:    milestones,
		spectatingRef: spectatingRef,
	}

	screen := ui.NewScreen(ui.WithController(simCtrl))

	title := willow.NewText("title", "Reactive: Watch Lifecycle", font)
	title.TextBlock.FontSize = sizeLarge
	title.TextBlock.Color = willow.RGBA(1, 1, 1, 1)
	title.SetPosition(24, 14)
	screen.AddNode(title)

	div := ui.NewDivider("divider", screenW)
	div.SetPosition(0, 46)
	screen.AddNode(div)

	// ── Left: live stats ──────────────────────────────────────────────────────
	addSection(screen, font, sizeSmall, "LIVE STATS", colLeft, 60)

	scoreLbl := addStatLabel(screen, font, sizeMedium, colLeft, 84)
	staminaLbl := addStatLabel(screen, font, sizeMedium, colLeft, 112)
	coinsLbl := addStatLabel(screen, font, sizeMedium, colLeft, 140)

	div2 := ui.NewDivider("divider-2", screenW)
	div2.SetPosition(0, 172)
	screen.AddNode(div2)
	addSection(screen, font, sizeSmall, "MILESTONES", colLeft, 184)

	milestoneLbl := ui.NewLabel("milestone", "No milestones yet.", font, sizeSmall)
	milestoneLbl.SetColor(willow.RGBA(1, 0.82, 0.28, 1))
	milestoneLbl.SetPosition(colLeft, 204)
	screen.Add(milestoneLbl)
	simCtrl.milestoneLbl = milestoneLbl

	noteLbl := ui.NewLabel("note",
		"Milestone watcher stays active even while tabbed out.",
		font, sizeSmall)
	noteLbl.SetColor(willow.RGBA(0.38, 0.44, 0.50, 1))
	noteLbl.SetPosition(colLeft, 228)
	screen.Add(noteLbl)

	// ── Right: spectator controls ─────────────────────────────────────────────
	addSection(screen, font, sizeSmall, "SPECTATOR", colRight, 60)

	tog := ui.NewToggle("spec")
	tog.BindValue(spectatingRef)
	tog.SetPosition(colRight, 88)
	screen.Add(tog)

	specLbl := ui.NewLabel("spec-lbl", "Spectating", font, sizeMedium)
	specLbl.SetColor(willow.RGBA(0.82, 0.82, 0.82, 1))
	specLbl.SetPosition(colRight+54, 91)
	screen.Add(specLbl)

	statusLbl := ui.NewLabel("status", "3 watchers active", font, sizeSmall)
	statusLbl.SetColor(willow.RGBA(0.35, 0.88, 0.50, 1))
	statusLbl.SetPosition(colRight, 130)
	screen.Add(statusLbl)

	div3 := ui.NewDivider("divider-3", 326)
	div3.SetPosition(colRight, 154)
	screen.AddNode(div3)

	newMatchBtn := ui.NewButton("new-match", "Reset", font, sizeSmall)
	newMatchBtn.SetSize(110, 28)
	newMatchBtn.SetPosition(colRight, 166)
	screen.Add(newMatchBtn)
	newMatchBtn.SetOnClick(func() {
		scoreRef.Set(0)
		staminaRef.Set(100)
		coinsRef.Set(0)
		*simCtrl.nextMilestone = 0
		milestoneLbl.SetText("No milestones yet.")
	})

	// ── Display watchers — stopped and restarted with spectating toggle ────────
	var scoreWatch, staminaWatch, coinsWatch ui.WatchHandle

	startWatching := func() {
		scoreWatch = ui.WatchEffect(func() {
			scoreLbl.SetText(fmt.Sprintf("Score:    %d pts", scoreRef.Get()))
		})
		staminaWatch = ui.WatchEffect(func() {
			staminaLbl.SetText(fmt.Sprintf("Stamina:  %d / 100", staminaRef.Get()))
		})
		coinsWatch = ui.WatchEffect(func() {
			coinsLbl.SetText(fmt.Sprintf("Coins:    %d", coinsRef.Get()))
		})
	}

	stopWatching := func() {
		scoreWatch.Stop()
		staminaWatch.Stop()
		coinsWatch.Stop()
	}

	startWatching()

	// Spectating toggle starts or stops all three display watchers.
	ui.WatchValue(spectatingRef, func(old, spectating bool) {
		if old == spectating {
			return // initial fire — already started above
		}
		if spectating {
			startWatching()
			statusLbl.SetText("3 watchers active")
			statusLbl.SetColor(willow.RGBA(0.35, 0.88, 0.50, 1))
		} else {
			stopWatching()
			statusLbl.SetText("tabbed out -- HUD paused")
			statusLbl.SetColor(willow.RGBA(0.92, 0.50, 0.28, 1))
		}
	})

	// ── Milestone watcher — always active, independent of spectating state ─────
	ui.WatchValue(coinsRef, func(old, newVal int) {
		if old == newVal {
			return // initial fire
		}
		for *simCtrl.nextMilestone < len(simCtrl.milestones) && newVal >= simCtrl.milestones[*simCtrl.nextMilestone] {
			milestoneLbl.SetText(fmt.Sprintf("%d coins collected! (while %s)",
				simCtrl.milestones[*simCtrl.nextMilestone],
				map[bool]string{true: "spectating", false: "tabbed out"}[spectatingRef.Peek()]))
			*simCtrl.nextMilestone++
		}
	})

	ui.Stage.Add(screen)
	ui.Setup(ui.StageConfig{
		Title:      "Reactive — Watch Lifecycle",
		Width:      screenW,
		Height:     screenH,
		ClearColor: willow.RGBA(0.08, 0.08, 0.10, 1),
	})
}

func addSection(screen *ui.Screen, font *willow.FontFamily, fontSize float64, text string, x, y float64) {
	n := willow.NewText("sec", text, font)
	n.TextBlock.FontSize = fontSize
	n.TextBlock.Color = willow.RGBA(0.45, 0.55, 0.65, 1)
	n.SetPosition(x, y)
	screen.AddNode(n)
}

func addStatLabel(screen *ui.Screen, font *willow.FontFamily, fontSize, x, y float64) *ui.Label {
	lbl := ui.NewLabel("stat", "--", font, fontSize)
	lbl.SetColor(willow.RGBA(0.85, 0.92, 0.98, 1))
	lbl.SetPosition(x, y)
	screen.Add(lbl)
	return lbl
}
