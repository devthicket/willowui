// Screens demonstrates WillowUI's ScreenManager — a navigation stack modelled
// after mobile app routing. Each screen is driven by a Controller with three
// lifecycle hooks:
//
//	OnCreate(s)   — called once when the screen is pushed; build your UI here
//	OnUpdate(dt)  — called every frame while the screen is active
//	OnDestroy()   — called when the screen is popped; release non-UI resources
//
// Three navigation flows are shown:
//
//	Menu → Settings → Back  shared audio Refs persist through the round-trip
//	Menu → Game     → Back  per-screen score lives in the OnCreate closure
//	Both screens receive shared appState (name, difficulty) via constructors
package main

import (
	"fmt"

	"github.com/devthicket/willow"
	ui "github.com/devthicket/willowui"
)

const (
	screenW = 800
	screenH = 600
)

var font *willow.FontFamily

const (
	sizeLarge  = 24.0
	sizeMedium = 16.0
	sizeSmall  = 16.0
)

// appState holds Refs shared across all screens. Because these are reactive
// references (not plain values), any screen that binds to them will always
// reflect the current value — even after navigating away and back.
type appState struct {
	playerName *ui.Ref[string]
	difficulty *ui.Ref[int] // 0=Easy 1=Normal 2=Hard
	musicOn    *ui.Ref[bool]
	sfxOn      *ui.Ref[bool]
}

var diffNames = []string{"Easy", "Normal", "Hard"}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func addSectionLabel(s *ui.Screen, text string, x, y float64) {
	n := willow.NewText("section-"+text, text, font)
	n.TextBlock.FontSize = sizeSmall
	n.TextBlock.Color = willow.RGBA(0.4, 0.5, 0.6, 1)
	n.SetPosition(x, y)
	s.AddNode(n)
}

// breadcrumb renders a dimmed navigation path (e.g. "Menu › Settings") so
// the screen stack is visible without any dynamic tracking.
func breadcrumb(s *ui.Screen, path string, y float64) {
	n := willow.NewText("breadcrumb", path, font)
	n.TextBlock.FontSize = sizeSmall
	n.TextBlock.Color = willow.RGBA(0.35, 0.4, 0.5, 1)
	n.SetPosition(float64(screenW)-300, y)
	s.AddNode(n)
}

// toggleStatus returns display text and a color for a bool toggle value.
func toggleStatus(on bool) (string, willow.Color) {
	if on {
		return "ON", willow.RGBA(0.3, 0.9, 0.5, 1)
	}
	return "OFF", willow.RGBA(0.75, 0.35, 0.35, 1)
}

// ---------------------------------------------------------------------------
// Menu screen
// ---------------------------------------------------------------------------

// menuController owns the menu screen. It stores sm so that button callbacks
// can push new screens onto the stack, and state so that player choices
// thread through to child screens.
type menuController struct {
	state *appState
}

