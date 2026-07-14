package navigator

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/bubbles/v2/key"
	"charm.land/lipgloss/v2"
	"github.com/stretchr/testify/require"
)

// testLabel is a non-focusable display row.
type testLabel string

func (l testLabel) Init() tea.Cmd {
	return nil
}

func (l testLabel) Update(_ tea.Msg) (tea.Model, tea.Cmd) {
	return l, nil
}

func (l testLabel) View() tea.View {
	return tea.NewView(string(l))
}

// testItem is a focusable, optionally disabled row.
type testItem struct {
	text     string
	focused  bool
	disabled bool
}

func (it *testItem) Init() tea.Cmd {
	return nil
}

func (it *testItem) Update(_ tea.Msg) (tea.Model, tea.Cmd) {
	return it, nil
}

func (it *testItem) View() tea.View {
	return tea.NewView(it.text)
}

func (it *testItem) Focus() tea.Cmd {
	it.focused = true
	return nil
}

func (it *testItem) Blur() tea.Cmd {
	it.focused = false
	return nil
}

func (it *testItem) Focused() bool {
	return it.focused
}

func (it *testItem) Enable() tea.Cmd {
	it.disabled = false
	return nil
}

func (it *testItem) Disable() tea.Cmd {
	it.disabled = true
	return nil
}

func (it *testItem) Disabled() bool {
	return it.disabled
}

func sendKey(t *testing.T, n *Model, key string) {
	t.Helper()

	var code rune
	switch key {
	case "up":
		code = tea.KeyUp
	case "down":
		code = tea.KeyDown
	case "space":
		code = tea.KeySpace
	default:
		code = rune(key[0])
	}

	_, _ = n.Update(tea.KeyPressMsg(tea.Key{Code: code}))
}

func TestFocusFirst(t *testing.T) {
	n := New(
		testLabel("header"),
		&testItem{text: "Alpha"},
		&testItem{text: "Beta"},
	)
	_ = n.FocusFirst()

	require.Equal(t, 1, n.focused)
	require.Equal(t, 1, n.CursorLine())
}

func TestFocusLast(t *testing.T) {
	n := New(
		testLabel("header"),
		&testItem{text: "Alpha"},
		&testItem{text: "Beta"},
	)
	_ = n.FocusLast()

	require.Equal(t, 2, n.focused)
	require.Equal(t, 2, n.CursorLine())
}

func TestOpenMode_KeepsFocusAtBoundaries(t *testing.T) {
	n := New(
		&testItem{text: "Alpha"},
		&testItem{text: "Beta"},
	)
	_ = n.FocusFirst()

	// Up at first focusable should keep focus on Alpha.
	sendKey(t, n, "up")
	require.Equal(t, 0, n.focused)
	require.True(t, n.IsAtFirstFocusable())

	// Move to Beta and press down; focus should stay on Beta.
	sendKey(t, n, "down")
	sendKey(t, n, "down")
	require.Equal(t, 1, n.focused)
	require.True(t, n.IsAtLastFocusable())
}

func TestClosedMode_WrapsAtBoundaries(t *testing.T) {
	n := New(
		&testItem{text: "Alpha"},
		&testItem{text: "Beta"},
	)
	n.Closed()
	_ = n.FocusFirst()

	// Up at first should wrap to last.
	sendKey(t, n, "up")
	require.Equal(t, 1, n.focused)

	// Down at last should wrap to first.
	sendKey(t, n, "down")
	require.Equal(t, 0, n.focused)
}

func TestDisabledRows_AreSkipped(t *testing.T) {
	n := New(
		&testItem{text: "Alpha"},
		&testItem{text: "Beta", disabled: true},
		&testItem{text: "Gamma"},
	)
	_ = n.FocusFirst()

	require.Equal(t, 0, n.focused)

	sendKey(t, n, "down")

	require.Equal(t, 2, n.focused)
}

func TestFocusMovement(t *testing.T) {
	n := New(
		&testItem{text: "Alpha"},
		&testItem{text: "Beta"},
		&testItem{text: "Gamma"},
	)
	_ = n.FocusFirst()

	sendKey(t, n, "down")
	require.Equal(t, 1, n.focused)

	sendKey(t, n, "down")
	require.Equal(t, 2, n.focused)

	sendKey(t, n, "up")
	require.Equal(t, 1, n.focused)
}

