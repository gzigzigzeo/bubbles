package menurow

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/require"
)

func newRow(name string, msg string) *Model[string] {
	return New(name, name, "", msg)
}

func TestController_MarkAndUnmark(t *testing.T) {
	c := NewController([]*Model[string]{
		newRow("a", "msg-a"),
		newRow("b", "msg-b"),
		newRow("c", "msg-c"),
	})

	c.Mark(1)
	require.True(t, c.IsMarked(1))
	require.Equal(t, []int{1}, c.Marked())
	require.Equal(t, []string{"b"}, c.MarkedValues())

	c.Unmark(1)
	require.False(t, c.IsMarked(1))
	require.Empty(t, c.Marked())
}

func TestController_Toggle(t *testing.T) {
	c := NewController([]*Model[string]{
		newRow("a", "msg-a"),
		newRow("b", "msg-b"),
	})

	c.Toggle(0)
	require.True(t, c.IsMarked(0))

	c.Toggle(0)
	require.False(t, c.IsMarked(0))
}

func TestController_MarkOnly(t *testing.T) {
	c := NewController([]*Model[string]{
		newRow("a", "msg-a"),
		newRow("b", "msg-b"),
		newRow("c", "msg-c"),
	})

	c.Mark(0)
	c.Mark(2)
	c.MarkOnly(1)

	require.False(t, c.IsMarked(0))
	require.True(t, c.IsMarked(1))
	require.False(t, c.IsMarked(2))
}

func TestController_UnmarkAll(t *testing.T) {
	c := NewController([]*Model[string]{
		newRow("a", "msg-a"),
		newRow("b", "msg-b"),
		newRow("c", "msg-c"),
	})

	c.Mark(0)
	c.Mark(2)
	c.UnmarkAll()

	require.Empty(t, c.Marked())
	require.False(t, c.rows[0].Marked())
	require.False(t, c.rows[2].Marked())
}

func TestController_FocusedIndex_tracksFocusState(t *testing.T) {
	a := newRow("a", "msg-a")
	b := newRow("b", "msg-b")
	c := NewController([]*Model[string]{a, b})

	require.Equal(t, -1, c.FocusedIndex())

	_ = b.Focus()
	require.Equal(t, 1, c.FocusedIndex())

	_ = b.Blur()
	require.Equal(t, -1, c.FocusedIndex())
}

func TestController_Update_selectsFocusedRow(t *testing.T) {
	rows := []*Model[string]{
		newRow("a", "msg-a"),
		newRow("b", "msg-b"),
	}
	c := NewController(rows)

	_ = rows[1].Focus()
	cmd := c.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyEnter}))

	require.NotNil(t, cmd)
	require.Equal(t, "msg-b", cmd())
}

func TestController_Update_ignoresSelectWhenBlurred(t *testing.T) {
	rows := []*Model[string]{
		newRow("a", "msg-a"),
	}
	c := NewController(rows)

	cmd := c.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyEnter}))

	require.Nil(t, cmd)
}

func TestController_Update_togglesMarkInMultiSelectMode(t *testing.T) {
	rows := []*Model[string]{
		newRow("a", "msg-a"),
		newRow("b", "msg-b"),
	}
	c := NewController(rows, WithMode[string](ModeMultiSelect))

	_ = rows[0].Focus()
	cmd := c.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeySpace}))

	require.Nil(t, cmd)
	require.True(t, c.IsMarked(0))
}

func TestController_Update_ignoresSpaceInSelectMode(t *testing.T) {
	rows := []*Model[string]{
		newRow("a", "msg-a"),
	}
	c := NewController(rows, WithMode[string](ModeSelect))

	_ = rows[0].Focus()
	cmd := c.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeySpace}))

	require.Nil(t, cmd)
	require.False(t, c.IsMarked(0))
}

func TestController_Update_usesCustomSelectKey(t *testing.T) {
	rows := []*Model[string]{
		newRow("a", "msg-a"),
	}
	c := NewController(rows, WithSelectKeys[string]("x"))

	_ = rows[0].Focus()
	cmd := c.Update(tea.KeyPressMsg(tea.Key{Code: 'x'}))

	require.NotNil(t, cmd)
	require.Equal(t, "msg-a", cmd())
}

func TestController_Update_usesCustomMarkKey(t *testing.T) {
	rows := []*Model[string]{
		newRow("a", "msg-a"),
	}
	c := NewController(rows,
		WithMode[string](ModeMultiSelect),
		WithMarkKeys[string]("m"),
	)

	_ = rows[0].Focus()
	cmd := c.Update(tea.KeyPressMsg(tea.Key{Code: 'm'}))

	require.Nil(t, cmd)
	require.True(t, c.IsMarked(0))
}

func TestController_outOfRangeIgnored(t *testing.T) {
	c := NewController([]*Model[string]{
		newRow("a", "msg-a"),
		newRow("b", "msg-b"),
	})

	c.Mark(10)
	c.Unmark(-1)
	c.Toggle(5)

	require.Empty(t, c.Marked())
	require.False(t, c.IsMarked(99))
}
