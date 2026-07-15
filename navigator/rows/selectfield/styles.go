package selectfield

import (
	"charm.land/lipgloss/v2"

	"github.com/gzigzigzeo/bubbles/navigator/row"
)

// PickerStyles holds the visual pieces of the inline dropdown rows.
type PickerStyles struct {
	Cursor lipgloss.Style
	Item   lipgloss.Style
}

// SelectStyles holds the visual styles for the inline value and the picker
// dropdown.
type SelectStyles struct {
	Label      lipgloss.Style
	Value      lipgloss.Style
	ArrowLeft  lipgloss.Style
	ArrowRight lipgloss.Style
	Error      lipgloss.Style
	Picker     PickerStyles
}

// Styles is the per-state style set for a select field.
type Styles = row.StateSet[SelectStyles]

// DefaultStyles returns a minimal, functional style set.
func DefaultStyles() Styles {
	neon := lipgloss.NewStyle().Foreground(lipgloss.Color("#53d1ff"))
	dim := lipgloss.NewStyle().Foreground(lipgloss.Color("#d3d3d9"))
	pale := lipgloss.NewStyle().Foreground(lipgloss.Color("#5C5C5C"))
	picker := PickerStyles{
		Cursor: lipgloss.NewStyle().SetString("▶ "),
		Item:   lipgloss.NewStyle(),
	}

	return Styles{
		Focused: SelectStyles{
			Label:      lipgloss.NewStyle().Bold(true).MarginRight(1),
			Value:      dim,
			ArrowLeft:  neon.SetString("  ◀ "),
			ArrowRight: neon.SetString("  ▶"),
			Error:      lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FF5F5F")).MarginRight(1),
			Picker:     picker,
		},
		Blurred: SelectStyles{
			Label:      lipgloss.NewStyle().MarginRight(1),
			Value:      dim.PaddingLeft(4),
			ArrowRight: lipgloss.NewStyle(),
			Error:      lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FF5F5F")).MarginRight(1),
			Picker:     picker,
		},
		Disabled: SelectStyles{
			Label:      lipgloss.NewStyle().Faint(true).MarginRight(1),
			Value:      pale,
			ArrowRight: lipgloss.NewStyle(),
			Error:      lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FF5F5F")).MarginRight(1),
			Picker:     picker,
		},
	}
}
