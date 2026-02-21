package integration

import (
	"testing"

	ui "github.com/devthicket/willowui"
	"github.com/devthicket/willowui/internal/template"
)

func TestCompileXML_SimpleLabel(t *testing.T) {
	xml := `<Label text="Hello" />`
	node, err := ui.CompileXML([]byte(xml))
	if err != nil {
		t.Fatal(err)
	}
	if node.ComponentType != "Label" {
		t.Errorf("type = %q, want Label", node.ComponentType)
	}
	if len(node.Attributes) != 1 {
		t.Fatalf("attrs = %d, want 1", len(node.Attributes))
	}
	if node.Attributes[0].Name != "text" || node.Attributes[0].Static != "Hello" {
		t.Errorf("attr = %+v, want text=Hello", node.Attributes[0])
	}
}

func TestCompileXML_BindAttribute(t *testing.T) {
	xml := `<Label bind:text="title" />`
	node, err := ui.CompileXML([]byte(xml))
	if err != nil {
		t.Fatal(err)
	}
	if len(node.Attributes) != 1 {
		t.Fatalf("attrs = %d, want 1", len(node.Attributes))
	}
	attr := node.Attributes[0]
	if attr.Name != "text" {
		t.Errorf("name = %q, want 'text'", attr.Name)
	}
	if attr.Expr == nil {
		t.Error("expected bind expression, got nil")
	}
	ref, ok := attr.Expr.(ui.ExprRef)
	if !ok {
		t.Fatalf("expr type = %T, want ExprRef", attr.Expr)
	}
	if ref.Path != "title" {
		t.Errorf("ref path = %q, want 'title'", ref.Path)
	}
}

func TestCompileXML_EventAttribute(t *testing.T) {
	xml := `<Button on:click="handleClick" />`
	node, err := ui.CompileXML([]byte(xml))
	if err != nil {
		t.Fatal(err)
	}
	if len(node.Attributes) != 1 {
		t.Fatalf("attrs = %d, want 1", len(node.Attributes))
	}
	attr := node.Attributes[0]
	if attr.Name != "click" || !attr.IsEvent {
		t.Errorf("attr = %+v, want click event", attr)
	}
	if attr.Static != "handleClick" {
		t.Errorf("value = %q, want 'handleClick'", attr.Static)
	}
}

func TestCompileXML_DirectiveIf(t *testing.T) {
	xml := `<Label ui:if="visible" text="hi" />`
	node, err := ui.CompileXML([]byte(xml))
	if err != nil {
		t.Fatal(err)
	}
	if len(node.Directives) != 1 {
		t.Fatalf("directives = %d, want 1", len(node.Directives))
	}
	if node.Directives[0].Type != ui.DirectiveIf {
		t.Errorf("directive type = %d, want DirectiveIf", node.Directives[0].Type)
	}
}

func TestCompileXML_DirectiveFor(t *testing.T) {
	xml := `<Label ui:for="item in items" bind:text="item.name" />`
	node, err := ui.CompileXML([]byte(xml))
	if err != nil {
		t.Fatal(err)
	}
	if len(node.Directives) != 1 {
		t.Fatalf("directives = %d, want 1", len(node.Directives))
	}
	dir := node.Directives[0]
	if dir.Type != ui.DirectiveFor {
		t.Errorf("type = %d, want DirectiveFor", dir.Type)
	}
	if dir.VarName != "item" {
		t.Errorf("varName = %q, want 'item'", dir.VarName)
	}
}

func TestCompileXML_DirectiveShow(t *testing.T) {
	xml := `<Panel ui:show="expanded" />`
	node, err := ui.CompileXML([]byte(xml))
	if err != nil {
		t.Fatal(err)
	}
	if len(node.Directives) != 1 {
		t.Fatalf("directives = %d, want 1", len(node.Directives))
	}
	if node.Directives[0].Type != ui.DirectiveShow {
		t.Errorf("type = %d, want DirectiveShow", node.Directives[0].Type)
	}
}

func TestCompileXML_NestedChildren(t *testing.T) {
	xml := `<Panel>
		<Label text="one" />
		<Button text="two" />
	</Panel>`
	node, err := ui.CompileXML([]byte(xml))
	if err != nil {
		t.Fatal(err)
	}
	if node.ComponentType != "Panel" {
		t.Errorf("type = %q, want Panel", node.ComponentType)
	}
	if len(node.Children) != 2 {
		t.Fatalf("children = %d, want 2", len(node.Children))
	}
	if node.Children[0].ComponentType != "Label" {
		t.Errorf("child[0] = %q, want Label", node.Children[0].ComponentType)
	}
	if node.Children[1].ComponentType != "Button" {
		t.Errorf("child[1] = %q, want Button", node.Children[1].ComponentType)
	}
}

