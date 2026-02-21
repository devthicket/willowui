package template

import (
	"testing"
)

// ── ParseExpression ─────────────────────────────────────────────────────────

func TestParseExpression_Number(t *testing.T) {
	node, err := ParseExpression("42")
	if err != nil {
		t.Fatal(err)
	}
	lit, ok := node.(ExprLiteral)
	if !ok {
		t.Fatalf("expected ExprLiteral, got %T", node)
	}
	if lit.Value != 42.0 {
		t.Errorf("got %v, want 42", lit.Value)
	}
}

func TestParseExpression_String(t *testing.T) {
	node, err := ParseExpression(`'hello'`)
	if err != nil {
		t.Fatal(err)
	}
	lit, ok := node.(ExprLiteral)
	if !ok {
		t.Fatalf("expected ExprLiteral, got %T", node)
	}
	if lit.Value != "hello" {
		t.Errorf("got %v, want hello", lit.Value)
	}
}

func TestParseExpression_Bool(t *testing.T) {
	for _, tt := range []struct {
		src  string
		want bool
	}{
		{"true", true},
		{"false", false},
	} {
		node, err := ParseExpression(tt.src)
		if err != nil {
			t.Fatalf("ParseExpression(%q): %v", tt.src, err)
		}
		lit := node.(ExprLiteral)
		if lit.Value != tt.want {
			t.Errorf("ParseExpression(%q) = %v, want %v", tt.src, lit.Value, tt.want)
		}
	}
}

func TestParseExpression_Nil(t *testing.T) {
	node, err := ParseExpression("nil")
	if err != nil {
		t.Fatal(err)
	}
	lit := node.(ExprLiteral)
	if lit.Value != nil {
		t.Errorf("got %v, want nil", lit.Value)
	}
}

func TestParseExpression_Ref(t *testing.T) {
	node, err := ParseExpression("user.name")
	if err != nil {
		t.Fatal(err)
	}
	ref, ok := node.(ExprRef)
	if !ok {
		t.Fatalf("expected ExprRef, got %T", node)
	}
	if ref.Path != "user.name" {
		t.Errorf("got path %q, want %q", ref.Path, "user.name")
	}
}

func TestParseExpression_Binary(t *testing.T) {
	node, err := ParseExpression("1 + 2")
	if err != nil {
		t.Fatal(err)
	}
	bin, ok := node.(ExprBinary)
	if !ok {
		t.Fatalf("expected ExprBinary, got %T", node)
	}
	if bin.Op != BinAdd {
		t.Errorf("got op %v, want BinAdd", bin.Op)
	}
}

func TestParseExpression_Comparison(t *testing.T) {
	ops := []struct {
		src string
		op  BinOp
	}{
		{"a == b", BinEq},
		{"a != b", BinNeq},
		{"a < b", BinLt},
		{"a <= b", BinLte},
		{"a > b", BinGt},
		{"a >= b", BinGte},
	}
	for _, tt := range ops {
		node, err := ParseExpression(tt.src)
		if err != nil {
			t.Fatalf("ParseExpression(%q): %v", tt.src, err)
		}
		bin := node.(ExprBinary)
		if bin.Op != tt.op {
			t.Errorf("ParseExpression(%q): got op %v, want %v", tt.src, bin.Op, tt.op)
		}
	}
}

func TestParseExpression_Unary(t *testing.T) {
	node, err := ParseExpression("!visible")
	if err != nil {
		t.Fatal(err)
	}
	un, ok := node.(ExprUnary)
	if !ok {
		t.Fatalf("expected ExprUnary, got %T", node)
	}
	if un.Op != UnaryNot {
		t.Errorf("got op %v, want UnaryNot", un.Op)
	}
}

func TestParseExpression_Negation(t *testing.T) {
	node, err := ParseExpression("-5")
	if err != nil {
		t.Fatal(err)
	}
	un, ok := node.(ExprUnary)
	if !ok {
		t.Fatalf("expected ExprUnary, got %T", node)
	}
	if un.Op != UnaryNeg {
		t.Errorf("got op %v, want UnaryNeg", un.Op)
	}
}

func TestParseExpression_Ternary(t *testing.T) {
	node, err := ParseExpression("x > 0 ? 'yes' : 'no'")
	if err != nil {
		t.Fatal(err)
	}
	_, ok := node.(ExprTernary)
	if !ok {
		t.Fatalf("expected ExprTernary, got %T", node)
	}
}

