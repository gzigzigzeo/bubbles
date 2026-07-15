package menu_test

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/stretchr/testify/require"

	"github.com/gzigzigzeo/bubbles/navigator/rows/menu"
)

func keyPress(code rune) tea.Msg {
	return tea.KeyPressMsg(tea.Key{
		Code:        code,
		Text:        "",
		Mod:         0,
		ShiftedCode: 0,
		BaseCode:    0,
		IsRepeat:    false,
	})
}

func TestModel_View_rendersCursorAndNameWhenFocused(t *testing.T) {
	m := menu.New("Alpha", "alpha", "first", "msg")
	_ = m.Focus()

	view := m.View().Content

	require.Contains(t, view, "Alpha")
	require.Contains(t, view, "▶")
}

func TestModel_View_rendersBlankCursorWhenBlurred(t *testing.T) {
	m := menu.New("Alpha", "alpha", "", "msg")

	view := m.View().Content

	require.Contains(t, view, "Alpha")
	require.NotContains(t, view, "▶")
}

func TestModel_View_rendersMarkWhenMarked(t *testing.T) {
	m := menu.New("Alpha", "alpha", "", "msg")
	m.SetMarked(true)

	view := m.View().Content

	require.Contains(t, view, "✓")
}

func TestModel_View_usesMarkedNameStyleWhenMarked(t *testing.T) {
	m := menu.New("Alpha", "alpha", "", "msg")
	m.SetMarked(true)

	styles := menu.DefaultStyles()
	styles.Focused.MarkedName = lipgloss.NewStyle().Foreground(lipgloss.Color("#ff0000"))
	m.SetStyles(styles)
	_ = m.Focus()

	view := m.View().Content

	require.Contains(t, view, "\u001b[38;2;255;0;0m")
}

func TestModel_View_rendersNoMarkWhenUnmarked(t *testing.T) {
	m := menu.New("Alpha", "alpha", "", "msg")

	styles := menu.DefaultStyles()
	styles.Blurred.NoMark = lipgloss.NewStyle().SetString("☐ ")
	m.SetStyles(styles)

	view := m.View().Content

	require.Contains(t, view, "☐")
}

func TestModel_View_rendersDescription(t *testing.T) {
	m := menu.New("Alpha", "alpha", "first option", "msg")

	view := m.View().Content

	require.Contains(t, view, "first option")
}

func TestModel_View_faintWhenDisabled(t *testing.T) {
	m := menu.New("Alpha", "alpha", "", "msg")
	_ = m.Disable()

	view := m.View().Content

	require.Contains(t, view, "Alpha")
	require.Contains(t, view, "\u001b[2m")
}

func TestModel_New_panicsWhenMsgIsNil(t *testing.T) {
	require.Panics(t, func() {
		_ = menu.New("Alpha", "alpha", "", nil)
	})
}

func TestModel_Focus_notifiesController(t *testing.T) {
	rows := []*menu.Model[string]{
		menu.New("Alpha", "alpha", "", "msg-a"),
		menu.New("Beta", "beta", "", "msg-b"),
	}
	ctrl := menu.NewController(rows)

	_ = rows[1].Focus()

	require.Equal(t, 1, ctrl.FocusedIndex())
}

func TestModel_Blur_notifiesController(t *testing.T) {
	rows := []*menu.Model[string]{
		menu.New("Alpha", "alpha", "", "msg-a"),
		menu.New("Beta", "beta", "", "msg-b"),
	}
	ctrl := menu.NewController(rows)

	_ = rows[0].Focus()
	_ = rows[0].Blur()

	require.Equal(t, -1, ctrl.FocusedIndex())
}

func TestModel_Update_forwardsKeysToController(t *testing.T) {
	rows := []*menu.Model[string]{
		menu.New("Alpha", "alpha", "", "msg-a"),
		menu.New("Beta", "beta", "", "msg-b"),
	}
	_ = menu.NewController(rows)

	_ = rows[0].Focus()
	_, cmd := rows[0].Update(keyPress(tea.KeyEnter))

	require.NotNil(t, cmd)
	require.Equal(t, "msg-a", cmd())
}

func TestModel_Update_ignoresKeysWhenDisabled(t *testing.T) {
	rows := []*menu.Model[string]{
		menu.New("Alpha", "alpha", "", "msg-a"),
	}
	_ = menu.NewController(rows)

	_ = rows[0].Focus()
	_ = rows[0].Disable()

	_, cmd := rows[0].Update(keyPress(tea.KeyEnter))

	require.Nil(t, cmd)
}

func TestModel_Update_ignoresKeysWhenNotInController(t *testing.T) {
	m := menu.New("Alpha", "alpha", "", "msg")
	_ = m.Focus()

	_, cmd := m.Update(keyPress(tea.KeyEnter))

	require.Nil(t, cmd)
}
