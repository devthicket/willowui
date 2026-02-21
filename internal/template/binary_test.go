package template

import (
	"testing"
)

// ── EncodeIR / DecodeIR roundtrip ───────────────────────────────────────────

func TestBinary_SimpleNode(t *testing.T) {
	ir := &IRNode{ComponentType: "Panel"}
	encoded, err := EncodeIR(ir)
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := DecodeIR(encoded)
	if err != nil {
		t.Fatal(err)
	}
	if decoded.ComponentType != "Panel" {
		t.Errorf("ComponentType = %q, want Panel", decoded.ComponentType)
	}
}

func TestBinary_NodeWithStaticAttributes(t *testing.T) {
	ir := &IRNode{
		ComponentType: "Panel",
		Attributes: []IRAttribute{
			{Name: "width", Static: "200"},
			{Name: "height", Static: "100"},
		},
	}
	decoded := roundtrip(t, ir)
	if len(decoded.Attributes) != 2 {
		t.Fatalf("got %d attrs, want 2", len(decoded.Attributes))
	}
	if decoded.Attributes[0].Name != "width" || decoded.Attributes[0].Static != "200" {
		t.Errorf("attr[0] = %+v", decoded.Attributes[0])
	}
	if decoded.Attributes[1].Name != "height" || decoded.Attributes[1].Static != "100" {
		t.Errorf("attr[1] = %+v", decoded.Attributes[1])
	}
}

func TestBinary_NodeWithBindExpr(t *testing.T) {
	ir := &IRNode{
		ComponentType: "Label",
		Attributes: []IRAttribute{
			{Name: "text", Expr: ExprRef{Path: "user.name"}},
		},
	}
	decoded := roundtrip(t, ir)
	if len(decoded.Attributes) != 1 {
		t.Fatalf("got %d attrs, want 1", len(decoded.Attributes))
	}
	ref, ok := decoded.Attributes[0].Expr.(ExprRef)
	if !ok {
		t.Fatalf("expected ExprRef, got %T", decoded.Attributes[0].Expr)
	}
	if ref.Path != "user.name" {
		t.Errorf("path = %q, want user.name", ref.Path)
	}
}

func TestBinary_NodeWithEvent(t *testing.T) {
	ir := &IRNode{
		ComponentType: "Button",
		Attributes: []IRAttribute{
			{Name: "click", Static: "handleClick", IsEvent: true},
		},
	}
	decoded := roundtrip(t, ir)
	attr := decoded.Attributes[0]
	if !attr.IsEvent {
		t.Error("expected IsEvent=true")
	}
	if attr.Static != "handleClick" {
		t.Errorf("static = %q, want handleClick", attr.Static)
	}
}

func TestBinary_NodeWithDirectives(t *testing.T) {
	ir := &IRNode{
		ComponentType: "Panel",
		Directives: []IRDirective{
			{Type: DirectiveIf, Expr: ExprRef{Path: "visible"}},
			{Type: DirectiveFor, Expr: ExprRef{Path: "items"}, VarName: "item"},
			{Type: DirectiveShow, Expr: ExprLiteral{Value: true}},
		},
	}
	decoded := roundtrip(t, ir)
	if len(decoded.Directives) != 3 {
		t.Fatalf("got %d directives, want 3", len(decoded.Directives))
	}
	if decoded.Directives[0].Type != DirectiveIf {
		t.Errorf("dir[0] type = %v, want DirectiveIf", decoded.Directives[0].Type)
	}
	if decoded.Directives[1].Type != DirectiveFor {
		t.Errorf("dir[1] type = %v, want DirectiveFor", decoded.Directives[1].Type)
	}
	if decoded.Directives[1].VarName != "item" {
		t.Errorf("dir[1] VarName = %q, want item", decoded.Directives[1].VarName)
	}
	if decoded.Directives[2].Type != DirectiveShow {
		t.Errorf("dir[2] type = %v, want DirectiveShow", decoded.Directives[2].Type)
	}
}

