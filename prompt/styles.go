package prompt

import "charm.land/lipgloss/v2"

// Styles configures the appearance of a Prompt.
type Styles struct {
	// Container wraps the entire prompt row. Its Width() defines the total
	// component width; the question column fills the remainder after the icon.
	Container lipgloss.Style

	// Icon is a pre-styled glyph column. Set the glyph with SetString and fix
	// the column width with Width so the question column aligns consistently,
	// e.g. lipgloss.NewStyle().Width(4).Foreground(color).SetString("⚠").
	Icon lipgloss.Style

	// Label styles the question text.
	Label lipgloss.Style

	// CursorStyle is the style applied to the cursor block.
	CursorStyle lipgloss.Style

	// CursorTextStyle is the style applied to the character inside the cursor
	// when it is in its blinking (off) state.
	CursorTextStyle lipgloss.Style

	// Echo styles the answer label rendered after the user responds.
	Echo lipgloss.Style
}

var (
	colorWarn    = lipgloss.Color("#ffffa0")
	colorCaution = lipgloss.Color("#ffb347")
	colorSuccess = lipgloss.Color("#a6e3a1")
)

// NewWarnStyles returns Styles for a warning prompt (yellow).
func NewWarnStyles() Styles {
	return Styles{
		Icon:            lipgloss.NewStyle().Width(4).Foreground(colorWarn).SetString("⚠"),
		Label:           lipgloss.NewStyle().Foreground(colorWarn),
		CursorStyle:     lipgloss.NewStyle().Foreground(colorWarn),
		CursorTextStyle: lipgloss.NewStyle().Foreground(colorWarn),
		Echo:            lipgloss.NewStyle().Foreground(colorWarn),
	}
}

// NewErrorStyles returns Styles for an error prompt (orange).
func NewErrorStyles() Styles {
	return Styles{
		Icon:            lipgloss.NewStyle().Width(4).Foreground(colorCaution).SetString("!"),
		Label:           lipgloss.NewStyle().Foreground(colorCaution),
		CursorStyle:     lipgloss.NewStyle().Foreground(colorCaution),
		CursorTextStyle: lipgloss.NewStyle().Foreground(colorCaution),
		Echo:            lipgloss.NewStyle().Foreground(colorCaution),
	}
}

// NewSuccessStyles returns Styles for a success prompt (green).
func NewSuccessStyles() Styles {
	return Styles{
		Icon:            lipgloss.NewStyle().Width(4).Foreground(colorSuccess).SetString("✓"),
		Label:           lipgloss.NewStyle().Foreground(colorSuccess),
		CursorStyle:     lipgloss.NewStyle().Foreground(colorSuccess),
		CursorTextStyle: lipgloss.NewStyle().Foreground(colorSuccess),
		Echo:            lipgloss.NewStyle().Foreground(colorSuccess),
	}
}

// NewInfoStyles returns Styles for an info prompt (default terminal colors).
func NewInfoStyles() Styles {
	return Styles{
		Icon: lipgloss.NewStyle().Width(4).SetString("i"),
	}
}
