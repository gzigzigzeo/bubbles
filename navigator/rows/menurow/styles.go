package menurow

import (
	"charm.land/lipgloss/v2"

	"github.com/gzigzigzeo/bubbles/navigator/rows/row"
)

// RowStyles holds the visual pieces of a single menu row.
type RowStyles struct {
	Name        lipgloss.Style
	Cursor      lipgloss.Style
	Mark        lipgloss.Style
	Description lipgloss.Style
}

// Styles is the per-state style set for a menu row.
type Styles = row.StateSet[RowStyles]

// DefaultStyles returns a minimal, functional style set: ▶ cursor, ✓ mark, and
// same-width blanks for non-focused states so rows stay aligned.
func DefaultStyles() Styles {
	cursor := lipgloss.NewStyle().SetString("▶ ")
	blank := lipgloss.NewStyle().SetString("  ")
	mark := lipgloss.NewStyle().SetString("✓ ")

	return Styles{
		Focused: RowStyles{
			Name:        lipgloss.NewStyle(),
			Cursor:      cursor,
			Mark:        mark,
			Description: lipgloss.NewStyle(),
		},
		Blurred: RowStyles{
			Name:        lipgloss.NewStyle(),
			Cursor:      blank,
			Mark:        mark,
			Description: lipgloss.NewStyle(),
		},
		Disabled: RowStyles{
			Name:        lipgloss.NewStyle().Faint(true),
			Cursor:      blank,
			Mark:        mark,
			Description: lipgloss.NewStyle().Faint(true),
		},
	}
}
