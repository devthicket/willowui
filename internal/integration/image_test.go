package integration

import (
	"testing"

	"github.com/devthicket/willow"
	"github.com/hajimehoshi/ebiten/v2"
	ui "github.com/devthicket/willowui"
)

// --- Construction ---

func TestNewImageCreates(t *testing.T) {
	resetScheduler()
	im := ui.NewImage("img")
	defer im.Dispose()

	if im.Node() == nil {
		t.Fatal("Node() should not be nil")
	}
	if im.Name() != "img" {
		t.Errorf("Name() = %q, want %q", im.Name(), "img")
	}
}

func TestNewImageDefaultScaleMode(t *testing.T) {
	resetScheduler()
	im := ui.NewImage("img")
	defer im.Dispose()

	if im.ScaleMode() != ui.ImageScaleStretch {
		t.Errorf("default ScaleMode = %d, want ImageScaleStretch", im.ScaleMode())
	}
}

func TestNewImageDefaultSize(t *testing.T) {
	resetScheduler()
	im := ui.NewImage("img")
	defer im.Dispose()

	if im.Width != 64 || im.Height != 64 {
		t.Errorf("default size = (%v,%v), want (64,64)", im.Width, im.Height)
	}
}

// --- SetSize ---

func TestImageSetSize(t *testing.T) {
	resetScheduler()
	im := ui.NewImage("img")
	defer im.Dispose()

	im.SetSize(128, 96)
	if im.Width != 128 || im.Height != 96 {
		t.Errorf("SetSize: got (%v,%v), want (128,96)", im.Width, im.Height)
	}
}

// --- Source ---

func TestImageSetImage(t *testing.T) {
	resetScheduler()
	im := ui.NewImage("img")
	defer im.Dispose()

	src := ebiten.NewImage(32, 32)
	im.SetImage(src)

	if im.Tint().A() == 0 {
		t.Error("Tint alpha should not be 0 after SetImage")
	}
}

func TestImageClearImageNoPanic(t *testing.T) {
	resetScheduler()
	im := ui.NewImage("img")
	defer im.Dispose()

	// ClearImage on an empty widget should not panic.
	im.ClearImage()

	src := ebiten.NewImage(16, 16)
	im.SetImage(src)
	im.ClearImage()
}

func TestImageSetRegion(t *testing.T) {
	resetScheduler()
	im := ui.NewImage("img")
	defer im.Dispose()

	region := willow.TextureRegion{Width: 64, Height: 64}
	im.SetRegion(region)
	// No panic and widget still valid.
	if im.Node() == nil {
		t.Fatal("Node() nil after SetRegion")
	}
}

// --- ScaleMode ---

func TestImageSetScaleMode(t *testing.T) {
	resetScheduler()
	im := ui.NewImage("img")
	defer im.Dispose()

	modes := []ui.ImageScaleMode{
		ui.ImageScaleStretch,
		ui.ImageScaleFit,
		ui.ImageScaleFill,
		ui.ImageScaleCenter,
		ui.ImageScaleTile,
	}
	for _, mode := range modes {
		im.SetScaleMode(mode)
		if im.ScaleMode() != mode {
			t.Errorf("ScaleMode() = %d, want %d", im.ScaleMode(), mode)
		}
	}
}

// --- Stretch layout ---

func TestImageStretchScalesNodeToWidgetSize(t *testing.T) {
	resetScheduler()
	im := ui.NewImage("img")
	defer im.Dispose()

	src := ebiten.NewImage(32, 16)
	im.SetImage(src)
	im.SetScaleMode(ui.ImageScaleStretch)
	im.SetSize(128, 64)

	w, h := im.ImgSize()
	if w != 128 || h != 64 {
		t.Errorf("Stretch size = (%v,%v), want (128,64)", w, h)
	}
}

// --- Fit layout ---

func TestImageFitUniformScaleLetterbox(t *testing.T) {
	resetScheduler()
	im := ui.NewImage("img")
	defer im.Dispose()

	// Image 100x50 in widget 200x200 → uniform scale 2, result 200x100, y-centered at 50.
	src := ebiten.NewImage(100, 50)
	im.SetImage(src)
	im.SetScaleMode(ui.ImageScaleFit)
	im.SetSize(200, 200)

	w, h := im.ImgSize()
	if w != 200 || h != 100 {
		t.Errorf("Fit size = (%v,%v), want (200,100)", w, h)
	}
	x, y := im.ImgPosition()
	if x != 0 || y != 50 {
		t.Errorf("Fit position = (%v,%v), want (0,50)", x, y)
	}
}

