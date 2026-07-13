package field

import "charm.land/lipgloss/v2"

// Styles is the form's own chrome: row container, cursor indicator,
// field labels, error text, and hint text.
type Styles struct {
	Row           lipgloss.Style
	CursorFocused lipgloss.Style
	CursorBlurred lipgloss.Style
	LabelFocused  lipgloss.Style
	LabelBlurred  lipgloss.Style
	LabelDisabled lipgloss.Style
	Gutter        lipgloss.Style
	ErrText       lipgloss.Style
	HintBlock     lipgloss.Style
	HintText      lipgloss.Style
	HintLink      lipgloss.Style
}

// DefaultStyles returns a usable Styles with no external theme dependency.
func DefaultStyles() Styles {
	labelFocused := lipgloss.NewStyle().Foreground(lipgloss.Color("#53d1ff"))
	labelBlurred := lipgloss.NewStyle().Foreground(lipgloss.Color("#d3d3d9"))
	labelDisabled := lipgloss.NewStyle().Foreground(lipgloss.Color("#5C5C5C"))
	pale := lipgloss.NewStyle().Foreground(lipgloss.Color("#a09caa"))

	return Styles{
		Row:           lipgloss.NewStyle(),
		CursorFocused: lipgloss.NewStyle().Foreground(lipgloss.Color("#53d1ff")).SetString("▶ "),
		CursorBlurred: lipgloss.NewStyle().SetString("  "),
		LabelFocused:  labelFocused,
		LabelBlurred:  labelBlurred,
		LabelDisabled: labelDisabled,
		Gutter:        lipgloss.NewStyle().Width(2),
		ErrText:       lipgloss.NewStyle().Foreground(lipgloss.Color("#ffaba0")),
		HintBlock:     lipgloss.NewStyle().MarginBottom(1),
		HintText:      pale,
		HintLink:      pale.Underline(true),
	}
}
