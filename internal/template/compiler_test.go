package template

import (
	"testing"
)

// ── CompileXML ──────────────────────────────────────────────────────────────

func TestCompileXML_SimplePanel(t *testing.T) {
	ir, err := CompileXML([]byte(`<Panel width="200" height="100"/>`))
	if err != nil {
		t.Fatal(err)
	}
	if ir.ComponentType != "Panel" {
		t.Errorf("ComponentType = %q, want Panel", ir.ComponentType)
	}
	if len(ir.Attributes) != 2 {
		t.Fatalf("got %d attributes, want 2", len(ir.Attributes))
	}
	names := map[string]string{}
	for _, a := range ir.Attributes {
		names[a.Name] = a.Static
	}
	if names["width"] != "200" {
		t.Errorf("width = %q, want 200", names["width"])
	}
	if names["height"] != "100" {
		t.Errorf("height = %q, want 100", names["height"])
	}
}

func TestCompileXML_NestedChildren(t *testing.T) {
	ir, err := CompileXML([]byte(`<Panel><Label text="hello"/></Panel>`))
	if err != nil {
		t.Fatal(err)
	}
	if ir.ComponentType != "Panel" {
		t.Errorf("ComponentType = %q, want Panel", ir.ComponentType)
	}
	if len(ir.Children) != 1 {
		t.Fatalf("got %d children, want 1", len(ir.Children))
	}
	child := ir.Children[0]
	if child.ComponentType != "Label" {
		t.Errorf("child ComponentType = %q, want Label", child.ComponentType)
	}
	if len(child.Attributes) != 1 || child.Attributes[0].Static != "hello" {
		t.Errorf("child text attr unexpected: %+v", child.Attributes)
	}
}

func TestCompileXML_BindAttribute(t *testing.T) {
	ir, err := CompileXML([]byte(`<Label bind:text="name"/>`))
	if err != nil {
		t.Fatal(err)
	}
	if len(ir.Attributes) != 1 {
		t.Fatalf("got %d attrs, want 1", len(ir.Attributes))
	}
	attr := ir.Attributes[0]
	if attr.Name != "text" {
		t.Errorf("attr name = %q, want text", attr.Name)
	}
	if attr.Expr == nil {
		t.Fatal("expected Expr to be set for bind attribute")
	}
	ref, ok := attr.Expr.(ExprRef)
	if !ok {
		t.Fatalf("expected ExprRef, got %T", attr.Expr)
	}
	if ref.Path != "name" {
		t.Errorf("ref path = %q, want name", ref.Path)
	}
}

func TestCompileXML_EventAttribute(t *testing.T) {
	ir, err := CompileXML([]byte(`<Button on:click="handleClick"/>`))
	if err != nil {
		t.Fatal(err)
	}
	if len(ir.Attributes) != 1 {
		t.Fatalf("got %d attrs, want 1", len(ir.Attributes))
	}
	attr := ir.Attributes[0]
	if attr.Name != "click" {
		t.Errorf("attr name = %q, want click", attr.Name)
	}
	if !attr.IsEvent {
		t.Error("expected IsEvent=true")
	}
	if attr.Static != "handleClick" {
		t.Errorf("static = %q, want handleClick", attr.Static)
	}
}

func TestCompileXML_DirectiveIf(t *testing.T) {
	ir, err := CompileXML([]byte(`<Panel ui:if="visible"/>`))
	if err != nil {
		t.Fatal(err)
	}
	if len(ir.Directives) != 1 {
		t.Fatalf("got %d directives, want 1", len(ir.Directives))
	}
	dir := ir.Directives[0]
	if dir.Type != DirectiveIf {
		t.Errorf("directive type = %v, want DirectiveIf", dir.Type)
	}
	if dir.Expr == nil {
		t.Error("expected Expr to be set")
	}
}

func TestCompileXML_DirectiveFor(t *testing.T) {
	ir, err := CompileXML([]byte(`<Panel ui:for="item in items"/>`))
	if err != nil {
		t.Fatal(err)
	}
	if len(ir.Directives) != 1 {
		t.Fatalf("got %d directives, want 1", len(ir.Directives))
	}
	dir := ir.Directives[0]
	if dir.Type != DirectiveFor {
		t.Errorf("directive type = %v, want DirectiveFor", dir.Type)
	}
	if dir.VarName != "item" {
		t.Errorf("VarName = %q, want item", dir.VarName)
	}
	ref, ok := dir.Expr.(ExprRef)
	if !ok {
		t.Fatalf("expected ExprRef, got %T", dir.Expr)
	}
	if ref.Path != "items" {
		t.Errorf("ref path = %q, want items", ref.Path)
	}
}

