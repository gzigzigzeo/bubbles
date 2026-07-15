package menu

import tea "charm.land/bubbletea/v2"

// ControllerBuilder configures a menu controller through chained method calls.
type ControllerBuilder[T comparable] struct {
	rows       []*Model[T]
	mode       Mode
	selectKeys []string
	markKeys   []string
}

// NewControllerBuilder creates a menu controller builder with defaults applied.
func NewControllerBuilder[T comparable]() *ControllerBuilder[T] {
	return &ControllerBuilder[T]{
		rows:       nil,
		mode:       ModeSelect,
		selectKeys: []string{"enter"},
		markKeys:   []string{"space"},
	}
}

// Add appends a menu row with the given name, value, description, and message.
func (b *ControllerBuilder[T]) Add(name string, value T, description string, msg tea.Msg) *ControllerBuilder[T] {
	b.rows = append(b.rows, New(name, value, description, msg))

	return b
}

// Mode sets the controller's selection mode.
func (b *ControllerBuilder[T]) Mode(mode Mode) *ControllerBuilder[T] {
	b.mode = mode

	return b
}

// SelectKeys sets the key binding that selects the focused row.
func (b *ControllerBuilder[T]) SelectKeys(keys ...string) *ControllerBuilder[T] {
	b.selectKeys = keys

	return b
}

// MarkKeys sets the key binding that toggles the focused row's mark in
// multi-select mode.
func (b *ControllerBuilder[T]) MarkKeys(keys ...string) *ControllerBuilder[T] {
	b.markKeys = keys

	return b
}

// Build returns the configured menu controller.
func (b *ControllerBuilder[T]) Build() *Controller[T] {
	return NewController(
		b.rows,
		WithMode[T](b.mode),
		WithSelectKeys[T](b.selectKeys...),
		WithMarkKeys[T](b.markKeys...),
	)
}
