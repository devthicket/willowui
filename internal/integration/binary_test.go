package integration

import (
	"testing"

	ui "github.com/devthicket/willowui"
)

func TestBinaryRoundTrip_SimpleLabel(t *testing.T) {
	ir := &ui.IRNode{
		ComponentType: "Label",
		Attributes: []ui.IRAttribute{
			{Name: "text", Static: "Hello"},
		},
	}
	data, err := ui.EncodeIR(ir)
	if err != nil {
		t.Fatal(err)
	}
	got, err := ui.DecodeIR(data)
	if err != nil {
		t.Fatal(err)
	}
	if got.ComponentType != "Label" {
		t.Errorf("ComponentType = %q, want Label", got.ComponentType)
	}
	if len(got.Attributes) != 1 {
		t.Fatalf("attrs = %d, want 1", len(got.Attributes))
	}
	if got.Attributes[0].Name != "text" || got.Attributes[0].Static != "Hello" {
		t.Errorf("attr = %+v, want text=Hello", got.Attributes[0])
	}
}

func TestBinaryRoundTrip_NestedPanel(t *testing.T) {
	ir := &ui.IRNode{
		ComponentType: "Panel",
		Attributes: []ui.IRAttribute{
			{Name: "layout", Static: "vbox"},
			{Name: "spacing", Static: "12"},
		},
		Children: []*ui.IRNode{
			{ComponentType: "Label", Attributes: []ui.IRAttribute{{Name: "text", Static: "A"}}},
			{ComponentType: "Button", Attributes: []ui.IRAttribute{{Name: "text", Static: "B"}}},
		},
	}
	data, err := ui.EncodeIR(ir)
	if err != nil {
		t.Fatal(err)
	}
	got, err := ui.DecodeIR(data)
	if err != nil {
		t.Fatal(err)
	}
	if got.ComponentType != "Panel" {
		t.Errorf("root type = %q", got.ComponentType)
	}
	if len(got.Children) != 2 {
		t.Fatalf("children = %d, want 2", len(got.Children))
	}
	if got.Children[0].ComponentType != "Label" {
		t.Errorf("child[0] = %q, want Label", got.Children[0].ComponentType)
	}
	if got.Children[1].ComponentType != "Button" {
		t.Errorf("child[1] = %q, want Button", got.Children[1].ComponentType)
	}
}

func TestBinaryRoundTrip_ExprRef(t *testing.T) {
	ir := &ui.IRNode{
		ComponentType: "Label",
		Attributes: []ui.IRAttribute{
			{Name: "text", Expr: ui.ExprRef{Path: "user.name"}},
		},
	}
	data, err := ui.EncodeIR(ir)
	if err != nil {
		t.Fatal(err)
	}
	got, err := ui.DecodeIR(data)
	if err != nil {
		t.Fatal(err)
	}
	ref, ok := got.Attributes[0].Expr.(ui.ExprRef)
	if !ok {
		t.Fatalf("expr type = %T, want ExprRef", got.Attributes[0].Expr)
	}
	if ref.Path != "user.name" {
		t.Errorf("ref path = %q, want user.name", ref.Path)
	}
}

func TestBinaryRoundTrip_ExprLiteralFloat(t *testing.T) {
	ir := &ui.IRNode{
		ComponentType: "Label",
		Attributes: []ui.IRAttribute{
			{Name: "val", Expr: ui.ExprLiteral{Value: 3.14}},
		},
	}
	data, err := ui.EncodeIR(ir)
	if err != nil {
		t.Fatal(err)
	}
	got, err := ui.DecodeIR(data)
	if err != nil {
		t.Fatal(err)
	}
	lit, ok := got.Attributes[0].Expr.(ui.ExprLiteral)
	if !ok {
		t.Fatalf("expr type = %T, want ExprLiteral", got.Attributes[0].Expr)
	}
	if lit.Value != 3.14 {
		t.Errorf("value = %v, want 3.14", lit.Value)
	}
}

func TestBinaryRoundTrip_ExprLiteralString(t *testing.T) {
	ir := &ui.IRNode{
		ComponentType: "Label",
		Attributes: []ui.IRAttribute{
			{Name: "val", Expr: ui.ExprLiteral{Value: "hello world"}},
		},
	}
	data, err := ui.EncodeIR(ir)
	if err != nil {
		t.Fatal(err)
	}
	got, err := ui.DecodeIR(data)
	if err != nil {
		t.Fatal(err)
	}
	lit := got.Attributes[0].Expr.(ui.ExprLiteral)
	if lit.Value != "hello world" {
		t.Errorf("value = %v, want hello world", lit.Value)
	}
}

