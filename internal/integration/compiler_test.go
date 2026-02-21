package integration

import (
	"testing"

	ui "github.com/devthicket/willowui"
	"github.com/devthicket/willowui/internal/template"
)

// testDataProvider implements DataProvider for testing.
type testDataProvider struct {
	refs    map[string]any
	methods map[string]bool
}

func (p *testDataProvider) LookupRef(path string) any {
	return p.refs[path]
}

func (p *testDataProvider) CallMethod(name string) bool {
	v, ok := p.methods[name]
	return ok && v
}

// xmlTestController implements Controller + DataProvider for XML tests.
type xmlTestController struct {
	provider *testDataProvider
	created  bool
}

func (c *xmlTestController) OnCreate(s *ui.Screen)  { c.created = true }
func (c *xmlTestController) OnUpdate(dt float64)    {}
func (c *xmlTestController) OnDestroy()             {}
func (c *xmlTestController) LookupRef(path string) any {
	return c.provider.LookupRef(path)
}
func (c *xmlTestController) CallMethod(name string) bool {
	return c.provider.CallMethod(name)
}

func TestTemplateRegistry_RegisterAndGet(t *testing.T) {
	reg := ui.NewTemplateRegistry()
	err := reg.RegisterXML("test", []byte(`<Panel><Label text="hi" /></Panel>`))
	if err != nil {
		t.Fatal(err)
	}
	ir := reg.Get("test")
	if ir == nil {
		t.Fatal("expected non-nil IR")
	}
	if ir.ComponentType != "Panel" {
		t.Errorf("root type = %q, want Panel", ir.ComponentType)
	}
}

func TestTemplateRegistry_RegisterError(t *testing.T) {
	reg := ui.NewTemplateRegistry()
	err := reg.RegisterXML("bad", []byte(`<Unknown />`))
	if err == nil {
		t.Error("expected error for unknown component")
	}
}

func TestTemplateRegistry_GetMissing(t *testing.T) {
	reg := ui.NewTemplateRegistry()
	if reg.Get("nope") != nil {
		t.Error("expected nil for missing template")
	}
}

func TestInstantiate_SimpleLabel(t *testing.T) {
	reg := ui.NewTemplateRegistry()
	reg.SetFonts(nil, nil)

	err := reg.RegisterXML("test", []byte(`<Label text="Hello" />`))
	if err != nil {
		t.Fatal(err)
	}

	ctrl := &xmlTestController{provider: &testDataProvider{refs: map[string]any{}}}
	comp, err := reg.Instantiate("test", ctrl, nil)
	if err != nil {
		t.Fatal(err)
	}
	if comp == nil {
		t.Fatal("expected non-nil component")
	}
	label := template.CompAsLabel(comp)
	if label == nil {
		t.Fatal("expected Label component")
	}
	if label.Text() != "Hello" {
		t.Errorf("text = %q, want 'Hello'", label.Text())
	}
}

func TestInstantiate_BindText(t *testing.T) {
	reg := ui.NewTemplateRegistry()
	reg.SetFonts(nil, nil)

	titleRef := ui.NewRef("initial")
	ctrl := &xmlTestController{
		provider: &testDataProvider{
			refs: map[string]any{"title": titleRef},
		},
	}

	err := reg.RegisterXML("test", []byte(`<Label bind:text="title" />`))
	if err != nil {
		t.Fatal(err)
	}

	comp, err := reg.Instantiate("test", ctrl, nil)
	if err != nil {
		t.Fatal(err)
	}
	label := template.CompAsLabel(comp)
	if label == nil {
		t.Fatal("expected Label component")
	}
	if label.Text() != "initial" {
		t.Errorf("text = %q, want 'initial'", label.Text())
	}
}

func TestInstantiate_PanelWithChildren(t *testing.T) {
	reg := ui.NewTemplateRegistry()
	reg.SetFonts(nil, nil)

	err := reg.RegisterXML("test", []byte(`
		<Panel layout="vbox">
			<Label text="one" />
			<Label text="two" />
		</Panel>
	`))
	if err != nil {
		t.Fatal(err)
	}

	ctrl := &xmlTestController{provider: &testDataProvider{refs: map[string]any{}}}
	comp, err := reg.Instantiate("test", ctrl, nil)
	if err != nil {
		t.Fatal(err)
	}
	if comp.NumChildren() != 2 {
		t.Errorf("children = %d, want 2", comp.NumChildren())
	}
}

