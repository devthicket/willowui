package template

import (
	"encoding/json"
	"fmt"
	"strings"
)

// knownComponents is the set of valid component type names in XML templates.
var knownComponents = map[string]bool{
	"Panel":            true,
	"Label":            true,
	"Button":           true,
	"IconButton":       true,
	"Toggle":           true,
	"Checkbox":         true,
	"TextInput":        true,
	"TextArea":         true,
	"Slider":           true,
	"ScrollBar":        true,
	"ProgressBar":      true,
	"MeterBar":         true,
	"List":             true,
	"TileList":         true,
	"TreeList":         true,
	"TabBar":           true,
"ScrollPanel":      true,
	"Window":           true,
	"RichText":         true,
	"Radio":            true,
	"Component":        true,
	"NumberStepper":    true,
	"ToggleButtonBar":  true,
	"AnchorLayout":     true,
	"TwoColumnLayout":  true,
	"Tab":              true,
	"Spacer":           true,
	"Select":           true,
	"OptionRotator":    true,
	"RadioButton":      true, // pseudo-element: only valid as a child of Radio
	"DragHandle":       true,
	"InputField":       true,
	"SearchBox":        true,
	"Badge":            true,
	"StatWeb":          true,
	"NavDrawer":        true,
	"ImageCropper":     true,
	"ColorPicker":      true,
	"ToolBar":          true,
	"SortableTreeList": true,
	"Image":            true,
	"AnimatedImage":    true,
	"CalendarSelector": true,
	"TimePicker":       true,
	"KeybindInput":     true,
	"MaskedInput":      true,
	"SortableList":     true,
	"Tag":              true,
	"TagBar":           true,
	"Accordion":        true,
	"DataTable":        true,
	"Column":           true,
	"Section":          true,
	"Tooltip":          true,
	"Popover":          true,
	"GradientEditor":   true,
	"MenuBar":          true,
}

// CompileXML parses XML template data and compiles it to an IR tree.
func CompileXML(data []byte) (*IRNode, error) {
	return CompileXMLWithTypes(data, nil)
}

// CompileXMLWithTypes parses XML template data and compiles it to an IR tree,
// accepting extra custom component type names in addition to built-in types.
func CompileXMLWithTypes(data []byte, extraTypes map[string]bool) (*IRNode, error) {
	elem, err := parseXML(data)
	if err != nil {
		return nil, err
	}
	return compileElement(elem, extraTypes)
}

func compileElement(elem *xmlElement, extraTypes map[string]bool) (*IRNode, error) {
	if !knownComponents[elem.Name] && !extraTypes[elem.Name] {
		return nil, fmt.Errorf("unknown component type %q", elem.Name)
	}

	node := &IRNode{ComponentType: elem.Name}

	// Process attributes
	for _, attr := range elem.Attrs {
		switch attr.Prefix {
		case "bind":
			expr, err := ParseExpression(attr.Value)
			if err != nil {
				return nil, fmt.Errorf("bind:%s expression error: %w", attr.Name, err)
			}
			node.Attributes = append(node.Attributes, IRAttribute{
				Name: attr.Name,
				Expr: expr,
			})
		case "on":
			node.Attributes = append(node.Attributes, IRAttribute{
				Name:    attr.Name,
				Static:  attr.Value,
				IsEvent: true,
			})
		case "ui":
			dir, err := parseDirective(attr.Name, attr.Value)
			if err != nil {
				return nil, err
			}
			node.Directives = append(node.Directives, dir)
		default:
			// Plain static attribute
			node.Attributes = append(node.Attributes, IRAttribute{
				Name:   attr.Name,
				Static: attr.Value,
			})
		}
	}

	// Process text content with interpolation
	if elem.Text != "" {
		if expr, found := parseInterpolation(elem.Text); found {
			node.Text = elem.Text
			node.Attributes = append(node.Attributes, IRAttribute{
				Name: "text",
				Expr: expr,
			})
		} else {
			node.Text = elem.Text
		}
	}

	// Process children — for RichText, check if children are markup tags
	// rather than real components. If so, serialize the mixed content as
	// a "markup" attribute instead of compiling children as components.
	if elem.Name == "RichText" && hasMarkupChildren(elem) {
		markup := serializeMixedContent(elem.MixedContent)
		node.Attributes = append(node.Attributes, IRAttribute{
			Name:   "markup",
			Static: markup,
		})
	} else {
		for _, child := range elem.Children {
			// <Theme> pseudo-element: extract JSON text and store as ThemePatch.
			if child.Name == "Theme" {
				text := strings.TrimSpace(child.Text)
				if text == "" {
					return nil, fmt.Errorf("<Theme> element must contain JSON text")
				}
				// Validate that the text is valid JSON.
				if !isValidJSON([]byte(text)) {
					return nil, fmt.Errorf("<Theme> element contains invalid JSON")
				}
				node.ThemePatch = []byte(text)
				continue
			}
			childNode, err := compileElement(child, extraTypes)
			if err != nil {
				return nil, err
			}
			node.Children = append(node.Children, childNode)
		}
	}

	return node, nil
}