func TestNestedNavigator_BoundaryAware_LeavesInner(t *testing.T) {
	inner := New(
		&testItem{text: "Echo"},
		&testItem{text: "Foxtrot"},
		&testItem{text: "Golf"},
	)

	outer := New(
		&testItem{text: "Alpha"},
		inner,
		&testItem{text: "Bravo"},
	)
	_ = outer.FocusFirst()
	sendKey(t, outer, "down") // focus inner -> Echo

	require.Equal(t, 1, outer.focused)
	require.Equal(t, 0, inner.focused)

	// Move to Golf (last inner item).
	sendKey(t, outer, "down")
	sendKey(t, outer, "down")
	require.Equal(t, 2, inner.focused)

	// Down at inner boundary should move focus to Bravo.
	sendKey(t, outer, "down")
	require.Equal(t, 2, outer.focused)
	require.True(t, outer.rows[2].(*testItem).focused)

	// Up from Bravo moves back into the inner navigator. Navigate to its first
	// item, then one more up crosses the inner boundary to Alpha.
	sendKey(t, outer, "up") // Bravo -> Golf
	sendKey(t, outer, "up") // Golf -> Foxtrot
	sendKey(t, outer, "up") // Foxtrot -> Echo
	require.Equal(t, 1, outer.focused)
	require.Equal(t, 0, inner.focused)

	sendKey(t, outer, "up") // Echo boundary -> Alpha
	require.Equal(t, 0, outer.focused)
}

func TestNestedNavigator_BoundaryAware_OuterBoundaryScrolls(t *testing.T) {
	inner := New(
		&testItem{text: "Echo"},
		&testItem{text: "Foxtrot"},
	)

	outer := New(
		testLabel("header"),
		inner,
		testLabel("footer"),
	)
	outer.ViewportCoordinator().SetHeight(2)
	_ = outer.FocusFirst()
	sendKey(t, outer, "down") // focus inner -> Echo

	require.Equal(t, 1, outer.focused)

	// Move to Foxtrot (last inner item).
	sendKey(t, outer, "down")
	require.Equal(t, 1, inner.focused)

	// Down at inner boundary. Outer has no focusable below inner, so focus stays
	// on Foxtrot and the viewport scrolls to reveal the footer.
	sendKey(t, outer, "down")
	require.Equal(t, 1, outer.focused)
	require.Equal(t, 1, inner.focused)
	require.Equal(t, 2, outer.ViewportCoordinator().YOffset())
}

func TestNestedNavigator_DelegatesToOuter(t *testing.T) {
	// Inner navigator that defocuses itself when reaching its last item.
	inner := &defocusingNavigator{
		rows: []tea.Model{
			&testItem{text: "Echo"},
			&testItem{text: "Foxtrot"},
		},
	}

	outer := New(
		&testItem{text: "Alpha"},
		inner,
		&testItem{text: "Bravo"},
	)
	_ = outer.FocusFirst()
	sendKey(t, outer, "down") // focus inner

	require.Equal(t, 1, outer.focused)

	// First down inside inner moves to Foxtrot.
	sendKey(t, outer, "down")
	require.Equal(t, 1, inner.focused)

	// Second down causes inner to defocus; outer moves to Bravo.
	sendKey(t, outer, "down")
	require.Equal(t, 2, outer.focused)
}

var defocusKeyDown = key.NewBinding(
	key.WithKeys("down"),
)

// defocusingNavigator is a test double that behaves like a FocusReceiver and
// defocuses when the user presses down at its last item.
type defocusingNavigator struct {
	rows    []tea.Model
	focused int
}

func (d *defocusingNavigator) Init() tea.Cmd {
	return nil
}

func (d *defocusingNavigator) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	km, ok := msg.(tea.KeyMsg)
	if !ok {
		return d, nil
	}

	if key.Matches(km, defocusKeyDown) && d.focused == len(d.rows)-1 {
		if f, ok := d.rows[d.focused].(Focusable); ok {
			_ = f.Blur()
		}
		d.focused = -1
		return d, nil
	}

	if key.Matches(km, defocusKeyDown) && d.focused+1 < len(d.rows) {
		if f, ok := d.rows[d.focused].(Focusable); ok {
			_ = f.Blur()
		}
		d.focused++
		if f, ok := d.rows[d.focused].(Focusable); ok {
			_ = f.Focus()
		}
	}

	return d, nil
}

func (d *defocusingNavigator) View() tea.View {
	rows := make([]string, len(d.rows))
	for i, r := range d.rows {
		rows[i] = r.View().Content
	}

	return tea.NewView(lipgloss.JoinVertical(lipgloss.Left, rows...))
}

func (d *defocusingNavigator) Focus() tea.Cmd {
	d.focused = 0
	if f, ok := d.rows[0].(Focusable); ok {
		_ = f.Focus()
	}

	return nil
}

func (d *defocusingNavigator) Blur() tea.Cmd {
	if d.focused >= 0 && d.focused < len(d.rows) {
		if f, ok := d.rows[d.focused].(Focusable); ok {
			_ = f.Blur()
		}
	}
	d.focused = -1

	return nil
}

func (d *defocusingNavigator) Focused() bool {
	return d.focused >= 0
}

func (d *defocusingNavigator) FocusFirst() tea.Cmd {
	return d.Focus()
}

func (d *defocusingNavigator) FocusLast() tea.Cmd {
	d.focused = len(d.rows) - 1
	if f, ok := d.rows[d.focused].(Focusable); ok {
		_ = f.Focus()
	}

	return nil
}

func (d *defocusingNavigator) CursorLine() int {
	if d.focused < 0 {
		return 0
	}

	return d.focused
}