func TestCompileXML_TextInterpolation(t *testing.T) {
	xml := `<Label>Hello {{name}}</Label>`
	node, err := ui.CompileXML([]byte(xml))
	if err != nil {
		t.Fatal(err)
	}
	if node.Text != "Hello {{name}}" {
		t.Errorf("text = %q", node.Text)
	}
	// Should have a text binding attribute from interpolation
	found := false
	for _, attr := range node.Attributes {
		if attr.Name == "text" && attr.Expr != nil {
			found = true
		}
	}
	if !found {
		t.Error("expected interpolated text binding attribute")
	}
}

func TestCompileXML_UnknownComponent(t *testing.T) {
	xml := `<Nonexistent />`
	_, err := ui.CompileXML([]byte(xml))
	if err == nil {
		t.Error("expected error for unknown component")
	}
}

func TestCompileXML_InvalidBindExpr(t *testing.T) {
	xml := `<Label bind:text="@invalid" />`
	_, err := ui.CompileXML([]byte(xml))
	if err == nil {
		t.Error("expected error for invalid expression")
	}
}

func TestCompileXML_InvalidForSyntax(t *testing.T) {
	xml := `<Label ui:for="broken" />`
	_, err := ui.CompileXML([]byte(xml))
	if err == nil {
		t.Error("expected error for invalid for syntax")
	}
}

func TestCompileXML_MultipleDirectives(t *testing.T) {
	xml := `<Label ui:if="visible" ui:show="expanded" text="hi" />`
	node, err := ui.CompileXML([]byte(xml))
	if err != nil {
		t.Fatal(err)
	}
	if len(node.Directives) != 2 {
		t.Fatalf("directives = %d, want 2", len(node.Directives))
	}
}

func TestCompileXML_MixedAttributes(t *testing.T) {
	xml := `<Button text="Click" bind:enabled="canClick" on:click="doClick" />`
	node, err := ui.CompileXML([]byte(xml))
	if err != nil {
		t.Fatal(err)
	}
	if len(node.Attributes) != 3 {
		t.Fatalf("attrs = %d, want 3", len(node.Attributes))
	}
	// Verify attribute classification
	var hasStatic, hasBind, hasEvent bool
	for _, a := range node.Attributes {
		if a.Name == "text" && a.Static == "Click" && a.Expr == nil && !a.IsEvent {
			hasStatic = true
		}
		if a.Name == "enabled" && a.Expr != nil && !a.IsEvent {
			hasBind = true
		}
		if a.Name == "click" && a.IsEvent {
			hasEvent = true
		}
	}
	if !hasStatic {
		t.Error("missing static text attribute")
	}
	if !hasBind {
		t.Error("missing bind:enabled attribute")
	}
	if !hasEvent {
		t.Error("missing on:click event attribute")
	}
}

func TestCompileXML_SortableTreeList(t *testing.T) {
	xml := `<SortableTreeList />`
	node, err := ui.CompileXML([]byte(xml))
	if err != nil {
		t.Fatal(err)
	}
	if node.ComponentType != "SortableTreeList" {
		t.Errorf("type = %q, want SortableTreeList", node.ComponentType)
	}
}

func TestCompileXML_EmptyDocument(t *testing.T) {
	_, err := ui.CompileXML([]byte(""))
	if err == nil {
		t.Error("expected error for empty document")
	}
}

// --- Missing attribute tests (spec 06) ---

func xmlInstantiate(t *testing.T, xmlData string) *ui.Component {
	t.Helper()
	reg := ui.NewTemplateRegistry()
	reg.SetFonts(nil, nil)
	if err := reg.RegisterXML("test", []byte(xmlData)); err != nil {
		t.Fatal(err)
	}
	ctrl := &xmlTestController{provider: &testDataProvider{refs: map[string]any{}}}
	comp, err := reg.Instantiate("test", ctrl, nil)
	if err != nil {
		t.Fatal(err)
	}
	return comp
}

func TestXMLAttr_ButtonAutoSize(t *testing.T) {
	comp := xmlInstantiate(t, `<Button text="OK" autoSize="true" />`)
	b := template.CompAsButton(comp)
	if b == nil {
		t.Fatal("expected Button")
	}
	if !b.AutoSize() {
		t.Error("autoSize should be true")
	}
}

func TestXMLAttr_LabelBoldItalicFontSize(t *testing.T) {
	comp := xmlInstantiate(t, `<Label text="styled" bold="true" italic="true" fontSize="24" />`)
	l := template.CompAsLabel(comp)
	if l == nil {
		t.Fatal("expected Label")
	}
	if l.Text() != "styled" {
		t.Errorf("text = %q, want styled", l.Text())
	}
}

func TestXMLAttr_WindowCloseable(t *testing.T) {
	xmlInstantiate(t, `<Window title="Win" closeable="false" minWidth="200" minHeight="150" escResult="cancel" enterResult="ok" />`)
}