// --- Fill layout ---

func TestImageFillUniformScaleCrop(t *testing.T) {
	resetScheduler()
	im := ui.NewImage("img")
	defer im.Dispose()

	// Image 100x50 in widget 200x200 → Fill pre-renders to a 200x200 clip canvas
	// positioned at (0,0); the scaled image is drawn into the canvas with cropping.
	src := ebiten.NewImage(100, 50)
	im.SetImage(src)
	im.SetScaleMode(ui.ImageScaleFill)
	im.SetSize(200, 200)

	w, h := im.ImgSize()
	if w != 200 || h != 200 {
		t.Errorf("Fill canvas size = (%v,%v), want (200,200)", w, h)
	}
	x, y := im.ImgPosition()
	if x != 0 || y != 0 {
		t.Errorf("Fill canvas position = (%v,%v), want (0,0)", x, y)
	}
}

// --- Center layout ---

func TestImageCenterNoScaleCentered(t *testing.T) {
	resetScheduler()
	im := ui.NewImage("img")
	defer im.Dispose()

	// Image 40x40 in widget 100x100 → Center pre-renders to a 100x100 clip canvas
	// with the image drawn centered (at 30,30) inside it.
	src := ebiten.NewImage(40, 40)
	im.SetImage(src)
	im.SetScaleMode(ui.ImageScaleCenter)
	im.SetSize(100, 100)

	w, h := im.ImgSize()
	if w != 100 || h != 100 {
		t.Errorf("Center canvas size = (%v,%v), want (100,100)", w, h)
	}
	x, y := im.ImgPosition()
	if x != 0 || y != 0 {
		t.Errorf("Center canvas position = (%v,%v), want (0,0)", x, y)
	}
}

// --- SizeToContent ---

func TestImageSizeToContentMatchesNativeDimensions(t *testing.T) {
	resetScheduler()
	im := ui.NewImage("img")
	defer im.Dispose()

	src := ebiten.NewImage(80, 60)
	im.SetImage(src)
	im.SizeToContent()

	if im.Width != 80 || im.Height != 60 {
		t.Errorf("SizeToContent: got (%v,%v), want (80,60)", im.Width, im.Height)
	}
}

func TestImageSizeToContentNoOpWhenNoImage(t *testing.T) {
	resetScheduler()
	im := ui.NewImage("img")
	defer im.Dispose()

	im.SetSize(64, 64)
	im.SizeToContent() // should be a no-op
	if im.Width != 64 || im.Height != 64 {
		t.Errorf("SizeToContent with no image changed size to (%v,%v)", im.Width, im.Height)
	}
}

// --- Tint ---

func TestImageSetTint(t *testing.T) {
	resetScheduler()
	im := ui.NewImage("img")
	defer im.Dispose()

	tint := willow.RGBA(1, 0, 0, 0.5)
	im.SetTint(tint)
	got := im.Tint()
	if got != tint {
		t.Errorf("Tint() = %v, want %v", got, tint)
	}
}

func TestImageSetAlpha(t *testing.T) {
	resetScheduler()
	im := ui.NewImage("img")
	defer im.Dispose()

	im.SetAlpha(0.3)
	tint := im.Tint()
	// Alpha should be approximately 0.3 (float32 conversion).
	if tint.A() < 0.29 || tint.A() > 0.31 {
		t.Errorf("SetAlpha(0.3) → Tint.A = %v, want ~0.3", tint.A())
	}
}

// --- CornerRadius ---

func TestImageSetCornerRadius(t *testing.T) {
	resetScheduler()
	im := ui.NewImage("img")
	defer im.Dispose()

	// Setting a positive corner radius should not panic.
	im.SetSize(100, 100)
	im.SetCornerRadius(8)
}

func TestImageSetCornerRadiusPill(t *testing.T) {
	resetScheduler()
	im := ui.NewImage("img")
	defer im.Dispose()

	// -1 = pill: corner radius should resolve to half of min(w,h).
	im.SetSize(100, 60)
	im.SetCornerRadius(-1) // should resolve to 30, no panic
}
