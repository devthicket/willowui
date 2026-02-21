package template

import (
	"fmt"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/devthicket/willowui/internal/sg"
	"github.com/devthicket/willowui/internal/widget"
)

// ExprNode is the interface for all expression AST nodes.
type ExprNode interface {
	exprNode()
}

// ExprRef is a dotted reference path like "user.name" or just "count".
type ExprRef struct {
	Path string
}

func (ExprRef) exprNode() {}

// ExprLiteral is a constant value (string, float64, bool, nil).
type ExprLiteral struct {
	Value any
}

func (ExprLiteral) exprNode() {}

// BinOp identifies a binary operator.
type BinOp int

const (
	BinAdd BinOp = iota
	BinSub
	BinMul
	BinDiv
	BinMod
	BinEq
	BinNeq
	BinLt
	BinLte
	BinGt
	BinGte
	BinAnd
	BinOr
)

// ExprBinary is a binary operation (e.g. a + b, x == y).
type ExprBinary struct {
	Op    BinOp
	Left  ExprNode
	Right ExprNode
}

func (ExprBinary) exprNode() {}

// UnaryOp identifies a unary operator.
type UnaryOp int

const (
	UnaryNot UnaryOp = iota
	UnaryNeg
)

// ExprUnary is a unary operation (e.g. !visible, -offset).
type ExprUnary struct {
	Op      UnaryOp
	Operand ExprNode
}

func (ExprUnary) exprNode() {}

// ExprTernary is a ternary conditional (cond ? then : else).
type ExprTernary struct {
	Cond ExprNode
	Then ExprNode
	Else ExprNode
}

func (ExprTernary) exprNode() {}

// ExprConcat is a string concatenation of parts (from interpolation).
type ExprConcat struct {
	Parts []ExprNode
}

func (ExprConcat) exprNode() {}

// --- Tokenizer ---

type tokenKind int

const (
	tokEOF tokenKind = iota
	tokNumber
	tokString
	tokIdent
	tokTrue
	tokFalse
	tokNil
	tokPlus
	tokMinus
	tokStar
	tokSlash
	tokPercent
	tokEqEq
	tokBangEq
	tokLt
	tokLte
	tokGt
	tokGte
	tokAmpAmp
	tokPipePipe
	tokBang
	tokQuestion
	tokColon
	tokDot
	tokLParen
	tokRParen
)

type token struct {
	kind tokenKind
	sval string
	nval float64
}

type lexer struct {
	src []rune
	pos int
}

func newLexer(src string) *lexer {
	return &lexer{src: []rune(src)}
}

func (l *lexer) peek() rune {
	if l.pos >= len(l.src) {
		return 0
	}
	return l.src[l.pos]
}

func (l *lexer) next() rune {
	r := l.peek()
	l.pos++
	return r
}

func (l *lexer) skipWhitespace() {
	for l.pos < len(l.src) && unicode.IsSpace(l.src[l.pos]) {
		l.pos++
	}
}

func (l *lexer) nextToken() (token, error) {
	l.skipWhitespace()
	if l.pos >= len(l.src) {
		return token{kind: tokEOF}, nil
	}

	ch := l.peek()

	// Numbers
	if ch >= '0' && ch <= '9' {
		return l.readNumber()
	}

	// Strings
	if ch == '\'' || ch == '"' {
		return l.readString()
	}

	// Identifiers and keywords
	if ch == '_' || unicode.IsLetter(ch) {
		return l.readIdent(), nil
	}

	// Two-character operators
	if l.pos+1 < len(l.src) {
		two := string(l.src[l.pos : l.pos+2])
		switch two {
		case "==":
			l.pos += 2
			return token{kind: tokEqEq}, nil
		case "!=":
			l.pos += 2
			return token{kind: tokBangEq}, nil
		case "<=":
			l.pos += 2
			return token{kind: tokLte}, nil
		case ">=":
			l.pos += 2
			return token{kind: tokGte}, nil
		case "&&":
			l.pos += 2
			return token{kind: tokAmpAmp}, nil
		case "||":
			l.pos += 2
			return token{kind: tokPipePipe}, nil
		}
	}

	// Single-character operators
	l.pos++
	switch ch {
	case '+':
		return token{kind: tokPlus}, nil
	case '-':
		return token{kind: tokMinus}, nil
	case '*':
		return token{kind: tokStar}, nil
	case '/':
		return token{kind: tokSlash}, nil
	case '%':
		return token{kind: tokPercent}, nil
	case '<':
		return token{kind: tokLt}, nil
	case '>':
		return token{kind: tokGt}, nil
	case '!':
		return token{kind: tokBang}, nil
	case '?':
		return token{kind: tokQuestion}, nil
	case ':':
		return token{kind: tokColon}, nil
	case '.':
		return token{kind: tokDot}, nil
	case '(':
		return token{kind: tokLParen}, nil
	case ')':
		return token{kind: tokRParen}, nil
	}

	return token{}, fmt.Errorf("unexpected character %q at position %d", ch, l.pos-1)
}

