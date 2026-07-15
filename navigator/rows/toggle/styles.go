package toggle

import (
	"charm.land/lipgloss/v2"

	"github.com/gzigzigzeo/bubbles/navigator/row"
)

// ToggleStyles defines the visual pieces of a toggle row.
type ToggleStyles struct {
	Label lipgloss.Style // style for the row label
	On    lipgloss.Style // rendered when the value is true
	Off   lipgloss.Style // rendered when the value is false
}

// Styles is the per-state style set for a toggle row.
type Styles = row.StateSet[ToggleStyles]

// DefaultStyles returns usable Styles with no external theme dependency,
// rendering onMsg when on and offMsg when off.
func DefaultStyles(onMsg, offMsg string) Styles {
	neon := lipgloss.NewStyle().Foreground(lipgloss.Color("#53d1ff"))
	dim := lipgloss.NewStyle().Foreground(lipgloss.Color("#d3d3d9"))
	pale := lipgloss.NewStyle().Foreground(lipgloss.Color("#5C5C5C"))

	onMsg = "● " + onMsg
	offMsg = "○ " + offMsg

	return Styles{
		Focused: ToggleStyles{
			Label: lipgloss.NewStyle().Bold(true).MarginRight(1),
			On:    neon.SetString(onMsg),
			Off:   dim.SetString(offMsg),
		},
		Blurred: ToggleStyles{
			Label: lipgloss.NewStyle().MarginRight(1),
			On:    dim.SetString(onMsg),
			Off:   dim.SetString(offMsg),
		},
		Disabled: ToggleStyles{
			Label: lipgloss.NewStyle().Faint(true).MarginRight(1),
			On:    pale.SetString(onMsg),
			Off:   pale.SetString(offMsg),
		},
	}
}
