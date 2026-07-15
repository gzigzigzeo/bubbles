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

	// Invalid styles the hint shown briefly after an unrecognized key is
	// pressed.
	Invalid lipgloss.Style
}

const iconWidth = 4

// NewWarnStyles returns Styles for a warning prompt (yellow).
func NewWarnStyles() Styles {
	colorWarn := lipgloss.Color("#ffffa0")
	colorInvalid := lipgloss.Color("#f38ba8")

	return Styles{
		Container:       lipgloss.NewStyle(),
		Icon:            lipgloss.NewStyle().Width(iconWidth).Foreground(colorWarn).SetString("⚠"),
		Label:           lipgloss.NewStyle().Foreground(colorWarn),
		CursorStyle:     lipgloss.NewStyle().Foreground(colorWarn),
		CursorTextStyle: lipgloss.NewStyle().Foreground(colorWarn),
		Echo:            lipgloss.NewStyle().Foreground(colorWarn),
		Invalid:         lipgloss.NewStyle().Foreground(colorInvalid),
	}
}

// WithWarnStyles applies the warning (yellow) style preset. Equivalent to
// WithStyles(NewWarnStyles()).
func WithWarnStyles() Option { return WithStyles(NewWarnStyles()) }

// NewErrorStyles returns Styles for an error prompt (orange).
func NewErrorStyles() Styles {
	colorCaution := lipgloss.Color("#ffb347")
	colorInvalid := lipgloss.Color("#f38ba8")

	return Styles{
		Container:       lipgloss.NewStyle(),
		Icon:            lipgloss.NewStyle().Width(iconWidth).Foreground(colorCaution).SetString("!"),
		Label:           lipgloss.NewStyle().Foreground(colorCaution),
		CursorStyle:     lipgloss.NewStyle().Foreground(colorCaution),
		CursorTextStyle: lipgloss.NewStyle().Foreground(colorCaution),
		Echo:            lipgloss.NewStyle().Foreground(colorCaution),
		Invalid:         lipgloss.NewStyle().Foreground(colorInvalid),
	}
}

// WithErrorStyles applies the error (orange) style preset. Equivalent to
// WithStyles(NewErrorStyles()).
func WithErrorStyles() Option { return WithStyles(NewErrorStyles()) }

// NewSuccessStyles returns Styles for a success prompt (green).
func NewSuccessStyles() Styles {
	colorSuccess := lipgloss.Color("#a6e3a1")
	colorInvalid := lipgloss.Color("#f38ba8")

	return Styles{
		Container:       lipgloss.NewStyle(),
		Icon:            lipgloss.NewStyle().Width(iconWidth).Foreground(colorSuccess).SetString("✓"),
		Label:           lipgloss.NewStyle().Foreground(colorSuccess),
		CursorStyle:     lipgloss.NewStyle().Foreground(colorSuccess),
		CursorTextStyle: lipgloss.NewStyle().Foreground(colorSuccess),
		Echo:            lipgloss.NewStyle().Foreground(colorSuccess),
		Invalid:         lipgloss.NewStyle().Foreground(colorInvalid),
	}
}

// WithSuccessStyles applies the success (green) style preset. Equivalent to
// WithStyles(NewSuccessStyles()).
func WithSuccessStyles() Option { return WithStyles(NewSuccessStyles()) }

// NewInfoStyles returns Styles for an info prompt (default terminal colors).
func NewInfoStyles() Styles {
	colorInvalid := lipgloss.Color("#f38ba8")

	return Styles{
		Container:       lipgloss.NewStyle(),
		Icon:            lipgloss.NewStyle().Width(iconWidth).SetString("i"),
		Label:           lipgloss.NewStyle(),
		CursorStyle:     lipgloss.NewStyle(),
		CursorTextStyle: lipgloss.NewStyle(),
		Echo:            lipgloss.NewStyle(),
		Invalid:         lipgloss.NewStyle().Foreground(colorInvalid),
	}
}

// WithInfoStyles applies the info (neutral) style preset. Equivalent to
// WithStyles(NewInfoStyles()).
func WithInfoStyles() Option { return WithStyles(NewInfoStyles()) }
