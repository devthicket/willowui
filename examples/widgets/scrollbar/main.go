// Reactive - ScrollBar: Custom Scroll Area
// Shows ScrollBar.BindScrollPos with Ref[float64] to wire two external
// scrollbars to a manually-clipped content viewport — without ScrollPanel.
// Demonstrates that refs can be driven from either side: dragging the
// scrollbar updates the ref, and programmatic Set() on the ref moves
// both the thumb and the content simultaneously.
package main

import (
	"fmt"

	"github.com/devthicket/willow"
	"github.com/hajimehoshi/ebiten/v2"
	ui "github.com/devthicket/willowui"
)

const (
	screenW = 800
	screenH = 500

	vpX     = 40.0  // viewport left
	vpY     = 62.0  // viewport top
	vpW     = 460.0 // visible width
	vpH     = 280.0 // visible height
	sbThick = 16.0  // scrollbar thickness

	// Content grid: 8 cols × 9 rows of tiles.
	cols    = 8
	rows    = 9
	tileW   = 90.0
	tileH   = 55.0
	tileGap = 8.0

	// Total content dimensions.
	contentW = cols*(tileW+tileGap) - tileGap // 776
	contentH = rows*(tileH+tileGap) - tileGap // 559
)

type scrollController struct {
	scrollY  *ui.Ref[float64]
	contentH float64
}

func (c *scrollController) OnCreate(s *ui.Screen) {}

func (c *scrollController) OnUpdate(dt float64) {
	// Mouse wheel scrolls vertically when the cursor is over the viewport.
	mx, my := ebiten.CursorPosition()
	if float64(mx) >= vpX && float64(mx) < vpX+vpW &&
		float64(my) >= vpY && float64(my) < vpY+vpH {
		_, wy := ebiten.Wheel()
		if wy != 0 {
			newY := c.scrollY.Peek() - wy*40
			if newY < 0 {
				newY = 0
			} else if newY > c.contentH-vpH {
				newY = c.contentH - vpH
			}
			c.scrollY.Set(newY)
		}
	}
}

func (c *scrollController) OnDestroy() {}

