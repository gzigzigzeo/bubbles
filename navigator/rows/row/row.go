// Package row provides the shared interfaces and composable state helpers used
// by rows consumed by [github.com/gzigzigzeo/bubbles/navigator.Model].
package row

import tea "charm.land/bubbletea/v2"

// Focusable is implemented by rows that can hold and release keyboard focus.
type Focusable interface {
	Focus() tea.Cmd
	Blur() tea.Cmd
	Focused() bool
}

// Disableable is implemented by rows that can be enabled or disabled.
type Disableable interface {
	Enable() tea.Cmd
	Disable() tea.Cmd
	Disabled() bool
}

// FocusReceiver is implemented by models that accept directed focus entry.
// [github.com/gzigzigzeo/bubbles/navigator.Model] implements this interface,
// enabling nested navigators.
type FocusReceiver interface {
	FocusFirst() tea.Cmd
	FocusLast() tea.Cmd
}

// BoundaryAware is implemented by focus receivers that can report whether the
// cursor is currently at their first or last focusable item. A parent navigator
// uses this to move focus out of a nested navigator when its boundary is
// reached, instead of forwarding the key and letting the nested navigator keep
// focus.
type BoundaryAware interface {
	IsAtFirstFocusable() bool
	IsAtLastFocusable() bool
}

// CursorAware is implemented by rows that have an internal cursor position
// within their own View() output. Navigator uses this to scroll the active
// line into view and implements it itself for parent navigators to consume.
type CursorAware interface {
	CursorLine() int
}

// FocusedState is a concrete focus state for rows that don't delegate focus to
// a child model.
type FocusedState struct {
	focused bool
}

// Focus marks the row as focused.
func (f *FocusedState) Focus() tea.Cmd {
	f.focused = true

	return nil
}

// Blur removes focus from the row.
func (f *FocusedState) Blur() tea.Cmd {
	f.focused = false

	return nil
}

// Focused reports whether the row is focused.
func (f *FocusedState) Focused() bool {
	return f.focused
}

// DisabledState is a concrete enabled/disabled state for rows that need to
// track it.
type DisabledState struct {
	disabled bool
}

// Enable marks the row as enabled.
func (d *DisabledState) Enable() tea.Cmd {
	d.disabled = false

	return nil
}

// Disable marks the row as disabled.
func (d *DisabledState) Disable() tea.Cmd {
	d.disabled = true

	return nil
}

// Disabled reports whether the row is disabled.
func (d *DisabledState) Disabled() bool {
	return d.disabled
}