func TestInstantiate_EventBinding(t *testing.T) {
	reg := ui.NewTemplateRegistry()
	reg.SetFonts(nil, nil)

	called := false
	ctrl := &xmlTestController{
		provider: &testDataProvider{
			refs:    map[string]any{},
			methods: map[string]bool{"handleClick": true},
		},
	}

	err := reg.RegisterXML("test", []byte(`<Button text="Go" on:click="handleClick" />`))
	if err != nil {
		t.Fatal(err)
	}

	comp, err := reg.Instantiate("test", ctrl, nil)
	if err != nil {
		t.Fatal(err)
	}

	// Trigger the button's click handler
	btn := template.CompAsButton(comp)
	if btn == nil {
		t.Fatal("expected Button component")
	}
	if btn.HasOnClickCallback() {
		btn.SimulateOnClick()
		called = ctrl.provider.CallMethod("handleClick")
	}
	if !called {
		t.Error("expected handleClick to be callable")
	}
}

func TestInstantiate_ShowDirective(t *testing.T) {
	reg := ui.NewTemplateRegistry()
	reg.SetFonts(nil, nil)

	visRef := ui.NewRef(true)
	ctrl := &xmlTestController{
		provider: &testDataProvider{
			refs: map[string]any{"visible": visRef},
		},
	}

	err := reg.RegisterXML("test", []byte(`<Panel ui:show="visible" />`))
	if err != nil {
		t.Fatal(err)
	}

	comp, err := reg.Instantiate("test", ctrl, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !comp.IsVisible() {
		t.Error("expected visible when ref is true")
	}
}

func TestInstantiate_MissingTemplate(t *testing.T) {
	reg := ui.NewTemplateRegistry()
	_, err := reg.Instantiate("nonexistent", nil, nil)
	if err == nil {
		t.Error("expected error for missing template")
	}
}

func TestInstantiate_NestedPanel(t *testing.T) {
	reg := ui.NewTemplateRegistry()
	reg.SetFonts(nil, nil)

	err := reg.RegisterXML("test", []byte(`
		<Panel>
			<Panel>
				<Label text="deep" />
			</Panel>
		</Panel>
	`))
	if err != nil {
		t.Fatal(err)
	}

	ctrl := &xmlTestController{provider: &testDataProvider{refs: map[string]any{}}}
	comp, err := reg.Instantiate("test", ctrl, nil)
	if err != nil {
		t.Fatal(err)
	}
	if comp.NumChildren() != 1 {
		t.Fatalf("outer children = %d, want 1", comp.NumChildren())
	}
	inner := comp.Children()[0]
	if inner.NumChildren() != 1 {
		t.Fatalf("inner children = %d, want 1", inner.NumChildren())
	}
}

func TestInstantiate_StaticEnabled(t *testing.T) {
	reg := ui.NewTemplateRegistry()
	reg.SetFonts(nil, nil)

	err := reg.RegisterXML("test", []byte(`<Button text="No" enabled="false" />`))
	if err != nil {
		t.Fatal(err)
	}

	ctrl := &xmlTestController{provider: &testDataProvider{refs: map[string]any{}}}
	comp, err := reg.Instantiate("test", ctrl, nil)
	if err != nil {
		t.Fatal(err)
	}
	if comp.IsEnabled() {
		t.Error("expected disabled button")
	}
}

func TestParseColor(t *testing.T) {
	tests := []struct {
		input      string
		r, g, b, a float64
	}{
		{"#ff0000", 1, 0, 0, 1},
		{"#00ff00ff", 0, 1, 0, 1},
		{"#000000", 0, 0, 0, 1},
		{"#ffffff", 1, 1, 1, 1},
	}
	for _, tt := range tests {
		c := template.ParseColor(tt.input)
		if c.R() != tt.r || c.G() != tt.g || c.B() != tt.b || c.A() != tt.a {
			t.Errorf("ParseColor(%q) = {%.0f,%.0f,%.0f,%.0f}, want {%.0f,%.0f,%.0f,%.0f}",
				tt.input, c.R(), c.G(), c.B(), c.A(), tt.r, tt.g, tt.b, tt.a)
		}
	}
}
