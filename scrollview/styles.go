package scrollview

import "charm.land/lipgloss/v2"

// Styles holds the lipgloss styles applied to the scrollbar column.
type Styles struct {
	// Track is applied to the non-thumb cells of the scrollbar column.
	// The character to render is taken from the style's embedded string
	// (set via lipgloss.Style.SetString); the default is " " (space).
	Track lipgloss.Style
	// Thumb is applied to the thumb cells of the scrollbar column.
	// The character to render is taken from the style's embedded string
	// (set via lipgloss.Style.SetString); the default is "▒".
	Thumb lipgloss.Style
	// MoreAbove is rendered at the top cell of the scrollbar column when
	// there are rows above the current viewport (YOffset > 0).
	MoreAbove lipgloss.Style
	// NoMoreAbove is rendered at the top cell of the scrollbar column when
	// the viewport is already scrolled to the very top (YOffset == 0).
	NoMoreAbove lipgloss.Style
	// MoreBelow is rendered at the bottom cell of the scrollbar column when
	// there are rows below the current viewport.
	MoreBelow lipgloss.Style
	// NoMoreBelow is rendered at the bottom cell of the scrollbar column when
	// the viewport is already scrolled to the very bottom.
	NoMoreBelow lipgloss.Style
	// HiddenBar is rendered in the scrollbar column on every row when the
	// content fits entirely within the viewport (no scrollbar is needed).
	// It should visually match the width of the active scrollbar column so
	// that the content area never shifts. The default is a single space.
	HiddenBar lipgloss.Style
}

// DefaultStyles returns a sensible default Styles for a new Model.
func DefaultStyles() Styles {
	gray := lipgloss.NewStyle().Foreground(lipgloss.Color("#585858"))
	white := lipgloss.NewStyle().Foreground(lipgloss.Color("#ffffff"))

	return Styles{
		Track:       gray.SetString(" "),
		Thumb:       lipgloss.NewStyle().SetString("▒").Foreground(lipgloss.Color("#c0c0c0")),
		MoreAbove:   white.SetString("▲"),
		NoMoreAbove: gray.SetString("▲"),
		MoreBelow:   white.SetString("▼"),
		NoMoreBelow: gray.SetString("▼"),
		HiddenBar:   lipgloss.NewStyle().SetString(" "),
	}
}
