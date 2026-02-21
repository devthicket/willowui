package markup

import (
	"encoding/xml"
	"fmt"
	"strconv"
	"strings"

	"github.com/devthicket/willowui/internal/sg"
)

// markupState tracks cumulative style at a point in the parse.
type markupState struct {
	bold          bool
	italic        bool
	underline     bool
	strikethrough bool
	color         sg.Color
	colorSet      bool
	outline       *Outline
	sizeOverride  float64
	linkURL       string
}

// listIndent is the left margin for list items (in pixels).
const listIndent = 12.0

// ParseMarkup parses XML-like markup into TextSpan slices.
// Supported tags: b, i, u, strike, color, size, outline, span, link, br,
// h1, h2, h3, ul, ol, li.
// The source parameter is accepted but not used during parsing; font resolution
// is deferred to render time via the RichText source.
func ParseMarkup(input string, source *sg.FontFamily, baseSize float64, headingScale [3]float64) ([]TextSpan, error) {
	// Wrap in a root element for valid XML.
	wrapped := "<_root>" + input + "</_root>"
	dec := xml.NewDecoder(strings.NewReader(wrapped))
	dec.Strict = false

	var spans []TextSpan
	var stack []markupState
	cur := markupState{}

	// List tracking.
	type listCtx struct {
		ordered bool
		index   int
	}
	var listStack []listCtx
	listDepth := 0 // nesting depth for whitespace suppression

	flush := func(text string) {
		if text == "" {
			return
		}
		spans = append(spans, TextSpan{
			Text:          text,
			Source:        nil, // inherit from RichText source
			Bold:          cur.bold,
			Italic:        cur.italic,
			Color:         cur.color,
			ColorSet:      cur.colorSet,
			Outline:       cur.outline,
			Underline:     cur.underline,
			Strikethrough: cur.strikethrough,
			SizeOverride:  cur.sizeOverride,
			LinkURL:       cur.linkURL,
		})
	}

	push := func() {
		stack = append(stack, cur)
	}

	pop := func() {
		if len(stack) > 0 {
			cur = stack[len(stack)-1]
			stack = stack[:len(stack)-1]
		}
	}

	for {
		tok, err := dec.Token()
		if err != nil {
			break
		}

		switch t := tok.(type) {
		case xml.StartElement:
			name := t.Name.Local
			switch name {
			case "_root":
				continue
			case "b":
				push()
				cur.bold = true
			case "i":
				push()
				cur.italic = true
			case "u":
				push()
				cur.underline = true
			case "strike":
				push()
				cur.strikethrough = true
			case "color":
				push()
				val := attrVal(t.Attr, "value")
				if c, cErr := ParseColor(val); cErr == nil {
					cur.color = c
					cur.colorSet = true
				}
			case "size":
				push()
				val := attrVal(t.Attr, "value")
				if s, sErr := strconv.ParseFloat(val, 64); sErr == nil && s > 0 {
					cur.sizeOverride = s
				}
			case "outline":
				push()
				th := 2.0
				if v := attrVal(t.Attr, "thickness"); v != "" {
					if parsed, pErr := strconv.ParseFloat(v, 64); pErr == nil {
						th = parsed
					}
				}
				oc := sg.RGBA(0, 0, 0, 1)
				if v := attrVal(t.Attr, "color"); v != "" {
					if c, cErr := ParseColor(v); cErr == nil {
						oc = c
					}
				}
				cur.outline = &Outline{Color: oc, Thickness: th}
			case "span":
				push()
				if v := attrVal(t.Attr, "color"); v != "" {
					if c, cErr := ParseColor(v); cErr == nil {
						cur.color = c
						cur.colorSet = true
					}
				}
				if v := attrVal(t.Attr, "size"); v != "" {
					if s, sErr := strconv.ParseFloat(v, 64); sErr == nil && s > 0 {
						cur.sizeOverride = s
					}
				}
				if v := attrVal(t.Attr, "font"); v != "" {
					switch v {
					case "bold":
						cur.bold = true
					case "italic":
						cur.italic = true
					case "bolditalic":
						cur.bold = true
						cur.italic = true
					case "regular":
						cur.bold = false
						cur.italic = false
					}
				}
			case "link":
				push()
				cur.linkURL = attrVal(t.Attr, "url")
				cur.underline = true
			case "br":
				flush("\n")
			case "h1", "h2", "h3":
				push()
				cur.bold = true
				idx := int(name[1] - '1') // 0, 1, 2
				if idx >= 0 && idx < 3 {
					cur.sizeOverride = baseSize * headingScale[idx]
				}
				flush("\n")
			case "ul":
				listStack = append(listStack, listCtx{ordered: false})
				listDepth++
			case "ol":
				listStack = append(listStack, listCtx{ordered: true})
				listDepth++
			case "li":
				if len(listStack) > 0 {
					lc := &listStack[len(listStack)-1]
					// Emit newline between items, or before the first
					// item only if the previous span doesn't already
					// end with a newline.
					if lc.index > 0 {
						flush("\n")
					} else if len(spans) > 0 && !endsWithNewline(spans) {
						flush("\n")
					}
					lc.index++

					// Build the bullet/number prefix.
					var prefix string
					if lc.ordered {
						prefix = fmt.Sprintf("%d. ", lc.index)
					} else {
						prefix = "\u2022 "
					}

					// Emit the prefix with Indent to position the bullet
					// and align wrapped continuation lines with the text
					// after the bullet.
					indent := listIndent * float64(listDepth)
					spans = append(spans, TextSpan{
						Text:     prefix,
						Bold:     cur.bold,
						Italic:   cur.italic,
						Color:    cur.color,
						ColorSet: cur.colorSet,
						Indent:   indent,
					})
				}
			default:
				return nil, fmt.Errorf("unknown markup tag <%s>", name)
			}

		case xml.EndElement:
			name := t.Name.Local
			switch name {
			case "_root":
				continue
			case "b", "i", "u", "strike", "color", "size", "outline", "span", "link":
				pop()
			case "h1", "h2", "h3":
				flush("\n")
				pop()
			case "ul", "ol":
				if len(listStack) > 0 {
					listStack = listStack[:len(listStack)-1]
					listDepth--
				}
			case "li":
				// no-op, content already flushed
			case "br":
				// self-closing, no end action
			}

		case xml.CharData:
			text := string(t)
			// Inside lists, suppress whitespace-only text between tags
			// (e.g. newlines between </li> and <li>).
			if listDepth > 0 && strings.TrimSpace(text) == "" {
				continue
			}
			flush(text)
		}
	}

	return spans, nil
}

// endsWithNewline returns true if the last span's text ends with a newline.
func endsWithNewline(spans []TextSpan) bool {
	if len(spans) == 0 {
		return false
	}
	last := spans[len(spans)-1].Text
	return len(last) > 0 && last[len(last)-1] == '\n'
}

// attrVal finds an attribute value by local name.
func attrVal(attrs []xml.Attr, name string) string {
	for _, a := range attrs {
		if a.Name.Local == name {
			return a.Value
		}
	}
	return ""
}
