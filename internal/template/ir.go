package template

// DirectiveType identifies a structural directive in a compiled template.
type DirectiveType int

const (
	DirectiveIf     DirectiveType = iota // ui:if="expr"
	DirectiveElseIf                      // ui:else-if="expr"
	DirectiveElse                        // ui:else
	DirectiveFor                         // ui:for="item in list"
	DirectiveKey                         // ui:key="expr"
	DirectiveShow                        // ui:show="expr"
)

// IRNode is the intermediate representation of a compiled XML template element.
type IRNode struct {
	ComponentType string
	Attributes    []IRAttribute
	Children      []*IRNode
	Directives    []IRDirective
	Text          string // interpolated text content
	ThemePatch    []byte // raw JSON from a <Theme> child element (root node only)
}

// IRAttribute represents a single attribute on an IR node.
type IRAttribute struct {
	Name    string   // attribute name (e.g. "text", "size")
	Static  string   // static value (empty when Expr is set)
	Expr    ExprNode // parsed expression for bind: attributes
	IsEvent bool     // true for on:click etc.
}

// IRDirective represents a structural directive attached to an IR node.
type IRDirective struct {
	Type    DirectiveType
	Expr    ExprNode // condition or collection expression
	VarName string   // loop variable name (for ui:for)
}
