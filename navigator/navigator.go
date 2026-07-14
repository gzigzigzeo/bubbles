// Package navigator manages an ordered list of rows with keyboard focus.
// Each row is a [tea.Model] that may additionally implement [row.Focusable],
// [row.Disableable], and/or [row.FocusReceiver].
//
// The navigator keeps the focused row visible through an internal
// [ViewportCoordinator]. Callers can access the coordinator via
// [Model.ViewportCoordinator] to configure the viewport or attach a [Viewport].
package navigator

import (
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/gzigzigzeo/bubbles/navigator/rows/row"
)

const focusCmdsCapacity = 2

// keyUpBinding returns the key binding for moving focus up.
func keyUpBinding() key.Binding {
	return key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", "up"),
	)
}

// keyDownBinding returns the key binding for moving focus down.
func keyDownBinding() key.Binding {
	return key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", "down"),
	)
}

// Model manages an ordered list of rows with keyboard navigation.
// It renders rows as a flat joined string and uses an internal
// [ViewportCoordinator] to keep the focused row visible. The zero value is not
// usable; use [New].
type Model struct {
	rows    []tea.Model
	focused int                  // index into rows; -1 = none focused
	wrap    bool                 // true: wrap at boundaries; false: keep focus at boundaries
	coord   *ViewportCoordinator // always non-nil after construction
}

// New creates a Navigator over rows. Focus does not wrap at boundaries by
// default. A default [ViewportCoordinator] is created automatically; replace it
// with [Model.SetViewportCoordinator] if needed. Call [FocusFirst] or
// [FocusLast] to give initial focus, then call [Init].
func New(rows ...tea.Model) *Model {
	nav := &Model{
		rows:    rows,
		focused: -1,
		wrap:    false,
		coord:   nil,
	}
	nav.coord = NewViewportCoordinator(nav)

	return nav
}

// Wrap enables wrap-at-boundaries focus mode: focus wraps from the last row
// back to the first (and vice versa) instead of keeping focus at boundaries.
func (n *Model) Wrap() {
	n.wrap = true
}

// ViewportCoordinator returns the navigator's internal viewport coordinator.
func (n *Model) ViewportCoordinator() *ViewportCoordinator {
	return n.coord
}

// FocusFirst focuses the first non-disabled [row.Focusable] row and scrolls it into
// view. Implements [row.FocusReceiver].
func (n *Model) FocusFirst() tea.Cmd {
	cmd := n.focusIndexDir(n.firstFocusable(), 1)
	n.scrollToFocus()

	return cmd
}

// FocusLast focuses the last non-disabled [row.Focusable] row and scrolls it into
// view. Implements [row.FocusReceiver].
func (n *Model) FocusLast() tea.Cmd {
	cmd := n.focusIndexDir(n.lastFocusable(), -1)
	n.scrollToFocus()

	return cmd
}

// Focus focuses the first non-disabled row. Implements [row.Focusable].
func (n *Model) Focus() tea.Cmd {
	return n.FocusFirst()
}

// Blur removes focus from the current row. Implements [row.Focusable].
func (n *Model) Blur() tea.Cmd {
	return n.blurCurrent()
}

// Focused reports whether any row holds focus. Implements [row.Focusable].
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
// row is a [row.FocusReceiver] (a nested Navigator), keys are passed through it;
// if it defocuses itself the outer Navigator shifts focus in the same cycle.
// The internal [ViewportCoordinator] is updated automatically so the focused
// row stays visible.
func (n *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		if n.focused >= 0 {
			cmd := n.updateFocused(msg)

			return n, cmd
		}

		return n, nil
	}

	if n.focused >= 0 {
		if _, isReceiver := n.rows[n.focused].(row.FocusReceiver); isReceiver {
			oldCursor := n.CursorLine()
			cmd := n.updateFocusedReceiver(keyMsg)
			n.scrollAfterMove(oldCursor, n.navigationDir(keyMsg))

			return n, cmd
		}
	}

	oldCursor := n.CursorLine()
	dir := n.navigationDir(keyMsg)

	var cmd tea.Cmd

	switch {
	case key.Matches(keyMsg, keyUpBinding()):
		cmd = n.move(-1)
	case key.Matches(keyMsg, keyDownBinding()):
		cmd = n.move(1)
	default:
		if n.focused >= 0 {
			cmd = n.updateFocused(keyMsg)

			return n, cmd
		}

		return n, nil
	}

	n.scrollAfterMove(oldCursor, dir)

	return n, cmd
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

// Height returns the total number of lines in the current View() output.
func (n *Model) Height() int {
	return n.totalLines()
}

// CursorLine returns the line within View() output at which the active cursor
// sits. For a focused nested Navigator it recurses, so a parent viewport can
// scroll the correct line into view. Implements [row.CursorAware].
func (n *Model) CursorLine() int {
	if n.focused < 0 {
		return 0
	}

	start := 0

	for _, r := range n.rows[:n.focused] {
		start += viewLineCount(r.View())
	}

	if ca, ok := n.rows[n.focused].(row.CursorAware); ok {
		return start + ca.CursorLine()
	}

	return start
}

// navigationDir returns -1 for up keys, 1 for down keys, and 0 otherwise.
func (n *Model) navigationDir(keyMsg tea.KeyMsg) int {
	switch {
	case key.Matches(keyMsg, keyUpBinding()):
		return -1
	case key.Matches(keyMsg, keyDownBinding()):
		return 1
	}

	return 0
}

