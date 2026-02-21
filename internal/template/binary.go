package template

import (
	"errors"
	"fmt"

	"github.com/devthicket/willowui/internal/binutil"
)

// Binary format constants for .xmlui files.
var xmluiMagic = [4]byte{'X', 'U', 'I', 0x00}

const xmluiVersion uint16 = 1

// ExprNode type tags for binary encoding.
const (
	exprTagNil     uint8 = 0
	exprTagRef     uint8 = 1
	exprTagLiteral uint8 = 2
	exprTagBinary  uint8 = 3
	exprTagUnary   uint8 = 4
	exprTagTernary uint8 = 5
	exprTagConcat  uint8 = 6
)

// ExprLiteral value type tags.
const (
	litTagNil       uint8 = 0
	litTagFloat64   uint8 = 1
	litTagString    uint8 = 2
	litTagBoolTrue  uint8 = 3
	litTagBoolFalse uint8 = 4
)

// EncodeIR serializes an IRNode tree to the .xmlui binary format.
func EncodeIR(node *IRNode) ([]byte, error) {
	if node == nil {
		return nil, errors.New("cannot encode nil IRNode")
	}
	w := &binWriter{}
	w.WriteBytes(xmluiMagic[:])
	w.WriteU16(xmluiVersion)
	w.writeIRNode(node)
	if w.Err != nil {
		return nil, w.Err
	}
	return w.Buf, nil
}

// DecodeIR deserializes an IRNode tree from the .xmlui binary format.
func DecodeIR(data []byte) (*IRNode, error) {
	if len(data) < 6 {
		return nil, errors.New("xmlui: data too short")
	}
	r := &binReader{Reader: binutil.Reader{Data: data}}
	// Validate magic bytes.
	magic := r.ReadBytes(4)
	if r.Err != nil {
		return nil, r.Err
	}
	if magic[0] != xmluiMagic[0] || magic[1] != xmluiMagic[1] ||
		magic[2] != xmluiMagic[2] || magic[3] != xmluiMagic[3] {
		return nil, errors.New("xmlui: invalid magic bytes")
	}
	// Validate version.
	ver := r.ReadU16()
	if r.Err != nil {
		return nil, r.Err
	}
	if ver != xmluiVersion {
		return nil, fmt.Errorf("xmlui: unsupported version %d", ver)
	}
	node := r.readIRNode()
	if r.Err != nil {
		return nil, r.Err
	}
	return node, nil
}

// --- Writer ---

type binWriter struct {
	binutil.Writer
}

func (w *binWriter) writeExprNode(node ExprNode) {
	if w.Err != nil {
		return
	}
	if node == nil {
		w.WriteU8(exprTagNil)
		return
	}
	switch n := node.(type) {
	case ExprRef:
		w.WriteU8(exprTagRef)
		w.WriteString(n.Path)
	case ExprLiteral:
		w.WriteU8(exprTagLiteral)
		w.writeLiteralValue(n.Value)
	case ExprBinary:
		w.WriteU8(exprTagBinary)
		w.WriteU8(uint8(n.Op))
		w.writeExprNode(n.Left)
		w.writeExprNode(n.Right)
	case ExprUnary:
		w.WriteU8(exprTagUnary)
		w.WriteU8(uint8(n.Op))
		w.writeExprNode(n.Operand)
	case ExprTernary:
		w.WriteU8(exprTagTernary)
		w.writeExprNode(n.Cond)
		w.writeExprNode(n.Then)
		w.writeExprNode(n.Else)
	case ExprConcat:
		w.WriteU8(exprTagConcat)
		w.WriteU32(uint32(len(n.Parts)))
		for _, part := range n.Parts {
			w.writeExprNode(part)
		}
	default:
		w.Err = fmt.Errorf("xmlui: unknown ExprNode type %T", node)
	}
}

func (w *binWriter) writeLiteralValue(val any) {
	if w.Err != nil {
		return
	}
	switch v := val.(type) {
	case nil:
		w.WriteU8(litTagNil)
	case float64:
		w.WriteU8(litTagFloat64)
		w.WriteFloat64(v)
	case string:
		w.WriteU8(litTagString)
		w.WriteString(v)
	case bool:
		if v {
			w.WriteU8(litTagBoolTrue)
		} else {
			w.WriteU8(litTagBoolFalse)
		}
	default:
		w.Err = fmt.Errorf("xmlui: unsupported literal type %T", val)
	}
}

