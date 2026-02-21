package markup

import "github.com/devthicket/willowui/internal/sg"

// Outline defines a text stroke rendered behind the fill.
type Outline struct {
	Color     sg.Color
	Thickness float64
}

// TextSpan represents a styled segment of text within a RichText component.
// Fields left at their zero values inherit from the parent RichText.
type TextSpan struct {
	Text          string
	Source        *sg.FontFamily // nil = inherit from RichText
	Bold          bool
	Italic        bool
	Color         sg.Color // used when ColorSet == true
	ColorSet      bool     // false = inherit
	Outline       *Outline // nil = inherit; &Outline{} = explicitly none
	Underline     bool
	Strikethrough bool
	SizeOverride  float64 // 0 = inherit displaySize
	LinkURL       string  // non-empty = clickable
	Indent        float64 // left indent for this line and wrapped continuations
}
