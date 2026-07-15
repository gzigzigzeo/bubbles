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
	pale := lipgloss.NewStyle().Foreground(lipgloss.Color("#5C5C5C"))

	disabled := base
	disabled.Focused.Text = pale
	disabled.Blurred.Text = pale

	return Styles{
		Focused: TextInputStyles{
			Label: lipgloss.NewStyle().Bold(true).MarginRight(1),
			Input: base,
			Error: lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FF5F5F")).MarginRight(1),
		},
		Blurred: TextInputStyles{
			Label: lipgloss.NewStyle().MarginRight(1),
			Input: base,
			Error: lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FF5F5F")).MarginRight(1),
		},
		Disabled: TextInputStyles{
			Label: lipgloss.NewStyle().Faint(true).MarginRight(1),
			Input: disabled,
			Error: lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FF5F5F")).MarginRight(1),
		},
	}
}