func TestParseExpression_Precedence(t *testing.T) {
	// 1 + 2 * 3 should parse as 1 + (2 * 3)
	node, err := ParseExpression("1 + 2 * 3")
	if err != nil {
		t.Fatal(err)
	}
	bin := node.(ExprBinary)
	if bin.Op != BinAdd {
		t.Fatalf("top-level should be Add, got %v", bin.Op)
	}
	right := bin.Right.(ExprBinary)
	if right.Op != BinMul {
		t.Errorf("right side should be Mul, got %v", right.Op)
	}
}

func TestParseExpression_Parens(t *testing.T) {
	node, err := ParseExpression("(1 + 2) * 3")
	if err != nil {
		t.Fatal(err)
	}
	bin := node.(ExprBinary)
	if bin.Op != BinMul {
		t.Fatalf("top-level should be Mul, got %v", bin.Op)
	}
	left := bin.Left.(ExprBinary)
	if left.Op != BinAdd {
		t.Errorf("left side should be Add, got %v", left.Op)
	}
}

func TestParseExpression_Errors(t *testing.T) {
	bad := []string{
		"",
		"+ +",
		"1 +",
		"(1 + 2",
	}
	for _, s := range bad {
		_, err := ParseExpression(s)
		if err == nil {
			t.Errorf("ParseExpression(%q) expected error", s)
		}
	}
}

// ── EvalExpression ──────────────────────────────────────────────────────────

func eval(t *testing.T, src string, locals map[string]any) any {
	t.Helper()
	node, err := ParseExpression(src)
	if err != nil {
		t.Fatalf("parse %q: %v", src, err)
	}
	ctx := &EvalContext{Locals: locals}
	result, err := EvalExpression(node, ctx)
	if err != nil {
		t.Fatalf("eval %q: %v", src, err)
	}
	return result
}

func TestEval_Arithmetic(t *testing.T) {
	tests := []struct {
		src  string
		want float64
	}{
		{"1 + 2", 3},
		{"10 - 3", 7},
		{"4 * 5", 20},
		{"10 / 4", 2.5},
		{"10 % 3", 1},
	}
	for _, tt := range tests {
		got := eval(t, tt.src, nil)
		f, ok := got.(float64)
		if !ok {
			t.Errorf("eval(%q) = %T, want float64", tt.src, got)
			continue
		}
		if f != tt.want {
			t.Errorf("eval(%q) = %v, want %v", tt.src, f, tt.want)
		}
	}
}

func TestEval_Comparison(t *testing.T) {
	tests := []struct {
		src  string
		want bool
	}{
		{"5 == 5", true},
		{"5 != 3", true},
		{"3 < 5", true},
		{"5 < 3", false},
		{"5 <= 5", true},
		{"5 > 3", true},
		{"3 >= 5", false},
	}
	for _, tt := range tests {
		got := eval(t, tt.src, nil)
		b, ok := got.(bool)
		if !ok {
			t.Errorf("eval(%q) = %T, want bool", tt.src, got)
			continue
		}
		if b != tt.want {
			t.Errorf("eval(%q) = %v, want %v", tt.src, b, tt.want)
		}
	}
}

func TestEval_Logical(t *testing.T) {
	tests := []struct {
		src  string
		want bool
	}{
		{"true && true", true},
		{"true && false", false},
		{"false || true", true},
		{"false || false", false},
	}
	for _, tt := range tests {
		got := eval(t, tt.src, nil)
		b := got.(bool)
		if b != tt.want {
			t.Errorf("eval(%q) = %v, want %v", tt.src, b, tt.want)
		}
	}
}

func TestEval_Unary(t *testing.T) {
	if got := eval(t, "!true", nil); got != false {
		t.Errorf("!true = %v", got)
	}
	if got := eval(t, "-5", nil).(float64); got != -5 {
		t.Errorf("-5 = %v", got)
	}
}

func TestEval_Ternary(t *testing.T) {
	got := eval(t, "true ? 'yes' : 'no'", nil)
	if got != "yes" {
		t.Errorf("got %v, want yes", got)
	}
	got = eval(t, "false ? 'yes' : 'no'", nil)
	if got != "no" {
		t.Errorf("got %v, want no", got)
	}
}

func TestEval_StringConcat(t *testing.T) {
	got := eval(t, "'hello' + ' ' + 'world'", nil)
	if got != "hello world" {
		t.Errorf("got %v, want 'hello world'", got)
	}
}

func TestEval_DivByZero(t *testing.T) {
	got := eval(t, "10 / 0", nil)
	if got != 0.0 {
		t.Errorf("10/0 = %v, want 0", got)
	}
}