func TestBinary_NodeWithChildren(t *testing.T) {
	ir := &IRNode{
		ComponentType: "Panel",
		Children: []*IRNode{
			{ComponentType: "Label"},
			{ComponentType: "Button"},
		},
	}
	decoded := roundtrip(t, ir)
	if len(decoded.Children) != 2 {
		t.Fatalf("got %d children, want 2", len(decoded.Children))
	}
	if decoded.Children[0].ComponentType != "Label" {
		t.Errorf("child[0] = %q, want Label", decoded.Children[0].ComponentType)
	}
	if decoded.Children[1].ComponentType != "Button" {
		t.Errorf("child[1] = %q, want Button", decoded.Children[1].ComponentType)
	}
}

func TestBinary_NodeWithText(t *testing.T) {
	ir := &IRNode{
		ComponentType: "Label",
		Text:          "Hello {{name}}",
	}
	decoded := roundtrip(t, ir)
	if decoded.Text != "Hello {{name}}" {
		t.Errorf("Text = %q, want 'Hello {{name}}'", decoded.Text)
	}
}

func TestBinary_NodeWithThemePatch(t *testing.T) {
	ir := &IRNode{
		ComponentType: "Panel",
		ThemePatch:    []byte(`{"colors":{"primary":"#ff0000"}}`),
	}
	decoded := roundtrip(t, ir)
	if string(decoded.ThemePatch) != `{"colors":{"primary":"#ff0000"}}` {
		t.Errorf("ThemePatch = %q", string(decoded.ThemePatch))
	}
}

func TestBinary_ComplexNestedTree(t *testing.T) {
	ir := &IRNode{
		ComponentType: "Panel",
		Attributes: []IRAttribute{
			{Name: "width", Static: "400"},
		},
		Directives: []IRDirective{
			{Type: DirectiveShow, Expr: ExprRef{Path: "visible"}},
		},
		ThemePatch: []byte(`{"bg":"#000"}`),
		Children: []*IRNode{
			{
				ComponentType: "Label",
				Text:          "Count: {{count}}",
				Attributes: []IRAttribute{
					{Name: "text", Expr: ExprConcat{Parts: []ExprNode{
						ExprLiteral{Value: "Count: "},
						ExprRef{Path: "count"},
					}}},
				},
			},
			{
				ComponentType: "Button",
				Attributes: []IRAttribute{
					{Name: "click", Static: "onAdd", IsEvent: true},
				},
				Children: []*IRNode{
					{ComponentType: "Label", Text: "Add"},
				},
			},
		},
	}
	decoded := roundtrip(t, ir)
	if decoded.ComponentType != "Panel" {
		t.Errorf("root type = %q", decoded.ComponentType)
	}
	if len(decoded.Children) != 2 {
		t.Fatalf("got %d children, want 2", len(decoded.Children))
	}
	if decoded.Children[0].Text != "Count: {{count}}" {
		t.Errorf("child[0] text = %q", decoded.Children[0].Text)
	}
	if len(decoded.Children[1].Children) != 1 {
		t.Fatalf("child[1] has %d children, want 1", len(decoded.Children[1].Children))
	}
	if string(decoded.ThemePatch) != `{"bg":"#000"}` {
		t.Errorf("ThemePatch = %q", string(decoded.ThemePatch))
	}
}

// ── Expression type roundtrips ──────────────────────────────────────────────

func TestBinary_ExprRef(t *testing.T) {
	ir := &IRNode{
		ComponentType: "Label",
		Attributes:    []IRAttribute{{Name: "x", Expr: ExprRef{Path: "a.b.c"}}},
	}
	decoded := roundtrip(t, ir)
	ref, ok := decoded.Attributes[0].Expr.(ExprRef)
	if !ok {
		t.Fatalf("expected ExprRef, got %T", decoded.Attributes[0].Expr)
	}
	if ref.Path != "a.b.c" {
		t.Errorf("path = %q, want a.b.c", ref.Path)
	}
}

