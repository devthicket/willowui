package integration

import (
	"testing"

	ui "github.com/devthicket/willowui"
	"github.com/devthicket/willowui/internal/template"
)

// --- Parse tests ---

func TestParseExpression_Literal(t *testing.T) {
	tests := []struct {
		src  string
		want any
	}{
		{"42", 42.0},
		{"3.14", 3.14},
		{"'hello'", "hello"},
		{`"world"`, "world"},
		{"true", true},
		{"false", false},
		{"nil", nil},
	}
	for _, tt := range tests {
		node, err := ui.ParseExpression(tt.src)
		if err != nil {
			t.Errorf("ParseExpression(%q) error: %v", tt.src, err)
			continue
		}
		lit, ok := node.(ui.ExprLiteral)
		if !ok {
			t.Errorf("ParseExpression(%q) = %T, want ExprLiteral", tt.src, node)
			continue
		}
		if lit.Value != tt.want {
			t.Errorf("ParseExpression(%q).Value = %v, want %v", tt.src, lit.Value, tt.want)
		}
	}
}

func TestParseExpression_Ref(t *testing.T) {
	node, err := ui.ParseExpression("user.name")
	if err != nil {
		t.Fatal(err)
	}
	ref, ok := node.(ui.ExprRef)
	if !ok {
		t.Fatalf("expected ExprRef, got %T", node)
	}
	if ref.Path != "user.name" {
		t.Errorf("path = %q, want %q", ref.Path, "user.name")
	}
}

func TestParseExpression_Binary(t *testing.T) {
	tests := []string{
		"a + b",
		"x - 1",
		"a * b",
		"x / 2",
		"n % 3",
		"a == b",
		"a != b",
		"a < b",
		"a <= b",
		"a > b",
		"a >= b",
		"a && b",
		"a || b",
	}
	for _, src := range tests {
		node, err := ui.ParseExpression(src)
		if err != nil {
			t.Errorf("ParseExpression(%q) error: %v", src, err)
			continue
		}
		if _, ok := node.(ui.ExprBinary); !ok {
			t.Errorf("ParseExpression(%q) = %T, want ExprBinary", src, node)
		}
	}
}

func TestParseExpression_Unary(t *testing.T) {
	node, err := ui.ParseExpression("!visible")
	if err != nil {
		t.Fatal(err)
	}
	un, ok := node.(ui.ExprUnary)
	if !ok {
		t.Fatalf("expected ExprUnary, got %T", node)
	}
	if un.Op != ui.UnaryNot {
		t.Errorf("op = %d, want UnaryNot", un.Op)
	}

	node, err = ui.ParseExpression("-x")
	if err != nil {
		t.Fatal(err)
	}
	un, ok = node.(ui.ExprUnary)
	if !ok {
		t.Fatalf("expected ExprUnary, got %T", node)
	}
	if un.Op != ui.UnaryNeg {
		t.Errorf("op = %d, want UnaryNeg", un.Op)
	}
}

func TestParseExpression_Ternary(t *testing.T) {
	node, err := ui.ParseExpression("active ? 'yes' : 'no'")
	if err != nil {
		t.Fatal(err)
	}
	ter, ok := node.(ui.ExprTernary)
	if !ok {
		t.Fatalf("expected ExprTernary, got %T", node)
	}
	if _, ok := ter.Cond.(ui.ExprRef); !ok {
		t.Errorf("cond = %T, want ExprRef", ter.Cond)
	}
}

func TestParseExpression_Precedence(t *testing.T) {
	// a + b * c should parse as a + (b * c)
	node, err := ui.ParseExpression("a + b * c")
	if err != nil {
		t.Fatal(err)
	}
	bin, ok := node.(ui.ExprBinary)
	if !ok {
		t.Fatalf("expected ExprBinary, got %T", node)
	}
	if bin.Op != ui.BinAdd {
		t.Errorf("outer op = %d, want BinAdd", bin.Op)
	}
	inner, ok := bin.Right.(ui.ExprBinary)
	if !ok {
		t.Fatalf("right = %T, want ExprBinary", bin.Right)
	}
	if inner.Op != ui.BinMul {
		t.Errorf("inner op = %d, want BinMul", inner.Op)
	}
}

