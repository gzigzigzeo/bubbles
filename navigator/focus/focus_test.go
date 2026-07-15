package focus_test

import (
	"testing"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/require"

	"github.com/gzigzigzeo/bubbles/navigator/focus"
	"github.com/gzigzigzeo/bubbles/navigator/row"
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
	ctrl := focus.New(
		&testItem{text: "Alpha", focused: false, disabled: false},
		&testItem{text: "Beta", focused: false, disabled: true},
		&testItem{text: "Gamma", focused: false, disabled: false},
	)

	_ = ctrl.FocusFirst()

	require.Equal(t, 0, ctrl.FocusedIndex())

	first, ok := ctrl.Items()[0].(*testItem)
	require.True(t, ok)
	require.True(t, first.focused)
}

func TestController_FocusLast(t *testing.T) {
	ctrl := focus.New(
		&testItem{text: "Alpha", focused: false, disabled: false},
		&testItem{text: "Beta", focused: false, disabled: true},
		&testItem{text: "Gamma", focused: false, disabled: false},
	)

	_ = ctrl.FocusLast()

	require.Equal(t, 2, ctrl.FocusedIndex())

	last, ok := ctrl.Items()[2].(*testItem)
	require.True(t, ok)
	require.True(t, last.focused)
}

func TestController_MoveNext_skipsDisabled(t *testing.T) {
	ctrl := focus.New(
		&testItem{text: "Alpha", focused: false, disabled: false},
		&testItem{text: "Beta", focused: false, disabled: true},
		&testItem{text: "Gamma", focused: false, disabled: false},
	)

	_ = ctrl.FocusFirst()
	_ = ctrl.MoveNext()

	require.Equal(t, 2, ctrl.FocusedIndex())
}

func TestController_MovePrev_skipsDisabled(t *testing.T) {
	ctrl := focus.New(
		&testItem{text: "Alpha", focused: false, disabled: false},
		&testItem{text: "Beta", focused: false, disabled: true},
		&testItem{text: "Gamma", focused: false, disabled: false},
	)

	_ = ctrl.FocusLast()
	_ = ctrl.MovePrev()

	require.Equal(t, 0, ctrl.FocusedIndex())
}

func TestController_MoveNext_stopsAtBoundaryWithoutWrap(t *testing.T) {
	ctrl := focus.New(
		&testItem{text: "Alpha", focused: false, disabled: false},
		&testItem{text: "Beta", focused: false, disabled: false},
	)

	_ = ctrl.FocusLast()
	_ = ctrl.MoveNext()

	require.Equal(t, 1, ctrl.FocusedIndex())
	require.True(t, ctrl.IsAtLastFocusable())
}

func TestController_MoveNext_wrapsAtBoundary(t *testing.T) {
	ctrl := focus.New(
		&testItem{text: "Alpha", focused: false, disabled: false},
		&testItem{text: "Beta", focused: false, disabled: false},
	)
	ctrl.SetWrap(true)

	_ = ctrl.FocusLast()
	_ = ctrl.MoveNext()

	require.Equal(t, 0, ctrl.FocusedIndex())
}

func TestController_Update_handlesNextKey(t *testing.T) {
	ctrl := focus.New(
		&testItem{text: "Alpha", focused: false, disabled: false},
		&testItem{text: "Beta", focused: false, disabled: false},
	)
	ctrl.SetNextKeys("down", "j")
	ctrl.SetPrevKeys("up", "k")

	_ = ctrl.FocusFirst()
	_ = ctrl.Update(tea.KeyPressMsg(tea.Key{
		Code:        tea.KeyDown,
		Text:        "",
		Mod:         0,
		ShiftedCode: 0,
		BaseCode:    0,
		IsRepeat:    false,
	}))

	require.Equal(t, 1, ctrl.FocusedIndex())
}

func TestController_Update_handlesPrevKey(t *testing.T) {
	ctrl := focus.New(
		&testItem{text: "Alpha", focused: false, disabled: false},
		&testItem{text: "Beta", focused: false, disabled: false},
	)
	ctrl.SetNextKeys("down", "j")
	ctrl.SetPrevKeys("up", "k")

	_ = ctrl.FocusLast()
	_ = ctrl.Update(tea.KeyPressMsg(tea.Key{
		Code:        tea.KeyUp,
		Text:        "",
		Mod:         0,
		ShiftedCode: 0,
		BaseCode:    0,
		IsRepeat:    false,
	}))

	require.Equal(t, 0, ctrl.FocusedIndex())
}

func TestController_Update_forwardsOtherKeysToFocusedItem(t *testing.T) {
	space := &spaceItem{text: "Alpha", focused: false, toggled: false}
	ctrl := focus.New(space)
	ctrl.SetNextKeys("down")
	ctrl.SetPrevKeys("up")

	_ = ctrl.FocusFirst()
	_ = ctrl.Update(tea.KeyPressMsg(tea.Key{
		Code:        tea.KeySpace,
		Text:        "",
		Mod:         0,
		ShiftedCode: 0,
		BaseCode:    0,
		IsRepeat:    false,
	}))

	require.True(t, space.toggled)
}

func TestController_Blur_removesFocus(t *testing.T) {
	ctrl := focus.New(&testItem{text: "Alpha", focused: false, disabled: false})

	_ = ctrl.FocusFirst()
	_ = ctrl.Blur()

	require.False(t, ctrl.Focused())
	require.Equal(t, -1, ctrl.FocusedIndex())
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

var _ row.Focusable = (*testItem)(nil)
var _ row.Disableable = (*testItem)(nil)
