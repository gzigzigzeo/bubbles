package menu

import "charm.land/lipgloss/v2"

// Styles holds the visual styles for a Menu's rows and leftmost gutter.
type Styles struct {
	ScrollUp      lipgloss.Style
	ScrollDown    lipgloss.Style
	CursorFocused lipgloss.Style
	CursorBlurred lipgloss.Style
	CursorMarked  lipgloss.Style
	LabelFocused  lipgloss.Style
	LabelBlurred  lipgloss.Style
	LabelMarked   lipgloss.Style
	Description   lipgloss.Style
}

// DefaultStyles returns a minimal, functional set of styles: the same glyph
// conventions used elsewhere in this UI kit (▶ cursor, ▲/▼ scroll arrows, ✓
// marker), no color. Callers embedded in a real layout will typically
// override Row (indentation/width) and the colors to match their own theme.
func DefaultStyles() Styles {
	return Styles{
		ScrollUp:      lipgloss.NewStyle().SetString("▲ "),
		ScrollDown:    lipgloss.NewStyle().SetString("▼ "),
		CursorFocused: lipgloss.NewStyle().SetString("▶ "),
		CursorBlurred: lipgloss.NewStyle().SetString("  "),
		CursorMarked:  lipgloss.NewStyle().SetString("✓ "),
	}
}
