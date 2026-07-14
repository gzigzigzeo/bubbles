package focus

import (
	"testing"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/require"

	"github.com/gzigzigzeo/bubbles/navigator/rows/row"
)

// testItem is a focusable, optionally disabled item.
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

func TestController_FocusFirst(t *testing.T) {
	c := New(
		&testItem{text: "Alpha"},
		&testItem{text: "Beta", disabled: true},
		&testItem{text: "Gamma"},
	)

	_ = c.FocusFirst()

	require.Equal(t, 0, c.FocusedIndex())
	require.True(t, c.Items()[0].(*testItem).focused)
}

func TestController_FocusLast(t *testing.T) {
	c := New(
		&testItem{text: "Alpha"},
		&testItem{text: "Beta", disabled: true},
		&testItem{text: "Gamma"},
	)

	_ = c.FocusLast()

	require.Equal(t, 2, c.FocusedIndex())
	require.True(t, c.Items()[2].(*testItem).focused)
}

func TestController_MoveNext_skipsDisabled(t *testing.T) {
	c := New(
		&testItem{text: "Alpha"},
		&testItem{text: "Beta", disabled: true},
		&testItem{text: "Gamma"},
	)

	_ = c.FocusFirst()
	_ = c.MoveNext()

	require.Equal(t, 2, c.FocusedIndex())
}

func TestController_MovePrev_skipsDisabled(t *testing.T) {
	c := New(
		&testItem{text: "Alpha"},
		&testItem{text: "Beta", disabled: true},
		&testItem{text: "Gamma"},
	)

	_ = c.FocusLast()
	_ = c.MovePrev()

	require.Equal(t, 0, c.FocusedIndex())
}

func TestController_MoveNext_stopsAtBoundaryWithoutWrap(t *testing.T) {
	c := New(
		&testItem{text: "Alpha"},
		&testItem{text: "Beta"},
	)

	_ = c.FocusLast()
	_ = c.MoveNext()

	require.Equal(t, 1, c.FocusedIndex())
	require.True(t, c.IsAtLastFocusable())
}

func TestController_MoveNext_wrapsAtBoundary(t *testing.T) {
	c := New(
		&testItem{text: "Alpha"},
		&testItem{text: "Beta"},
	)
	c.SetWrap(true)

	_ = c.FocusLast()
	_ = c.MoveNext()

	require.Equal(t, 0, c.FocusedIndex())
}

func TestController_Update_handlesNextKey(t *testing.T) {
	c := New(
		&testItem{text: "Alpha"},
		&testItem{text: "Beta"},
	)
	c.SetNextKeys("down", "j")
	c.SetPrevKeys("up", "k")

	_ = c.FocusFirst()
	_ = c.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyDown}))

	require.Equal(t, 1, c.FocusedIndex())
}

func TestController_Update_handlesPrevKey(t *testing.T) {
	c := New(
		&testItem{text: "Alpha"},
		&testItem{text: "Beta"},
	)
	c.SetNextKeys("down", "j")
	c.SetPrevKeys("up", "k")

	_ = c.FocusLast()
	_ = c.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyUp}))

	require.Equal(t, 0, c.FocusedIndex())
}

func TestController_Update_forwardsOtherKeysToFocusedItem(t *testing.T) {
	space := &spaceItem{text: "Alpha"}
	c := New(space)
	c.SetNextKeys("down")
	c.SetPrevKeys("up")

	_ = c.FocusFirst()
	_ = c.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeySpace}))

	require.True(t, space.toggled)
}

func TestController_Blur_removesFocus(t *testing.T) {
	c := New(&testItem{text: "Alpha"})

	_ = c.FocusFirst()
	_ = c.Blur()

	require.False(t, c.Focused())
	require.Equal(t, -1, c.FocusedIndex())
}

// spaceItem records whether it received a space key.
type spaceItem struct {
	text    string
	focused bool
	toggled bool
}

func (s *spaceItem) Init() tea.Cmd {
	return nil
}

func (s *spaceItem) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	km, ok := msg.(tea.KeyMsg)
	if !ok || !s.focused {
		return s, nil
	}

	spaceKey := key.NewBinding(key.WithKeys("space"))
	if key.Matches(km, spaceKey) {
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

var _ row.Focusable = (*testItem)(nil)
var _ row.Disableable = (*testItem)(nil)
