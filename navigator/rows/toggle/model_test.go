package toggle_test

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/require"

	"github.com/gzigzigzeo/bubbles/navigator/rows/toggle"
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

func TestModel_View_rendersLabelWhenFocused(t *testing.T) {
	m := toggle.New("Notifications")
	_ = m.Focus()

	require.Contains(t, m.View().Content, "Notifications")
}

func TestModel_View_rendersLabelWhenBlurred(t *testing.T) {
	m := toggle.New("Notifications")

	require.Contains(t, m.View().Content, "Notifications")
}

func TestModel_View_rendersOnWhenTrue(t *testing.T) {
	m := toggle.New("Notifications", toggle.WithValue(true))

	require.Contains(t, m.View().Content, "● On")
}

func TestModel_View_rendersOffWhenFalse(t *testing.T) {
	m := toggle.New("Notifications")

	require.Contains(t, m.View().Content, "○ Off")
}

func TestModel_View_faintWhenDisabled(t *testing.T) {
	m := toggle.New("Notifications")
	_ = m.Disable()

	require.Contains(t, m.View().Content, "\u001b[2m")
}

func TestModel_Update_togglesOnSpaceWhenFocused(t *testing.T) {
	m := toggle.New("Notifications")
	_ = m.Focus()

	require.False(t, m.Get())

	_, _ = m.Update(keyPress(tea.KeySpace))

	require.True(t, m.Get())
}

func TestModel_Update_togglesBackOnSpace(t *testing.T) {
	m := toggle.New("Notifications", toggle.WithValue(true))
	_ = m.Focus()

	require.True(t, m.Get())

	_, _ = m.Update(keyPress(tea.KeySpace))

	require.False(t, m.Get())
}

func TestModel_Update_ignoresSpaceWhenBlurred(t *testing.T) {
	m := toggle.New("Notifications", toggle.WithValue(false))

	_, _ = m.Update(keyPress(tea.KeySpace))

	require.False(t, m.Get())
}

func TestModel_Update_ignoresSpaceWhenDisabled(t *testing.T) {
	m := toggle.New("Notifications")
	_ = m.Focus()
	_ = m.Disable()

	_, _ = m.Update(keyPress(tea.KeySpace))

	require.False(t, m.Get())
}

func TestModel_Update_usesCustomToggleKey(t *testing.T) {
	m := toggle.New("Notifications", toggle.WithToggleKeys("x"))
	_ = m.Focus()

	_, _ = m.Update(keyPress('x'))

	require.True(t, m.Get())
}

func TestModel_Keys_returnsToggleBinding(t *testing.T) {
	m := toggle.New("Notifications")

	keys := m.Keys()

	require.Len(t, keys, 1)
	require.Equal(t, []string{"space"}, keys[0].Keys())
	require.Equal(t, "space", keys[0].Help().Key)
	require.Equal(t, "toggle", keys[0].Help().Desc)
}

func TestModel_Label(t *testing.T) {
	m := toggle.New("Notifications")

	require.Equal(t, "Notifications", m.Label())

	m.SetLabel("Alerts")

	require.Equal(t, "Alerts", m.Label())
}
