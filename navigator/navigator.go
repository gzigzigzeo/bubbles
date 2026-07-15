// Package navigator manages an ordered list of rows with keyboard focus.
// Each row is a [tea.Model] that may additionally implement [row.Focusable],
// [row.Disableable], and/or [row.FocusReceiver].
//
// The navigator keeps the focused row visible through an internal
// [ViewportController]. Callers can access the controller via
// [Model.ViewportController] to configure the viewport or attach a [Viewport].
package navigator

import (
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/gzigzigzeo/bubbles/navigator/focus"
	"github.com/gzigzigzeo/bubbles/navigator/row"
)

// keyUpBinding is the default key binding for moving focus up.
var keyUpBinding = key.NewBinding(
	key.WithKeys("up", "k"),
	key.WithHelp("↑/k", "up"),
)

// keyDownBinding is the default key binding for moving focus down.
var keyDownBinding = key.NewBinding(
	key.WithKeys("down", "j"),
	key.WithHelp("↓/j", "down"),
)

// Model manages an ordered list of rows with keyboard navigation.
// It renders rows as a flat joined string and uses an internal
// [ViewportController] to keep the focused row visible. The zero value is not
// usable; use [New].
type Model struct {
	ctrl  *focus.Controller
	coord *ViewportController // always non-nil after construction
}

// New creates a Navigator over rows. Focus does not wrap at boundaries by
// default. A default [ViewportController] is created automatically. Call
// [FocusFirst] or [FocusLast] to give initial focus, then call [Init].
func New(rows ...tea.Model) *Model {
	ctrl := focus.New(rows...)
	ctrl.SetPrevKeys("up", "k")
	ctrl.SetNextKeys("down", "j")

	nav := &Model{
		ctrl:  ctrl,
		coord: nil,
	}
	nav.coord = NewViewportController(nav)

	return nav
}

// Wrap enables wrap-at-boundaries focus mode: focus wraps from the last row
// back to the first (and vice versa) instead of keeping focus at boundaries.
func (n *Model) Wrap() {
	n.ctrl.SetWrap(true)
}

// ViewportController returns the navigator's internal viewport controller.
func (n *Model) ViewportController() *ViewportController {
	return n.coord
}

// FocusFirst focuses the first non-disabled [row.Focusable] row and scrolls it
// into view. Implements [row.FocusReceiver].
func (n *Model) FocusFirst() tea.Cmd {
	cmd := n.ctrl.FocusFirst()
	n.scrollToFocus()

	return cmd
}

// FocusLast focuses the last non-disabled [row.Focusable] row and scrolls it
// into view. Implements [row.FocusReceiver].
func (n *Model) FocusLast() tea.Cmd {
	cmd := n.ctrl.FocusLast()
	n.scrollToFocus()

	return cmd
}

// FocusIndex focuses the row at idx if it is focusable and scrolls it into
// view.
func (n *Model) FocusIndex(idx int) tea.Cmd {
	cmd := n.ctrl.FocusIndex(idx)
	n.scrollToFocus()

	return cmd
}

// Focus focuses the first non-disabled row. Implements [row.Focusable].
func (n *Model) Focus() tea.Cmd {
	return n.FocusFirst()
}

// Blur removes focus from the current row. Implements [row.Focusable].
func (n *Model) Blur() tea.Cmd {
	return n.ctrl.Blur()
}

// Focused reports whether any row holds focus. Implements [row.Focusable].
func (n *Model) Focused() bool {
	return n.ctrl.Focused()
}

// Init initializes all rows.
func (n *Model) Init() tea.Cmd {
	return n.ctrl.Init()
}

// Update handles keyboard navigation and routes other messages to the focused
// row. Up/Down (and vi aliases k/j) move focus between rows. When the focused
// row is a [row.FocusReceiver] (a nested Navigator), keys are passed through it;
// if it defocuses itself the outer Navigator shifts focus in the same cycle.
// The internal [ViewportController] is updated automatically so the focused
// row stays visible.
func (n *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return n, n.ctrl.Update(msg)
	}

	if n.ctrl.Focused() {
		if _, isReceiver := n.ctrl.CurrentItem().(row.FocusReceiver); isReceiver {
			oldCursor := n.CursorLine()
			cmd := n.updateFocusedReceiver(keyMsg)
			n.scrollAfterMove(oldCursor, n.navigationDir(keyMsg))

			return n, cmd
		}
	}

	oldCursor := n.CursorLine()
	dir := n.navigationDir(keyMsg)
	cmd := n.ctrl.Update(keyMsg)
	n.scrollAfterMove(oldCursor, dir)

	return n, cmd
}

