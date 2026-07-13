package textinputfield

import (
	"charm.land/bubbles/v2/textinput"
	"charm.land/lipgloss/v2"

	"github.com/gzigzigzeo/bubbles/form/field"
)

// Styles defines the per-state (focused/blurred/disabled) styles for a
// textinputfield.State and the textfield.Model/numberfield.Model that embed it.
type Styles = field.EntryStyles[textinput.Styles]

// DefaultStyles returns usable Styles with no external theme dependency.
func DefaultStyles() Styles {
	base := textinput.DefaultStyles(true)
	pale := lipgloss.NewStyle().Foreground(lipgloss.Color("#5C5C5C"))

	disabled := base
	disabled.Focused.Text = pale
	disabled.Blurred.Text = pale

	return Styles{
		Focused:  base,
		Blurred:  base,
		Disabled: disabled,
	}
}
