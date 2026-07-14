package selectfield

import (
	"charm.land/lipgloss/v2"

	"github.com/gzigzigzeo/bubbles/form/field"
	"github.com/gzigzigzeo/bubbles/menu"
)

// SelectStyles holds the visual styles for the inline view and the picker dropdown.
type SelectStyles struct {
	// Inline view
	Value      lipgloss.Style
	ArrowLeft  lipgloss.Style
	ArrowRight lipgloss.Style
	// Picker dropdown
	Picker menu.Styles
}

// Styles is the field.EntryStyles container for Model's Focused, Blurred, and Disabled variants.
type Styles = field.EntryStyles[SelectStyles]

// DefaultStyles returns usable Styles with no external theme dependency.
func DefaultStyles() Styles {
	neon := lipgloss.NewStyle().Foreground(lipgloss.Color("#53d1ff"))
	dim := lipgloss.NewStyle().Foreground(lipgloss.Color("#d3d3d9"))
	pale := lipgloss.NewStyle().Foreground(lipgloss.Color("#5C5C5C"))
	picker := menu.DefaultStyles()

	return Styles{
		Focused: SelectStyles{
			Value:      dim,
			ArrowLeft:  neon.SetString("  ◀ "),
			ArrowRight: neon.SetString("  ▶"),
			Picker:     picker,
		},
		Blurred: SelectStyles{
			Value:  dim.PaddingLeft(4),
			Picker: picker,
		},
		Disabled: SelectStyles{
			Value:  pale,
			Picker: picker,
		},
	}
}
