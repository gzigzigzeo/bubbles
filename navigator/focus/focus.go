// Package focus provides a reusable focus controller for an ordered list of
// [tea.Model] items. It tracks the focused index, skips disabled items, supports
// wrap-around, and interprets configurable next/previous key bindings. It does
// not render or manage viewport scrolling.
package focus

import (
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"

	"github.com/gzigzigzeo/bubbles/navigator/row"
)

// Controller manages an ordered list of items and which one holds focus.
type Controller struct {
	items   []tea.Model
	focused int
	wrap    bool
	nextKey key.Binding
	prevKey key.Binding
}

const initialCmdCapacity = 2

// New creates a Controller over items. Focus starts at -1 (nothing focused) and
// wrap is disabled. Next and previous keys are unbound until configured.
func New(items ...tea.Model) *Controller {
	return &Controller{
		items:   items,
		focused: -1,
		wrap:    false,
		nextKey: key.NewBinding(),
		prevKey: key.NewBinding(),
	}
}

// SetWrap enables or disables wrap-at-boundaries focus movement.
func (c *Controller) SetWrap(wrap bool) {
	c.wrap = wrap
}

// SetNextKeys sets the key binding that moves focus to the next focusable item.
func (c *Controller) SetNextKeys(keys ...string) {
	c.nextKey = key.NewBinding(key.WithKeys(keys...), key.WithHelp(keys[0], "next"))
}

// SetPrevKeys sets the key binding that moves focus to the previous focusable
// item.
func (c *Controller) SetPrevKeys(keys ...string) {
	c.prevKey = key.NewBinding(key.WithKeys(keys...), key.WithHelp(keys[0], "prev"))
}

// Init initializes all items.
func (c *Controller) Init() tea.Cmd {
	cmds := make([]tea.Cmd, 0, len(c.items))

	for _, it := range c.items {
		cmds = append(cmds, it.Init())
	}

	return tea.Batch(cmds...)
}

// Update handles next/previous navigation keys and forwards all other messages
// to the currently focused item.
func (c *Controller) Update(msg tea.Msg) tea.Cmd {
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return c.updateFocused(msg)
	}

	switch {
	case key.Matches(keyMsg, c.nextKey):
		return c.MoveNext()
	case key.Matches(keyMsg, c.prevKey):
		return c.MovePrev()
	default:
		return c.updateFocused(msg)
	}
}

// Focus focuses the first focusable item.
func (c *Controller) Focus() tea.Cmd {
	return c.FocusFirst()
}

// Blur removes focus from the current item and resets the focused index.
func (c *Controller) Blur() tea.Cmd {
	if c.focused < 0 || c.focused >= len(c.items) {
		c.focused = -1

		return nil
	}

	idx := c.focused
	c.focused = -1

	if f, ok := c.items[idx].(row.Focusable); ok {
		return f.Blur()
	}

	return nil
}

// Focused reports whether any item holds focus.
func (c *Controller) Focused() bool {
	return c.focused >= 0
}

// FocusFirst focuses the first non-disabled focusable item.
func (c *Controller) FocusFirst() tea.Cmd {
	return c.focusIndexDir(c.firstFocusable(), 1)
}

// FocusLast focuses the last non-disabled focusable item.
func (c *Controller) FocusLast() tea.Cmd {
	return c.focusIndexDir(c.lastFocusable(), -1)
}

// FocusIndex focuses the item at idx if it is focusable and non-disabled.
func (c *Controller) FocusIndex(idx int) tea.Cmd {
	if !c.isFocusable(idx) {
		return nil
	}

	return c.focusIndexDir(idx, 1)
}

// MoveNext moves focus to the next focusable item, wrapping if enabled.
func (c *Controller) MoveNext() tea.Cmd {
	next := c.nextFocusable(c.focused, 1)

	if next < 0 {
		if !c.wrap {
			return nil
		}

		next = c.firstFocusable()
		if next < 0 {
			return nil
		}
	}

	return c.focusIndexDir(next, 1)
}