func (c *menuController) OnCreate(s *ui.Screen) {
	title := willow.NewText("title", "WillowUI: Screen Manager Demo", font)
	title.TextBlock.FontSize = sizeLarge
	title.TextBlock.Color = willow.RGBA(1, 1, 1, 1)
	title.SetPosition(24, 16)
	s.AddNode(title)

	div := ui.NewDivider("divider", float64(screenW-48))
	div.SetPosition(24, 48)
	s.AddNode(div)

	// ── Player setup ──────────────────────────────────────────────────────────
	addSectionLabel(s, "Player Setup", 40, 64)

	nameLabel := ui.NewLabel("name-lbl", "Player Name", font, sizeSmall)
	nameLabel.SetColor(willow.RGBA(0.55, 0.65, 0.75, 1))
	nameLabel.SetPosition(40, 86)
	s.Add(nameLabel)

	// BindValue links the input to c.state.playerName — changes are
	// immediately visible to any other screen reading the same Ref.
	nameInput := ui.NewTextInput("name-input", font, sizeMedium)
	nameInput.BindValue(c.state.playerName)
	nameInput.SetPlaceholder("Enter name...")
	nameInput.SetWidth(240)
	nameInput.SetPosition(40, 106)
	s.Add(nameInput)

	diffLabel := ui.NewLabel("diff-lbl", "Difficulty", font, sizeSmall)
	diffLabel.SetColor(willow.RGBA(0.55, 0.65, 0.75, 1))
	diffLabel.SetPosition(40, 150)
	s.Add(diffLabel)

	diffBar := ui.NewToggleButtonBar("diff-bar", font, sizeMedium)
	diffBar.SetSize(300, 36)
	for _, n := range diffNames {
		diffBar.AddButton(n)
	}
	diffBar.BindSelected(c.state.difficulty)
	diffBar.SetPosition(40, 170)
	s.Add(diffBar)

	// ── Navigation ────────────────────────────────────────────────────────────
	settingsBtn := ui.NewButton("settings-btn", "Settings", font, sizeMedium)
	settingsBtn.SetSize(130, 40)
	settingsBtn.SetOnClick(func() {
		// Push adds a new screen on top of the stack.
		ui.Stage.Add(ui.NewScreen(ui.WithController(&settingsController{
			state: c.state,
		})))
	})
	settingsBtn.SetPosition(40, 230)
	s.Add(settingsBtn)

	startBtn := ui.NewButton("start-btn", "Start Game", font, sizeMedium)
	startBtn.SetSize(140, 40)
	startBtn.SetOnClick(func() {
		// Peek reads the value without registering a reactive dependency.
		// In a button callback there is no reactive context, so Peek and
		// Get are equivalent — Peek makes the intent explicit.
		name := c.state.playerName.Peek()
		if name == "" {
			name = "Player"
		}
		ui.Stage.Add(ui.NewScreen(ui.WithController(&gameController{
			state: c.state,
			name:  name,
		})))
	})
	startBtn.SetPosition(186, 230)
	s.Add(startBtn)

	hint := ui.NewLabel("hint",
		"Tip: change your name or difficulty, visit Settings, then come back -- your choices will still be here.",
		font, sizeSmall)
	hint.SetColor(willow.RGBA(0.4, 0.5, 0.6, 1))
	hint.SetWrapWidth(680)
	hint.SetPosition(40, 290)
	s.Add(hint)
}

// OnUpdate is called every frame. Use it for per-frame menu logic such as
// animating a background or polling for a gamepad "start" press.
func (c *menuController) OnUpdate(dt float64) {}

// OnDestroy is called when the screen is popped off the stack. Release any
// non-UI resources here (timers, audio handles, network connections, etc.).
// The UI node tree is cleaned up automatically by the Screen.
func (c *menuController) OnDestroy() {}

// ---------------------------------------------------------------------------
// Settings screen
// ---------------------------------------------------------------------------

type settingsController struct {
	state *appState
}

