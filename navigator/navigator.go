// Package navigator manages an ordered list of rows with keyboard focus.
// Each row is a [tea.Model] that may additionally implement [Focusable],
// [Disableable], and/or [FocusReceiver].
//
// The navigator keeps the focused row visible through an internal
// [ViewportCoordinator]. Callers can access the coordinator via
// [Model.ViewportCoordinator] to configure the viewport or attach a [Viewport].
package navigator

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/bubbles/v2/key"
	"charm.land/lipgloss/v2"
)

var menuKeyUp = key.NewBinding(
	key.WithKeys("up", "k"),
	key.WithHelp("↑/k", "up"),
)

var menuKeyDown = key.NewBinding(
	key.WithKeys("down", "j"),
	key.WithHelp("↓/j", "down"),
)

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
// [Model] implements this interface, enabling nested navigators.
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

// Model manages an ordered list of rows with keyboard navigation.
// It renders rows as a flat joined string and uses an internal
// [ViewportCoordinator] to keep the focused row visible. The zero value is not
// usable; use [New].
type Model struct {
	rows    []tea.Model
	focused int  // index into rows; -1 = none focused
	closed  bool // true: wrap at boundaries; false: keep focus at boundaries
	coord   *ViewportCoordinator
}

// New creates a Navigator over rows. Focus is open by default (loses focus at
// boundaries). A default [ViewportCoordinator] is created automatically; replace
// it with [Model.SetViewportCoordinator] if needed. Call [FocusFirst] or
// [FocusLast] to give initial focus, then call [Init].
func New(rows ...tea.Model) *Model {
	return &Model{
		rows:  rows,
		focused: -1,
		coord:   NewViewportCoordinator(),
	}
}

// Closed enables closed-off focus mode: focus wraps from the last row back
// to the first (and vice versa) instead of keeping focus at boundaries.
func (n *Model) Closed() {
	n.closed = true
}

// ViewportCoordinator returns the navigator's internal viewport coordinator.
func (n *Model) ViewportCoordinator() *ViewportCoordinator {
	return n.coord
}

// SetViewportCoordinator replaces the navigator's viewport coordinator. The
// coordinator must not be nil.
func (n *Model) SetViewportCoordinator(c *ViewportCoordinator) {
	n.coord = c
}

// FocusFirst focuses the first non-disabled [Focusable] row and scrolls it into
// view. Implements [FocusReceiver].
func (n *Model) FocusFirst() tea.Cmd {
	cmd := n.focusIndexDir(n.firstFocusable(), 1)
	n.scrollToFocus()

	return cmd
}

// FocusLast focuses the last non-disabled [Focusable] row and scrolls it into
// view. Implements [FocusReceiver].
func (n *Model) FocusLast() tea.Cmd {
	cmd := n.focusIndexDir(n.lastFocusable(), -1)
	n.scrollToFocus()

	return cmd
}

// Focus focuses the first non-disabled row. Implements [Focusable].
func (n *Model) Focus() tea.Cmd {
	return n.FocusFirst()
}

// Blur removes focus from the current row. Implements [Focusable].
func (n *Model) Blur() tea.Cmd {
	return n.blurCurrent()
}

// Focused reports whether any row holds focus. Implements [Focusable].
func (n *Model) Focused() bool {
	return n.focused >= 0
}

// Init initializes all rows.
func (n *Model) Init() tea.Cmd {
	cmds := make([]tea.Cmd, 0, len(n.rows))

	for _, r := range n.rows {
		cmds = append(cmds, r.Init())
	}

	return tea.Batch(cmds...)
}

