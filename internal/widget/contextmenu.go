package widget

import "github.com/devthicket/willowui/internal/sg"

// ContextMenu is a list of items that can be shown at an arbitrary position,
// typically on right-click. Attach it to a Component via SetContextMenu.
type ContextMenu struct {
	items       []MenuItem
	source      *sg.FontFamily
	font        *sg.FontFamily
	displaySize float64
	variant     Variant
	popup       *MenuPopup
}

// NewContextMenu creates a ContextMenu with the given font source and display size.
func NewContextMenu(source *sg.FontFamily, displaySize float64) *ContextMenu {
	var font *sg.FontFamily
	if source != nil {
		font = source
	}
	return &ContextMenu{
		source:      source,
		font:        font,
		displaySize: displaySize,
	}
}

// SetItems replaces the item list.
func (c *ContextMenu) SetItems(items []MenuItem) {
	c.items = items
}

// SetVariant sets the theme variant for the popup.
func (c *ContextMenu) SetVariant(v Variant) {
	c.variant = v
}

// ShowAt opens the menu with its top-left at (x, y) in world/screen coordinates.
func (c *ContextMenu) ShowAt(x, y float64) {
	if c.popup == nil {
		c.popup = NewMenuPopup(c.source, c.displaySize)
		c.popup.SetVariant(c.variant)
	}
	c.popup.SetItems(c.items)
	DefaultMenuPopupManager.ShowAt(c.popup, x, y)
}

// Hide closes the context menu if it is currently open.
func (c *ContextMenu) Hide() {
	DefaultMenuPopupManager.Hide()
}
