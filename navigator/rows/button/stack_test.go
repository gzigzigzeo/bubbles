package button

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/require"

	"github.com/gzigzigzeo/bubbles/navigator"
)

func TestStack_View_rendersButtonsHorizontally(t *testing.T) {
	stack := NewStack(
		New("A", "msg-a"),
		New("B", "msg-b"),
	)

	view := stack.View().Content

	require.Contains(t, view, "A")
	require.Contains(t, view, "B")
}

func TestStack_Focus_firstButton(t *testing.T) {
	stack := NewStack(
		New("A", "msg-a"),
		New("B", "msg-b"),
	)

	_ = stack.Focus()

	require.True(t, stack.Focused())
	require.Equal(t, 0, stack.Controller.FocusedIndex())
}

func TestStack_Update_movesFocusRight(t *testing.T) {
	stack := NewStack(
		New("A", "msg-a"),
		New("B", "msg-b"),
	)

	_ = stack.Focus()
	_, _ = stack.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyRight}))

	require.Equal(t, 1, stack.Controller.FocusedIndex())
}

func TestStack_Update_movesFocusLeft(t *testing.T) {
	stack := NewStack(
		New("A", "msg-a"),
		New("B", "msg-b"),
	)

	_ = stack.Focus()
	_, _ = stack.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyRight}))
	_, _ = stack.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyLeft}))

	require.Equal(t, 0, stack.Controller.FocusedIndex())
}

func TestStack_Update_forwardsPressToFocusedButton(t *testing.T) {
	stack := NewStack(
		New("A", "msg-a"),
		New("B", "msg-b"),
	)

	_ = stack.Focus()
	_, cmd := stack.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyEnter}))

	require.NotNil(t, cmd)
	require.Equal(t, "msg-a", cmd())
}

func TestStack_Blur_removesFocus(t *testing.T) {
	stack := NewStack(
		New("A", "msg-a"),
	)

	_ = stack.Focus()
	_ = stack.Blur()

	require.False(t, stack.Focused())
	require.False(t, stack.Controller.Focused())
}

func TestStack_IsAtFirstFocusable(t *testing.T) {
	stack := NewStack(
		New("A", "msg-a"),
		New("B", "msg-b"),
	)

	_ = stack.Focus()

	require.True(t, stack.IsAtFirstFocusable())
	require.False(t, stack.IsAtLastFocusable())
}

func TestStack_IsAtLastFocusable(t *testing.T) {
	stack := NewStack(
		New("A", "msg-a"),
		New("B", "msg-b"),
	)

	_ = stack.Controller.FocusLast()

	require.False(t, stack.IsAtFirstFocusable())
	require.True(t, stack.IsAtLastFocusable())
}

func TestStack_CursorLine(t *testing.T) {
	stack := NewStack(
		New("A", "msg-a"),
	)

	require.Equal(t, 0, stack.CursorLine())
}

func TestStack_Controller_isPublic(t *testing.T) {
	stack := NewStack(
		New("A", "msg-a"),
		New("B", "msg-b"),
	)

	stack.Controller.SetPrevKeys("p")
	stack.Controller.SetNextKeys("n")

	_ = stack.Focus()
	_, _ = stack.Update(tea.KeyPressMsg(tea.Key{Code: 'n'}))

	require.Equal(t, 1, stack.Controller.FocusedIndex())

	_, _ = stack.Update(tea.KeyPressMsg(tea.Key{Code: 'p'}))

	require.Equal(t, 0, stack.Controller.FocusedIndex())
}

// testRow is a simple focusable row used to verify outer navigator interaction.
type testRow struct {
	text    string
	focused bool
}

func (r *testRow) Init() tea.Cmd {
	return nil
}

func (r *testRow) Update(_ tea.Msg) (tea.Model, tea.Cmd) {
	return r, nil
}

func (r *testRow) View() tea.View {
	return tea.NewView(r.text)
}

func (r *testRow) Focus() tea.Cmd {
	r.focused = true

	return nil
}

func (r *testRow) Blur() tea.Cmd {
	r.focused = false

	return nil
}

func (r *testRow) Focused() bool {
	return r.focused
}

func TestStack_Navigator_ExitsAtBoundary(t *testing.T) {
	row := &testRow{text: "Row"}
	stack := NewStack(
		New("A", "msg-a"),
		New("B", "msg-b"),
	)

	nav := navigator.New(row, stack)
	_ = nav.FocusFirst()

	require.True(t, row.Focused())

	// Move down into the stack.
	_, _ = nav.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyDown}))

	require.False(t, row.Focused())
	require.True(t, stack.Focused())

	// Move up out of the stack from the first button.
	_, _ = nav.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyUp}))

	require.True(t, row.Focused())
	require.False(t, stack.Focused())
}

func TestStack_Navigator_ExitsFromMiddleButton(t *testing.T) {
	row := &testRow{text: "Row"}
	stack := NewStack(
		New("A", "msg-a"),
		New("B", "msg-b"),
		New("C", "msg-c"),
	)

	nav := navigator.New(row, stack)
	_ = nav.FocusFirst()
	_, _ = nav.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyDown}))
	_, _ = stack.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyRight}))

	require.Equal(t, 1, stack.Controller.FocusedIndex())
	require.True(t, stack.Focused())

	// Move up out of the stack from the middle button.
	_, _ = nav.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyUp}))

	require.True(t, row.Focused())
	require.False(t, stack.Focused())
}
