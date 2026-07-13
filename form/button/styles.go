package button

import (
	"charm.land/lipgloss/v2"

	"github.com/gzigzigzeo/bubbles/form/field"
)

// Styles defines the visual appearance of a button for each of its three states.
type Styles = field.EntryStyles[lipgloss.Style]

// DefaultStyles returns usable Styles with no external theme dependency.
func DefaultStyles() Styles {
	return Styles{
		Focused:  lipgloss.NewStyle().Foreground(lipgloss.Color("#53d1ff")).Bold(true).MarginRight(1),
		Blurred:  lipgloss.NewStyle().Foreground(lipgloss.Color("#d3d3d9")).MarginRight(1),
		Disabled: lipgloss.NewStyle().Foreground(lipgloss.Color("#5C5C5C")).MarginRight(1),
	}
}