func TestEval_Locals(t *testing.T) {
	locals := map[string]any{"x": 10.0, "y": 3.0}
	got := eval(t, "x + y", locals)
	if got != 13.0 {
		t.Errorf("x+y = %v, want 13", got)
	}
}

func TestEval_NilExpr(t *testing.T) {
	_, err := EvalExpression(nil, &EvalContext{})
	if err == nil {
		t.Error("expected error for nil node")
	}
}

// ── ToBool ──────────────────────────────────────────────────────────────────

func TestToBool(t *testing.T) {
	tests := []struct {
		val  any
		want bool
	}{
		{nil, false},
		{false, false},
		{true, true},
		{0.0, false},
		{1.0, true},
		{-1.0, true},
		{0, false},
		{1, true},
		{"", false},
		{"hello", true},
	}
	for _, tt := range tests {
		got := ToBool(tt.val)
		if got != tt.want {
			t.Errorf("ToBool(%v) = %v, want %v", tt.val, got, tt.want)
		}
	}
}

// ── ParseInterpolation ──────────────────────────────────────────────────────

func TestParseInterpolation_NoInterp(t *testing.T) {
	_, found := ParseInterpolation("plain text")
	if found {
		t.Error("expected no interpolation in plain text")
	}
}

func TestParseInterpolation_Simple(t *testing.T) {
	node, found := ParseInterpolation("{{name}}")
	if !found {
		t.Fatal("expected interpolation")
	}
	ref, ok := node.(ExprRef)
	if !ok {
		t.Fatalf("expected ExprRef, got %T", node)
	}
	if ref.Path != "name" {
		t.Errorf("got path %q, want %q", ref.Path, "name")
	}
}

func TestParseInterpolation_Mixed(t *testing.T) {
	node, found := ParseInterpolation("Hello {{name}}!")
	if !found {
		t.Fatal("expected interpolation")
	}
	concat, ok := node.(ExprConcat)
	if !ok {
		t.Fatalf("expected ExprConcat, got %T", node)
	}
	if len(concat.Parts) != 3 {
		t.Fatalf("got %d parts, want 3", len(concat.Parts))
	}
}

func TestParseInterpolation_Unterminated(t *testing.T) {
	node, found := ParseInterpolation("{{oops")
	if !found {
		t.Fatal("expected found=true even for unterminated")
	}
	// Unterminated treated as literal
	lit, ok := node.(ExprLiteral)
	if !ok {
		t.Fatalf("expected ExprLiteral, got %T", node)
	}
	if lit.Value != "{{oops" {
		t.Errorf("got %v, want {{oops", lit.Value)
	}
}

// ── toFloat conversion (tested via eval) ────────────────────────────────────

func TestEval_ToFloat_StringToFloat(t *testing.T) {
	// "5" + 0 should convert "5" to 5.0 via toFloat
	got := eval(t, "'5' + 0", nil)
	// With string on left, BinAdd does string concat: "5" + "0" = "50"
	// Actually — only if either side is string does it concat. "0" is float64
	// so left is string, triggers string concat path: "5" + fmt.Sprint(0) = "50"
	if got != "50" {
		t.Errorf("'5' + 0 = %v (%T), want '50'", got, got)
	}
}

func TestEval_ToFloat_BoolToFloat(t *testing.T) {
	// true in arithmetic context: true * 3 → 1 * 3 = 3
	got := eval(t, "true * 3", nil)
	if got != 3.0 {
		t.Errorf("true * 3 = %v, want 3", got)
	}
	got = eval(t, "false * 3", nil)
	if got != 0.0 {
		t.Errorf("false * 3 = %v, want 0", got)
	}
}

func TestEval_ToFloat_NilToFloat(t *testing.T) {
	// nil + 0 → toFloat(nil)=0 + 0 = 0
	got := eval(t, "nil + 0", nil)
	if got != 0.0 {
		t.Errorf("nil + 0 = %v, want 0", got)
	}
}

func TestEval_ToFloat_IntToFloat(t *testing.T) {
	// Test via locals with int value
	locals := map[string]any{"x": 5}
	got := eval(t, "x * 2", locals)
	if got != 10.0 {
		t.Errorf("x*2 = %v, want 10", got)
	}
}

// ── evalConcat (tested via ParseInterpolation + eval) ───────────────────────