// MovePrev moves focus to the previous focusable item, wrapping if enabled.
func (c *Controller) MovePrev() tea.Cmd {
	prev := c.nextFocusable(c.focused, -1)

	if prev < 0 {
		if !c.wrap {
			return nil
		}

		prev = c.lastFocusable()
		if prev < 0 {
			return nil
		}
	}

	return c.focusIndexDir(prev, -1)
}

// IsAtFirstFocusable reports whether the currently focused item is the first
// focusable item.
func (c *Controller) IsAtFirstFocusable() bool {
	return c.focused == c.firstFocusable()
}

// IsAtLastFocusable reports whether the currently focused item is the last
// focusable item.
func (c *Controller) IsAtLastFocusable() bool {
	return c.focused == c.lastFocusable()
}

// CurrentItem returns the currently focused item, or nil if none is focused.
func (c *Controller) CurrentItem() tea.Model {
	if c.focused < 0 || c.focused >= len(c.items) {
		return nil
	}

	return c.items[c.focused]
}

// Items returns the items in the controller.
func (c *Controller) Items() []tea.Model {
	return c.items
}

// FocusedIndex returns the index of the currently focused item, or -1 if none
// is focused.
func (c *Controller) FocusedIndex() int {
	return c.focused
}

// updateFocused forwards msg to the currently focused item and returns its
// command.
func (c *Controller) updateFocused(msg tea.Msg) tea.Cmd {
	if c.focused < 0 || c.focused >= len(c.items) {
		return nil
	}

	updated, cmd := c.items[c.focused].Update(msg)
	c.items[c.focused] = updated

	return cmd
}

// focusIndexDir blurs the current item and focuses the item at idx, using dir
// to select FocusFirst (dir >= 0) vs FocusLast (dir < 0) for nested
// [row.FocusReceiver] items.
func (c *Controller) focusIndexDir(idx int, dir int) tea.Cmd {
	if idx < 0 || idx >= len(c.items) {
		return nil
	}

	cmds := make([]tea.Cmd, 0, initialCmdCapacity)

	if blur := c.Blur(); blur != nil {
		cmds = append(cmds, blur)
	}

	c.focused = idx

	if focusCmd := c.focusCmdForItem(idx, dir); focusCmd != nil {
		cmds = append(cmds, focusCmd)
	}

	return tea.Batch(cmds...)
}

// focusCmdForItem returns the command to focus the item at idx, choosing
// FocusFirst (dir >= 0) vs FocusLast (dir < 0) for nested [row.FocusReceiver]
// items.
func (c *Controller) focusCmdForItem(idx int, dir int) tea.Cmd {
	focusable, ok := c.items[idx].(row.Focusable)
	if !ok {
		return nil
	}

	if receiver, ok := focusable.(row.FocusReceiver); ok {
		if dir >= 0 {
			return receiver.FocusFirst()
		}

		return receiver.FocusLast()
	}

	return focusable.Focus()
}

// isFocusable reports whether the item at idx can receive focus.
func (c *Controller) isFocusable(idx int) bool {
	if idx < 0 || idx >= len(c.items) {
		return false
	}

	item := c.items[idx]

	if _, ok := item.(row.Focusable); !ok {
		return false
	}

	if d, ok := item.(row.Disableable); ok && d.Disabled() {
		return false
	}

	return true
}

// nextFocusable returns the first focusable index beyond from in direction dir,
// or -1 if none exists.
func (c *Controller) nextFocusable(from, dir int) int {
	pos := from + dir

	for pos >= 0 && pos < len(c.items) {
		if c.isFocusable(pos) {
			return pos
		}

		pos += dir
	}

	return -1
}

// firstFocusable returns the index of the first focusable item, or -1.
func (c *Controller) firstFocusable() int {
	for i := range c.items {
		if c.isFocusable(i) {
			return i
		}
	}

	return -1
}

// lastFocusable returns the index of the last focusable item, or -1.
func (c *Controller) lastFocusable() int {
	for i := len(c.items) - 1; i >= 0; i-- {
		if c.isFocusable(i) {
			return i
		}
	}

	return -1
}