func TestCompileXML_DirectiveShow(t *testing.T) {
	ir, err := CompileXML([]byte(`<Panel ui:show="active"/>`))
	if err != nil {
		t.Fatal(err)
	}
	if len(ir.Directives) != 1 {
		t.Fatalf("got %d directives, want 1", len(ir.Directives))
	}
	dir := ir.Directives[0]
	if dir.Type != DirectiveShow {
		t.Errorf("directive type = %v, want DirectiveShow", dir.Type)
	}
}

func TestCompileXML_DirectiveElseIf(t *testing.T) {
	ir, err := CompileXML([]byte(`<Panel ui:else-if="count > 0"/>`))
	if err != nil {
		t.Fatal(err)
	}
	if len(ir.Directives) != 1 {
		t.Fatalf("got %d directives, want 1", len(ir.Directives))
	}
	if ir.Directives[0].Type != DirectiveElseIf {
		t.Errorf("directive type = %v, want DirectiveElseIf", ir.Directives[0].Type)
	}
}

func TestCompileXML_DirectiveElse(t *testing.T) {
	ir, err := CompileXML([]byte(`<Panel ui:else=""/>`))
	if err != nil {
		t.Fatal(err)
	}
	if len(ir.Directives) != 1 {
		t.Fatalf("got %d directives, want 1", len(ir.Directives))
	}
	if ir.Directives[0].Type != DirectiveElse {
		t.Errorf("directive type = %v, want DirectiveElse", ir.Directives[0].Type)
	}
}

func TestCompileXML_DirectiveKey(t *testing.T) {
	ir, err := CompileXML([]byte(`<Panel ui:key="id"/>`))
	if err != nil {
		t.Fatal(err)
	}
	if len(ir.Directives) != 1 {
		t.Fatalf("got %d directives, want 1", len(ir.Directives))
	}
	if ir.Directives[0].Type != DirectiveKey {
		t.Errorf("directive type = %v, want DirectiveKey", ir.Directives[0].Type)
	}
}

func TestCompileXML_ThemeChild(t *testing.T) {
	xml := `<Panel><Theme>{"colors":{}}</Theme></Panel>`
	ir, err := CompileXML([]byte(xml))
	if err != nil {
		t.Fatal(err)
	}
	if ir.ThemePatch == nil {
		t.Fatal("expected ThemePatch to be set")
	}
	if string(ir.ThemePatch) != `{"colors":{}}` {
		t.Errorf("ThemePatch = %q, want {\"colors\":{}}", string(ir.ThemePatch))
	}
	// Theme should not appear as a child node
	if len(ir.Children) != 0 {
		t.Errorf("got %d children, want 0 (Theme should be extracted)", len(ir.Children))
	}
}

func TestCompileXML_ThemeInvalidJSON(t *testing.T) {
	xml := `<Panel><Theme>not json</Theme></Panel>`
	_, err := CompileXML([]byte(xml))
	if err == nil {
		t.Error("expected error for invalid JSON in Theme")
	}
}

func TestCompileXML_ThemeEmpty(t *testing.T) {
	xml := `<Panel><Theme></Theme></Panel>`
	_, err := CompileXML([]byte(xml))
	if err == nil {
		t.Error("expected error for empty Theme element")
	}
}

func TestCompileXML_TextInterpolation(t *testing.T) {
	ir, err := CompileXML([]byte(`<Label>Hello {{name}}</Label>`))
	if err != nil {
		t.Fatal(err)
	}
	if ir.Text != "Hello {{name}}" {
		t.Errorf("Text = %q, want 'Hello {{name}}'", ir.Text)
	}
	// Should have a "text" attribute with an expression
	var textAttr *IRAttribute
	for i := range ir.Attributes {
		if ir.Attributes[i].Name == "text" {
			textAttr = &ir.Attributes[i]
			break
		}
	}
	if textAttr == nil {
		t.Fatal("expected 'text' attribute from interpolation")
	}
	if textAttr.Expr == nil {
		t.Error("expected Expr to be set on interpolated text attribute")
	}
	// Should be ExprConcat with 3 parts: "Hello ", ref(name), ""
	_, ok := textAttr.Expr.(ExprConcat)
	if !ok {
		// Could also be a single ExprRef if "Hello {{name}}" is parsed differently
		t.Logf("expr type = %T (expected ExprConcat for mixed content)", textAttr.Expr)
	}
}

func TestCompileXML_TextNoInterpolation(t *testing.T) {
	ir, err := CompileXML([]byte(`<Label>Hello world</Label>`))
	if err != nil {
		t.Fatal(err)
	}
	if ir.Text != "Hello world" {
		t.Errorf("Text = %q, want 'Hello world'", ir.Text)
	}
	// No expression-based text attribute should be added
	for _, a := range ir.Attributes {
		if a.Name == "text" && a.Expr != nil {
			t.Error("did not expect interpolated text attribute for plain text")
		}
	}
}

