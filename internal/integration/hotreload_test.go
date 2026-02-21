//go:build hotreload

package integration

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	ui "github.com/devthicket/willowui"
	"github.com/devthicket/willowui/internal/template"
)

func TestHotReloader_Reload(t *testing.T) {
	resetScheduler()

	// Write initial XML to a temp file.
	dir := t.TempDir()
	xmlPath := filepath.Join(dir, "template.xml")
	initial := []byte(`<Panel layout="vbox"><Label text="Hello" /></Panel>`)
	if err := os.WriteFile(xmlPath, initial, 0644); err != nil {
		t.Fatal(err)
	}

	reg := ui.NewTemplateRegistry()
	reg.SetFonts(nil, nil)

	ctrl := &xmlTestController{
		provider: &testDataProvider{refs: map[string]any{}},
	}

	// Register and instantiate the initial template.
	xmlData, _ := os.ReadFile(xmlPath)
	if err := reg.RegisterXML("test", xmlData); err != nil {
		t.Fatal(err)
	}

	screen := ui.NewScreen(ui.WithController(ctrl))

	comp, err := reg.Instantiate("test", ctrl, screen)
	if err != nil {
		t.Fatal(err)
	}
	screen.Add(comp)

	// Verify initial state.
	if screen.NumChildren() != 1 {
		t.Fatalf("initial children = %d, want 1", screen.NumChildren())
	}
	panel := screen.Children()[0]
	if panel.NumChildren() != 1 {
		t.Fatalf("initial panel children = %d, want 1", panel.NumChildren())
	}

	// Overwrite with modified XML.
	modified := []byte(`<Panel layout="vbox"><Label text="World" /><Button text="New" /></Panel>`)
	if err := os.WriteFile(xmlPath, modified, 0644); err != nil {
		t.Fatal(err)
	}

	// Create a HotReloader and call Reload directly (skip file watcher).
	hr := ui.NewHotReloaderDirect(reg, screen, ctrl, "test", xmlPath)

	if err := hr.Reload(); err != nil {
		t.Fatal(err)
	}

	// Verify updated state.
	if screen.NumChildren() != 1 {
		t.Fatalf("reloaded children = %d, want 1", screen.NumChildren())
	}
	newPanel := screen.Children()[0]
	if newPanel.NumChildren() != 2 {
		t.Fatalf("reloaded panel children = %d, want 2", newPanel.NumChildren())
	}
}

func TestHotReloader_ReloadSkipsUnchanged(t *testing.T) {
	resetScheduler()

	dir := t.TempDir()
	xmlPath := filepath.Join(dir, "template.xml")
	xml := []byte(`<Panel layout="vbox"><Label text="Hello" /></Panel>`)
	if err := os.WriteFile(xmlPath, xml, 0644); err != nil {
		t.Fatal(err)
	}

	reg := ui.NewTemplateRegistry()
	reg.SetFonts(nil, nil)

	ctrl := &xmlTestController{
		provider: &testDataProvider{refs: map[string]any{}},
	}

	if err := reg.RegisterXML("test", xml); err != nil {
		t.Fatal(err)
	}

	screen := ui.NewScreen(ui.WithController(ctrl))

	comp, err := reg.Instantiate("test", ctrl, screen)
	if err != nil {
		t.Fatal(err)
	}
	screen.Add(comp)

	hr, err := ui.NewHotReloader(reg, screen, ctrl, "test", xmlPath)
	if err != nil {
		t.Fatal(err)
	}
	defer hr.Stop()

	// Grab a reference to the current child.
	originalChild := screen.Children()[0]

	// Reload with identical content — tree should NOT be swapped.
	if err := hr.Reload(); err != nil {
		t.Fatal(err)
	}

	if screen.Children()[0] != originalChild {
		t.Error("expected tree to be unchanged when content is identical")
	}

	// Now write different content and reload — tree SHOULD be swapped.
	modified := []byte(`<Panel layout="vbox"><Label text="Changed" /></Panel>`)
	if err := os.WriteFile(xmlPath, modified, 0644); err != nil {
		t.Fatal(err)
	}

	if err := hr.Reload(); err != nil {
		t.Fatal(err)
	}

	if screen.Children()[0] == originalChild {
		t.Error("expected tree to be replaced when content changed")
	}
}