func TestEval_Concat(t *testing.T) {
	node, found := ParseInterpolation("Hello {{name}}, you are {{age}}!")
	if !found {
		t.Fatal("expected interpolation")
	}
	ctx := &EvalContext{Locals: map[string]any{"name": "Alice", "age": 30.0}}
	result, err := EvalExpression(node, ctx)
	if err != nil {
		t.Fatal(err)
	}
	if result != "Hello Alice, you are 30!" {
		t.Errorf("got %q, want 'Hello Alice, you are 30!'", result)
	}
}

func TestEval_ConcatSingleInterp(t *testing.T) {
	node, found := ParseInterpolation("{{greeting}}")
	if !found {
		t.Fatal("expected interpolation")
	}
	ctx := &EvalContext{Locals: map[string]any{"greeting": "hi"}}
	result, err := EvalExpression(node, ctx)
	if err != nil {
		t.Fatal(err)
	}
	if result != "hi" {
		t.Errorf("got %q, want 'hi'", result)
	}
}

// ── readString with double quotes ───────────────────────────────────────────

func TestParseExpression_DoubleQuotedString(t *testing.T) {
	node, err := ParseExpression(`"hello"`)
	if err != nil {
		t.Fatal(err)
	}
	lit, ok := node.(ExprLiteral)
	if !ok {
		t.Fatalf("expected ExprLiteral, got %T", node)
	}
	if lit.Value != "hello" {
		t.Errorf("got %v, want hello", lit.Value)
	}
}

func TestParseExpression_DoubleQuotedEscape(t *testing.T) {
	node, err := ParseExpression(`"line\nnext"`)
	if err != nil {
		t.Fatal(err)
	}
	lit := node.(ExprLiteral)
	if lit.Value != "line\nnext" {
		t.Errorf("got %q, want %q", lit.Value, "line\nnext")
	}
}

func TestParseExpression_StringEscapeTab(t *testing.T) {
	node, err := ParseExpression(`'a\tb'`)
	if err != nil {
		t.Fatal(err)
	}
	lit := node.(ExprLiteral)
	if lit.Value != "a\tb" {
		t.Errorf("got %q, want %q", lit.Value, "a\tb")
	}
}

func TestParseExpression_StringEscapeBackslash(t *testing.T) {
	node, err := ParseExpression(`'a\\b'`)
	if err != nil {
		t.Fatal(err)
	}
	lit := node.(ExprLiteral)
	if lit.Value != "a\\b" {
		t.Errorf("got %q, want %q", lit.Value, "a\\b")
	}
}

// ── Dotted paths ────────────────────────────────────────────────────────────

func TestParseExpression_DottedPath(t *testing.T) {
	node, err := ParseExpression("a.b.c")
	if err != nil {
		t.Fatal(err)
	}
	ref, ok := node.(ExprRef)
	if !ok {
		t.Fatalf("expected ExprRef, got %T", node)
	}
	if ref.Path != "a.b.c" {
		t.Errorf("got path %q, want a.b.c", ref.Path)
	}
}

func TestParseExpression_DottedPathInExpression(t *testing.T) {
	node, err := ParseExpression("user.score > 100")
	if err != nil {
		t.Fatal(err)
	}
	bin, ok := node.(ExprBinary)
	if !ok {
		t.Fatalf("expected ExprBinary, got %T", node)
	}
	ref := bin.Left.(ExprRef)
	if ref.Path != "user.score" {
		t.Errorf("got path %q, want user.score", ref.Path)
	}
}

// ── Modulo operator ─────────────────────────────────────────────────────────

func TestEval_Modulo(t *testing.T) {
	got := eval(t, "7 % 3", nil)
	if got != 1.0 {
		t.Errorf("7 %% 3 = %v, want 1", got)
	}
}

func TestEval_ModuloByZero(t *testing.T) {
	got := eval(t, "7 % 0", nil)
	if got != 0.0 {
		t.Errorf("7 %% 0 = %v, want 0", got)
	}
}

func TestParseExpression_Modulo(t *testing.T) {
	node, err := ParseExpression("7 % 3")
	if err != nil {
		t.Fatal(err)
	}
	bin, ok := node.(ExprBinary)
	if !ok {
		t.Fatalf("expected ExprBinary, got %T", node)
	}
	if bin.Op != BinMod {
		t.Errorf("op = %v, want BinMod", bin.Op)
	}
}

// ── String + number concatenation ───────────────────────────────────────────

func TestEval_StringNumberConcat(t *testing.T) {
	got := eval(t, "'count: ' + 5", nil)
	if got != "count: 5" {
		t.Errorf("got %v, want 'count: 5'", got)
	}
}

