package menu_test

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/require"

	"github.com/gzigzigzeo/bubbles/navigator/rows/menu"
)

func newRow(name string, msg string) *menu.Model[string] {
	return menu.New(name, name, "", msg)
}

func TestController_MarkAndUnmark(t *testing.T) {
	ctrl := menu.NewController([]*menu.Model[string]{
		newRow("a", "msg-a"),
		newRow("b", "msg-b"),
		newRow("c", "msg-c"),
	})

	ctrl.Mark(1)
	require.True(t, ctrl.IsMarked(1))
	require.Equal(t, []int{1}, ctrl.Marked())
	require.Equal(t, []string{"b"}, ctrl.MarkedValues())

	ctrl.Unmark(1)
	require.False(t, ctrl.IsMarked(1))
	require.Empty(t, ctrl.Marked())
}

func TestController_Toggle(t *testing.T) {
	ctrl := menu.NewController([]*menu.Model[string]{
		newRow("a", "msg-a"),
		newRow("b", "msg-b"),
	})

	ctrl.Toggle(0)
	require.True(t, ctrl.IsMarked(0))

	ctrl.Toggle(0)
	require.False(t, ctrl.IsMarked(0))
}

func TestController_MarkOnly(t *testing.T) {
	ctrl := menu.NewController([]*menu.Model[string]{
		newRow("a", "msg-a"),
		newRow("b", "msg-b"),
		newRow("c", "msg-c"),
	})

	ctrl.Mark(0)
	ctrl.Mark(2)
	ctrl.MarkOnly(1)

	require.False(t, ctrl.IsMarked(0))
	require.True(t, ctrl.IsMarked(1))
	require.False(t, ctrl.IsMarked(2))
}

func TestController_UnmarkAll(t *testing.T) {
	ctrl := menu.NewController([]*menu.Model[string]{
		newRow("a", "msg-a"),
		newRow("b", "msg-b"),
		newRow("c", "msg-c"),
	})

	ctrl.Mark(0)
	ctrl.Mark(2)
	ctrl.UnmarkAll()

	require.Empty(t, ctrl.Marked())
	require.False(t, ctrl.Rows()[0].Marked())
	require.False(t, ctrl.Rows()[2].Marked())
}

func TestController_FocusedIndex_tracksFocusState(t *testing.T) {
	a := newRow("a", "msg-a")
	row := newRow("b", "msg-b")
	ctrl := menu.NewController([]*menu.Model[string]{a, row})

	require.Equal(t, -1, ctrl.FocusedIndex())

	_ = row.Focus()

	require.Equal(t, 1, ctrl.FocusedIndex())

	_ = row.Blur()

	require.Equal(t, -1, ctrl.FocusedIndex())
}

func TestController_Update_selectsFocusedRow(t *testing.T) {
	rows := []*menu.Model[string]{
		newRow("a", "msg-a"),
		newRow("b", "msg-b"),
	}
	ctrl := menu.NewController(rows)

	_ = rows[1].Focus()
	cmd := ctrl.Update(tea.KeyPressMsg(tea.Key{
		Code:        tea.KeyEnter,
		Text:        "",
		Mod:         0,
		ShiftedCode: 0,
		BaseCode:    0,
		IsRepeat:    false,
	}))

	require.NotNil(t, cmd)
	require.Equal(t, "msg-b", cmd())
}

func TestController_Update_ignoresSelectWhenBlurred(t *testing.T) {
	rows := []*menu.Model[string]{
		newRow("a", "msg-a"),
	}
	ctrl := menu.NewController(rows)

	cmd := ctrl.Update(tea.KeyPressMsg(tea.Key{
		Code:        tea.KeyEnter,
		Text:        "",
		Mod:         0,
		ShiftedCode: 0,
		BaseCode:    0,
		IsRepeat:    false,
	}))

	require.Nil(t, cmd)
}

func TestController_Update_togglesMarkInMultiSelectMode(t *testing.T) {
	rows := []*menu.Model[string]{
		newRow("a", "msg-a"),
		newRow("b", "msg-b"),
	}
	ctrl := menu.NewController(rows, menu.WithMode[string](menu.ModeMultiSelect))

	_ = rows[0].Focus()
	cmd := ctrl.Update(tea.KeyPressMsg(tea.Key{
		Code:        tea.KeySpace,
		Text:        "",
		Mod:         0,
		ShiftedCode: 0,
		BaseCode:    0,
		IsRepeat:    false,
	}))

	require.Nil(t, cmd)
	require.True(t, ctrl.IsMarked(0))
}

func TestController_Update_ignoresSpaceInSelectMode(t *testing.T) {
	rows := []*menu.Model[string]{
		newRow("a", "msg-a"),
	}
	ctrl := menu.NewController(rows, menu.WithMode[string](menu.ModeSelect))

	_ = rows[0].Focus()
	cmd := ctrl.Update(tea.KeyPressMsg(tea.Key{
		Code:        tea.KeySpace,
		Text:        "",
		Mod:         0,
		ShiftedCode: 0,
		BaseCode:    0,
		IsRepeat:    false,
	}))

	require.Nil(t, cmd)
	require.False(t, ctrl.IsMarked(0))
}

func TestController_Update_usesCustomSelectKey(t *testing.T) {
	rows := []*menu.Model[string]{
		newRow("a", "msg-a"),
	}
	ctrl := menu.NewController(rows, menu.WithSelectKeys[string]("x"))

	_ = rows[0].Focus()
	cmd := ctrl.Update(tea.KeyPressMsg(tea.Key{
		Code:        'x',
		Text:        "",
		Mod:         0,
		ShiftedCode: 0,
		BaseCode:    0,
		IsRepeat:    false,
	}))

	require.NotNil(t, cmd)
	require.Equal(t, "msg-a", cmd())
}

func TestController_Update_usesCustomMarkKey(t *testing.T) {
	rows := []*menu.Model[string]{
		newRow("a", "msg-a"),
	}
	ctrl := menu.NewController(rows,
		menu.WithMode[string](menu.ModeMultiSelect),
		menu.WithMarkKeys[string]("m"),
	)

	_ = rows[0].Focus()
	cmd := ctrl.Update(tea.KeyPressMsg(tea.Key{
		Code:        'm',
		Text:        "",
		Mod:         0,
		ShiftedCode: 0,
		BaseCode:    0,
		IsRepeat:    false,
	}))

	require.Nil(t, cmd)
	require.True(t, ctrl.IsMarked(0))
}

func TestController_outOfRangeIgnored(t *testing.T) {
	ctrl := menu.NewController([]*menu.Model[string]{
		newRow("a", "msg-a"),
		newRow("b", "msg-b"),
	})

	ctrl.Mark(10)
	ctrl.Unmark(-1)
	ctrl.Toggle(5)

	require.Empty(t, ctrl.Marked())
	require.False(t, ctrl.IsMarked(99))
}
