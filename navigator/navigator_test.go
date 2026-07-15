package navigator_test

import (
	"testing"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/stretchr/testify/require"

	"github.com/gzigzigzeo/bubbles/navigator"
	"github.com/gzigzigzeo/bubbles/navigator/row"
	"github.com/gzigzigzeo/bubbles/navigator/rows/selectfield"
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

func sendKey(t *testing.T, nav *navigator.Model, key string) {
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

	_, _ = nav.Update(tea.KeyPressMsg(tea.Key{
		Code:        code,
		Text:        "",
		Mod:         0,
		ShiftedCode: 0,
		BaseCode:    0,
		IsRepeat:    false,
	}))
}

func TestFocusFirst(t *testing.T) {
	nav := navigator.New(
		testLabel("header"),
		&testItem{text: "Alpha", focused: false, disabled: false},
		&testItem{text: "Beta", focused: false, disabled: false},
	)
	_ = nav.FocusFirst()

	require.Equal(t, 1, nav.FocusedIndex())
	require.Equal(t, 1, nav.CursorLine())
}

func TestFocusLast(t *testing.T) {
	nav := navigator.New(
		testLabel("header"),
		&testItem{text: "Alpha", focused: false, disabled: false},
		&testItem{text: "Beta", focused: false, disabled: false},
	)
	_ = nav.FocusLast()

	require.Equal(t, 2, nav.FocusedIndex())
	require.Equal(t, 2, nav.CursorLine())
}

func TestOpenMode_KeepsFocusAtBoundaries(t *testing.T) {
	nav := navigator.New(
		&testItem{text: "Alpha", focused: false, disabled: false},
		&testItem{text: "Beta", focused: false, disabled: false},
	)
	_ = nav.FocusFirst()

	// Up at first focusable should keep focus on Alpha.
	sendKey(t, nav, "up")
	require.Equal(t, 0, nav.FocusedIndex())
	require.True(t, nav.IsAtFirstFocusable())

	// Move to Beta and press down; focus should stay on Beta.
	sendKey(t, nav, "down")
	sendKey(t, nav, "down")
	require.Equal(t, 1, nav.FocusedIndex())
	require.True(t, nav.IsAtLastFocusable())
}

func TestWrapMode_WrapsAtBoundaries(t *testing.T) {
	nav := navigator.New(
		&testItem{text: "Alpha", focused: false, disabled: false},
		&testItem{text: "Beta", focused: false, disabled: false},
	)
	nav.Wrap()
	_ = nav.FocusFirst()

	// Up at first should wrap to last.
	sendKey(t, nav, "up")
	require.Equal(t, 1, nav.FocusedIndex())

	// Down at last should wrap to first.
	sendKey(t, nav, "down")
	require.Equal(t, 0, nav.FocusedIndex())
}

func TestDisabledRows_AreSkipped(t *testing.T) {
	nav := navigator.New(
		&testItem{text: "Alpha", focused: false, disabled: false},
		&testItem{text: "Beta", focused: false, disabled: true},
		&testItem{text: "Gamma", focused: false, disabled: false},
	)
	_ = nav.FocusFirst()

	require.Equal(t, 0, nav.FocusedIndex())

	sendKey(t, nav, "down")

	require.Equal(t, 2, nav.FocusedIndex())
}

func TestFocusMovement(t *testing.T) {
	nav := navigator.New(
		&testItem{text: "Alpha", focused: false, disabled: false},
		&testItem{text: "Beta", focused: false, disabled: false},
		&testItem{text: "Gamma", focused: false, disabled: false},
	)
	_ = nav.FocusFirst()

	sendKey(t, nav, "down")
	require.Equal(t, 1, nav.FocusedIndex())

	sendKey(t, nav, "down")
	require.Equal(t, 2, nav.FocusedIndex())

	sendKey(t, nav, "up")
	require.Equal(t, 1, nav.FocusedIndex())
}

func TestNestedNavigator_BoundaryAware_LeavesInner(t *testing.T) {
	inner := navigator.New(
		&testItem{text: "Echo", focused: false, disabled: false},
		&testItem{text: "Foxtrot", focused: false, disabled: false},
		&testItem{text: "Golf", focused: false, disabled: false},
	)

	outer := navigator.New(
		&testItem{text: "Alpha", focused: false, disabled: false},
		inner,
		&testItem{text: "Bravo", focused: false, disabled: false},
	)
	_ = outer.FocusFirst()
	sendKey(t, outer, "down") // focus inner -> Echo

	require.Equal(t, 1, outer.FocusedIndex())
	require.Equal(t, 0, inner.FocusedIndex())

	// Move to Golf (last inner item).
	sendKey(t, outer, "down")
	sendKey(t, outer, "down")
	require.Equal(t, 2, inner.FocusedIndex())

	// Down at inner boundary should move focus to Bravo.
	sendKey(t, outer, "down")
	require.Equal(t, 2, outer.FocusedIndex())

	bravo, ok := outer.Items()[2].(*testItem)
	require.True(t, ok)
	require.True(t, bravo.focused)

	// Up from Bravo moves back into the inner navigator. Navigate to its first
	// item, then one more up crosses the inner boundary to Alpha.
	sendKey(t, outer, "up") // Bravo -> Golf
	sendKey(t, outer, "up") // Golf -> Foxtrot
	sendKey(t, outer, "up") // Foxtrot -> Echo
	require.Equal(t, 1, outer.FocusedIndex())
	require.Equal(t, 0, inner.FocusedIndex())

	sendKey(t, outer, "up") // Echo boundary -> Alpha
	require.Equal(t, 0, outer.FocusedIndex())
}

func TestNestedNavigator_BoundaryAware_OuterBoundaryScrolls(t *testing.T) {
	inner := navigator.New(
		&testItem{text: "Echo", focused: false, disabled: false},
		&testItem{text: "Foxtrot", focused: false, disabled: false},
	)

	outer := navigator.New(
		testLabel("header"),
		inner,
		testLabel("footer"),
	)
	outer.ViewportController().SetHeight(2)
	_ = outer.FocusFirst()
	sendKey(t, outer, "down") // focus inner -> Echo

	require.Equal(t, 1, outer.FocusedIndex())

	// Move to Foxtrot (last inner item).
	sendKey(t, outer, "down")
	require.Equal(t, 1, inner.FocusedIndex())

	// Down at inner boundary. Outer has no focusable below inner, so focus stays
	// on Foxtrot and the viewport scrolls to reveal the footer.
	sendKey(t, outer, "down")
	require.Equal(t, 1, outer.FocusedIndex())
	require.Equal(t, 1, inner.FocusedIndex())
	require.Equal(t, 2, outer.ViewportController().YOffset())
}

func TestNestedNavigator_DelegatesToOuter(t *testing.T) {
	// Inner navigator that defocuses itself when reaching its last item.
	inner := &defocusingNavigator{
		rows: []tea.Model{
			&testItem{text: "Echo", focused: false, disabled: false},
			&testItem{text: "Foxtrot", focused: false, disabled: false},
		},
		focused: -1,
	}

	outer := navigator.New(
		&testItem{text: "Alpha", focused: false, disabled: false},
		inner,
		&testItem{text: "Bravo", focused: false, disabled: false},
	)
	_ = outer.FocusFirst()
	sendKey(t, outer, "down") // focus inner

	require.Equal(t, 1, outer.FocusedIndex())

	// First down inside inner moves to Foxtrot.
	sendKey(t, outer, "down")
	require.Equal(t, 1, inner.focused)

	// Second down causes inner to defocus; outer moves to Bravo.
	sendKey(t, outer, "down")
	require.Equal(t, 2, outer.FocusedIndex())
}

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
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return d, nil
	}

	defocusKeyDown := key.NewBinding(key.WithKeys("down"))

	if key.Matches(keyMsg, defocusKeyDown) && d.focused == len(d.rows)-1 {
		if f, ok := d.rows[d.focused].(row.Focusable); ok {
			_ = f.Blur()
		}

		d.focused = -1

		return d, nil
	}

	if key.Matches(keyMsg, defocusKeyDown) && d.focused+1 < len(d.rows) {
		if f, ok := d.rows[d.focused].(row.Focusable); ok {
			_ = f.Blur()
		}

		d.focused++

		if f, ok := d.rows[d.focused].(row.Focusable); ok {
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
	if f, ok := d.rows[0].(row.Focusable); ok {
		_ = f.Focus()
	}

	return nil
}

func (d *defocusingNavigator) Blur() tea.Cmd {
	if d.focused >= 0 && d.focused < len(d.rows) {
		if f, ok := d.rows[d.focused].(row.Focusable); ok {
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
	if f, ok := d.rows[d.focused].(row.Focusable); ok {
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

// spaceItem is a focusable row that records whether it received a space key.
type spaceItem struct {
	text    string
	focused bool
	toggled bool
}

func (s *spaceItem) Init() tea.Cmd {
	return nil
}

func (s *spaceItem) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok || !s.focused {
		return s, nil
	}

	spaceKey := key.NewBinding(key.WithKeys("space"))
	if key.Matches(keyMsg, spaceKey) {
		s.toggled = true
	}

	return s, nil
}

func (s *spaceItem) View() tea.View {
	return tea.NewView(s.text)
}

func (s *spaceItem) Focus() tea.Cmd {
	s.focused = true

	return nil
}

func (s *spaceItem) Blur() tea.Cmd {
	s.focused = false

	return nil
}

func (s *spaceItem) Focused() bool {
	return s.focused
}

func TestSpaceForwardsToFocusedRow(t *testing.T) {
	row := &spaceItem{text: "Alpha", focused: false, toggled: false}
	nav := navigator.New(testLabel("header"), row)
	_ = nav.FocusFirst()

	sendKey(t, nav, "space")

	require.True(t, row.toggled)
}

func TestSelectField_OpenPicker_ScrollsSelectedIntoView(t *testing.T) {
	// Viewport height 3, select field at line 2 with 6 options. The committed
	// option is far enough down that the picker dropdown would extend below the
	// viewport; opening it must scroll so the selected option is visible.
	selectRow := selectfield.NewFromStrings([]string{"a", "b", "c", "d", "e", "f"})
	selectRow.Set("e")

	nav := navigator.NewBuilder().
		WithItems(testLabel("header"), testLabel("spacer"), selectRow).
		WithControllerItems(selectRow.Controller()).
		WithHeight(3).
		Build()
	_ = nav.FocusFirst()

	// Move focus onto the select field (line 2).
	sendKey(t, nav, "down")
	sendKey(t, nav, "down")

	require.Equal(t, 2, nav.FocusedIndex())

	// Open the picker. The select field returns a LockFocusMsg command that the
	// navigator handles in the next update.
	_, cmd := nav.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyEnter}))
	require.NotNil(t, cmd)

	_, _ = nav.Update(cmd())

	// The committed option "e" is at picker index 4, so the outer cursor line
	// becomes 2 + 1 + 4 = 7. Total content becomes 2 + 1 + 6 = 9.
	cursor := nav.CursorLine()
	offset := nav.ViewportController().YOffset()
	height := nav.ViewportController().Height()

	require.Greater(t, cursor, 2)
	require.GreaterOrEqual(t, cursor, offset)
	require.Less(t, cursor, offset+height)
}
