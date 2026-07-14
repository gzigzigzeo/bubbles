package menurow

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/require"
)

func TestModel_View_rendersCursorAndNameWhenFocused(t *testing.T) {
	m := New("Alpha", "alpha", "first", "msg")
	_ = m.Focus()

	view := m.View().Content

	require.Contains(t, view, "Alpha")
	require.Contains(t, view, "▶")
}

func TestModel_View_rendersBlankCursorWhenBlurred(t *testing.T) {
	m := New("Alpha", "alpha", "", "msg")

	view := m.View().Content

	require.Contains(t, view, "Alpha")
	require.NotContains(t, view, "▶")
}

func TestModel_View_rendersMarkWhenMarked(t *testing.T) {
	m := New("Alpha", "alpha", "", "msg")
	m.SetMarked(true)

	view := m.View().Content

	require.Contains(t, view, "✓")
}

func TestModel_View_rendersDescription(t *testing.T) {
	m := New("Alpha", "alpha", "first option", "msg")

	view := m.View().Content

	require.Contains(t, view, "first option")
}

func TestModel_View_faintWhenDisabled(t *testing.T) {
	m := New("Alpha", "alpha", "", "msg")
	_ = m.Disable()

	view := m.View().Content

	require.Contains(t, view, "Alpha")
	require.True(t, strings.Contains(view, "\u001b[2m"), "disabled row should render faint")
}

func TestModel_New_panicsWhenMsgIsNil(t *testing.T) {
	require.Panics(t, func() {
		_ = New("Alpha", "alpha", "", nil)
	})
}

func TestModel_Focus_notifiesCollection(t *testing.T) {
	rows := []*Model[string]{
		New("Alpha", "alpha", "", "msg-a"),
		New("Beta", "beta", "", "msg-b"),
	}
	c := NewCollection(rows)

	_ = rows[1].Focus()

	require.Equal(t, 1, c.FocusedIndex())
}

func TestModel_Blur_notifiesCollection(t *testing.T) {
	rows := []*Model[string]{
		New("Alpha", "alpha", "", "msg-a"),
		New("Beta", "beta", "", "msg-b"),
	}
	_ = NewCollection(rows)

	_ = rows[0].Focus()
	_ = rows[0].Blur()

	require.Equal(t, -1, rows[0].collection.FocusedIndex())
}

func TestModel_Update_forwardsKeysToCollection(t *testing.T) {
	rows := []*Model[string]{
		New("Alpha", "alpha", "", "msg-a"),
		New("Beta", "beta", "", "msg-b"),
	}
	_ = NewCollection(rows)

	_ = rows[0].Focus()
	_, cmd := rows[0].Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyEnter}))

	require.NotNil(t, cmd)
	require.Equal(t, "msg-a", cmd())
}

func TestModel_Update_ignoresKeysWhenDisabled(t *testing.T) {
	rows := []*Model[string]{
		New("Alpha", "alpha", "", "msg-a"),
	}
	_ = NewCollection(rows)

	_ = rows[0].Focus()
	_ = rows[0].Disable()

	_, cmd := rows[0].Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyEnter}))

	require.Nil(t, cmd)
}

func TestModel_Update_ignoresKeysWhenNotInCollection(t *testing.T) {
	m := New("Alpha", "alpha", "", "msg")
	_ = m.Focus()

	_, cmd := m.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyEnter}))

	require.Nil(t, cmd)
}
