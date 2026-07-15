package textinput

import (
	"charm.land/bubbles/v2/textinput"
	"charm.land/lipgloss/v2"

	"github.com/gzigzigzeo/bubbles/navigator/row"
)

// TextInputStyles defines the visual pieces of a text input row.
type TextInputStyles struct {
	Label lipgloss.Style
	Input textinput.Styles
	Error lipgloss.Style
}

// Styles is the per-state style set for a text input row.
type Styles = row.StateSet[TextInputStyles]

// DefaultStyles returns usable Styles with no external theme dependency.
func DefaultStyles() Styles {
	base := textinput.DefaultStyles(true)
	pale := lipgloss.NewStyle().Foreground(row.ColorPale)

	disabled := base
	disabled.Focused.Text = pale
	disabled.Blurred.Text = pale

	return Styles{
		Focused: TextInputStyles{
			Label: row.DefaultLabelStyle().Bold(true),
			Input: base,
			Error: lipgloss.NewStyle().Bold(true).Foreground(row.ColorError).MarginRight(1),
		},
		Blurred: TextInputStyles{
			Label: row.DefaultLabelStyle(),
			Input: base,
			Error: lipgloss.NewStyle().Bold(true).Foreground(row.ColorError).MarginRight(1),
		},
		Disabled: TextInputStyles{
			Label: row.DefaultLabelStyle().Faint(true),
			Input: disabled,
			Error: lipgloss.NewStyle().Bold(true).Foreground(row.ColorError).MarginRight(1),
		},
	}
}
