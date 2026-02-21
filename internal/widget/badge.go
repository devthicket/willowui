package widget

import (
	"fmt"

	"github.com/devthicket/willowui/internal/core"
	"github.com/devthicket/willowui/internal/render"
	"github.com/devthicket/willowui/internal/sg"
	"github.com/devthicket/willowui/internal/theme"
)

// Badge is a small pill or dot overlay that displays a count, label, or
// status indicator. It supports two display modes: label mode (short text
// in a pill shape) and dot mode (small filled circle, no text).
type Badge struct {
	Component
	textNode    *sg.Node
	font        *sg.FontFamily
	displaySize float64
	text        string
	maxCount    int
	dotMode     bool
	padOverride bool // true when SetPadding has been called
}

// NewBadge creates a Badge with the given name, font source, and display size.
func NewBadge(name string, source *sg.FontFamily, displaySize float64) *Badge {
	var font *sg.FontFamily
	if source != nil {
		font = source
	}
	b := &Badge{
		font:        font,
		displaySize: displaySize,
		maxCount:    99,
	}
	initComponent(&b.Component, name)
	b.initBackground(name)

	b.textNode = sg.NewText(name+"-text", "", font)
	b.textNode.TextBlock.FontSize = displaySize
	b.node.AddChild(b.textNode)

	b.onThemeChange = func() { b.UpdateVisuals() }

	return b
}

// Text returns the current text content.
func (b *Badge) Text() string {
	return b.text
}

// SetText updates the displayed text and re-sizes if in auto-size mode.
func (b *Badge) SetText(text string) {
	b.text = text
	b.syncText()
}

// SetCount sets the text to the string representation of n, clamped by MaxCount.
func (b *Badge) SetCount(n int) {
	if b.maxCount > 0 && n > b.maxCount {
		b.text = fmt.Sprintf("%d+", b.maxCount)
	} else {
		b.text = fmt.Sprint(n)
	}
	b.syncText()
}

// SetMaxCount sets the maximum displayed count. Values above this show "N+".
// Default is 99. Set to 0 to disable truncation.
func (b *Badge) SetMaxCount(n int) {
	b.maxCount = n
}

// SetDotMode enables or disables dot mode (small circle, no text).
func (b *Badge) SetDotMode(v bool) {
	b.dotMode = v
	b.syncText()
}

// SetPadding overrides the theme padding with per-instance values.
func (b *Badge) SetPadding(top, right, bottom, left float64) {
	b.Padding = render.Insets{Top: top, Right: right, Bottom: bottom, Left: left}
	b.padOverride = true
}

// SetSize overrides the auto-size dimensions.
func (b *Badge) SetSize(w, h float64) {
	b.Width = w
	b.Height = h
	b.resizeBackground(w, h)
	b.node.HitShape = sg.HitRect{X: 0, Y: 0, Width: w, Height: h}
	b.positionText()
	b.UpdateVisuals()
	b.MarkLayoutDirty()
}

// SizeToContent auto-sizes the badge to fit its content plus padding.
func (b *Badge) SizeToContent() {
	group := b.EffectiveTheme().Badge.Group(b.Variant())

	if b.dotMode {
		dotSize := group.DotSize
		if dotSize <= 0 {
			dotSize = 8
		}
		b.Width = dotSize
		b.Height = dotSize
	} else {
		pad := b.resolvedPadding(group)
		tw, th := measureDisplay(b.font, b.textNode.TextBlock.Content, b.displaySize)
		b.Width = tw + pad.Horizontal()
		b.Height = th + pad.Vertical()
		// Ensure minimum pill width (at least as wide as tall).
		if b.Width < b.Height {
			b.Width = b.Height
		}
	}

	b.resizeBackground(b.Width, b.Height)
	b.node.HitShape = sg.HitRect{X: 0, Y: 0, Width: b.Width, Height: b.Height}
	b.positionText()
	b.UpdateVisuals()
	b.MarkLayoutDirty()
}

// UpdateVisuals applies theme colors and corner radius.
func (b *Badge) UpdateVisuals() {
	th := b.EffectiveTheme()
	group := th.Badge.Group(b.Variant())

	// Resolve corner radius (pill by default).
	cr := resolveCornerRadius(group.CornerRadius, b.Height)
	b.applyCornerRadius(cr)

	// Background.
	b.applyBackground(group.Background[core.StateDefault])

	// Text color.
	b.textNode.SetTextColor(group.TextColor.Resolve(core.StateDefault))

	// Hide text in dot mode.
	b.textNode.SetVisible(!b.dotMode)

	b.MarkDrawDirty()
}

// resolvedPadding returns the per-instance padding if set, otherwise the theme padding.
func (b *Badge) resolvedPadding(group *theme.BadgeGroup) render.Insets {
	if b.padOverride {
		return b.Padding
	}
	return group.Padding
}

// syncText updates the text node content based on current state.
func (b *Badge) syncText() {
	display := b.text
	if b.dotMode {
		display = ""
	}
	b.textNode.SetContent(display)
	b.MarkDrawDirty()
}

// positionText centers the text node within the badge.
func (b *Badge) positionText() {
	if b.font == nil {
		return
	}
	tw, th := measureDisplay(b.font, b.textNode.TextBlock.Content, b.displaySize)
	b.textNode.SetPosition(
		(b.Width-tw)/2,
		(b.Height-th)/2,
	)
}
