package menurow

import (
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
)

// Mode controls how a [Collection] responds to activation keys.
type Mode int

const (
	// ModeSelect is a single-select menu: Enter selects the focused row and
	// Space is unbound.
	ModeSelect Mode = iota

	// ModeMultiSelect is a multi-select menu: Space toggles the focused row's
	// mark and Enter selects it.
	ModeMultiSelect
)

// Collection manages a set of menu rows, their marks, and activation keys. It
// is the controller for its rows: rows are data sources and transparently
// forward keys to the collection.
type Collection[T comparable] struct {
	mode         Mode
	rows         []*Model[T]
	marked       map[int]struct{}
	focusedIndex int
	selectKey    key.Binding
	markKey      key.Binding
}

// CollectionOption configures a [Collection].
type CollectionOption[T comparable] func(*Collection[T])

// WithMode sets the collection's selection mode.
func WithMode[T comparable](mode Mode) CollectionOption[T] {
	return func(c *Collection[T]) {
		c.mode = mode
	}
}

// WithSelectKeys sets the key binding that selects the focused row.
func WithSelectKeys[T comparable](keys ...string) CollectionOption[T] {
	return func(c *Collection[T]) {
		c.selectKey = key.NewBinding(key.WithKeys(keys...), key.WithHelp(keys[0], "select"))
	}
}

// WithMarkKeys sets the key binding that toggles the focused row's mark in
// [ModeMultiSelect].
func WithMarkKeys[T comparable](keys ...string) CollectionOption[T] {
	return func(c *Collection[T]) {
		c.markKey = key.NewBinding(key.WithKeys(keys...), key.WithHelp(keys[0], "mark"))
	}
}

// NewCollection creates a collection over the given rows and applies options.
func NewCollection[T comparable](rows []*Model[T], opts ...CollectionOption[T]) *Collection[T] {
	c := &Collection[T]{
		rows:         rows,
		marked:       make(map[int]struct{}),
		focusedIndex: -1,
		selectKey:    key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "select")),
		markKey:      key.NewBinding(key.WithKeys("space"), key.WithHelp("space", "mark")),
	}

	for i, r := range rows {
		r.collection = c
		r.index = i
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

// Rows returns the rows in the collection.
func (c *Collection[T]) Rows() []*Model[T] {
	return c.rows
}

// FocusedIndex returns the index of the row that currently has focus, or -1 if
// none of the collection's rows are focused.
func (c *Collection[T]) FocusedIndex() int {
	return c.focusedIndex
}

// setFocused records that the row at idx has received focus.
func (c *Collection[T]) setFocused(idx int) {
	c.focusedIndex = idx
}

// clearFocus records that the row at idx has lost focus, but only if it was
// the currently focused row.
func (c *Collection[T]) clearFocus(idx int) {
	if c.focusedIndex == idx {
		c.focusedIndex = -1
	}
}

// updateForRow handles keys forwarded by a row. It returns a command if the key
// activated the row, or nil if the key was ignored.
func (c *Collection[T]) updateForRow(row *Model[T], msg tea.Msg) tea.Cmd {
	km, ok := msg.(tea.KeyMsg)
	if !ok || !row.Focused() || c.focusedIndex != row.index {
		return nil
	}

	switch {
	case c.mode == ModeMultiSelect && key.Matches(km, c.markKey):
		c.Toggle(c.focusedIndex)

		return nil
	case key.Matches(km, c.selectKey):
		return func() tea.Msg {
			return c.rows[c.focusedIndex].msg
		}
	}

	return nil
}

// Mark marks the row at index. Out-of-range indices are ignored.
func (c *Collection[T]) Mark(index int) {
	if !c.valid(index) {
		return
	}

	c.rows[index].SetMarked(true)
	c.marked[index] = struct{}{}
}

// Unmark unmarks the row at index. Out-of-range indices are ignored.
func (c *Collection[T]) Unmark(index int) {
	if !c.valid(index) {
		return
	}

	c.rows[index].SetMarked(false)
	delete(c.marked, index)
}

// Toggle marks the row at index if it is unmarked, or unmarks it if marked.
func (c *Collection[T]) Toggle(index int) {
	if c.IsMarked(index) {
		c.Unmark(index)

		return
	}

	c.Mark(index)
}

// MarkOnly unmarks every row and then marks the row at index.
func (c *Collection[T]) MarkOnly(index int) {
	c.UnmarkAll()
	c.Mark(index)
}

// UnmarkAll clears every mark.
func (c *Collection[T]) UnmarkAll() {
	for index := range c.marked {
		c.rows[index].SetMarked(false)
	}

	clear(c.marked)
}

// Marked returns the indices of all marked rows in ascending order.
func (c *Collection[T]) Marked() []int {
	indices := make([]int, 0, len(c.marked))

	for i := range c.rows {
		if _, ok := c.marked[i]; ok {
			indices = append(indices, i)
		}
	}

	return indices
}

// MarkedValues returns the values of all marked rows in ascending index order.
func (c *Collection[T]) MarkedValues() []T {
	values := make([]T, 0, len(c.marked))

	for i := range c.rows {
		if _, ok := c.marked[i]; ok {
			values = append(values, c.rows[i].Value())
		}
	}

	return values
}

// IsMarked reports whether the row at index is marked.
func (c *Collection[T]) IsMarked(index int) bool {
	if !c.valid(index) {
		return false
	}

	return c.rows[index].Marked()
}

// valid reports whether index is inside the rows slice.
func (c *Collection[T]) valid(index int) bool {
	return index >= 0 && index < len(c.rows)
}