func (l *lexer) readNumber() (token, error) {
	start := l.pos
	for l.pos < len(l.src) && (l.src[l.pos] >= '0' && l.src[l.pos] <= '9') {
		l.pos++
	}
	if l.pos < len(l.src) && l.src[l.pos] == '.' {
		l.pos++
		for l.pos < len(l.src) && (l.src[l.pos] >= '0' && l.src[l.pos] <= '9') {
			l.pos++
		}
	}
	s := string(l.src[start:l.pos])
	n, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return token{}, fmt.Errorf("invalid number %q", s)
	}
	return token{kind: tokNumber, nval: n}, nil
}

func (l *lexer) readString() (token, error) {
	quote := l.next()
	var buf strings.Builder
	for l.pos < len(l.src) {
		ch := l.next()
		if ch == quote {
			return token{kind: tokString, sval: buf.String()}, nil
		}
		if ch == '\\' && l.pos < len(l.src) {
			esc := l.next()
			switch esc {
			case 'n':
				buf.WriteByte('\n')
			case 't':
				buf.WriteByte('\t')
			case '\\':
				buf.WriteByte('\\')
			default:
				buf.WriteRune(esc)
			}
			continue
		}
		buf.WriteRune(ch)
	}
	return token{}, fmt.Errorf("unterminated string")
}

func (l *lexer) readIdent() token {
	start := l.pos
	for l.pos < len(l.src) && (l.src[l.pos] == '_' || unicode.IsLetter(l.src[l.pos]) || unicode.IsDigit(l.src[l.pos])) {
		l.pos++
	}
	s := string(l.src[start:l.pos])
	switch s {
	case "true":
		return token{kind: tokTrue}
	case "false":
		return token{kind: tokFalse}
	case "nil":
		return token{kind: tokNil}
	default:
		return token{kind: tokIdent, sval: s}
	}
}

// --- Parser ---

type parser struct {
	lex     *lexer
	current token
	err     error
}

func newParser(src string) *parser {
	p := &parser{lex: newLexer(src)}
	p.advance()
	return p
}

func (p *parser) advance() {
	if p.err != nil {
		return
	}
	t, err := p.lex.nextToken()
	if err != nil {
		p.err = err
		return
	}
	p.current = t
}

func (p *parser) expect(k tokenKind) {
	if p.current.kind != k {
		p.err = fmt.Errorf("expected token %d, got %d", k, p.current.kind)
		return
	}
	p.advance()
}

// ParseExpression parses an expression string into an AST.
func ParseExpression(src string) (ExprNode, error) {
	p := newParser(src)
	node := p.parseTernary()
	if p.err != nil {
		return nil, p.err
	}
	if p.current.kind != tokEOF {
		return nil, fmt.Errorf("unexpected token after expression")
	}
	return node, nil
}

func (p *parser) parseTernary() ExprNode {
	cond := p.parseOr()
	if p.err != nil || p.current.kind != tokQuestion {
		return cond
	}
	p.advance() // skip ?
	then := p.parseTernary()
	p.expect(tokColon)
	els := p.parseTernary()
	return ExprTernary{Cond: cond, Then: then, Else: els}
}