// Update handles keyboard navigation and routes other messages to the focused
// row. Up/Down (and vi aliases k/j) move focus between rows. When the focused
// row is a [FocusReceiver] (a nested Navigator), keys are passed through it;
// if it defocuses itself the outer Navigator shifts focus in the same cycle.
// The internal [ViewportCoordinator] is updated automatically so the focused
// row stays visible.
func (n *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if km, ok := msg.(tea.KeyMsg); ok {
		oldCursor := n.CursorLine()
		dir := n.navigationDir(km)

		if n.focused >= 0 {
			if _, isReceiver := n.rows[n.focused].(FocusReceiver); isReceiver {
				_, cmd := n.updateFocusedReceiver(km)
				n.scrollAfterMove(oldCursor, dir)

				return n, cmd
			}
		}

		var cmd tea.Cmd

		switch {
		case key.Matches(km, menuKeyUp):
			cmd = n.move(-1)
		case key.Matches(km, menuKeyDown):
			cmd = n.move(1)
		}

		n.scrollAfterMove(oldCursor, dir)

		return n, cmd
	}

	if n.focused >= 0 {
		return n.updateFocused(msg)
	}

	return n, nil
}

// View renders all rows as a flat joined string. Pair with a viewport (via the
// internal [ViewportCoordinator]) to get height-clipped scrollable display.
func (n *Model) View() tea.View {
	rows := make([]string, len(n.rows))

	for i, r := range n.rows {
		rows[i] = r.View().Content
	}

	return tea.NewView(lipgloss.JoinVertical(lipgloss.Left, rows...))
}

// navigationDir returns -1 for up keys, 1 for down keys, and 0 otherwise.
func (n *Model) navigationDir(km tea.KeyMsg) int {
	switch {
	case key.Matches(km, menuKeyUp):
		return -1
	case key.Matches(km, menuKeyDown):
		return 1
	}

	return 0
}

// scrollAfterMove updates the viewport coordinator after a potential focus
// change. If the cursor moved the viewport scrolls to keep it visible; if the
// cursor stayed at a boundary the viewport scrolls one line in dir.
func (n *Model) scrollAfterMove(oldCursor, dir int) {
	if n.coord == nil {
		return
	}

	newCursor := n.CursorLine()
	total := n.totalLines()

	if newCursor != oldCursor {
		n.coord.scrollToFocus(newCursor, total)
	} else if dir != 0 {
		n.coord.scrollAtBoundary(newCursor, total, dir)
	}
}

// scrollToFocus scrolls the internal coordinator to the current cursor line.
func (n *Model) scrollToFocus() {
	if n.coord == nil {
		return
	}

	n.coord.scrollToFocus(n.CursorLine(), n.totalLines())
}

// updateFocused forwards msg to the currently focused row and returns.
func (n *Model) updateFocused(msg tea.Msg) (tea.Model, tea.Cmd) {
	updated, cmd := n.rows[n.focused].Update(msg)
	n.rows[n.focused] = updated

	return n, cmd
}

// updateFocusedReceiver forwards a key to the focused [FocusReceiver] row.
// If the row reports it is at a boundary via [BoundaryAware], the parent
// handles the key instead so focus can leave the nested navigator. If the row
// defocuses itself as a result, focus is shifted within the same Update cycle —
// no message round-trip needed.
func (n *Model) updateFocusedReceiver(km tea.KeyMsg) (tea.Model, tea.Cmd) {
	f, isFocusable := n.rows[n.focused].(Focusable)
	wasFocused := isFocusable && f.Focused()

	// If the nested navigator is at the relevant boundary, let the parent move
	// focus out instead of forwarding the key.
	if ba, ok := n.rows[n.focused].(BoundaryAware); ok {
		switch {
		case key.Matches(km, menuKeyUp):
			if ba.IsAtFirstFocusable() {
				return n, n.move(-1)
			}
		case key.Matches(km, menuKeyDown):
			if ba.IsAtLastFocusable() {
				return n, n.move(1)
			}
		}
	}

	updated, cmd := n.rows[n.focused].Update(km)
	n.rows[n.focused] = updated

	if wasFocused {
		if nf, ok := updated.(Focusable); ok && !nf.Focused() {
			dir := 1
			if key.Matches(km, menuKeyUp) {
				dir = -1
			}

			return n, tea.Batch(cmd, n.move(dir))
		}
	}

	return n, cmd
}

