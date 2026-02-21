package integration

import (
	"testing"

	ui "github.com/devthicket/willowui"
	"golang.org/x/image/font/gofont/goregular"
)

func TestInstantiateStaticNilController(t *testing.T) {
	resetScheduler()

	reg := ui.NewTemplateRegistry()
	font := ui.MustLoadDefaultFont()
	reg.SetFonts(nil, font)
	reg.SetFontSize(14)

	xml := []byte(`<Panel layout="vbox" size="200,100">
		<Label text="Hello" />
	</Panel>`)

	if err := reg.RegisterXML("test", xml); err != nil {
		t.Fatalf("RegisterXML: %v", err)
	}

	comp, err := reg.InstantiateStatic("test", nil)
	if err != nil {
		t.Fatalf("InstantiateStatic: %v", err)
	}
	defer comp.Dispose()

	if comp.Width != 200 || comp.Height != 100 {
		t.Errorf("size = (%v, %v), want (200, 100)", comp.Width, comp.Height)
	}
}

func TestInstantiateNilControllerDirect(t *testing.T) {
	resetScheduler()

	reg := ui.NewTemplateRegistry()
	font := ui.MustLoadDefaultFont()
	reg.SetFonts(nil, font)
	reg.SetFontSize(14)

	xml := []byte(`<Button text="OK" />`)

	if err := reg.RegisterXML("btn", xml); err != nil {
		t.Fatalf("RegisterXML: %v", err)
	}

	comp, err := reg.Instantiate("btn", nil, nil)
	if err != nil {
		t.Fatalf("Instantiate with nil ctrl: %v", err)
	}
	defer comp.Dispose()

	if comp.Width == 0 || comp.Height == 0 {
		t.Error("button should have auto-sized to non-zero dimensions")
	}
}

func TestNewTemplateRegistryWithFont(t *testing.T) {
	resetScheduler()

	reg, err := ui.NewTemplateRegistryWithFont(goregular.TTF, 14)
	if err != nil {
		t.Fatalf("NewTemplateRegistryWithFont: %v", err)
	}

	xml := []byte(`<Label text="test" />`)
	if err := reg.RegisterXML("lbl", xml); err != nil {
		t.Fatalf("RegisterXML: %v", err)
	}

	comp, err := reg.InstantiateStatic("lbl", nil)
	if err != nil {
		t.Fatalf("InstantiateStatic: %v", err)
	}
	defer comp.Dispose()
}

func TestFillAttributeInTemplate(t *testing.T) {
	resetScheduler()

	reg := ui.NewTemplateRegistry()
	font := ui.MustLoadDefaultFont()
	reg.SetFonts(nil, font)
	reg.SetFontSize(14)

	// Panel has default 8px padding on each side, so content width = 400 - 16 = 384.
	xml := []byte(`<Panel layout="vbox" size="400,300">
		<Panel fill="width" height="50" />
	</Panel>`)

	if err := reg.RegisterXML("fill", xml); err != nil {
		t.Fatalf("RegisterXML: %v", err)
	}

	comp, err := reg.InstantiateStatic("fill", nil)
	if err != nil {
		t.Fatalf("Instantiate: %v", err)
	}
	defer comp.Dispose()

	comp.UpdateLayout()

	children := comp.Children()
	if len(children) == 0 {
		t.Fatal("expected at least 1 child")
	}
	child := children[0]
	// Panel default padding is 8 per side → content width = 400 - 16 = 384.
	if child.Width != 384 {
		t.Errorf("child.Width = %v, want 384 (fill=width, minus 8px padding each side)", child.Width)
	}
}

func TestGrowAttributeInTemplate(t *testing.T) {
	resetScheduler()

	reg := ui.NewTemplateRegistry()
	font := ui.MustLoadDefaultFont()
	reg.SetFonts(nil, font)
	reg.SetFontSize(14)

	// Panel default padding is 8 per side → content width = 300 - 16 = 284.
	// Fixed child = 100, so grow child gets 284 - 100 = 184.
	xml := []byte(`<Panel layout="hbox" size="300,50">
		<Panel width="100" height="30" />
		<Panel grow="1" height="30" />
	</Panel>`)

	if err := reg.RegisterXML("grow", xml); err != nil {
		t.Fatalf("RegisterXML: %v", err)
	}

	comp, err := reg.InstantiateStatic("grow", nil)
	if err != nil {
		t.Fatalf("Instantiate: %v", err)
	}
	defer comp.Dispose()

	comp.UpdateLayout()

	children := comp.Children()
	if len(children) < 2 {
		t.Fatal("expected at least 2 children")
	}
	// Content width = 300 - 8 - 8 = 284. Fixed = 100. Grow gets 184.
	if children[1].Width != 184 {
		t.Errorf("grow child width = %v, want 184 (300 - 16 padding - 100 fixed)", children[1].Width)
	}
}

func TestThemePatchCompile(t *testing.T) {
	resetScheduler()

	reg := ui.NewTemplateRegistry()
	font := ui.MustLoadDefaultFont()
	reg.SetFonts(nil, font)
	reg.SetFontSize(14)

	// Must set base theme JSON for inline patches to work.
	baseTheme := []byte(`{
		"button": {
			"primary": { "background": "#6366f1" },
			"neutral": { "background": "#374151" }
		}
	}`)
	if err := reg.SetThemeJSON(baseTheme); err != nil {
		t.Fatalf("SetThemeJSON: %v", err)
	}

	xml := []byte(`<Panel layout="vbox" size="200,100">
		<Theme>{"button": {"variants": {"submit": {"background": "#4caf50"}}}}</Theme>
		<Button text="OK" variant="submit" />
	</Panel>`)

	if err := reg.RegisterXML("themed", xml); err != nil {
		t.Fatalf("RegisterXML: %v", err)
	}

	comp, err := reg.InstantiateStatic("themed", nil)
	if err != nil {
		t.Fatalf("Instantiate: %v", err)
	}
	defer comp.Dispose()
}

func TestThemePatchRequiresBaseTheme(t *testing.T) {
	resetScheduler()

	reg := ui.NewTemplateRegistry()
	font := ui.MustLoadDefaultFont()
	reg.SetFonts(nil, font)
	reg.SetFontSize(14)
	// Deliberately not calling SetThemeJSON.

	xml := []byte(`<Panel>
		<Theme>{"button": {"variants": {"submit": {"background": "#4caf50"}}}}</Theme>
	</Panel>`)

	if err := reg.RegisterXML("notheme", xml); err != nil {
		t.Fatalf("RegisterXML: %v", err)
	}

	_, err := reg.InstantiateStatic("notheme", nil)
	if err == nil {
		t.Fatal("expected error when using <Theme> patch without SetThemeJSON")
	}
}

func TestThemePatchInvalidJSON(t *testing.T) {
	resetScheduler()

	reg := ui.NewTemplateRegistry()

	xml := []byte(`<Panel>
		<Theme>not valid json</Theme>
	</Panel>`)

	err := reg.RegisterXML("bad", xml)
	if err == nil {
		t.Fatal("expected error for invalid JSON in <Theme>")
	}
}