func TestBinaryRoundTrip_ExprLiteralBool(t *testing.T) {
	ir := &ui.IRNode{
		ComponentType: "Label",
		Attributes: []ui.IRAttribute{
			{Name: "a", Expr: ui.ExprLiteral{Value: true}},
			{Name: "b", Expr: ui.ExprLiteral{Value: false}},
		},
	}
	data, err := ui.EncodeIR(ir)
	if err != nil {
		t.Fatal(err)
	}
	got, err := ui.DecodeIR(data)
	if err != nil {
		t.Fatal(err)
	}
	litA := got.Attributes[0].Expr.(ui.ExprLiteral)
	litB := got.Attributes[1].Expr.(ui.ExprLiteral)
	if litA.Value != true {
		t.Errorf("a = %v, want true", litA.Value)
	}
	if litB.Value != false {
		t.Errorf("b = %v, want false", litB.Value)
	}
}

func TestBinaryRoundTrip_ExprLiteralNil(t *testing.T) {
	ir := &ui.IRNode{
		ComponentType: "Label",
		Attributes: []ui.IRAttribute{
			{Name: "val", Expr: ui.ExprLiteral{Value: nil}},
		},
	}
	data, err := ui.EncodeIR(ir)
	if err != nil {
		t.Fatal(err)
	}
	got, err := ui.DecodeIR(data)
	if err != nil {
		t.Fatal(err)
	}
	lit := got.Attributes[0].Expr.(ui.ExprLiteral)
	if lit.Value != nil {
		t.Errorf("value = %v, want nil", lit.Value)
	}
}

func TestBinaryRoundTrip_ExprBinary(t *testing.T) {
	ir := &ui.IRNode{
		ComponentType: "Label",
		Attributes: []ui.IRAttribute{
			{Name: "val", Expr: ui.ExprBinary{
				Op:    ui.BinAdd,
				Left:  ui.ExprRef{Path: "a"},
				Right: ui.ExprLiteral{Value: 1.0},
			}},
		},
	}
	data, err := ui.EncodeIR(ir)
	if err != nil {
		t.Fatal(err)
	}
	got, err := ui.DecodeIR(data)
	if err != nil {
		t.Fatal(err)
	}
	bin, ok := got.Attributes[0].Expr.(ui.ExprBinary)
	if !ok {
		t.Fatalf("expr type = %T, want ExprBinary", got.Attributes[0].Expr)
	}
	if bin.Op != ui.BinAdd {
		t.Errorf("op = %d, want BinAdd", bin.Op)
	}
	if ref, ok := bin.Left.(ui.ExprRef); !ok || ref.Path != "a" {
		t.Errorf("left = %+v, want ExprRef{a}", bin.Left)
	}
	if lit, ok := bin.Right.(ui.ExprLiteral); !ok || lit.Value != 1.0 {
		t.Errorf("right = %+v, want ExprLiteral{1.0}", bin.Right)
	}
}

func TestBinaryRoundTrip_ExprUnary(t *testing.T) {
	ir := &ui.IRNode{
		ComponentType: "Label",
		Attributes: []ui.IRAttribute{
			{Name: "val", Expr: ui.ExprUnary{Op: ui.UnaryNot, Operand: ui.ExprRef{Path: "visible"}}},
		},
	}
	data, err := ui.EncodeIR(ir)
	if err != nil {
		t.Fatal(err)
	}
	got, err := ui.DecodeIR(data)
	if err != nil {
		t.Fatal(err)
	}
	un, ok := got.Attributes[0].Expr.(ui.ExprUnary)
	if !ok {
		t.Fatalf("expr type = %T, want ExprUnary", got.Attributes[0].Expr)
	}
	if un.Op != ui.UnaryNot {
		t.Errorf("op = %d, want UnaryNot", un.Op)
	}
}

func TestBinaryRoundTrip_ExprTernary(t *testing.T) {
	ir := &ui.IRNode{
		ComponentType: "Label",
		Attributes: []ui.IRAttribute{
			{Name: "val", Expr: ui.ExprTernary{
				Cond: ui.ExprRef{Path: "flag"},
				Then: ui.ExprLiteral{Value: "yes"},
				Else: ui.ExprLiteral{Value: "no"},
			}},
		},
	}
	data, err := ui.EncodeIR(ir)
	if err != nil {
		t.Fatal(err)
	}
	got, err := ui.DecodeIR(data)
	if err != nil {
		t.Fatal(err)
	}
	ter, ok := got.Attributes[0].Expr.(ui.ExprTernary)
	if !ok {
		t.Fatalf("expr type = %T, want ExprTernary", got.Attributes[0].Expr)
	}
	if _, ok := ter.Cond.(ui.ExprRef); !ok {
		t.Errorf("cond = %T, want ExprRef", ter.Cond)
	}
}

