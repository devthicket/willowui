package widget

// AnchorLayout is a Panel whose Layout is pre-set to LayoutAnchor.
// It is a convenience wrapper around Panel + LayoutAnchor; all anchor
// functionality lives on Component.
//
// Prefer using a plain Panel (or Component) with Layout = LayoutAnchor and
// Component.AddAnchoredChild for new code. AnchorLayout is kept for
// compatibility with existing code and XML templates.
type AnchorLayout struct {
	Panel
}

// NewAnchorLayout creates an AnchorLayout container.
func NewAnchorLayout(name string) *AnchorLayout {
	al := &AnchorLayout{}
	initComponent(&al.Component, name)
	al.initBackground(name)
	al.initBorder(name)

	// Wire theme — reuse PanelGroup.
	al.onThemeChange = func() { al.applyThemeColors() }
	al.applyThemeColors()

	al.Layout = LayoutAnchor
	return al
}