func TestHotReloader_ReloadUpdatesBindings(t *testing.T) {
	resetScheduler()

	dir := t.TempDir()
	xmlPath := filepath.Join(dir, "template.xml")
	initial := []byte(`<Label bind:text="title" />`)
	if err := os.WriteFile(xmlPath, initial, 0644); err != nil {
		t.Fatal(err)
	}

	titleRef := ui.NewRef("first")
	reg := ui.NewTemplateRegistry()
	reg.SetFonts(nil, nil)

	ctrl := &xmlTestController{
		provider: &testDataProvider{
			refs: map[string]any{"title": titleRef},
		},
	}

	xmlData, _ := os.ReadFile(xmlPath)
	if err := reg.RegisterXML("test", xmlData); err != nil {
		t.Fatal(err)
	}

	screen := ui.NewScreen(ui.WithController(ctrl))

	comp, err := reg.Instantiate("test", ctrl, screen)
	if err != nil {
		t.Fatal(err)
	}
	screen.Add(comp)

	label := template.CompAsLabel(screen.Children()[0])
	if label == nil {
		t.Fatal("expected Label")
	}
	if label.Text() != "first" {
		t.Errorf("initial text = %q, want first", label.Text())
	}

	// Change ref and reload with new static text.
	modified := []byte(`<Label text="static-text" />`)
	if err := os.WriteFile(xmlPath, modified, 0644); err != nil {
		t.Fatal(err)
	}

	hr := ui.NewHotReloaderDirect(reg, screen, ctrl, "test", xmlPath)

	if err := hr.Reload(); err != nil {
		t.Fatal(err)
	}

	newLabel := template.CompAsLabel(screen.Children()[0])
	if newLabel == nil {
		t.Fatal("expected Label after reload")
	}
	if newLabel.Text() != "static-text" {
		t.Errorf("reloaded text = %q, want static-text", newLabel.Text())
	}
}

