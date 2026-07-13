package togglefield

import (
	"charm.land/lipgloss/v2"

	"github.com/gzigzigzeo/bubbles/form/field"
)

// OnOffStyles defines the styles for the "on" and "off" states of a toggle field.
type OnOffStyles struct {
	On  lipgloss.Style
	Off lipgloss.Style
}

// Styles defines the styles for a toggle field across its focused, blurred,
// and disabled states.
type Styles = field.EntryStyles[OnOffStyles]

// DefaultStyles returns usable Styles with no external theme dependency,
// rendering yesMsg when on and noMsg when off.
func DefaultStyles(yesMsg, noMsg string) Styles {
	neon := lipgloss.NewStyle().Foreground(lipgloss.Color("#53d1ff"))
	dim := lipgloss.NewStyle().Foreground(lipgloss.Color("#d3d3d9"))
	pale := lipgloss.NewStyle().Foreground(lipgloss.Color("#5C5C5C"))

	yesMsg = "● " + yesMsg
	noMsg = "○ " + noMsg

	return Styles{
		Focused: OnOffStyles{
			On:  neon.SetString(yesMsg),
			Off: dim.SetString(noMsg),
		},
		Blurred: OnOffStyles{
			On:  dim.SetString(yesMsg),
			Off: dim.SetString(noMsg),
		},
		Disabled: OnOffStyles{
			On:  pale.SetString(yesMsg),
			Off: pale.SetString(noMsg),
		},
	}
}