func TestXMLAttr_InputFieldValidation(t *testing.T) {
	xmlInstantiate(t, `<InputField label="Email" maxLength="50" labelPosition="left" validationState="error" validationMessage="Invalid email" />`)
}

func TestXMLAttr_BadgeCountMaxCount(t *testing.T) {
	xmlInstantiate(t, `<Badge count="42" maxCount="99" />`)
}

func TestXMLAttr_RadioColumns(t *testing.T) {
	xmlInstantiate(t, `<Radio columns="2" verticalFirst="true">
		<RadioButton label="A" />
		<RadioButton label="B" />
	</Radio>`)
}

func TestXMLAttr_ProgressBarRangeFillColor(t *testing.T) {
	xmlInstantiate(t, `<ProgressBar value="0.5" range="0,100" fillColor="#ff0000" showLabel="true" />`)
}

func TestXMLAttr_AccordionExpanded(t *testing.T) {
	xmlInstantiate(t, `<Accordion exclusive="true" expanded="sec1">
		<Section id="sec1" label="Section 1">
			<Label text="Content 1" />
		</Section>
		<Section id="sec2" label="Section 2">
			<Label text="Content 2" />
		</Section>
	</Accordion>`)
}

func TestXMLAttr_TileListColumns(t *testing.T) {
	xmlInstantiate(t, `<TileList tileWidth="64" tileHeight="64" columns="3" size="300,300" />`)
}

func TestXMLAttr_TreeListSelectable(t *testing.T) {
	xmlInstantiate(t, `<TreeList itemHeight="30" selectable="true" leafOnlySelection="true" size="200,300" />`)
}

func TestXMLAttr_SortableTreeListSelected(t *testing.T) {
	xmlInstantiate(t, `<SortableTreeList selected="0" size="200,300" />`)
}

func TestXMLAttr_RichTextHeadingScale(t *testing.T) {
	xmlInstantiate(t, `<RichText headingScale="2.0,1.5,1.2" markup="hello" />`)
}

func TestXMLEvent_LinkClick(t *testing.T) {
	xml := `<RichText on:linkClick="handleLink" />`
	node, err := ui.CompileXML([]byte(xml))
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, attr := range node.Attributes {
		if attr.Name == "linkClick" && attr.IsEvent {
			found = true
		}
	}
	if !found {
		t.Error("expected linkClick event attribute")
	}
}

// --- Newly wired widgets (spec 07) ---

func TestXMLCompile_Tooltip(t *testing.T) {
	xml := `<Tooltip text="help" anchor="above" showDelay="15" />`
	node, err := ui.CompileXML([]byte(xml))
	if err != nil {
		t.Fatal(err)
	}
	if node.ComponentType != "Tooltip" {
		t.Errorf("type = %q, want Tooltip", node.ComponentType)
	}
}

func TestXMLInstantiate_Tooltip(t *testing.T) {
	comp := xmlInstantiate(t, `<Tooltip text="help" anchor="below" showDelay="10" hideDelay="5" fadeIn="0.2" fadeOut="0.1" clampToScreen="false" size="200,40" />`)
	tt, ok := comp.UserData().(*ui.Tooltip)
	if !ok {
		t.Fatal("expected *Tooltip in UserData")
	}
	if tt.ShowDelay != 10 {
		t.Errorf("ShowDelay = %d, want 10", tt.ShowDelay)
	}
	if tt.HideDelay != 5 {
		t.Errorf("HideDelay = %d, want 5", tt.HideDelay)
	}
	if tt.ClampToScreen {
		t.Error("ClampToScreen should be false")
	}
}

func TestXMLCompile_Popover(t *testing.T) {
	xml := `<Popover title="Details" preferredSide="above" showCloseButton="true" />`
	node, err := ui.CompileXML([]byte(xml))
	if err != nil {
		t.Fatal(err)
	}
	if node.ComponentType != "Popover" {
		t.Errorf("type = %q, want Popover", node.ComponentType)
	}
}

func TestXMLInstantiate_Popover(t *testing.T) {
	xmlInstantiate(t, `<Popover title="Info" preferredSide="right" contentSize="300,200" showCloseButton="true" />`)
}

func TestXMLInstantiate_ToolBar(t *testing.T) {
	comp := xmlInstantiate(t, `<ToolBar orientation="vertical" overflowMode="wrap" size="300,40" />`)
	if comp == nil {
		t.Fatal("expected non-nil component")
	}
}

func TestXMLInstantiate_ToolBarHorizontal(t *testing.T) {
	xmlInstantiate(t, `<ToolBar orientation="horizontal" wrap="true" size="400,40" />`)
}