func TestHotReloader_WatcherDirectWrite(t *testing.T) {
	resetScheduler()

	dir := t.TempDir()
	xmlPath := filepath.Join(dir, "template.xml")
	initial := []byte(`<Panel layout="vbox"><Label text="Before" /></Panel>`)
	if err := os.WriteFile(xmlPath, initial, 0644); err != nil {
		t.Fatal(err)
	}

	reg := ui.NewTemplateRegistry()
	reg.SetFonts(nil, nil)

	ctrl := &xmlTestController{
		provider: &testDataProvider{refs: map[string]any{}},
	}

	xmlData, _ := os.ReadFile(xmlPath)
	if err := reg.RegisterXML("test", xmlData); err != nil {
		t.Fatal(err)
	}

	screen := ui.NewScreen(ui.WithController(ctrl))

	comp, err := reg.Instantiate("test", ctrl, screen)
	if err != nil {
		t.Fatal(err)
	}
	screen.Add(comp)

	// Start the real file watcher.
	hr, err := ui.NewHotReloader(reg, screen, ctrl, "test", xmlPath)
	if err != nil {
		t.Fatal(err)
	}
	defer hr.Stop()

	// Direct write to the file (like echo > file).
	modified := []byte(`<Panel layout="vbox"><Label text="After" /><Button text="New" /></Panel>`)
	if err := os.WriteFile(xmlPath, modified, 0644); err != nil {
		t.Fatal(err)
	}

	// Wait for debounce (200ms) + processing time.
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if screen.NumChildren() == 1 {
			panel := screen.Children()[0]
			if panel.NumChildren() == 2 {
				return // success
			}
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatal("watcher did not reload after direct write within timeout")
}

func TestHotReloader_WatcherAtomicSave(t *testing.T) {
	resetScheduler()

	dir := t.TempDir()
	xmlPath := filepath.Join(dir, "template.xml")
	initial := []byte(`<Panel layout="vbox"><Label text="Before" /></Panel>`)
	if err := os.WriteFile(xmlPath, initial, 0644); err != nil {
		t.Fatal(err)
	}

	reg := ui.NewTemplateRegistry()
	reg.SetFonts(nil, nil)

	ctrl := &xmlTestController{
		provider: &testDataProvider{refs: map[string]any{}},
	}

	xmlData, _ := os.ReadFile(xmlPath)
	if err := reg.RegisterXML("test", xmlData); err != nil {
		t.Fatal(err)
	}

	screen := ui.NewScreen(ui.WithController(ctrl))

	comp, err := reg.Instantiate("test", ctrl, screen)
	if err != nil {
		t.Fatal(err)
	}
	screen.Add(comp)

	hr, err := ui.NewHotReloader(reg, screen, ctrl, "test", xmlPath)
	if err != nil {
		t.Fatal(err)
	}
	defer hr.Stop()

	// Simulate atomic save: write to temp file, then rename over original.
	modified := []byte(`<Panel layout="vbox"><Label text="Atomic" /><Button text="Added" /></Panel>`)
	tmpPath := xmlPath + ".tmp"
	if err := os.WriteFile(tmpPath, modified, 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.Rename(tmpPath, xmlPath); err != nil {
		t.Fatal(err)
	}

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if screen.NumChildren() == 1 {
			panel := screen.Children()[0]
			if panel.NumChildren() == 2 {
				return // success
			}
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatal("watcher did not reload after atomic save within timeout")
}

func TestHotReloader_WatcherGoLandSave(t *testing.T) {
	resetScheduler()

	dir := t.TempDir()
	xmlPath := filepath.Join(dir, "template.xml")
	initial := []byte(`<Panel layout="vbox"><Label text="Before" /></Panel>`)
	if err := os.WriteFile(xmlPath, initial, 0644); err != nil {
		t.Fatal(err)
	}

	reg := ui.NewTemplateRegistry()
	reg.SetFonts(nil, nil)

	ctrl := &xmlTestController{
		provider: &testDataProvider{refs: map[string]any{}},
	}

	xmlData, _ := os.ReadFile(xmlPath)
	if err := reg.RegisterXML("test", xmlData); err != nil {
		t.Fatal(err)
	}

	screen := ui.NewScreen(ui.WithController(ctrl))

	comp, err := reg.Instantiate("test", ctrl, screen)
	if err != nil {
		t.Fatal(err)
	}
	screen.Add(comp)

	hr, err := ui.NewHotReloader(reg, screen, ctrl, "test", xmlPath)
	if err != nil {
		t.Fatal(err)
	}
	defer hr.Stop()

	// Simulate GoLand's three-rename safe write:
	// 1. Write new content to template.xml___jb_tmp___
	// 2. Rename template.xml -> template.xml___jb_old___
	// 3. Rename template.xml___jb_tmp___ -> template.xml
	// 4. Delete template.xml___jb_old___
	modified := []byte(`<Panel layout="vbox"><Label text="GoLand" /><Button text="Added" /></Panel>`)
	tmpPath := xmlPath + "___jb_tmp___"
	oldPath := xmlPath + "___jb_old___"

	if err := os.WriteFile(tmpPath, modified, 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.Rename(xmlPath, oldPath); err != nil {
		t.Fatal(err)
	}
	if err := os.Rename(tmpPath, xmlPath); err != nil {
		t.Fatal(err)
	}
	os.Remove(oldPath)

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if screen.NumChildren() == 1 {
			panel := screen.Children()[0]
			if panel.NumChildren() == 2 {
				return // success
			}
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatal("watcher did not reload after GoLand-style safe write within timeout")
}

func TestHotReloader_ReloadTheme(t *testing.T) {
	resetScheduler()

	dir := t.TempDir()
	xmlPath := filepath.Join(dir, "template.xml")
	xmlData := []byte(`<Panel layout="vbox"><Label text="Hello" /></Panel>`)
	if err := os.WriteFile(xmlPath, xmlData, 0644); err != nil {
		t.Fatal(err)
	}
	jsonPath := filepath.Join(dir, "theme.json")
	themeV1 := []byte(`{}`)
	if err := os.WriteFile(jsonPath, themeV1, 0644); err != nil {
		t.Fatal(err)
	}

	reg := ui.NewTemplateRegistry()
	reg.SetFonts(nil, nil)

	ctrl := &xmlTestController{
		provider: &testDataProvider{refs: map[string]any{}},
	}
	if err := reg.RegisterXML("test", xmlData); err != nil {
		t.Fatal(err)
	}

	screen := ui.NewScreen(ui.WithController(ctrl))
	comp, err := reg.Instantiate("test", ctrl, screen)
	if err != nil {
		t.Fatal(err)
	}
	screen.Add(comp)

	hr := ui.NewHotReloaderDirect(reg, screen, ctrl, "test", xmlPath)
	if err := hr.WatchTheme(jsonPath); err != nil {
		t.Fatal(err)
	}

	// Grab initial theme.
	initialTheme := screen.Children()[0].EffectiveTheme()

	// Write modified theme and reload.
	themeV2 := []byte(`{"components":{}}`)
	if err := os.WriteFile(jsonPath, themeV2, 0644); err != nil {
		t.Fatal(err)
	}
	if err := hr.ReloadTheme(); err != nil {
		t.Fatal(err)
	}

	newTheme := screen.Children()[0].EffectiveTheme()
	if newTheme == initialTheme {
		t.Error("expected theme to change after ReloadTheme")
	}
}

func TestHotReloader_ReloadThemeSkipsUnchanged(t *testing.T) {
	resetScheduler()

	dir := t.TempDir()
	xmlPath := filepath.Join(dir, "template.xml")
	xmlData := []byte(`<Panel layout="vbox"><Label text="Hello" /></Panel>`)
	if err := os.WriteFile(xmlPath, xmlData, 0644); err != nil {
		t.Fatal(err)
	}
	jsonPath := filepath.Join(dir, "theme.json")
	themeJSON := []byte(`{}`)
	if err := os.WriteFile(jsonPath, themeJSON, 0644); err != nil {
		t.Fatal(err)
	}

	reg := ui.NewTemplateRegistry()
	reg.SetFonts(nil, nil)

	ctrl := &xmlTestController{
		provider: &testDataProvider{refs: map[string]any{}},
	}
	if err := reg.RegisterXML("test", xmlData); err != nil {
		t.Fatal(err)
	}

	screen := ui.NewScreen(ui.WithController(ctrl))
	comp, err := reg.Instantiate("test", ctrl, screen)
	if err != nil {
		t.Fatal(err)
	}
	screen.Add(comp)

	hr := ui.NewHotReloaderDirect(reg, screen, ctrl, "test", xmlPath)
	if err := hr.WatchTheme(jsonPath); err != nil {
		t.Fatal(err)
	}

	// First reload sets the theme.
	if err := hr.ReloadTheme(); err != nil {
		t.Fatal(err)
	}
	themeAfterFirst := screen.Children()[0].EffectiveTheme()

	// Second reload with identical content — theme should remain the same.
	if err := hr.ReloadTheme(); err != nil {
		t.Fatal(err)
	}
	themeAfterSecond := screen.Children()[0].EffectiveTheme()

	if themeAfterFirst != themeAfterSecond {
		t.Error("expected theme to remain unchanged when JSON is identical")
	}
}

func TestHotReloader_WatcherThemeFile(t *testing.T) {
	resetScheduler()

	dir := t.TempDir()
	xmlPath := filepath.Join(dir, "template.xml")
	xmlData := []byte(`<Panel layout="vbox"><Label text="Hello" /></Panel>`)
	if err := os.WriteFile(xmlPath, xmlData, 0644); err != nil {
		t.Fatal(err)
	}
	jsonPath := filepath.Join(dir, "theme.json")
	themeV1 := []byte(`{}`)
	if err := os.WriteFile(jsonPath, themeV1, 0644); err != nil {
		t.Fatal(err)
	}

	reg := ui.NewTemplateRegistry()
	reg.SetFonts(nil, nil)

	ctrl := &xmlTestController{
		provider: &testDataProvider{refs: map[string]any{}},
	}
	if err := reg.RegisterXML("test", xmlData); err != nil {
		t.Fatal(err)
	}

	screen := ui.NewScreen(ui.WithController(ctrl))
	comp, err := reg.Instantiate("test", ctrl, screen)
	if err != nil {
		t.Fatal(err)
	}
	screen.Add(comp)

	hr, err := ui.NewHotReloader(reg, screen, ctrl, "test", xmlPath)
	if err != nil {
		t.Fatal(err)
	}
	defer hr.Stop()

	if err := hr.WatchTheme(jsonPath); err != nil {
		t.Fatal(err)
	}

	initialTheme := screen.Children()[0].EffectiveTheme()

	// Write modified theme — watcher should pick it up.
	themeV2 := []byte(`{"components":{}}`)
	if err := os.WriteFile(jsonPath, themeV2, 0644); err != nil {
		t.Fatal(err)
	}

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if screen.Children()[0].EffectiveTheme() != initialTheme {
			return // success
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatal("watcher did not reload theme after write within timeout")
}