func (c *settingsController) OnCreate(s *ui.Screen) {
	title := willow.NewText("title", "Settings", font)
	title.TextBlock.FontSize = sizeLarge
	title.TextBlock.Color = willow.RGBA(1, 1, 1, 1)
	title.SetPosition(24, 16)
	s.AddNode(title)

	breadcrumb(s, "Menu  ›  Settings", 22)

	div := ui.NewDivider("divider", float64(screenW-48))
	div.SetPosition(24, 48)
	s.AddNode(div)

	addSectionLabel(s, "Audio", 40, 64)

	// ── Music row ─────────────────────────────────────────────────────────────
	musicLabel := ui.NewLabel("music-lbl", "Music", font, sizeMedium)
	musicLabel.SetColor(willow.RGBA(0.85, 0.85, 0.85, 1))
	musicLabel.SetPosition(40, 92)
	s.Add(musicLabel)

	musicToggle := ui.NewToggle("music-toggle")
	musicToggle.BindValue(c.state.musicOn)
	musicToggle.SetPosition(190, 90)
	s.Add(musicToggle)

	// WatchValue fires immediately with the current value, then again on every
	// change. s.TrackRef ensures the watcher is disposed when this screen pops.
	initMusicText, initMusicColor := toggleStatus(c.state.musicOn.Peek())
	musicStatus := ui.NewLabel("music-status", initMusicText, font, sizeSmall)
	musicStatus.SetColor(initMusicColor)
	musicStatus.SetPosition(252, 96)
	s.Add(musicStatus)
	s.TrackRef(ui.WatchValue(c.state.musicOn, func(_, v bool) {
		text, color := toggleStatus(v)
		musicStatus.SetText(text)
		musicStatus.SetColor(color)
	}))

	// ── SFX row ───────────────────────────────────────────────────────────────
	sfxLabel := ui.NewLabel("sfx-lbl", "Sound Effects", font, sizeMedium)
	sfxLabel.SetColor(willow.RGBA(0.85, 0.85, 0.85, 1))
	sfxLabel.SetPosition(40, 136)
	s.Add(sfxLabel)

	sfxToggle := ui.NewToggle("sfx-toggle")
	sfxToggle.BindValue(c.state.sfxOn)
	sfxToggle.SetPosition(190, 134)
	s.Add(sfxToggle)

	initSfxText, initSfxColor := toggleStatus(c.state.sfxOn.Peek())
	sfxStatus := ui.NewLabel("sfx-status", initSfxText, font, sizeSmall)
	sfxStatus.SetColor(initSfxColor)
	sfxStatus.SetPosition(252, 140)
	s.Add(sfxStatus)
	s.TrackRef(ui.WatchValue(c.state.sfxOn, func(_, v bool) {
		text, color := toggleStatus(v)
		sfxStatus.SetText(text)
		sfxStatus.SetColor(color)
	}))

	note := ui.NewLabel("note",
		"Tip: toggle these off, press Back, then Start Game -- you'll see the changes on the game screen.",
		font, sizeSmall)
	note.SetColor(willow.RGBA(0.4, 0.5, 0.6, 1))
	note.SetWrapWidth(560)
	note.SetPosition(40, 185)
	s.Add(note)

	backBtn := ui.NewButton("back-btn", "← Back", font, sizeMedium)
	backBtn.SetSize(120, 40)
	backBtn.SetOnClick(func() {
		// Pop removes this screen and returns to the one below (the menu).
		ui.Stage.Remove(s)
	})
	backBtn.SetPosition(40, 240)
	s.Add(backBtn)
}

func (c *settingsController) OnUpdate(dt float64) {}
func (c *settingsController) OnDestroy()          {}

// ---------------------------------------------------------------------------
// Game screen
// ---------------------------------------------------------------------------

type gameController struct {
	state *appState
	name  string
}

