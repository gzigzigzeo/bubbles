package menu

import (
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
)

// Mode controls how a [Controller] responds to activation keys.
type Mode int

const (
	// ModeSelect is a single-select menu: Enter selects the focused row and
	// Space is unbound.
	ModeSelect Mode = iota

	// ModeMultiSelect is a multi-select menu: Space toggles the focused row's
	// mark and Enter selects it.
	ModeMultiSelect
)

// Controller manages a set of menu rows, their marks, and activation keys. It
// is the controller for its rows: rows are data sources and transparently
// forward keys to the controller.
type Controller[T comparable] struct {
	mode         Mode
	rows         []*Model[T]
	marked       map[int]struct{}
	focusedIndex int
	selectKey    key.Binding
	markKey      key.Binding
}

// ControllerOption configures a [Controller].
type ControllerOption[T comparable] func(*Controller[T])

// WithMode sets the controller's selection mode.
func WithMode[T comparable](mode Mode) ControllerOption[T] {
	return func(c *Controller[T]) {
		c.mode = mode
	}
}

// WithSelectKeys sets the key binding that selects the focused row.
func WithSelectKeys[T comparable](keys ...string) ControllerOption[T] {
	return func(c *Controller[T]) {
		c.selectKey = key.NewBinding(key.WithKeys(keys...), key.WithHelp(keys[0], "select"))
	}
}

// WithMarkKeys sets the key binding that toggles the focused row's mark in
// [ModeMultiSelect].
func WithMarkKeys[T comparable](keys ...string) ControllerOption[T] {
	return func(c *Controller[T]) {
		c.markKey = key.NewBinding(key.WithKeys(keys...), key.WithHelp(keys[0], "mark"))
	}
}

// NewController creates a controller over the given rows and applies options.
func NewController[T comparable](rows []*Model[T], opts ...ControllerOption[T]) *Controller[T] {
	ctrl := &Controller[T]{
		mode:         ModeSelect,
		rows:         rows,
		marked:       make(map[int]struct{}),
		focusedIndex: -1,
		selectKey:    key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "select")),
		markKey:      key.NewBinding(key.WithKeys("space"), key.WithHelp("space", "mark")),
	}

	for i, r := range rows {
		r.controller = ctrl
		r.index = i
	}

	for _, opt := range opts {
		opt(ctrl)
	}

	return ctrl
}

// Rows returns the rows in the controller.
func (c *Controller[T]) Rows() []*Model[T] {
	return c.rows
}

// FocusedIndex returns the index of the row that currently has focus, or -1 if
// none of the controller's rows are focused.
func (c *Controller[T]) FocusedIndex() int {
	return c.focusedIndex
}

// setFocused records that the row at idx has received focus.
func (c *Controller[T]) setFocused(idx int) {
	c.focusedIndex = idx
}

// clearFocus records that the row at idx has lost focus, but only if it was
// the currently focused row.
func (c *Controller[T]) clearFocus(idx int) {
	if c.focusedIndex == idx {
		c.focusedIndex = -1
	}
}

// Update handles keys forwarded by a row. It returns a command if the key
// activated the row, or nil if the key was ignored.
func (c *Controller[T]) Update(msg tea.Msg) tea.Cmd {
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok || c.focusedIndex < 0 {
		return nil
	}

	row := c.rows[c.focusedIndex]
	if !row.Focused() {
		return nil
	}

	switch {
	case c.mode == ModeMultiSelect && key.Matches(keyMsg, c.markKey):
		c.Toggle(c.focusedIndex)

		return nil
	case key.Matches(keyMsg, c.selectKey):
		return func() tea.Msg {
			return c.rows[c.focusedIndex].msg
		}
	}

	return nil
}

// Mark marks the row at index. Out-of-range indices are ignored.
func (c *Controller[T]) Mark(index int) {
	if !c.valid(index) {
		return
	}

	c.rows[index].SetMarked(true)
	c.marked[index] = struct{}{}
}

// Unmark unmarks the row at index. Out-of-range indices are ignored.
func (c *Controller[T]) Unmark(index int) {
	if !c.valid(index) {
		return
	}

	c.rows[index].SetMarked(false)
	delete(c.marked, index)
}

// Toggle marks the row at index if it is unmarked, or unmarks it if marked.
func (c *Controller[T]) Toggle(index int) {
	if c.IsMarked(index) {
		c.Unmark(index)

		return
	}

	c.Mark(index)
}

// MarkOnly unmarks every row and then marks the row at index.
func (c *Controller[T]) MarkOnly(index int) {
	c.UnmarkAll()
	c.Mark(index)
}

// UnmarkAll clears every mark.
func (c *Controller[T]) UnmarkAll() {
	for index := range c.marked {
		c.rows[index].SetMarked(false)
	}

	clear(c.marked)
}

// Marked returns the indices of all marked rows in ascending order.
func (c *Controller[T]) Marked() []int {
	indices := make([]int, 0, len(c.marked))

	for i := range c.rows {
		if _, ok := c.marked[i]; ok {
			indices = append(indices, i)
		}
	}

	return indices
}

// MarkedValues returns the values of all marked rows in ascending index order.
func (c *Controller[T]) MarkedValues() []T {
	values := make([]T, 0, len(c.marked))

	for i := range c.rows {
		if _, ok := c.marked[i]; ok {
			values = append(values, c.rows[i].Value())
		}
	}

	return values
}

// IsMarked reports whether the row at index is marked.
func (c *Controller[T]) IsMarked(index int) bool {
	if !c.valid(index) {
		return false
	}

	return c.rows[index].Marked()
}

// valid reports whether index is inside the rows slice.
func (c *Controller[T]) valid(index int) bool {
	return index >= 0 && index < len(c.rows)
}
