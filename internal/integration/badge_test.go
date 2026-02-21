package integration

import (
	"testing"

	ui "github.com/devthicket/willowui"
)

func TestBadgeSetCountWithMaxCountDisplaysTruncated(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	b := ui.NewBadge("b", font, 0)
	defer b.Dispose()

	b.SetMaxCount(99)
	b.SetCount(100)

	if b.Text() != "99+" {
		t.Errorf("Text() = %q, want %q", b.Text(), "99+")
	}
}

func TestBadgeSetCountBelowMaxDisplaysExact(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	b := ui.NewBadge("b", font, 0)
	defer b.Dispose()

	b.SetMaxCount(99)
	b.SetCount(50)

	if b.Text() != "50" {
		t.Errorf("Text() = %q, want %q", b.Text(), "50")
	}
}

func TestBadgeSizeToContentSizesToTextPlusPadding(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	b := ui.NewBadge("b", font, 0)
	defer b.Dispose()

	b.SetText("42")
	b.SizeToContent()

	textW, textH := font.MeasureString("42", 0, false, false)
	pad := ui.DefaultTheme.Badge.Group(ui.Primary).Padding
	wantW := textW + pad.Left + pad.Right
	wantH := textH + pad.Top + pad.Bottom
	if wantW < wantH {
		wantW = wantH // min width = height
	}
	if b.Width != wantW {
		t.Errorf("Width = %f, want %f", b.Width, wantW)
	}
	if b.Height != wantH {
		t.Errorf("Height = %f, want %f", b.Height, wantH)
	}

	// Verify that wider text produces wider badge.
	b.SetText("9999")
	b.SizeToContent()
	widerW, _ := font.MeasureString("9999", 0, false, false)
	wantWider := widerW + pad.Left + pad.Right
	if wantWider < wantH {
		wantWider = wantH
	}
	if b.Width != wantWider {
		t.Errorf("Width = %f, want %f after wider text", b.Width, wantWider)
	}
}

func TestBadgeDotModeHidesText(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	b := ui.NewBadge("b", font, 0)
	defer b.Dispose()

	b.SetText("42")
	b.SetDotMode(true)
	b.SizeToContent()

	// Dot mode: size should be DotSize (default 8)
	if b.Width != 8 {
		t.Errorf("Width = %f, want 8 (dot mode)", b.Width)
	}
	if b.Height != 8 {
		t.Errorf("Height = %f, want 8 (dot mode)", b.Height)
	}
}

func TestBadgeSetVariantDoesNotPanic(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	b := ui.NewBadge("b", font, 0)
	defer b.Dispose()

	variants := []ui.Variant{
		ui.Primary,
		ui.Success,
		ui.Warning,
		ui.Danger,
		ui.Neutral,
	}
	for _, v := range variants {
		b.SetVariant(v)
	}
}

func TestBadgeDefaultMaxCountIs99(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	b := ui.NewBadge("b", font, 0)
	defer b.Dispose()

	b.SetCount(100)
	if b.Text() != "99+" {
		t.Errorf("Text() = %q, want %q (default max 99)", b.Text(), "99+")
	}
}

func TestBadgeMaxCountZeroDisablesTruncation(t *testing.T) {
	resetScheduler()
	font := newTestFont()
	b := ui.NewBadge("b", font, 0)
	defer b.Dispose()

	b.SetMaxCount(0)
	b.SetCount(999)

	if b.Text() != "999" {
		t.Errorf("Text() = %q, want %q (no truncation)", b.Text(), "999")
	}
}
