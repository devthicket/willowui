// AnimatedImage widget demo.
// Loads sprite sheet PNGs and a GIF, showing all three play modes plus GIF playback.
package main

import (
	"fmt"
	"image/gif"
	"image/png"
	"log"
	"os"

	"github.com/devthicket/willow"
	"github.com/hajimehoshi/ebiten/v2"
	ui "github.com/devthicket/willowui"
)

const (
	screenW = 940
	screenH = 380
)

func loadPNG(path string) *ebiten.Image {
	f, err := os.Open(path)
	if err != nil {
		log.Fatalf("open %s: %v", path, err)
	}
	defer f.Close()
	decoded, err := png.Decode(f)
	if err != nil {
		log.Fatalf("decode %s: %v", path, err)
	}
	return ebiten.NewImageFromImage(decoded)
}

func loadGIF(path string) *gif.GIF {
	f, err := os.Open(path)
	if err != nil {
		log.Fatalf("open %s: %v", path, err)
	}
	defer f.Close()
	g, err := gif.DecodeAll(f)
	if err != nil {
		log.Fatalf("decode %s: %v", path, err)
	}
	return g
}

func main() {
	font := ui.MustLoadDefaultFont()

	const (
		fontLarge = 22.0
		fontSmall = 13.0
		fontXS    = 11.5
		frameSize = 64
		animSize  = 96.0
	)

	// Load sprite sheets and GIF from disk.
	dir := "examples/widgets/animated-image/"
	colorSheet := loadPNG(dir + "color-strip.png")
	pulseSheet := loadPNG(dir + "pulse-strip.png")
	spinnerSheet := loadPNG(dir + "spinner-strip.png")
	dragonGIF := loadGIF("examples/_assets/dragon-welpling.gif")

	screen := ui.NewScreen()

	// Title.
	title := willow.NewText("title", "AnimatedImage Widget  /  Play Modes", font)
	title.TextBlock.FontSize = fontLarge
	title.TextBlock.Color = willow.RGBA(1, 1, 1, 1)
	title.SetPosition(24, 14)
	screen.AddNode(title)

	div := ui.NewDivider("divider", float64(screenW))
	div.SetPosition(0, 46)
	screen.AddNode(div)

	// Layout constants.
	const (
		startX = 16.0
		startY = 68.0
		colW   = 210.0
		gapX   = 12.0
		cardH  = 260.0
		padX   = 14.0
		padY   = 12.0
	)

	col := func(n int) float64 { return startX + float64(n)*(colW+gapX) }

	headerColor := willow.RGBA(0.65, 0.75, 0.85, 1)
	captionColor := willow.RGBA(0.5, 0.58, 0.66, 1)
	statusColor := willow.RGBA(0.6, 0.9, 0.5, 1)
	cardBg := willow.RGBA(0.11, 0.11, 0.15, 1)

	// Card background helper.
	addCard := func(n int) {
		bg := willow.NewRect("card-bg-"+fmt.Sprint(n), colW, cardH, cardBg)
		bg.SetPosition(col(n), startY)
		screen.AddNode(bg)
	}

	addLabel := func(name, text string, x, y float64, size float64, c willow.Color) {
		lbl := willow.NewText(name, text, font)
		lbl.TextBlock.FontSize = size
		lbl.TextBlock.Color = c
		lbl.SetPosition(x, y)
		lbl.ZIndex_ = 2
		lbl.Invalidate()
		screen.AddNode(lbl)
	}

	// ── 0. Loop mode (color cycle) ──────────────────────────────────────
	addCard(0)
	addLabel("hdr-loop", "Loop", col(0)+padX, startY+padY, fontSmall, headerColor)

	loopAnim := ui.NewAnimatedImage("anim-loop")
	loopAnim.SetAtlas(colorSheet, frameSize, frameSize)
	loopAnim.SetFPS(8)
	loopAnim.SetPlayMode(ui.AnimPlayLoop)
	loopAnim.Play()
	loopAnim.SetSize(animSize, animSize)
	loopAnim.SetPosition(col(0)+(colW-animSize)/2, startY+40)
	screen.Add(loopAnim)

	loopStatus := ui.NewLabel("loop-status", "Frame: 0 / 15", font, fontSmall)
	loopStatus.SetColor(statusColor)
	loopStatus.SetPosition(col(0)+padX, startY+152)
	screen.Add(loopStatus)

	loopAnim.SetOnFrameChanged(func(f int) {
		loopStatus.SetText(fmt.Sprintf("Frame: %d / %d", f, loopAnim.FrameCount()-1))
	})

	addLabel("cap-loop", "Cycles through hues.\nRestarts from frame 0\nafter last frame.", col(0)+padX, startY+180, fontXS, captionColor)

	// ── 1. Ping-Pong mode (pulse) ───────────────────────────────────────
	addCard(1)
	addLabel("hdr-pp", "Ping-Pong", col(1)+padX, startY+padY, fontSmall, headerColor)

	ppAnim := ui.NewAnimatedImage("anim-pingpong")
	ppAnim.SetAtlas(pulseSheet, frameSize, frameSize)
	ppAnim.SetFPS(12)
	ppAnim.SetPlayMode(ui.AnimPlayPingPong)
	ppAnim.Play()
	ppAnim.SetSize(animSize, animSize)
	ppAnim.SetPosition(col(1)+(colW-animSize)/2, startY+40)
	screen.Add(ppAnim)

	ppStatus := ui.NewLabel("pp-status", "Frame: 0 / 15", font, fontSmall)
	ppStatus.SetColor(statusColor)
	ppStatus.SetPosition(col(1)+padX, startY+152)
	screen.Add(ppStatus)

	ppAnim.SetOnFrameChanged(func(f int) {
		ppStatus.SetText(fmt.Sprintf("Frame: %d / %d", f, ppAnim.FrameCount()-1))
	})

	addLabel("cap-pp", "Reverses direction at\neach end. Smooth back\nand forth.", col(1)+padX, startY+180, fontXS, captionColor)

	// ── 2. Play Once (spinner) ──────────────────────────────────────────
	addCard(2)
	addLabel("hdr-once", "Play Once", col(2)+padX, startY+padY, fontSmall, headerColor)

	onceAnim := ui.NewAnimatedImage("anim-once")
	onceAnim.SetAtlas(spinnerSheet, frameSize, frameSize)
	onceAnim.SetFPS(8)
	onceAnim.SetPlayMode(ui.AnimPlayOnce)
	onceAnim.Play()
	onceAnim.SetSize(animSize, animSize)
	onceAnim.SetPosition(col(2)+(colW-animSize)/2, startY+40)
	screen.Add(onceAnim)

	onceStatus := ui.NewLabel("once-status", "Playing...", font, fontSmall)
	onceStatus.SetColor(statusColor)
	onceStatus.SetPosition(col(2)+padX, startY+152)
	screen.Add(onceStatus)

	onceAnim.SetOnFrameChanged(func(f int) {
		onceStatus.SetText(fmt.Sprintf("Frame: %d / %d", f, onceAnim.FrameCount()-1))
	})
	onceAnim.SetOnComplete(func() {
		onceStatus.SetText("Done! Click to replay.")
		onceStatus.SetColor(willow.RGBA(1.0, 0.55, 0.35, 1))
	})
	onceAnim.OnClick(func(_ willow.ClickContext) {
		onceAnim.Stop()
		onceAnim.Play()
		onceStatus.SetText("Playing...")
		onceStatus.SetColor(statusColor)
	})

	addLabel("cap-once", "Plays forward once.\nClick to replay after\nit completes.", col(2)+padX, startY+180, fontXS, captionColor)

	// ── 3. GIF (dragon whelpling) ───────────────────────────────────────
	addCard(3)
	addLabel("hdr-gif", "GIF", col(3)+padX, startY+padY, fontSmall, headerColor)

	gifAnim := ui.NewAnimatedImage("anim-gif")
	gifAnim.LoadGIF(dragonGIF)
	gifAnim.Play()
	gifAnim.SetSize(animSize, animSize)
	gifAnim.SetPosition(col(3)+(colW-animSize)/2, startY+40)
	screen.Add(gifAnim)

	gifStatus := ui.NewLabel("gif-status", fmt.Sprintf("Frames: %d", len(dragonGIF.Image)), font, fontSmall)
	gifStatus.SetColor(statusColor)
	gifStatus.SetPosition(col(3)+padX, startY+152)
	screen.Add(gifStatus)

	gifAnim.SetOnFrameChanged(func(f int) {
		gifStatus.SetText(fmt.Sprintf("Frame: %d / %d", f, gifAnim.FrameCount()-1))
	})

	addLabel("cap-gif", "Loaded from a .gif file.\nUses per-frame delays\nfrom the GIF metadata.", col(3)+padX, startY+180, fontXS, captionColor)

	// ── Footer ──────────────────────────────────────────────────────────
	addLabel("footer", "Sprite sheets from PNG, dragon from GIF -- all loaded from disk.",
		startX, startY+cardH+12, fontXS, willow.RGBA(0.35, 0.38, 0.45, 1))

	ui.Stage.Add(screen)
	ui.Setup(ui.StageConfig{
		Title:      "AnimatedImage Widget",
		Width:      screenW,
		Height:     screenH,
		ClearColor: willow.RGBA(0.07, 0.07, 0.09, 1),
	})
}