func main() {
	font := ui.MustLoadDefaultFont()

	const (
		sizeLarge  = 22.0
		sizeMedium = 16.0
		sizeSmall  = 16.0
	)

	// ── Reactive scroll positions ──────────────────────────────────────────
	scrollX := ui.NewRef(0.0)
	scrollY := ui.NewRef(0.0)
	hoveredTile := ui.NewRef("")

	ctrl := &scrollController{
		scrollY:  scrollY,
		contentH: contentH,
	}
	screen := ui.NewScreen(ui.WithController(ctrl))

	// Title.
	title := willow.NewText("title", "Reactive - ScrollBar: Custom Scroll Area", font)
	title.TextBlock.FontSize = sizeLarge
	title.TextBlock.Color = willow.RGBA(1, 1, 1, 1)
	title.SetPosition(24, 14)
	screen.AddNode(title)

	div := ui.NewDivider("divider", screenW)
	div.SetPosition(0, 46)
	screen.AddNode(div)

	// ── Viewport: a masked container that clips the content ────────────────
	viewport := willow.NewContainer("viewport")
	viewport.Interactable = true
	viewport.SetPosition(vpX, vpY)
	screen.AddNode(viewport)

	// Mask sprite (white, scaled to viewport size) clips children.
	maskRoot := willow.NewContainer("mask-root")
	maskSprite := willow.NewSprite("mask-rect", willow.TextureRegion{})
	maskSprite.SetColor(willow.RGBA(1, 1, 1, 1))
	maskSprite.SetScale(vpW, vpH)
	maskRoot.AddChild(maskSprite)
	viewport.SetMask(maskRoot)

	// ── Scrollable content ─────────────────────────────────────────────────
	content := willow.NewContainer("content")
	content.Interactable = true
	viewport.AddChild(content)

	tileColors := []willow.Color{
		willow.RGBA(0.22, 0.35, 0.55, 1),
		willow.RGBA(0.28, 0.48, 0.32, 1),
		willow.RGBA(0.52, 0.28, 0.30, 1),
		willow.RGBA(0.50, 0.38, 0.18, 1),
		willow.RGBA(0.38, 0.26, 0.52, 1),
		willow.RGBA(0.18, 0.42, 0.50, 1),
	}

	for row := 0; row < rows; row++ {
		for col := 0; col < cols; col++ {
			x := float64(col) * (tileW + tileGap)
			y := float64(row) * (tileH + tileGap)
			r, c := row, col

			tile := willow.NewSprite(fmt.Sprintf("tile-%d-%d", row, col), willow.TextureRegion{})
			tile.SetColor(tileColors[(row*cols+col)%len(tileColors)])
			tile.SetPosition(x, y)
			tile.SetScale(tileW, tileH)
			tile.Interactable = true
			tile.HitShape = willow.HitRect{X: 0, Y: 0, Width: tileW, Height: tileH}
			tileName := fmt.Sprintf("R%d C%d", r+1, c+1)
			inViewport := func(ctx willow.PointerContext) bool {
				return ctx.GlobalX >= vpX && ctx.GlobalX < vpX+vpW &&
					ctx.GlobalY >= vpY && ctx.GlobalY < vpY+vpH
			}
			tile.OnPointerEnter(func(ctx willow.PointerContext) {
				if !inViewport(ctx) {
					return
				}
				hoveredTile.Set(tileName)
			})
			// OnPointerMove catches the case where enter was rejected (pointer
			// was outside the viewport) and the pointer subsequently crosses
			// into the viewport while still over the same tile — no new enter
			// fires in that case, so we pick it up on the next move event.
			tile.OnPointerMove(func(ctx willow.PointerContext) {
				if !inViewport(ctx) || hoveredTile.Peek() == tileName {
					return
				}
				hoveredTile.Set(tileName)
			})
			tile.OnPointerLeave(func(_ willow.PointerContext) {
				hoveredTile.Set("")
			})
			content.AddChild(tile)

			lbl := willow.NewText(fmt.Sprintf("lbl-%d-%d", row, col),
				fmt.Sprintf("R%d C%d", row+1, col+1), font)
			lbl.TextBlock.FontSize = sizeSmall
			lbl.TextBlock.Color = willow.RGBA(1, 1, 1, 0.85)
			lbl.SetPosition(x+6, y+6)
			content.AddChild(lbl)
		}
	}

	// WatchEffect moves the content node whenever either scroll ref changes.
	ui.WatchEffect(func() {
		content.SetPosition(-scrollX.Get(), -scrollY.Get())
	})

	// ── Vertical scrollbar ─────────────────────────────────────────────────
	vScroll := ui.NewScrollBar("v-scroll")
	vScroll.SetOrientation(ui.Vertical)
	vScroll.SetSize(sbThick, vpH)
	vScroll.SetContentSize(contentH, vpH)
	vScroll.SetPosition(vpX+vpW+2, vpY)
	vScroll.BindScrollPos(scrollY)
	screen.Add(vScroll)

	// ── Horizontal scrollbar ───────────────────────────────────────────────
	hScroll := ui.NewScrollBar("h-scroll")
	hScroll.SetOrientation(ui.Horizontal)
	hScroll.SetSize(vpW, sbThick)
	hScroll.SetContentSize(contentW, vpW)
	hScroll.SetPosition(vpX, vpY+vpH+2)
	hScroll.BindScrollPos(scrollX)
	screen.Add(hScroll)

	// ── Status panel ───────────────────────────────────────────────────────
	const statusX = 545.0
	statusY := vpY

	addHeader(screen, font, sizeSmall, "Reactive state", statusX, statusY)
	statusY += 22

	xLbl := ui.NewLabel("x-lbl", "scrollX: 0.0", font, sizeSmall)
	xLbl.SetColor(willow.RGBA(0.7, 0.8, 0.5, 1))
	xLbl.SetPosition(statusX, statusY)
	screen.Add(xLbl)
	ui.WatchValue(scrollX, func(_, v float64) {
		xLbl.SetText(fmt.Sprintf("scrollX: %.1f", v))
	})
	statusY += 18

	yLbl := ui.NewLabel("y-lbl", "scrollY: 0.0", font, sizeSmall)
	yLbl.SetColor(willow.RGBA(0.7, 0.8, 0.5, 1))
	yLbl.SetPosition(statusX, statusY)
	screen.Add(yLbl)
	ui.WatchValue(scrollY, func(_, v float64) {
		yLbl.SetText(fmt.Sprintf("scrollY: %.1f", v))
	})
	statusY += 30

	// Computed: visible range summary.
	addHeader(screen, font, sizeSmall, "Visible range", statusX, statusY)
	statusY += 22

	xRangeLbl := ui.NewLabel("x-range", "X: 0 – 460 / 776", font, sizeSmall)
	xRangeLbl.SetColor(willow.RGBA(0.8, 0.7, 0.5, 1))
	xRangeLbl.SetPosition(statusX, statusY)
	screen.Add(xRangeLbl)
	statusY += 18

	yRangeLbl := ui.NewLabel("y-range", "Y: 0 – 280 / 559", font, sizeSmall)
	yRangeLbl.SetColor(willow.RGBA(0.8, 0.7, 0.5, 1))
	yRangeLbl.SetPosition(statusX, statusY)
	screen.Add(yRangeLbl)
	statusY += 38

	ui.WatchEffect(func() {
		x0 := scrollX.Get()
		y0 := scrollY.Get()
		xRangeLbl.SetText(fmt.Sprintf("X: %.0f – %.0f / %.0f", x0, x0+vpW, contentW))
		yRangeLbl.SetText(fmt.Sprintf("Y: %.0f – %.0f / %.0f", y0, y0+vpH, contentH))
	})

	// ── Programmatic control ───────────────────────────────────────────────
	addHeader(screen, font, sizeSmall, "Programmatic Set()", statusX, statusY)
	statusY += 22

	centerBtn := ui.NewButton("center-btn", "Jump to Center", font, sizeMedium)
	centerBtn.SetSize(180, 36)
	centerBtn.SetPosition(statusX, statusY)
	centerBtn.SetOnClick(func() {
		scrollX.Set((contentW - vpW) / 2)
		scrollY.Set((contentH - vpH) / 2)
	})
	screen.Add(centerBtn)
	statusY += 46

	resetBtn := ui.NewButton("reset-btn", "Reset to Origin", font, sizeMedium)
	resetBtn.SetSize(180, 36)
	resetBtn.SetPosition(statusX, statusY)
	resetBtn.SetOnClick(func() {
		scrollX.Set(0)
		scrollY.Set(0)
	})
	screen.Add(resetBtn)
	statusY += 46

	// ── Hovered tile ───────────────────────────────────────────────────────
	addHeader(screen, font, sizeSmall, "Hovered tile", statusX, statusY)
	statusY += 22

	hoveredLbl := ui.NewLabel("hovered-lbl", "--", font, sizeMedium)
	hoveredLbl.SetColor(willow.RGBA(1, 0.85, 0.5, 1))
	hoveredLbl.SetPosition(statusX, statusY)
	screen.Add(hoveredLbl)

	ui.WatchValue(hoveredTile, func(_, v string) {
		if v == "" {
			hoveredLbl.SetText("--")
		} else {
			hoveredLbl.SetText(v)
		}
	})

	ui.Stage.Add(screen)
	ui.Setup(ui.StageConfig{
		Title:      "Reactive - ScrollBar",
		Width:      screenW,
		Height:     screenH,
		ClearColor: willow.RGBA(0.08, 0.08, 0.10, 1),
	})
}

func addHeader(screen *ui.Screen, font *willow.FontFamily, fontSize float64, text string, x, y float64) {
	n := willow.NewText("hdr", text, font)
	n.TextBlock.FontSize = fontSize
	n.TextBlock.Color = willow.RGBA(0.45, 0.55, 0.65, 1)
	n.SetPosition(x, y)
	screen.AddNode(n)
}