func TestXMLInstantiate_StatWeb(t *testing.T) {
	comp := xmlInstantiate(t, `<StatWeb editable="true" fillEnabled="false" size="200,200" />`)
	sw, ok := comp.UserData().(*ui.StatWeb)
	if !ok {
		t.Fatal("expected *StatWeb in UserData")
	}
	if !sw.IsEditable() {
		t.Error("expected editable = true")
	}
}

func TestXMLInstantiate_GradientEditor(t *testing.T) {
	xmlInstantiate(t, `<GradientEditor showModeSelector="false" size="200,100" />`)
}

func TestXMLInstantiate_MenuBar(t *testing.T) {
	xmlInstantiate(t, `<MenuBar size="400,30" />`)
}

func TestXMLInstantiate_ColorPickerExtended(t *testing.T) {
	xmlInstantiate(t, `<ColorPicker showAlpha="false" defaultMode="hsv" size="40,40" />`)
}

func TestXMLEvent_PopoverClose(t *testing.T) {
	xml := `<Popover on:close="handleClose" />`
	node, err := ui.CompileXML([]byte(xml))
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, attr := range node.Attributes {
		if attr.Name == "close" && attr.IsEvent {
			found = true
		}
	}
	if !found {
		t.Error("expected close event attribute")
	}
}

func TestXMLEvent_PopoverOpen(t *testing.T) {
	xml := `<Popover on:open="handleOpen" />`
	node, err := ui.CompileXML([]byte(xml))
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, attr := range node.Attributes {
		if attr.Name == "open" && attr.IsEvent {
			found = true
		}
	}
	if !found {
		t.Error("expected open event attribute")
	}
}

func TestXMLEvent_ColorPickerCommit(t *testing.T) {
	xml := `<ColorPicker on:commit="handleCommit" />`
	node, err := ui.CompileXML([]byte(xml))
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, attr := range node.Attributes {
		if attr.Name == "commit" && attr.IsEvent {
			found = true
		}
	}
	if !found {
		t.Error("expected commit event attribute")
	}
}

// TestXMLAttr_FontSizePerWidget verifies that the fontSize attribute is
// accepted on all widgets listed in spec 108 and does not break construction.
func TestXMLAttr_FontSizePerWidget(t *testing.T) {
	cases := []struct {
		name string
		xml  string
	}{
		{"Button", `<Button text="Go" fontSize="20" />`},
		{"Checkbox", `<Checkbox text="On" fontSize="18" />`},
		{"NumberStepper", `<NumberStepper fontSize="16" />`},
		{"Select", `<Select options="A,B" fontSize="14" />`},
		{"Badge", `<Badge fontSize="12" />`},
		{"Tag", `<Tag fontSize="11" />`},
		{"TagBar", `<TagBar fontSize="13" />`},
		{"MaskedInput", `<MaskedInput fontSize="15" />`},
		{"KeybindInput", `<KeybindInput fontSize="14" />`},
		{"SearchBox", `<SearchBox fontSize="16" />`},
		{"CalendarSelector", `<CalendarSelector fontSize="14" />`},
		{"TimePicker", `<TimePicker fontSize="13" />`},
		{"OptionRotator", `<OptionRotator options="X,Y" fontSize="18" />`},
		{"ToggleButtonBar", `<ToggleButtonBar buttons="A,B" fontSize="16" />`},
		{"SortableTreeList", `<SortableTreeList fontSize="14" />`},
		{"Accordion", `<Accordion fontSize="15"><Section title="S1" /></Accordion>`},
		{"Label", `<Label text="hi" fontSize="24" />`},
		{"TextInput", `<TextInput fontSize="16" />`},
		{"TextArea", `<TextArea fontSize="14" />`},
		{"TabBar", `<TabBar fontSize="13" />`},
		{"Window", `<Window title="W" fontSize="18" />`},
		{"RichText", `<RichText fontSize="16" />`},
		{"InputField", `<InputField fontSize="14" />`},
		{"ColorPicker", `<ColorPicker fontSize="12" />`},
		{"StatWeb", `<StatWeb fontSize="14" />`},
		{"GradientEditor", `<GradientEditor fontSize="13" />`},
		{"MenuBar", `<MenuBar fontSize="14" />`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			xmlInstantiate(t, tc.xml)
		})
	}
}

// TestXMLAttr_FontSizeFallback verifies that omitting fontSize still works
// (falls back to registry default).
func TestXMLAttr_FontSizeFallback(t *testing.T) {
	cases := []struct {
		name string
		xml  string
	}{
		{"Button", `<Button text="Go" />`},
		{"Checkbox", `<Checkbox text="On" />`},
		{"Badge", `<Badge />`},
		{"Tag", `<Tag />`},
		{"Select", `<Select options="A,B" />`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			xmlInstantiate(t, tc.xml)
		})
	}
}