func (p *parser) parseOr() ExprNode {
	left := p.parseAnd()
	for p.err == nil && p.current.kind == tokPipePipe {
		p.advance()
		right := p.parseAnd()
		left = ExprBinary{Op: BinOr, Left: left, Right: right}
	}
	return left
}

func (p *parser) parseAnd() ExprNode {
	left := p.parseEquality()
	for p.err == nil && p.current.kind == tokAmpAmp {
		p.advance()
		right := p.parseEquality()
		left = ExprBinary{Op: BinAnd, Left: left, Right: right}
	}
	return left
}

func (p *parser) parseEquality() ExprNode {
	left := p.parseComparison()
	for p.err == nil {
		switch p.current.kind {
		case tokEqEq:
			p.advance()
			right := p.parseComparison()
			left = ExprBinary{Op: BinEq, Left: left, Right: right}
		case tokBangEq:
			p.advance()
			right := p.parseComparison()
			left = ExprBinary{Op: BinNeq, Left: left, Right: right}
		default:
			return left
		}
	}
	return left
}

func (p *parser) parseComparison() ExprNode {
	left := p.parseAddition()
	for p.err == nil {
		switch p.current.kind {
		case tokLt:
			p.advance()
			right := p.parseAddition()
			left = ExprBinary{Op: BinLt, Left: left, Right: right}
		case tokLte:
			p.advance()
			right := p.parseAddition()
			left = ExprBinary{Op: BinLte, Left: left, Right: right}
		case tokGt:
			p.advance()
			right := p.parseAddition()
			left = ExprBinary{Op: BinGt, Left: left, Right: right}
		case tokGte:
			p.advance()
			right := p.parseAddition()
			left = ExprBinary{Op: BinGte, Left: left, Right: right}
		default:
			return left
		}
	}
	return left
}

func (p *parser) parseAddition() ExprNode {
	left := p.parseMultiply()
	for p.err == nil {
		switch p.current.kind {
		case tokPlus:
			p.advance()
			right := p.parseMultiply()
			left = ExprBinary{Op: BinAdd, Left: left, Right: right}
		case tokMinus:
			p.advance()
			right := p.parseMultiply()
			left = ExprBinary{Op: BinSub, Left: left, Right: right}
		default:
			return left
		}
	}
	return left
}

func (p *parser) parseMultiply() ExprNode {
	left := p.parseUnary()
	for p.err == nil {
		switch p.current.kind {
		case tokStar:
			p.advance()
			right := p.parseUnary()
			left = ExprBinary{Op: BinMul, Left: left, Right: right}
		case tokSlash:
			p.advance()
			right := p.parseUnary()
			left = ExprBinary{Op: BinDiv, Left: left, Right: right}
		case tokPercent:
			p.advance()
			right := p.parseUnary()
			left = ExprBinary{Op: BinMod, Left: left, Right: right}
		default:
			return left
		}
	}
	return left
}

func (p *parser) parseUnary() ExprNode {
	if p.err != nil {
		return nil
	}
	switch p.current.kind {
	case tokBang:
		p.advance()
		operand := p.parseUnary()
		return ExprUnary{Op: UnaryNot, Operand: operand}
	case tokMinus:
		p.advance()
		operand := p.parseUnary()
		return ExprUnary{Op: UnaryNeg, Operand: operand}
	}
	return p.parsePrimary()
}

func (p *parser) parsePrimary() ExprNode {
	if p.err != nil {
		return nil
	}
	switch p.current.kind {
	case tokNumber:
		n := p.current.nval
		p.advance()
		return ExprLiteral{Value: n}
	case tokString:
		s := p.current.sval
		p.advance()
		return ExprLiteral{Value: s}
	case tokTrue:
		p.advance()
		return ExprLiteral{Value: true}
	case tokFalse:
		p.advance()
		return ExprLiteral{Value: false}
	case tokNil:
		p.advance()
		return ExprLiteral{Value: nil}
	case tokIdent:
		name := p.current.sval
		p.advance()
		// Build dotted path
		for p.err == nil && p.current.kind == tokDot {
			p.advance()
			if p.current.kind != tokIdent {
				p.err = fmt.Errorf("expected identifier after '.'")
				return nil
			}
			name += "." + p.current.sval
			p.advance()
		}
		return ExprRef{Path: name}
	case tokLParen:
		p.advance()
		expr := p.parseTernary()
		p.expect(tokRParen)
		return expr
	}
	p.err = fmt.Errorf("unexpected token in expression")
	return nil
}