func TestEval_NumberStringConcat(t *testing.T) {
	got := eval(t, "5 + ' items'", nil)
	if got != "5 items" {
		t.Errorf("got %v, want '5 items'", got)
	}
}

// ── Additional edge cases ───────────────────────────────────────────────────

func TestEval_NestedTernary(t *testing.T) {
	got := eval(t, "true ? false ? 'a' : 'b' : 'c'", nil)
	if got != "b" {
		t.Errorf("nested ternary = %v, want b", got)
	}
}

func TestEval_UnaryNegOfRef(t *testing.T) {
	locals := map[string]any{"x": 5.0}
	got := eval(t, "-x", locals)
	if got != -5.0 {
		t.Errorf("-x = %v, want -5", got)
	}
}

func TestEval_ComplexExpression(t *testing.T) {
	locals := map[string]any{"a": 10.0, "b": 3.0}
	got := eval(t, "(a + b) * 2 - 1", locals)
	if got != 25.0 {
		t.Errorf("(a+b)*2-1 = %v, want 25", got)
	}
}

func TestEval_ParentContext(t *testing.T) {
	parent := &EvalContext{Locals: map[string]any{"x": 10.0}}
	child := &EvalContext{Locals: map[string]any{"y": 20.0}, Parent: parent}
	node, err := ParseExpression("x + y")
	if err != nil {
		t.Fatal(err)
	}
	result, err := EvalExpression(node, child)
	if err != nil {
		t.Fatal(err)
	}
	if result != 30.0 {
		t.Errorf("x+y with parent context = %v, want 30", result)
	}
}

func TestParseExpression_UnterminatedString(t *testing.T) {
	_, err := ParseExpression(`'unterminated`)
	if err == nil {
		t.Error("expected error for unterminated string")
	}
}

func TestParseExpression_FloatNumber(t *testing.T) {
	node, err := ParseExpression("3.14")
	if err != nil {
		t.Fatal(err)
	}
	lit := node.(ExprLiteral)
	if lit.Value != 3.14 {
		t.Errorf("got %v, want 3.14", lit.Value)
	}
}

func TestParseExpression_UnexpectedChar(t *testing.T) {
	_, err := ParseExpression("@")
	if err == nil {
		t.Error("expected error for unexpected character")
	}
}

func TestParseInterpolation_MultipleInterps(t *testing.T) {
	node, found := ParseInterpolation("{{a}} and {{b}}")
	if !found {
		t.Fatal("expected interpolation")
	}
	concat, ok := node.(ExprConcat)
	if !ok {
		t.Fatalf("expected ExprConcat, got %T", node)
	}
	if len(concat.Parts) != 3 {
		t.Fatalf("got %d parts, want 3", len(concat.Parts))
	}
}

func TestParseInterpolation_ExpressionInside(t *testing.T) {
	node, found := ParseInterpolation("{{x + 1}}")
	if !found {
		t.Fatal("expected interpolation")
	}
	bin, ok := node.(ExprBinary)
	if !ok {
		t.Fatalf("expected ExprBinary, got %T", node)
	}
	if bin.Op != BinAdd {
		t.Errorf("op = %v, want BinAdd", bin.Op)
	}
}

func TestParseInterpolation_BadExprInsideTreatedAsLiteral(t *testing.T) {
	node, found := ParseInterpolation("{{1 +}}")
	if !found {
		t.Fatal("expected found=true")
	}
	// Bad expression inside {{ }} is treated as literal text
	lit, ok := node.(ExprLiteral)
	if !ok {
		t.Fatalf("expected ExprLiteral for bad expression, got %T", node)
	}
	if lit.Value != "{{1 +}}" {
		t.Errorf("got %v, want '{{1 +}}'", lit.Value)
	}
}

func TestEval_ShortCircuitAnd(t *testing.T) {
	// false && anything should short-circuit and not evaluate right side
	got := eval(t, "false && true", nil)
	if got != false {
		t.Errorf("false && true = %v, want false", got)
	}
}

func TestEval_ShortCircuitOr(t *testing.T) {
	// true || anything should short-circuit
	got := eval(t, "true || false", nil)
	if got != true {
		t.Errorf("true || false = %v, want true", got)
	}
}

func TestEval_DoubleNegation(t *testing.T) {
	got := eval(t, "!!true", nil)
	if got != true {
		t.Errorf("!!true = %v, want true", got)
	}
}

func TestEval_RefMissing(t *testing.T) {
	// Missing ref should return nil
	got := eval(t, "missing", nil)
	if got != nil {
		t.Errorf("missing ref = %v, want nil", got)
	}
}
