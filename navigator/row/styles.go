package row

import "charm.land/lipgloss/v2"

// Shared colors used by default row styles.
var (
	ColorAccent = lipgloss.Color("#53d1ff")
	ColorDim    = lipgloss.Color("#d3d3d9")
	ColorPale   = lipgloss.Color("#5C5C5C")
	ColorError  = lipgloss.Color("#FF5F5F")
)

// DefaultLabelWidth is the default width applied to row labels so values from
// adjacent rows line up vertically.
const DefaultLabelWidth = 15

// DefaultLabelStyle returns a base style for row labels with a fixed width,
// right alignment, and a small right margin.
func DefaultLabelStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Width(DefaultLabelWidth).
		Align(lipgloss.Right).
		MarginRight(1)
}

// StateSet holds a Focused, Blurred, and Disabled variant of some style bundle.
type StateSet[T any] struct {
	Focused  T
	Blurred  T
	Disabled T
}

// StatefulStyles stores a StateSet and selects the right variant for a state.
// It does not track disabled/focused state itself.
type StatefulStyles[T any] struct {
	styles StateSet[T]
}

// SetStyles replaces the current styles.
func (s *StatefulStyles[T]) SetStyles(styles StateSet[T]) {
	s.styles = styles
}

// StateStyles returns the Focused, Blurred, or Disabled style variant for the
// given state.
func (s *StatefulStyles[T]) StateStyles(disabled, focused bool) T {
	switch {
	case disabled:
		return s.styles.Disabled
	case focused:
		return s.styles.Focused
	default:
		return s.styles.Blurred
	}
}

// Styles returns the stored style set.
func (s *StatefulStyles[T]) Styles() StateSet[T] {
	return s.styles
}