func TestBinary_ExprLiteralNil(t *testing.T) {
	ir := &IRNode{
		ComponentType: "Label",
		Attributes:    []IRAttribute{{Name: "x", Expr: ExprLiteral{Value: nil}}},
	}
	decoded := roundtrip(t, ir)
	lit, ok := decoded.Attributes[0].Expr.(ExprLiteral)
	if !ok {
		t.Fatalf("expected ExprLiteral, got %T", decoded.Attributes[0].Expr)
	}
	if lit.Value != nil {
		t.Errorf("value = %v, want nil", lit.Value)
	}
}

func TestBinary_ExprLiteralFloat(t *testing.T) {
	ir := &IRNode{
		ComponentType: "Label",
		Attributes:    []IRAttribute{{Name: "x", Expr: ExprLiteral{Value: 3.14}}},
	}
	decoded := roundtrip(t, ir)
	lit := decoded.Attributes[0].Expr.(ExprLiteral)
	if lit.Value != 3.14 {
		t.Errorf("value = %v, want 3.14", lit.Value)
	}
}

func TestBinary_ExprLiteralString(t *testing.T) {
	ir := &IRNode{
		ComponentType: "Label",
		Attributes:    []IRAttribute{{Name: "x", Expr: ExprLiteral{Value: "hello"}}},
	}
	decoded := roundtrip(t, ir)
	lit := decoded.Attributes[0].Expr.(ExprLiteral)
	if lit.Value != "hello" {
		t.Errorf("value = %v, want hello", lit.Value)
	}
}

func TestBinary_ExprLiteralBool(t *testing.T) {
	for _, v := range []bool{true, false} {
		ir := &IRNode{
			ComponentType: "Label",
			Attributes:    []IRAttribute{{Name: "x", Expr: ExprLiteral{Value: v}}},
		}
		decoded := roundtrip(t, ir)
		lit := decoded.Attributes[0].Expr.(ExprLiteral)
		if lit.Value != v {
			t.Errorf("value = %v, want %v", lit.Value, v)
		}
	}
}

func TestBinary_ExprBinary(t *testing.T) {
	ir := &IRNode{
		ComponentType: "Label",
		Attributes: []IRAttribute{{Name: "x", Expr: ExprBinary{
			Op:    BinAdd,
			Left:  ExprLiteral{Value: 1.0},
			Right: ExprLiteral{Value: 2.0},
		}}},
	}
	decoded := roundtrip(t, ir)
	bin, ok := decoded.Attributes[0].Expr.(ExprBinary)
	if !ok {
		t.Fatalf("expected ExprBinary, got %T", decoded.Attributes[0].Expr)
	}
	if bin.Op != BinAdd {
		t.Errorf("op = %v, want BinAdd", bin.Op)
	}
}

func TestBinary_ExprUnary(t *testing.T) {
	ir := &IRNode{
		ComponentType: "Label",
		Attributes: []IRAttribute{{Name: "x", Expr: ExprUnary{
			Op:      UnaryNot,
			Operand: ExprRef{Path: "flag"},
		}}},
	}
	decoded := roundtrip(t, ir)
	un, ok := decoded.Attributes[0].Expr.(ExprUnary)
	if !ok {
		t.Fatalf("expected ExprUnary, got %T", decoded.Attributes[0].Expr)
	}
	if un.Op != UnaryNot {
		t.Errorf("op = %v, want UnaryNot", un.Op)
	}
}

func TestBinary_ExprTernary(t *testing.T) {
	ir := &IRNode{
		ComponentType: "Label",
		Attributes: []IRAttribute{{Name: "x", Expr: ExprTernary{
			Cond: ExprRef{Path: "flag"},
			Then: ExprLiteral{Value: "yes"},
			Else: ExprLiteral{Value: "no"},
		}}},
	}
	decoded := roundtrip(t, ir)
	ter, ok := decoded.Attributes[0].Expr.(ExprTernary)
	if !ok {
		t.Fatalf("expected ExprTernary, got %T", decoded.Attributes[0].Expr)
	}
	then := ter.Then.(ExprLiteral)
	if then.Value != "yes" {
		t.Errorf("then = %v, want yes", then.Value)
	}
}