func TestParseExpression_Parens(t *testing.T) {
	node, err := ui.ParseExpression("(a + b) * c")
	if err != nil {
		t.Fatal(err)
	}
	bin, ok := node.(ui.ExprBinary)
	if !ok {
		t.Fatalf("expected ExprBinary, got %T", node)
	}
	if bin.Op != ui.BinMul {
		t.Errorf("outer op = %d, want BinMul", bin.Op)
	}
}

func TestParseExpression_Errors(t *testing.T) {
	tests := []string{
		"",
		"+ +",
		"a.",
		"@invalid",
	}
	for _, src := range tests {
		_, err := ui.ParseExpression(src)
		if err == nil {
			t.Errorf("ParseExpression(%q) expected error", src)
		}
	}
}

func TestParseExpression_StringEscape(t *testing.T) {
	node, err := ui.ParseExpression(`'hello\nworld'`)
	if err != nil {
		t.Fatal(err)
	}
	lit := node.(ui.ExprLiteral)
	if lit.Value != "hello\nworld" {
		t.Errorf("got %q, want %q", lit.Value, "hello\nworld")
	}
}

// --- Eval tests ---

type mockProvider struct {
	refs    map[string]any
	methods map[string]bool
}

func (m *mockProvider) LookupRef(path string) any {
	return m.refs[path]
}

func (m *mockProvider) CallMethod(name string) bool {
	v, ok := m.methods[name]
	return ok && v
}

func newMockCtx(refs map[string]any) *ui.EvalContext {
	return &ui.EvalContext{
		Provider: &mockProvider{refs: refs},
	}
}

func TestEvalExpression_Arithmetic(t *testing.T) {
	ctx := newMockCtx(map[string]any{
		"a": 10.0,
		"b": 3.0,
	})

	tests := []struct {
		src  string
		want float64
	}{
		{"a + b", 13},
		{"a - b", 7},
		{"a * b", 30},
		{"a / b", 10.0 / 3.0},
		{"a % b", 1},
	}
	for _, tt := range tests {
		node, err := ui.ParseExpression(tt.src)
		if err != nil {
			t.Fatal(err)
		}
		result, err := ui.EvalExpression(node, ctx)
		if err != nil {
			t.Errorf("eval(%q) error: %v", tt.src, err)
			continue
		}
		got, ok := result.(float64)
		if !ok {
			t.Errorf("eval(%q) = %T, want float64", tt.src, result)
			continue
		}
		if got != tt.want {
			t.Errorf("eval(%q) = %v, want %v", tt.src, got, tt.want)
		}
	}
}

func TestEvalExpression_Comparison(t *testing.T) {
	ctx := newMockCtx(map[string]any{"x": 5.0, "y": 10.0})

	tests := []struct {
		src  string
		want bool
	}{
		{"x < y", true},
		{"x > y", false},
		{"x == y", false},
		{"x != y", true},
		{"x <= 5", true},
		{"y >= 10", true},
	}
	for _, tt := range tests {
		node, _ := ui.ParseExpression(tt.src)
		result, err := ui.EvalExpression(node, ctx)
		if err != nil {
			t.Errorf("eval(%q) error: %v", tt.src, err)
			continue
		}
		if result != tt.want {
			t.Errorf("eval(%q) = %v, want %v", tt.src, result, tt.want)
		}
	}
}

func TestEvalExpression_Logical(t *testing.T) {
	ctx := newMockCtx(map[string]any{"a": true, "b": false})

	tests := []struct {
		src  string
		want bool
	}{
		{"a && b", false},
		{"a || b", true},
		{"!a", false},
		{"!b", true},
	}
	for _, tt := range tests {
		node, _ := ui.ParseExpression(tt.src)
		result, err := ui.EvalExpression(node, ctx)
		if err != nil {
			t.Errorf("eval(%q) error: %v", tt.src, err)
			continue
		}
		if result != tt.want {
			t.Errorf("eval(%q) = %v, want %v", tt.src, result, tt.want)
		}
	}
}

