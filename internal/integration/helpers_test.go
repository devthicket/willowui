package integration

import (
	"log"

	"github.com/devthicket/willow"
	"github.com/devthicket/willowui/internal/reactive"
	"golang.org/x/image/font/gofont/goregular"
)

var _testFont *willow.FontFamily

func init() {
	f, err := willow.NewFontFamilyFromTTF(willow.FontFamilyConfig{
		Regular: goregular.TTF,
	})
	if err != nil {
		log.Fatalf("failed to create test font: %v", err)
	}
	_testFont = f
}

func newTestFont() *willow.FontFamily {
	return _testFont
}

func newLargeTestFont() *willow.FontFamily {
	// Same font family — display size controls the rendered size.
	return _testFont
}

// resetScheduler clears the default scheduler between tests.
func resetScheduler() {
	reactive.DefaultScheduler = reactive.Scheduler{}
	reactive.TrackingStack = reactive.TrackingStack[:0]
}

func newTestScene() *willow.Scene {
	return willow.NewScene()
}
