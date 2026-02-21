package widget

// Spacer is an invisible fixed-size gap for use in VBox/HBox layouts.
// It occupies space without rendering anything.
type Spacer struct {
	Component
}

// NewSpacer creates a Spacer with the given dimensions.
func NewSpacer(name string, w, h float64) *Spacer {
	s := &Spacer{}
	initComponent(&s.Component, name)
	s.Width = w
	s.Height = h
	s.MarkLayoutDirty()
	return s
}

// SetSize updates the spacer's dimensions.
func (s *Spacer) SetSize(w, h float64) {
	s.Width = w
	s.Height = h
	s.MarkLayoutDirty()
}
