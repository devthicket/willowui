package widget

import "github.com/devthicket/willowui/internal/markup"

// TextSpan represents a styled segment of text within a RichText component.
// Fields left at their zero values inherit from the parent RichText.
type TextSpan = markup.TextSpan

// Outline defines a text stroke rendered behind the fill.
type Outline = markup.Outline

var (
	// ParseMarkup parses XML-like markup into TextSpan slices.
	ParseMarkup = markup.ParseMarkup

	// ParseColor parses a color string in any supported format.
	ParseColor = markup.ParseColor
)