func (w *binWriter) writeIRNode(node *IRNode) {
	if w.Err != nil {
		return
	}
	w.WriteString(node.ComponentType)
	w.WriteString(node.Text)

	// Attributes
	w.WriteU32(uint32(len(node.Attributes)))
	for _, attr := range node.Attributes {
		w.WriteString(attr.Name)
		w.WriteString(attr.Static)
		w.writeExprNode(attr.Expr)
		w.WriteBool(attr.IsEvent)
	}

	// Directives
	w.WriteU32(uint32(len(node.Directives)))
	for _, dir := range node.Directives {
		w.WriteU8(uint8(dir.Type))
		w.writeExprNode(dir.Expr)
		w.WriteString(dir.VarName)
	}

	// Children (depth-first)
	w.WriteU32(uint32(len(node.Children)))
	for _, child := range node.Children {
		w.writeIRNode(child)
	}

	// ThemePatch (length-prefixed bytes; 0 means none)
	w.WriteU32(uint32(len(node.ThemePatch)))
	w.WriteBytes(node.ThemePatch)
}

// --- Reader ---

type binReader struct {
	binutil.Reader
}

func (r *binReader) readExprNode() ExprNode {
	if r.Err != nil {
		return nil
	}
	tag := r.ReadU8()
	if r.Err != nil {
		return nil
	}
	switch tag {
	case exprTagNil:
		return nil
	case exprTagRef:
		path := r.ReadString()
		return ExprRef{Path: path}
	case exprTagLiteral:
		val := r.readLiteralValue()
		return ExprLiteral{Value: val}
	case exprTagBinary:
		op := BinOp(r.ReadU8())
		left := r.readExprNode()
		right := r.readExprNode()
		return ExprBinary{Op: op, Left: left, Right: right}
	case exprTagUnary:
		op := UnaryOp(r.ReadU8())
		operand := r.readExprNode()
		return ExprUnary{Op: op, Operand: operand}
	case exprTagTernary:
		cond := r.readExprNode()
		then := r.readExprNode()
		els := r.readExprNode()
		return ExprTernary{Cond: cond, Then: then, Else: els}
	case exprTagConcat:
		count := r.ReadU32()
		parts := make([]ExprNode, count)
		for i := range parts {
			parts[i] = r.readExprNode()
		}
		return ExprConcat{Parts: parts}
	default:
		r.Err = fmt.Errorf("xmlui: unknown expr tag %d", tag)
		return nil
	}
}

func (r *binReader) readLiteralValue() any {
	if r.Err != nil {
		return nil
	}
	tag := r.ReadU8()
	switch tag {
	case litTagNil:
		return nil
	case litTagFloat64:
		return r.ReadFloat64()
	case litTagString:
		return r.ReadString()
	case litTagBoolTrue:
		return true
	case litTagBoolFalse:
		return false
	default:
		r.Err = fmt.Errorf("xmlui: unknown literal tag %d", tag)
		return nil
	}
}

func (r *binReader) readIRNode() *IRNode {
	if r.Err != nil {
		return nil
	}
	node := &IRNode{}
	node.ComponentType = r.ReadString()
	node.Text = r.ReadString()

	// Attributes
	attrCount := r.ReadU32()
	if r.Err != nil {
		return nil
	}
	node.Attributes = make([]IRAttribute, attrCount)
	for i := range node.Attributes {
		node.Attributes[i].Name = r.ReadString()
		node.Attributes[i].Static = r.ReadString()
		node.Attributes[i].Expr = r.readExprNode()
		node.Attributes[i].IsEvent = r.ReadBool()
	}

	// Directives
	dirCount := r.ReadU32()
	if r.Err != nil {
		return nil
	}
	node.Directives = make([]IRDirective, dirCount)
	for i := range node.Directives {
		node.Directives[i].Type = DirectiveType(r.ReadU8())
		node.Directives[i].Expr = r.readExprNode()
		node.Directives[i].VarName = r.ReadString()
	}

	// Children
	childCount := r.ReadU32()
	if r.Err != nil {
		return nil
	}
	node.Children = make([]*IRNode, childCount)
	for i := range node.Children {
		node.Children[i] = r.readIRNode()
	}

	// ThemePatch
	patchLen := r.ReadU32()
	if r.Err != nil {
		return nil
	}
	if patchLen > 0 {
		node.ThemePatch = r.ReadBytes(int(patchLen))
	}

	if r.Err != nil {
		return nil
	}
	return node
}