// richTextMarkupTags is the set of tags recognized as inline markup inside
// a <RichText> element (as opposed to child components).
var richTextMarkupTags = map[string]bool{
	"b": true, "i": true, "u": true, "strike": true,
	"color": true, "size": true, "outline": true, "span": true,
	"link": true, "br": true,
	"h1": true, "h2": true, "h3": true,
	"ul": true, "ol": true, "li": true,
}

// hasMarkupChildren returns true if the element has children and all of them
// are markup tags (not known component types).
func hasMarkupChildren(elem *xmlElement) bool {
	if len(elem.Children) == 0 && len(elem.MixedContent) == 0 {
		return false
	}
	for _, child := range elem.Children {
		if !richTextMarkupTags[child.Name] {
			return false
		}
	}
	return true
}

// serializeMixedContent converts MixedContent back into a markup string.
func serializeMixedContent(content []xmlContentNode) string {
	var sb strings.Builder
	for _, node := range content {
		if node.Element != nil {
			serializeElement(&sb, node.Element)
		} else {
			sb.WriteString(node.Text)
		}
	}
	return sb.String()
}

// serializeElement writes an XML element and its children back as markup.
func serializeElement(sb *strings.Builder, elem *xmlElement) {
	sb.WriteByte('<')
	sb.WriteString(elem.Name)
	for _, attr := range elem.Attrs {
		sb.WriteByte(' ')
		sb.WriteString(attr.Name)
		sb.WriteString(`="`)
		sb.WriteString(attr.Value)
		sb.WriteByte('"')
	}
	if elem.Name == "br" && len(elem.Children) == 0 && elem.Text == "" {
		sb.WriteString("/>")
		return
	}
	sb.WriteByte('>')
	if len(elem.MixedContent) > 0 {
		for _, node := range elem.MixedContent {
			if node.Element != nil {
				serializeElement(sb, node.Element)
			} else {
				sb.WriteString(node.Text)
			}
		}
	} else if elem.Text != "" {
		sb.WriteString(elem.Text)
	}
	sb.WriteString("</")
	sb.WriteString(elem.Name)
	sb.WriteByte('>')
}

// isValidJSON reports whether data is valid JSON.
func isValidJSON(data []byte) bool {
	var v any
	return json.Unmarshal(data, &v) == nil
}

func parseDirective(name, value string) (IRDirective, error) {
	switch name {
	case "if":
		expr, err := ParseExpression(value)
		if err != nil {
			return IRDirective{}, fmt.Errorf("ui:if expression error: %w", err)
		}
		return IRDirective{Type: DirectiveIf, Expr: expr}, nil
	case "else-if":
		expr, err := ParseExpression(value)
		if err != nil {
			return IRDirective{}, fmt.Errorf("ui:else-if expression error: %w", err)
		}
		return IRDirective{Type: DirectiveElseIf, Expr: expr}, nil
	case "else":
		return IRDirective{Type: DirectiveElse}, nil
	case "for":
		varName, expr, err := parseForDirective(value)
		if err != nil {
			return IRDirective{}, err
		}
		return IRDirective{Type: DirectiveFor, Expr: expr, VarName: varName}, nil
	case "key":
		expr, err := ParseExpression(value)
		if err != nil {
			return IRDirective{}, fmt.Errorf("ui:key expression error: %w", err)
		}
		return IRDirective{Type: DirectiveKey, Expr: expr}, nil
	case "show":
		expr, err := ParseExpression(value)
		if err != nil {
			return IRDirective{}, fmt.Errorf("ui:show expression error: %w", err)
		}
		return IRDirective{Type: DirectiveShow, Expr: expr}, nil
	case "ref":
		// ui:ref is handled during instantiation, not as a directive
		return IRDirective{}, fmt.Errorf("ui:ref should not be processed as a directive")
	default:
		return IRDirective{}, fmt.Errorf("unknown directive ui:%s", name)
	}
}

// parseForDirective parses "item in collection" syntax.
func parseForDirective(value string) (string, ExprNode, error) {
	parts := strings.SplitN(value, " in ", 2)
	if len(parts) != 2 {
		return "", nil, fmt.Errorf("ui:for expects 'variable in expression', got %q", value)
	}
	varName := strings.TrimSpace(parts[0])
	if varName == "" {
		return "", nil, fmt.Errorf("ui:for variable name is empty")
	}
	expr, err := ParseExpression(strings.TrimSpace(parts[1]))
	if err != nil {
		return "", nil, fmt.Errorf("ui:for expression error: %w", err)
	}
	return varName, expr, nil
}
