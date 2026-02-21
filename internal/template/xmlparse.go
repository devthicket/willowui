package template

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"strings"
)

// xmlContentNode represents either a text run or a child element in mixed content.
type xmlContentNode struct {
	Text    string
	Element *xmlElement
}

// xmlElement is a parsed XML element with its attributes and children.
type xmlElement struct {
	Name         string
	Attrs        []xmlAttr
	Children     []*xmlElement
	Text         string
	MixedContent []xmlContentNode // preserves interleaved text+element ordering
}

// xmlAttr is a parsed XML attribute with an optional namespace prefix.
type xmlAttr struct {
	Prefix string // "bind", "on", "ui", or "" for plain attributes
	Name   string
	Value  string
}

// parseXML parses XML data into an element tree using raw token mode.
// Attribute names are split on ':' to extract prefixes (bind:text, on:click, ui:if).
func parseXML(data []byte) (*xmlElement, error) {
	dec := xml.NewDecoder(bytes.NewReader(data))
	dec.Strict = false

	var stack []*xmlElement
	var root *xmlElement

	for {
		tok, err := dec.Token()
		if err != nil {
			if root != nil && len(stack) == 0 {
				break
			}
			return nil, fmt.Errorf("xml parse error: %w", err)
		}

		switch t := tok.(type) {
		case xml.StartElement:
			elem := &xmlElement{Name: t.Name.Local}
			for _, a := range t.Attr {
				attr := splitAttr(a)
				elem.Attrs = append(elem.Attrs, attr)
			}
			if len(stack) > 0 {
				parent := stack[len(stack)-1]
				parent.Children = append(parent.Children, elem)
				parent.MixedContent = append(parent.MixedContent, xmlContentNode{Element: elem})
			} else {
				root = elem
			}
			stack = append(stack, elem)

		case xml.EndElement:
			if len(stack) == 0 {
				return nil, fmt.Errorf("unexpected closing tag </%s>", t.Name.Local)
			}
			stack = stack[:len(stack)-1]

		case xml.CharData:
			text := strings.TrimSpace(string(t))
			if text != "" && len(stack) > 0 {
				current := stack[len(stack)-1]
				if current.Text == "" {
					current.Text = text
				} else {
					current.Text += " " + text
				}
				current.MixedContent = append(current.MixedContent, xmlContentNode{Text: text})
			}
		}
	}

	if root == nil {
		return nil, fmt.Errorf("empty XML document")
	}
	if len(stack) != 0 {
		return nil, fmt.Errorf("unclosed element <%s>", stack[len(stack)-1].Name)
	}

	return root, nil
}

// splitAttr splits an xml.Attr name on ':' to extract the prefix.
func splitAttr(a xml.Attr) xmlAttr {
	name := a.Name.Local
	// If the attribute has an XML namespace space, use it as prefix
	if a.Name.Space != "" {
		return xmlAttr{Prefix: a.Name.Space, Name: name, Value: a.Value}
	}
	// Otherwise split on ':'
	if idx := strings.IndexByte(name, ':'); idx > 0 {
		return xmlAttr{
			Prefix: name[:idx],
			Name:   name[idx+1:],
			Value:  a.Value,
		}
	}
	return xmlAttr{Name: name, Value: a.Value}
}