func TestBinaryRoundTrip_ExprConcat(t *testing.T) {
	ir := &ui.IRNode{
		ComponentType: "Label",
		Attributes: []ui.IRAttribute{
			{Name: "text", Expr: ui.ExprConcat{
				Parts: []ui.ExprNode{
					ui.ExprLiteral{Value: "Hello "},
					ui.ExprRef{Path: "name"},
					ui.ExprLiteral{Value: "!"},
				},
			}},
		},
	}
	data, err := ui.EncodeIR(ir)
	if err != nil {
		t.Fatal(err)
	}
	got, err := ui.DecodeIR(data)
	if err != nil {
		t.Fatal(err)
	}
	concat, ok := got.Attributes[0].Expr.(ui.ExprConcat)
	if !ok {
		t.Fatalf("expr type = %T, want ExprConcat", got.Attributes[0].Expr)
	}
	if len(concat.Parts) != 3 {
		t.Fatalf("parts = %d, want 3", len(concat.Parts))
	}
}

func TestBinaryRoundTrip_Directives(t *testing.T) {
	ir := &ui.IRNode{
		ComponentType: "Panel",
		Directives: []ui.IRDirective{
			{Type: ui.DirectiveIf, Expr: ui.ExprRef{Path: "visible"}},
			{Type: ui.DirectiveFor, Expr: ui.ExprRef{Path: "items"}, VarName: "item"},
			{Type: ui.DirectiveShow, Expr: ui.ExprLiteral{Value: true}},
		},
	}
	data, err := ui.EncodeIR(ir)
	if err != nil {
		t.Fatal(err)
	}
	got, err := ui.DecodeIR(data)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Directives) != 3 {
		t.Fatalf("directives = %d, want 3", len(got.Directives))
	}
	if got.Directives[0].Type != ui.DirectiveIf {
		t.Errorf("dir[0] type = %d, want DirectiveIf", got.Directives[0].Type)
	}
	if got.Directives[1].Type != ui.DirectiveFor {
		t.Errorf("dir[1] type = %d, want DirectiveFor", got.Directives[1].Type)
	}
	if got.Directives[1].VarName != "item" {
		t.Errorf("dir[1] varName = %q, want item", got.Directives[1].VarName)
	}
	if got.Directives[2].Type != ui.DirectiveShow {
		t.Errorf("dir[2] type = %d, want DirectiveShow", got.Directives[2].Type)
	}
}

func TestBinaryRoundTrip_EventAttribute(t *testing.T) {
	ir := &ui.IRNode{
		ComponentType: "Button",
		Attributes: []ui.IRAttribute{
			{Name: "click", Static: "handleClick", IsEvent: true},
		},
	}
	data, err := ui.EncodeIR(ir)
	if err != nil {
		t.Fatal(err)
	}
	got, err := ui.DecodeIR(data)
	if err != nil {
		t.Fatal(err)
	}
	if !got.Attributes[0].IsEvent {
		t.Error("IsEvent should be true")
	}
	if got.Attributes[0].Static != "handleClick" {
		t.Errorf("static = %q, want handleClick", got.Attributes[0].Static)
	}
}

func TestBinaryRoundTrip_TextContent(t *testing.T) {
	ir := &ui.IRNode{
		ComponentType: "Label",
		Text:          "Count: {{counter}}",
		Attributes: []ui.IRAttribute{
			{Name: "text", Expr: ui.ExprConcat{
				Parts: []ui.ExprNode{
					ui.ExprLiteral{Value: "Count: "},
					ui.ExprRef{Path: "counter"},
				},
			}},
		},
	}
	data, err := ui.EncodeIR(ir)
	if err != nil {
		t.Fatal(err)
	}
	got, err := ui.DecodeIR(data)
	if err != nil {
		t.Fatal(err)
	}
	if got.Text != "Count: {{counter}}" {
		t.Errorf("text = %q, want Count: {{counter}}", got.Text)
	}
}

