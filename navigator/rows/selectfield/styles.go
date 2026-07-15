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
	neon := lipgloss.NewStyle().Foreground(row.ColorAccent)
	dim := lipgloss.NewStyle().Foreground(row.ColorDim)
	pale := lipgloss.NewStyle().Foreground(row.ColorPale)
	picker := PickerStyles{
		Cursor: lipgloss.NewStyle().SetString("▶ "),
		Item:   lipgloss.NewStyle(),
	}

	return Styles{
		Focused: SelectStyles{
			Label:      row.DefaultLabelStyle().Bold(true),
			Value:      dim,
			ArrowLeft:  neon.SetString("  ◀ "),
			ArrowRight: neon.SetString("  ▶"),
			Error:      lipgloss.NewStyle().Bold(true).Foreground(row.ColorError).MarginRight(1),
			Picker:     picker,
		},
		Blurred: SelectStyles{
			Label:      row.DefaultLabelStyle(),
			Value:      dim.PaddingLeft(4),
			ArrowRight: lipgloss.NewStyle(),
			Error:      lipgloss.NewStyle().Bold(true).Foreground(row.ColorError).MarginRight(1),
			Picker:     picker,
		},
		Disabled: SelectStyles{
			Label:      row.DefaultLabelStyle().Faint(true),
			Value:      pale,
			ArrowRight: lipgloss.NewStyle(),
			Error:      lipgloss.NewStyle().Bold(true).Foreground(row.ColorError).MarginRight(1),
			Picker:     picker,
		},
	}
}