// scrollAfterMove updates the viewport coordinator after a potential focus
// change. If the cursor moved the viewport scrolls to keep it visible; if the
// cursor stayed at a boundary the viewport scrolls one line in dir.
func (n *Model) scrollAfterMove(oldCursor, dir int) {
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
	n.coord.scrollToFocus(n.CursorLine(), n.totalLines())
}

// updateFocused forwards msg to the currently focused row and returns its
// command.
func (n *Model) updateFocused(msg tea.Msg) tea.Cmd {
	updated, cmd := n.rows[n.focused].Update(msg)
	n.rows[n.focused] = updated

	return cmd
}

// updateFocusedReceiver forwards a key to the focused [row.FocusReceiver] row.
// If the row reports it is at a boundary via [row.BoundaryAware], the parent
// handles the key instead so focus can leave the nested navigator. If the row
// defocuses itself as a result, focus is shifted within the same Update cycle —
// no message round-trip needed.
func (n *Model) updateFocusedReceiver(keyMsg tea.KeyMsg) tea.Cmd {
	if cmd, ok := n.exitAtBoundary(keyMsg); ok {
		return cmd
	}

	focusable, isFocusable := n.rows[n.focused].(row.Focusable)
	wasFocused := isFocusable && focusable.Focused()

	updated, cmd := n.rows[n.focused].Update(keyMsg)
	n.rows[n.focused] = updated

	if recoverCmd := n.recoverFocusIfDefocused(keyMsg, updated, wasFocused); recoverCmd != nil {
		return tea.Batch(cmd, recoverCmd)
	}

	return cmd
}

// exitAtBoundary returns a command to move focus out of the focused row when
// the row implements [row.BoundaryAware] and is at the relevant boundary.
func (n *Model) exitAtBoundary(keyMsg tea.KeyMsg) (tea.Cmd, bool) {
	boundary, ok := n.rows[n.focused].(row.BoundaryAware)
	if !ok {
		return nil, false
	}

	switch {
	case key.Matches(keyMsg, keyUpBinding()):
		if boundary.IsAtFirstFocusable() {
			return n.move(-1), true
		}
	case key.Matches(keyMsg, keyDownBinding()):
		if boundary.IsAtLastFocusable() {
			return n.move(1), true
		}
	}

	return nil, false
}

// recoverFocusIfDefocused returns a command to shift focus in the same
// direction when the focused row defocused itself during the Update.
func (n *Model) recoverFocusIfDefocused(keyMsg tea.KeyMsg, updated tea.Model, wasFocused bool) tea.Cmd {
	if !wasFocused {
		return nil
	}

	if nf, ok := updated.(row.Focusable); !ok || nf.Focused() {
		return nil
	}

	dir := 1
	if key.Matches(keyMsg, keyUpBinding()) {
		dir = -1
	}

	return n.move(dir)
}

// move handles an up/down navigation key in dir (+1 = down, -1 = up).
// When a focusable row exists in that direction, focus moves to it. At a
// boundary in non-wrapping mode focus stays put; in wrapping mode focus wraps.
func (n *Model) move(dir int) tea.Cmd {
	next := n.nextFocusable(n.focused, dir)

	if next < 0 {
		if !n.wrap {
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

// focusIndexDir blurs the current row and focuses the row at idx, using dir
// to select FocusFirst (dir >= 0) vs FocusLast (dir < 0) for nested
// [row.FocusReceiver] rows.
func (n *Model) focusIndexDir(idx int, dir int) tea.Cmd {
	if idx < 0 || idx >= len(n.rows) {
		return nil
	}

	cmds := make([]tea.Cmd, 0, focusCmdsCapacity)

	if blur := n.blurCurrent(); blur != nil {
		cmds = append(cmds, blur)
	}

	n.focused = idx

	if focusCmd := n.focusCmdForRow(idx, dir); focusCmd != nil {
		cmds = append(cmds, focusCmd)
	}

	return tea.Batch(cmds...)
}

// focusCmdForRow returns the command to focus the row at idx, choosing
// FocusFirst (dir >= 0) vs FocusLast (dir < 0) for nested [row.FocusReceiver] rows.
func (n *Model) focusCmdForRow(idx int, dir int) tea.Cmd {
	focusable, ok := n.rows[idx].(row.Focusable)
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

// blurCurrent blurs the currently focused row and returns its command.
func (n *Model) blurCurrent() tea.Cmd {
	if n.focused < 0 || n.focused >= len(n.rows) {
		return nil
	}

	if f, ok := n.rows[n.focused].(row.Focusable); ok {
		return f.Blur()
	}

	return nil
}

// isFocusable reports whether the row at idx can receive focus: it must
// implement [row.Focusable] and must not be disabled.
func (n *Model) isFocusable(idx int) bool {
	if idx < 0 || idx >= len(n.rows) {
		return false
	}

	r := n.rows[idx]

	if _, ok := r.(row.Focusable); !ok {
		return false
	}

	if d, ok := r.(row.Disableable); ok && d.Disabled() {
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

// viewLineCount returns the rendered height of a single row view.
func viewLineCount(v tea.View) int {
	return lipgloss.Height(v.Content)
}

// totalLines returns the total number of lines in the joined row output.
func (n *Model) totalLines() int {
	total := 0

	for _, r := range n.rows {
		total += viewLineCount(r.View())
	}

	return total
}