// move handles an up/down navigation key in dir (+1 = down, -1 = up).
// When a focusable row exists in that direction, focus moves to it. At a
// boundary in open mode focus stays put; in closed mode focus wraps.
func (n *Model) move(dir int) tea.Cmd {
	next := n.nextFocusable(n.focused, dir)

	if next < 0 {
		if !n.closed {
			return nil
		}

		if dir > 0 {
			next = n.firstFocusable()
		} else {
			next = n.lastFocusable()
		}

		if next < 0 {
			return nil
		}
	}

	return n.focusIndexDir(next, dir)
}

// IsAtFirstFocusable reports whether the currently focused row is the first
// non-disabled focusable row.
func (n *Model) IsAtFirstFocusable() bool {
	return n.focused == n.firstFocusable()
}

// IsAtLastFocusable reports whether the currently focused row is the last
// non-disabled focusable row.
func (n *Model) IsAtLastFocusable() bool {
	return n.focused == n.lastFocusable()
}

// focusIndexDir blurs the current row and focuses the row at idx, using dir
// to select FocusFirst (dir >= 0) vs FocusLast (dir < 0) for nested
// [FocusReceiver] rows.
func (n *Model) focusIndexDir(idx int, dir int) tea.Cmd {
	if idx < 0 || idx >= len(n.rows) {
		return nil
	}

	cmds := make([]tea.Cmd, 0, 2)

	if blur := n.blurCurrent(); blur != nil {
		cmds = append(cmds, blur)
	}

	n.focused = idx

	if f, ok := n.rows[idx].(Focusable); ok {
		var focusCmd tea.Cmd

		if fr, ok := f.(FocusReceiver); ok {
			if dir >= 0 {
				focusCmd = fr.FocusFirst()
			} else {
				focusCmd = fr.FocusLast()
			}
		} else {
			focusCmd = f.Focus()
		}

		if focusCmd != nil {
			cmds = append(cmds, focusCmd)
		}
	}

	return tea.Batch(cmds...)
}

// blurCurrent blurs the currently focused row and returns its command.
func (n *Model) blurCurrent() tea.Cmd {
	if n.focused < 0 || n.focused >= len(n.rows) {
		return nil
	}

	if f, ok := n.rows[n.focused].(Focusable); ok {
		return f.Blur()
	}

	return nil
}

// isFocusable reports whether the row at idx can receive focus: it must
// implement [Focusable] and must not be disabled.
func (n *Model) isFocusable(idx int) bool {
	if idx < 0 || idx >= len(n.rows) {
		return false
	}

	r := n.rows[idx]

	if _, ok := r.(Focusable); !ok {
		return false
	}

	if d, ok := r.(Disableable); ok && d.Disabled() {
		return false
	}

	return true
}

// nextFocusable returns the first focusable index beyond from in direction
// dir, or -1 if none exists.
func (n *Model) nextFocusable(from, dir int) int {
	pos := from + dir

	for pos >= 0 && pos < len(n.rows) {
		if n.isFocusable(pos) {
			return pos
		}

		pos += dir
	}

	return -1
}

// firstFocusable returns the index of the first focusable row, or -1.
func (n *Model) firstFocusable() int {
	for i := range n.rows {
		if n.isFocusable(i) {
			return i
		}
	}

	return -1
}

// lastFocusable returns the index of the last focusable row, or -1.
func (n *Model) lastFocusable() int {
	for i := len(n.rows) - 1; i >= 0; i-- {
		if n.isFocusable(i) {
			return i
		}
	}

	return -1
}

// totalLines returns the total number of lines in the joined row output.
func (n *Model) totalLines() int {
	total := 0

	for _, r := range n.rows {
		total += strings.Count(r.View().Content, "\n") + 1
	}

	return total
}

// Height returns the total number of lines in the current View() output.
func (n *Model) Height() int {
	return n.totalLines()
}

// CursorLine returns the line within View() output at which the active cursor
// sits. For a focused nested Navigator it recurses, so a parent viewport can
// scroll the correct line into view. Implements [CursorAware].
func (n *Model) CursorLine() int {
	if n.focused < 0 {
		return 0
	}

	start := 0

	for _, r := range n.rows[:n.focused] {
		start += strings.Count(r.View().Content, "\n") + 1
	}

	if ca, ok := n.rows[n.focused].(CursorAware); ok {
		return start + ca.CursorLine()
	}

	return start
}