// --- Interpolation ---

// ParseInterpolation is the exported equivalent of parseInterpolation, for use
// by the root package tests which cannot call the unexported version.
func ParseInterpolation(text string) (ExprNode, bool) { return parseInterpolation(text) }

// parseInterpolation parses text containing {{expr}} interpolations.
// Returns the expression node and true if interpolation was found,
// or nil and false if the text contains no {{ }}.
func parseInterpolation(text string) (ExprNode, bool) {
	if !strings.Contains(text, "{{") {
		return nil, false
	}

	var parts []ExprNode
	remaining := text
	for {
		idx := strings.Index(remaining, "{{")
		if idx < 0 {
			if len(remaining) > 0 {
				parts = append(parts, ExprLiteral{Value: remaining})
			}
			break
		}
		if idx > 0 {
			parts = append(parts, ExprLiteral{Value: remaining[:idx]})
		}
		remaining = remaining[idx+2:]
		end := strings.Index(remaining, "}}")
		if end < 0 {
			// Unterminated — treat rest as literal
			parts = append(parts, ExprLiteral{Value: "{{" + remaining})
			break
		}
		exprSrc := strings.TrimSpace(remaining[:end])
		expr, err := ParseExpression(exprSrc)
		if err != nil {
			// On parse error, treat as literal
			parts = append(parts, ExprLiteral{Value: "{{" + remaining[:end] + "}}"})
		} else {
			parts = append(parts, expr)
		}
		remaining = remaining[end+2:]
	}

	if len(parts) == 1 {
		return parts[0], true
	}
	return ExprConcat{Parts: parts}, true
}

// --- Evaluator ---

// DataProvider is implemented by controllers that support XML template data binding.
type DataProvider interface {
	LookupRef(path string) any
	CallMethod(name string) bool
}

// EvalContext provides the evaluation environment for expressions.
type EvalContext struct {
	Provider DataProvider
	Locals   map[string]any
	Parent   *EvalContext
}

// EvalExpression evaluates an expression AST node in the given context.
func EvalExpression(node ExprNode, ctx *EvalContext) (any, error) {
	if node == nil {
		return nil, fmt.Errorf("nil expression node")
	}
	switch n := node.(type) {
	case ExprLiteral:
		return n.Value, nil
	case ExprRef:
		return evalRef(n.Path, ctx)
	case ExprBinary:
		return evalBinary(n, ctx)
	case ExprUnary:
		return evalUnary(n, ctx)
	case ExprTernary:
		return evalTernary(n, ctx)
	case ExprConcat:
		return evalConcat(n, ctx)
	default:
		return nil, fmt.Errorf("unknown expression node type %T", node)
	}
}

func evalRef(path string, ctx *EvalContext) (any, error) {
	// Check locals first (walk up parent chain)
	for c := ctx; c != nil; c = c.Parent {
		if c.Locals != nil {
			if v, ok := c.Locals[path]; ok {
				return unwrapReactive(v), nil
			}
		}
	}
	// Fall back to provider
	if ctx.Provider != nil {
		v := ctx.Provider.LookupRef(path)
		return unwrapReactive(v), nil
	}
	return nil, nil
}

// unwrapReactive extracts the current value from Ref[T] or Computed[T] types.
func unwrapReactive(v any) any {
	if v == nil {
		return nil
	}
	// Type-switch on known reactive types
	switch r := v.(type) {
	case *widget.Ref[string]:
		return r.Get()
	case *widget.Ref[float64]:
		return r.Get()
	case *widget.Ref[int]:
		return r.Get()
	case *widget.Ref[bool]:
		return r.Get()
	case *widget.Computed[string]:
		return r.Get()
	case *widget.Computed[float64]:
		return r.Get()
	case *widget.Computed[int]:
		return r.Get()
	case *widget.Computed[bool]:
		return r.Get()
	case *widget.Ref[sg.Color]:
		return r.Get()
	case *widget.Computed[sg.Color]:
		return r.Get()
	case *widget.Ref[time.Time]:
		return r.Get()
	case *widget.Computed[time.Time]:
		return r.Get()
	}
	return v
}