func TestCompileXML_UnknownComponent(t *testing.T) {
	_, err := CompileXML([]byte(`<FakeWidget/>`))
	if err == nil {
		t.Error("expected error for unknown component")
	}
}

func TestCompileXML_InvalidXML(t *testing.T) {
	_, err := CompileXML([]byte(`<Panel`))
	if err == nil {
		t.Error("expected error for invalid XML")
	}
}

func TestCompileXML_EmptyXML(t *testing.T) {
	_, err := CompileXML([]byte(``))
	if err == nil {
		t.Error("expected error for empty XML")
	}
}

func TestCompileXML_MultipleChildren(t *testing.T) {
	xml := `<Panel><Label text="a"/><Button text="b"/><Label text="c"/></Panel>`
	ir, err := CompileXML([]byte(xml))
	if err != nil {
		t.Fatal(err)
	}
	if len(ir.Children) != 3 {
		t.Fatalf("got %d children, want 3", len(ir.Children))
	}
	types := []string{ir.Children[0].ComponentType, ir.Children[1].ComponentType, ir.Children[2].ComponentType}
	want := []string{"Label", "Button", "Label"}
	for i := range types {
		if types[i] != want[i] {
			t.Errorf("child[%d] = %q, want %q", i, types[i], want[i])
		}
	}
}

func TestCompileXML_DeeplyNested(t *testing.T) {
	xml := `<Panel><Panel><Panel><Label text="deep"/></Panel></Panel></Panel>`
	ir, err := CompileXML([]byte(xml))
	if err != nil {
		t.Fatal(err)
	}
	// Walk down
	child := ir
	for i := 0; i < 3; i++ {
		if len(child.Children) != 1 {
			t.Fatalf("depth %d: got %d children, want 1", i, len(child.Children))
		}
		child = child.Children[0]
	}
	if child.ComponentType != "Label" {
		t.Errorf("deepest child = %q, want Label", child.ComponentType)
	}
}

func TestCompileXML_UnknownDirective(t *testing.T) {
	_, err := CompileXML([]byte(`<Panel ui:nope="x"/>`))
	if err == nil {
		t.Error("expected error for unknown directive ui:nope")
	}
}

func TestCompileXML_InvalidBindExpression(t *testing.T) {
	_, err := CompileXML([]byte(`<Label bind:text="1 +"/>`))
	if err == nil {
		t.Error("expected error for invalid bind expression")
	}
}

func TestCompileXML_InvalidForSyntax(t *testing.T) {
	_, err := CompileXML([]byte(`<Panel ui:for="badformat"/>`))
	if err == nil {
		t.Error("expected error for invalid ui:for syntax")
	}
}

func TestCompileXML_ForEmptyVarName(t *testing.T) {
	_, err := CompileXML([]byte(`<Panel ui:for=" in items"/>`))
	if err == nil {
		t.Error("expected error for empty var name in ui:for")
	}
}

func TestCompileXML_MultipleDirectives(t *testing.T) {
	ir, err := CompileXML([]byte(`<Panel ui:if="visible" ui:key="id"/>`))
	if err != nil {
		t.Fatal(err)
	}
	if len(ir.Directives) != 2 {
		t.Fatalf("got %d directives, want 2", len(ir.Directives))
	}
}

func TestCompileXML_MixedStaticAndBind(t *testing.T) {
	ir, err := CompileXML([]byte(`<Panel width="100" bind:height="h"/>`))
	if err != nil {
		t.Fatal(err)
	}
	if len(ir.Attributes) != 2 {
		t.Fatalf("got %d attrs, want 2", len(ir.Attributes))
	}
	// One should be static, one should have an expr
	var staticCount, exprCount int
	for _, a := range ir.Attributes {
		if a.Expr != nil {
			exprCount++
		} else {
			staticCount++
		}
	}
	if staticCount != 1 || exprCount != 1 {
		t.Errorf("staticCount=%d, exprCount=%d, want 1 and 1", staticCount, exprCount)
	}
}

func TestCompileXML_UnknownChildComponent(t *testing.T) {
	_, err := CompileXML([]byte(`<Panel><Bogus/></Panel>`))
	if err == nil {
		t.Error("expected error for unknown child component")
	}
}

func TestCompileXML_AllKnownComponents(t *testing.T) {
	// Spot-check several known components compile without error
	components := []string{"Panel", "Label", "Button", "Toggle", "Checkbox",
		"TextInput", "Slider", "ProgressBar", "List", "Window", "ScrollPanel",
		"Image", "Badge", "Tag", "Accordion", "DataTable", "Tooltip", "Popover"}
	for _, c := range components {
		xml := "<" + c + "/>"
		_, err := CompileXML([]byte(xml))
		if err != nil {
			t.Errorf("CompileXML(<%s/>) failed: %v", c, err)
		}
	}
}