// FocusedIndex returns the index of the currently focused row, or -1 if none
// is focused.
func (n *Model) FocusedIndex() int {
	return n.ctrl.FocusedIndex()
}

// Items returns the navigator's rows.
func (n *Model) Items() []tea.Model {
	return n.ctrl.Items()
}

// View renders all rows as a flat joined string. Pair with a viewport (via the
// internal [ViewportController]) to get height-clipped scrollable display.
func (n *Model) View() tea.View {
	items := n.ctrl.Items()
	rows := make([]string, len(items))

	for i, r := range items {
		rows[i] = r.View().Content
	}

	return tea.NewView(lipgloss.JoinVertical(lipgloss.Left, rows...))
}

// IsAtFirstFocusable reports whether the currently focused row is the first
// non-disabled focusable row.
func (n *Model) IsAtFirstFocusable() bool {
	return n.ctrl.IsAtFirstFocusable()
}

// IsAtLastFocusable reports whether the currently focused row is the last
// non-disabled focusable row.
func (n *Model) IsAtLastFocusable() bool {
	return n.ctrl.IsAtLastFocusable()
}

// Height returns the total number of lines in the current View() output.
func (n *Model) Height() int {
	return n.totalLines()
}

// CursorLine returns the line within View() output at which the active cursor
// sits. For a focused nested Navigator it recurses, so a parent viewport can
// scroll the correct line into view. Implements [row.CursorAware].
func (n *Model) CursorLine() int {
	focused := n.ctrl.FocusedIndex()
	if focused < 0 {
		return 0
	}

	items := n.ctrl.Items()
	start := 0

	for _, r := range items[:focused] {
		start += viewLineCount(r.View())
	}

	if ca, ok := items[focused].(row.CursorAware); ok {
		return start + ca.CursorLine()
	}

	return start
}

// navigationDir returns -1 for up keys, 1 for down keys, and 0 otherwise.
func (n *Model) navigationDir(keyMsg tea.KeyMsg) int {
	switch {
	case key.Matches(keyMsg, keyUpBinding):
		return -1
	case key.Matches(keyMsg, keyDownBinding):
		return 1
	}

	return 0
}

// scrollAfterMove updates the viewport controller after a potential focus
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

// scrollToFocus scrolls the internal controller to the current cursor line.
func (n *Model) scrollToFocus() {
	n.coord.scrollToFocus(n.CursorLine(), n.totalLines())
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

	current := n.ctrl.CurrentItem()
	focusable, isFocusable := current.(row.Focusable)
	wasFocused := isFocusable && focusable.Focused()

	updated, cmd := current.Update(keyMsg)
	items := n.ctrl.Items()
	focusedIdx := n.ctrl.FocusedIndex()
	items[focusedIdx] = updated

	if recoverCmd := n.recoverFocusIfDefocused(keyMsg, updated, wasFocused); recoverCmd != nil {
		return tea.Batch(cmd, recoverCmd)
	}

	return cmd
}

// exitAtBoundary returns a command to move focus out of the focused row when
// the row implements [row.BoundaryAware] and is at the relevant boundary.
func (n *Model) exitAtBoundary(keyMsg tea.KeyMsg) (tea.Cmd, bool) {
	boundary, ok := n.ctrl.CurrentItem().(row.BoundaryAware)
	if !ok {
		return nil, false
	}

	switch {
	case key.Matches(keyMsg, keyUpBinding):
		if boundary.IsAtFirstFocusable() {
			return n.move(-1), true
		}
	case key.Matches(keyMsg, keyDownBinding):
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
	if key.Matches(keyMsg, keyUpBinding) {
		dir = -1
	}

	return n.move(dir)
}

// move handles an up/down navigation key in dir (+1 = down, -1 = up).
func (n *Model) move(dir int) tea.Cmd {
	if dir > 0 {
		return n.ctrl.MoveNext()
	}

	return n.ctrl.MovePrev()
}

// viewLineCount returns the rendered height of a single row view.
func viewLineCount(v tea.View) int {
	return lipgloss.Height(v.Content)
}

// totalLines returns the total number of lines in the joined row output.
func (n *Model) totalLines() int {
	total := 0

	for _, r := range n.ctrl.Items() {
		total += viewLineCount(r.View())
	}

	return total
}