func evalBinary(n ExprBinary, ctx *EvalContext) (any, error) {
	left, err := EvalExpression(n.Left, ctx)
	if err != nil {
		return nil, err
	}

	// Short-circuit for logical operators
	if n.Op == BinAnd {
		if !toBool(left) {
			return false, nil
		}
		right, err := EvalExpression(n.Right, ctx)
		if err != nil {
			return nil, err
		}
		return toBool(right), nil
	}
	if n.Op == BinOr {
		if toBool(left) {
			return true, nil
		}
		right, err := EvalExpression(n.Right, ctx)
		if err != nil {
			return nil, err
		}
		return toBool(right), nil
	}

	right, err := EvalExpression(n.Right, ctx)
	if err != nil {
		return nil, err
	}

	switch n.Op {
	case BinEq:
		return toFloat(left) == toFloat(right), nil
	case BinNeq:
		return toFloat(left) != toFloat(right), nil
	case BinLt:
		return toFloat(left) < toFloat(right), nil
	case BinLte:
		return toFloat(left) <= toFloat(right), nil
	case BinGt:
		return toFloat(left) > toFloat(right), nil
	case BinGte:
		return toFloat(left) >= toFloat(right), nil
	case BinAdd:
		// String concatenation if either side is a string
		ls, lIsStr := left.(string)
		rs, rIsStr := right.(string)
		if lIsStr || rIsStr {
			if !lIsStr {
				ls = fmt.Sprint(left)
			}
			if !rIsStr {
				rs = fmt.Sprint(right)
			}
			return ls + rs, nil
		}
		return toFloat(left) + toFloat(right), nil
	case BinSub:
		return toFloat(left) - toFloat(right), nil
	case BinMul:
		return toFloat(left) * toFloat(right), nil
	case BinDiv:
		d := toFloat(right)
		if d == 0 {
			return 0.0, nil
		}
		return toFloat(left) / d, nil
	case BinMod:
		d := int64(toFloat(right))
		if d == 0 {
			return 0.0, nil
		}
		return float64(int64(toFloat(left)) % d), nil
	}
	return nil, fmt.Errorf("unknown binary op %d", n.Op)
}

func evalUnary(n ExprUnary, ctx *EvalContext) (any, error) {
	operand, err := EvalExpression(n.Operand, ctx)
	if err != nil {
		return nil, err
	}
	switch n.Op {
	case UnaryNot:
		return !toBool(operand), nil
	case UnaryNeg:
		return -toFloat(operand), nil
	}
	return nil, fmt.Errorf("unknown unary op %d", n.Op)
}

func evalTernary(n ExprTernary, ctx *EvalContext) (any, error) {
	cond, err := EvalExpression(n.Cond, ctx)
	if err != nil {
		return nil, err
	}
	if toBool(cond) {
		return EvalExpression(n.Then, ctx)
	}
	return EvalExpression(n.Else, ctx)
}

func evalConcat(n ExprConcat, ctx *EvalContext) (any, error) {
	var buf strings.Builder
	for _, part := range n.Parts {
		v, err := EvalExpression(part, ctx)
		if err != nil {
			return nil, err
		}
		buf.WriteString(fmt.Sprint(v))
	}
	return buf.String(), nil
}

// ToBool is the exported equivalent of toBool, for use by the root package tests.
func ToBool(v any) bool { return toBool(v) }

// toBool converts any value to a boolean.
func toBool(v any) bool {
	if v == nil {
		return false
	}
	switch val := v.(type) {
	case bool:
		return val
	case float64:
		return val != 0
	case int:
		return val != 0
	case string:
		return val != ""
	}
	return true
}

// toFloat converts any value to float64.
func toFloat(v any) float64 {
	if v == nil {
		return 0
	}
	switch val := v.(type) {
	case float64:
		return val
	case int:
		return float64(val)
	case bool:
		if val {
			return 1
		}
		return 0
	case string:
		f, _ := strconv.ParseFloat(val, 64)
		return f
	}
	return 0
}