func TestEvalExpression_Ternary(t *testing.T) {
	ctx := newMockCtx(map[string]any{"flag": true})

	node, _ := ui.ParseExpression("flag ? 'yes' : 'no'")
	result, _ := ui.EvalExpression(node, ctx)
	if result != "yes" {
		t.Errorf("got %v, want 'yes'", result)
	}

	ctx.Provider.(*mockProvider).refs["flag"] = false
	result, _ = ui.EvalExpression(node, ctx)
	if result != "no" {
		t.Errorf("got %v, want 'no'", result)
	}
}

func TestEvalExpression_StringConcat(t *testing.T) {
	ctx := newMockCtx(map[string]any{"name": "world"})

	node, _ := ui.ParseExpression("'hello ' + name")
	result, _ := ui.EvalExpression(node, ctx)
	if result != "hello world" {
		t.Errorf("got %v, want 'hello world'", result)
	}
}

func TestEvalExpression_ReactiveRef(t *testing.T) {
	ref := ui.NewRef("reactive value")
	ctx := newMockCtx(map[string]any{"label": ref})

	node, _ := ui.ParseExpression("label")
	result, _ := ui.EvalExpression(node, ctx)
	if result != "reactive value" {
		t.Errorf("got %v, want 'reactive value'", result)
	}
}

func TestEvalExpression_Locals(t *testing.T) {
	ctx := &ui.EvalContext{
		Provider: &mockProvider{refs: map[string]any{"x": 100.0}},
		Locals:   map[string]any{"x": 42.0},
	}
	node, _ := ui.ParseExpression("x")
	result, _ := ui.EvalExpression(node, ctx)
	if result != 42.0 {
		t.Errorf("locals should shadow provider: got %v, want 42", result)
	}
}

func TestEvalExpression_DivByZero(t *testing.T) {
	ctx := newMockCtx(nil)
	node, _ := ui.ParseExpression("10 / 0")
	result, err := ui.EvalExpression(node, ctx)
	if err != nil {
		t.Fatal(err)
	}
	if result != 0.0 {
		t.Errorf("div by zero should return 0, got %v", result)
	}
}

// --- Interpolation tests ---

func TestParseInterpolation_NoInterpolation(t *testing.T) {
	_, found := template.ParseInterpolation("plain text")
	if found {
		t.Error("should not find interpolation in plain text")
	}
}

func TestParseInterpolation_SingleExpr(t *testing.T) {
	node, found := template.ParseInterpolation("{{name}}")
	if !found {
		t.Fatal("should find interpolation")
	}
	ref, ok := node.(ui.ExprRef)
	if !ok {
		t.Fatalf("expected ExprRef, got %T", node)
	}
	if ref.Path != "name" {
		t.Errorf("path = %q, want 'name'", ref.Path)
	}
}

func TestParseInterpolation_Mixed(t *testing.T) {
	node, found := template.ParseInterpolation("Hello {{name}}!")
	if !found {
		t.Fatal("should find interpolation")
	}
	concat, ok := node.(ui.ExprConcat)
	if !ok {
		t.Fatalf("expected ExprConcat, got %T", node)
	}
	if len(concat.Parts) != 3 {
		t.Fatalf("expected 3 parts, got %d", len(concat.Parts))
	}
}

func TestParseInterpolation_MultipleExprs(t *testing.T) {
	node, found := template.ParseInterpolation("{{first}} and {{second}}")
	if !found {
		t.Fatal("should find interpolation")
	}
	concat, ok := node.(ui.ExprConcat)
	if !ok {
		t.Fatalf("expected ExprConcat, got %T", node)
	}
	if len(concat.Parts) != 3 {
		t.Fatalf("expected 3 parts, got %d", len(concat.Parts))
	}
}

func TestToBool(t *testing.T) {
	tests := []struct {
		val  any
		want bool
	}{
		{nil, false},
		{true, true},
		{false, false},
		{0.0, false},
		{1.0, true},
		{"", false},
		{"text", true},
		{42, true},
		{0, false},
	}
	for _, tt := range tests {
		got := template.ToBool(tt.val)
		if got != tt.want {
			t.Errorf("ToBool(%v) = %v, want %v", tt.val, got, tt.want)
		}
	}
}
