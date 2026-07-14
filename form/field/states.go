package field

import tea "charm.land/bubbletea/v2"

// DisabledState is a concrete enabled/disabled state, composed by widgets
// that need to track it.
type DisabledState struct {
	disabled bool
}

// Enable marks the component as enabled.
func (d *DisabledState) Enable() tea.Cmd {
	d.disabled = false

	return nil
}

// Disable marks the component as disabled.
func (d *DisabledState) Disable() tea.Cmd {
	d.disabled = true

	return nil
}

// Disabled returns true if the component is disabled.
func (d *DisabledState) Disabled() bool {
	return d.disabled
}

// EntryStyles is a generic container for a widget's per-state look: Focused,
// Blurred, and Disabled variants.
type EntryStyles[T any] struct {
	Focused  T
	Blurred  T
	Disabled T
}

// StyledState holds a widget's per-state styles; it doesn't track
// disabled/focused state itself.
type StyledState[T any] struct {
	styles EntryStyles[T]
}

// SetStyles replaces the current styles.
func (s *StyledState[T]) SetStyles(styles EntryStyles[T]) {
	s.styles = styles
}

// StateStyles returns the Focused, Blurred, or Disabled style variant for the
// given state.
func (s *StyledState[T]) StateStyles(disabled, focused bool) T {
	switch {
	case disabled:
		return s.styles.Disabled
	case focused:
		return s.styles.Focused
	default:
		return s.styles.Blurred
	}
}

// FocusedState is a concrete focus state for components that don't delegate
// focus to a child model.
type FocusedState struct {
	focused bool
}

// Focus sets the field as focused and returns a command to activate focus.
func (f *FocusedState) Focus() tea.Cmd {
	f.focused = true

	return nil
}

// Blur sets the field as not focused.
func (f *FocusedState) Blur() tea.Cmd {
	f.focused = false

	return nil
}

// Focused returns true if the field is focused.
func (f *FocusedState) Focused() bool {
	return f.focused
}

// NoopInit is embedded by widget types whose Init needs no asynchronous setup.
type NoopInit struct{}

// Init returns nil.
func (NoopInit) Init() tea.Cmd {
	return nil
}

// Init initializes every model and batches its non-nil commands.
func Init[T interface{ Init() tea.Cmd }](models ...T) tea.Cmd {
	cmds := make([]tea.Cmd, 0, len(models))

	for _, m := range models {
		if cmd := m.Init(); cmd != nil {
			cmds = append(cmds, cmd)
		}
	}

	return tea.Batch(cmds...)
}

// NoopWidth is embedded by widget types with no adjustable display width.
type NoopWidth struct{}

// SetWidth is a no-op.
func (NoopWidth) SetWidth(_ int) {
}

// ValueState is a generic composable value holder.
type ValueState[T any] struct {
	val T
}

// Get returns the current value of the field.
func (v *ValueState[T]) Get() T {
	return v.val
}

// Set updates the value of the field.
func (v *ValueState[T]) Set(t T) {
	v.val = t
}