func (c *gameController) OnCreate(s *ui.Screen) {
	diff := c.state.difficulty.Peek()
	diffName := diffNames[diff]
	pointsPerClick := []int{5, 10, 20}[diff]

	// ── Header ────────────────────────────────────────────────────────────────
	title := ui.NewLabel("game-title", fmt.Sprintf("Welcome, %s!", c.name), font, sizeLarge)
	title.SetColor(willow.RGBA(0.3, 1, 0.5, 1))
	title.SetPosition(40, 18)
	s.Add(title)

	breadcrumb(s, "Menu  ›  Game", 22)

	diffLine := ui.NewLabel("diff-line",
		fmt.Sprintf("Difficulty: %s  (+%d pts per click)", diffName, pointsPerClick),
		font, sizeSmall)
	diffLine.SetColor(willow.RGBA(0.5, 0.75, 0.45, 1))
	diffLine.SetPosition(40, 50)
	s.Add(diffLine)

	div := ui.NewDivider("divider", float64(screenW-48))
	div.SetPosition(24, 70)
	s.AddNode(div)

	// ── Shared-state proof ────────────────────────────────────────────────────
	musicText, _ := toggleStatus(c.state.musicOn.Peek())
	sfxText, _ := toggleStatus(c.state.sfxOn.Peek())
	bothOn := c.state.musicOn.Peek() && c.state.sfxOn.Peek()
	audioLineColor := willow.RGBA(0.3, 0.9, 0.5, 1)
	if !bothOn {
		audioLineColor = willow.RGBA(0.8, 0.5, 0.35, 1)
	}
	audioLine := ui.NewLabel("audio-line",
		fmt.Sprintf("Your settings →  Music: %s   SFX: %s", musicText, sfxText),
		font, sizeSmall)
	audioLine.SetColor(audioLineColor)
	audioLine.SetPosition(40, 86)
	s.Add(audioLine)

	// ── Score attack ──────────────────────────────────────────────────────────
	addSectionLabel(s, "Score Attack: reach 100 points", 40, 112)

	// score and scoreText are declared here in OnCreate and captured by the
	// button closures below. They live as long as this screen is on the stack
	// and are garbage-collected when it is popped — no manual cleanup needed.
	score := ui.NewRef(0)
	scoreText := ui.NewRef("Score: 0 / 100")

	scoreLabel := ui.NewLabel("score", "", font, sizeLarge)
	scoreLabel.SetColor(willow.RGBA(1, 0.85, 0.2, 1))
	scoreLabel.BindText(scoreText)
	scoreLabel.SetPosition(40, 134)
	s.Add(scoreLabel)

	progressRef := ui.NewRef(0.0)
	progress := ui.NewProgressBar("progress")
	progress.SetSize(400, 18)
	progress.BindValue(progressRef)
	progress.SetPosition(40, 170)
	s.Add(progress)

	clickBtn := ui.NewButton("click-btn",
		fmt.Sprintf("Click Me!   +%d pts", pointsPerClick), font, sizeMedium)
	clickBtn.SetSize(220, 44)
	clickBtn.SetOnClick(func() {
		v := score.Get() + pointsPerClick
		if v > 100 {
			v = 100
		}
		score.Set(v)
		scoreText.Set(fmt.Sprintf("Score: %d / 100", v))
		progressRef.Set(float64(v) / 100.0)
	})
	clickBtn.SetPosition(40, 206)
	s.Add(clickBtn)

	resetBtn := ui.NewButton("reset-btn", "Reset", font, sizeMedium)
	resetBtn.SetSize(100, 44)
	resetBtn.SetOnClick(func() {
		score.Set(0)
		scoreText.Set("Score: 0 / 100")
		progressRef.Set(0)
	})
	resetBtn.SetPosition(276, 206)
	s.Add(resetBtn)

	backBtn := ui.NewButton("back-btn", "← Back to Menu", font, sizeMedium)
	backBtn.SetSize(180, 40)
	backBtn.SetOnClick(func() { ui.Stage.Remove(s) })
	backBtn.SetPosition(40, 268)
	s.Add(backBtn)

	note := ui.NewLabel("note",
		"Score is local to this screen -- go back, return, and it resets to zero.",
		font, sizeSmall)
	note.SetColor(willow.RGBA(0.4, 0.5, 0.6, 1))
	note.SetPosition(40, 326)
	s.Add(note)
}

func (c *gameController) OnUpdate(dt float64) {}
func (c *gameController) OnDestroy()          {}

// ---------------------------------------------------------------------------
// Main
// ---------------------------------------------------------------------------

func main() {
	font = ui.MustLoadDefaultFont()

	// appState is created once at startup and passed into every screen via
	// constructors. Screens share values by holding pointers to the same Refs.
	state := &appState{
		playerName: ui.NewRef("Player1"),
		difficulty: ui.NewRef(1), // Normal
		musicOn:    ui.NewRef(true),
		sfxOn:      ui.NewRef(true),
	}

	ui.Stage.Add(ui.NewScreen(ui.WithController(&menuController{state: state})))

	ui.Setup(ui.StageConfig{
		Title:      "WillowUI — Screen Manager Demo",
		Width:      screenW,
		Height:     screenH,
		ClearColor: willow.RGBA(0.08, 0.08, 0.10, 1),
	})
}