func TestBinaryRoundTrip_FromXML(t *testing.T) {
	xmlData := `<Panel layout="vbox" spacing="12">
		<Label bind:text="title" />
		<Button text="Click" on:click="handleClick" />
		<Panel ui:show="showExtra">
			<Label text="Extra" />
		</Panel>
	</Panel>`

	ir, err := ui.CompileXML([]byte(xmlData))
	if err != nil {
		t.Fatal(err)
	}
	data, err := ui.EncodeIR(ir)
	if err != nil {
		t.Fatal(err)
	}
	got, err := ui.DecodeIR(data)
	if err != nil {
		t.Fatal(err)
	}
	if got.ComponentType != "Panel" {
		t.Errorf("root type = %q, want Panel", got.ComponentType)
	}
	if len(got.Children) != 3 {
		t.Fatalf("children = %d, want 3", len(got.Children))
	}
	if got.Children[0].ComponentType != "Label" {
		t.Errorf("child[0] = %q, want Label", got.Children[0].ComponentType)
	}
	if got.Children[1].ComponentType != "Button" {
		t.Errorf("child[1] = %q, want Button", got.Children[1].ComponentType)
	}
	// Verify the show directive survived
	if len(got.Children[2].Directives) != 1 {
		t.Fatalf("child[2] directives = %d, want 1", len(got.Children[2].Directives))
	}
	if got.Children[2].Directives[0].Type != ui.DirectiveShow {
		t.Errorf("directive type = %d, want DirectiveShow", got.Children[2].Directives[0].Type)
	}
}

func TestDecodeIR_BadMagic(t *testing.T) {
	data := []byte{0xFF, 0xFF, 0xFF, 0x00, 0x01, 0x00}
	_, err := ui.DecodeIR(data)
	if err == nil {
		t.Fatal("expected error for bad magic")
	}
}

func TestDecodeIR_TruncatedData(t *testing.T) {
	_, err := ui.DecodeIR([]byte{0x55, 0x49})
	if err == nil {
		t.Fatal("expected error for truncated data")
	}
}

func TestDecodeIR_BadVersion(t *testing.T) {
	data := []byte{'U', 'I', 'B', 0x00, 0xFF, 0xFF}
	_, err := ui.DecodeIR(data)
	if err == nil {
		t.Fatal("expected error for bad version")
	}
}

func TestEncodeIR_NilNode(t *testing.T) {
	_, err := ui.EncodeIR(nil)
	if err == nil {
		t.Fatal("expected error for nil node")
	}
}

func TestBinaryRoundTrip_EmptyNode(t *testing.T) {
	ir := &ui.IRNode{ComponentType: "Component"}
	data, err := ui.EncodeIR(ir)
	if err != nil {
		t.Fatal(err)
	}
	got, err := ui.DecodeIR(data)
	if err != nil {
		t.Fatal(err)
	}
	if got.ComponentType != "Component" {
		t.Errorf("type = %q, want Component", got.ComponentType)
	}
	if len(got.Attributes) != 0 {
		t.Errorf("attrs = %d, want 0", len(got.Attributes))
	}
	if len(got.Children) != 0 {
		t.Errorf("children = %d, want 0", len(got.Children))
	}
}

func TestBinaryRoundTrip_NilExprAttribute(t *testing.T) {
	ir := &ui.IRNode{
		ComponentType: "Label",
		Attributes: []ui.IRAttribute{
			{Name: "text", Static: "static", Expr: nil},
		},
	}
	data, err := ui.EncodeIR(ir)
	if err != nil {
		t.Fatal(err)
	}
	got, err := ui.DecodeIR(data)
	if err != nil {
		t.Fatal(err)
	}
	if got.Attributes[0].Expr != nil {
		t.Errorf("expr should be nil, got %T", got.Attributes[0].Expr)
	}
	if got.Attributes[0].Static != "static" {
		t.Errorf("static = %q, want static", got.Attributes[0].Static)
	}
}

func TestBinaryRoundTrip_AllBinOps(t *testing.T) {
	ops := []ui.BinOp{ui.BinAdd, ui.BinSub, ui.BinMul, ui.BinDiv, ui.BinMod, ui.BinEq, ui.BinNeq,
		ui.BinLt, ui.BinLte, ui.BinGt, ui.BinGte, ui.BinAnd, ui.BinOr}
	for _, op := range ops {
		ir := &ui.IRNode{
			ComponentType: "Label",
			Attributes: []ui.IRAttribute{
				{Name: "val", Expr: ui.ExprBinary{Op: op, Left: ui.ExprLiteral{Value: 1.0}, Right: ui.ExprLiteral{Value: 2.0}}},
			},
		}
		data, err := ui.EncodeIR(ir)
		if err != nil {
			t.Fatalf("encode op %d: %v", op, err)
		}
		got, err := ui.DecodeIR(data)
		if err != nil {
			t.Fatalf("decode op %d: %v", op, err)
		}
		bin := got.Attributes[0].Expr.(ui.ExprBinary)
		if bin.Op != op {
			t.Errorf("op = %d, want %d", bin.Op, op)
		}
	}
}