func TestBinary_ExprConcat(t *testing.T) {
	ir := &IRNode{
		ComponentType: "Label",
		Attributes: []IRAttribute{{Name: "x", Expr: ExprConcat{
			Parts: []ExprNode{
				ExprLiteral{Value: "Hello "},
				ExprRef{Path: "name"},
				ExprLiteral{Value: "!"},
			},
		}}},
	}
	decoded := roundtrip(t, ir)
	concat, ok := decoded.Attributes[0].Expr.(ExprConcat)
	if !ok {
		t.Fatalf("expected ExprConcat, got %T", decoded.Attributes[0].Expr)
	}
	if len(concat.Parts) != 3 {
		t.Fatalf("got %d parts, want 3", len(concat.Parts))
	}
}

func TestBinary_NilExpr(t *testing.T) {
	// An attribute with no Expr should roundtrip with nil Expr
	ir := &IRNode{
		ComponentType: "Label",
		Attributes:    []IRAttribute{{Name: "text", Static: "hello"}},
	}
	decoded := roundtrip(t, ir)
	if decoded.Attributes[0].Expr != nil {
		t.Errorf("expected nil Expr, got %T", decoded.Attributes[0].Expr)
	}
}

// ── Error cases ─────────────────────────────────────────────────────────────

func TestBinary_EncodeNil(t *testing.T) {
	_, err := EncodeIR(nil)
	if err == nil {
		t.Error("expected error for nil IRNode")
	}
}

func TestBinary_DecodeNil(t *testing.T) {
	_, err := DecodeIR(nil)
	if err == nil {
		t.Error("expected error for nil data")
	}
}

func TestBinary_DecodeShortData(t *testing.T) {
	_, err := DecodeIR([]byte("bad"))
	if err == nil {
		t.Error("expected error for short data")
	}
}

func TestBinary_DecodeWrongMagic(t *testing.T) {
	data := []byte{'B', 'A', 'D', 0x00, 0x01, 0x00}
	_, err := DecodeIR(data)
	if err == nil {
		t.Error("expected error for wrong magic bytes")
	}
}

func TestBinary_DecodeWrongVersion(t *testing.T) {
	// Correct magic, wrong version (99)
	data := []byte{'X', 'U', 'I', 0x00, 99, 0x00}
	_, err := DecodeIR(data)
	if err == nil {
		t.Error("expected error for wrong version")
	}
}

func TestBinary_DecodeTruncatedData(t *testing.T) {
	// Encode a valid node, then truncate the data
	ir := &IRNode{
		ComponentType: "Panel",
		Attributes:    []IRAttribute{{Name: "width", Static: "200"}},
	}
	encoded, err := EncodeIR(ir)
	if err != nil {
		t.Fatal(err)
	}
	// Truncate to half
	_, err = DecodeIR(encoded[:len(encoded)/2])
	if err == nil {
		t.Error("expected error for truncated data")
	}
}

func TestBinary_FullCompileRoundtrip(t *testing.T) {
	// Compile XML, encode to binary, decode, and verify
	xml := `<Panel width="200" bind:height="h" ui:if="visible"><Label text="hello"/></Panel>`
	ir, err := CompileXML([]byte(xml))
	if err != nil {
		t.Fatal(err)
	}
	decoded := roundtrip(t, ir)
	if decoded.ComponentType != "Panel" {
		t.Errorf("type = %q, want Panel", decoded.ComponentType)
	}
	if len(decoded.Children) != 1 {
		t.Fatalf("children = %d, want 1", len(decoded.Children))
	}
	if decoded.Children[0].ComponentType != "Label" {
		t.Errorf("child type = %q, want Label", decoded.Children[0].ComponentType)
	}
	if len(decoded.Directives) != 1 {
		t.Fatalf("directives = %d, want 1", len(decoded.Directives))
	}
}

// ── Helpers ─────────────────────────────────────────────────────────────────

func roundtrip(t *testing.T, ir *IRNode) *IRNode {
	t.Helper()
	encoded, err := EncodeIR(ir)
	if err != nil {
		t.Fatalf("EncodeIR: %v", err)
	}
	decoded, err := DecodeIR(encoded)
	if err != nil {
		t.Fatalf("DecodeIR: %v", err)
	}
	return decoded
}
