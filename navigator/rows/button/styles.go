package button

import (
	"charm.land/lipgloss/v2"

	"github.com/gzigzigzeo/bubbles/navigator/row"
)

// Styles is the per-state style set for a button.
type Styles = row.StateSet[lipgloss.Style]

// DefaultStyles returns a minimal, functional style set: reversed foreground and
// background for the focused state, plain for the blurred state, and faint for
// the disabled state. Each state has a right margin so stacked buttons are
// visually separated.
func DefaultStyles() Styles {
	return Styles{
		Focused:  lipgloss.NewStyle().Reverse(true).PaddingLeft(1).PaddingRight(1).MarginRight(1),
		Blurred:  lipgloss.NewStyle().PaddingLeft(1).PaddingRight(1).MarginRight(1),
		Disabled: lipgloss.NewStyle().Faint(true).PaddingLeft(1).PaddingRight(1).MarginRight(1),
	}
}
