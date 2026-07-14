package button

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/require"
)

func TestModel_View_rendersLabelWhenFocused(t *testing.T) {
	m := New("OK", "msg")
	_ = m.Focus()

	require.Contains(t, m.View().Content, "OK")
}

func TestModel_View_rendersLabelWhenBlurred(t *testing.T) {
	m := New("OK", "msg")

	require.Contains(t, m.View().Content, "OK")
}

func TestModel_View_faintWhenDisabled(t *testing.T) {
	m := New("OK", "msg")
	_ = m.Disable()

	require.True(t, strings.Contains(m.View().Content, "\u001b[2m"), "disabled button should render faint")
}

func TestModel_New_panicsWhenMsgIsNil(t *testing.T) {
	require.Panics(t, func() {
		_ = New("OK", nil)
	})
}

func TestModel_Update_emitsMsgOnPress(t *testing.T) {
	m := New("OK", "pressed")
	_ = m.Focus()

	_, cmd := m.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyEnter}))

	require.NotNil(t, cmd)
	require.Equal(t, "pressed", cmd())
}

func TestModel_Update_ignoresPressWhenBlurred(t *testing.T) {
	m := New("OK", "pressed")

	_, cmd := m.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyEnter}))

	require.Nil(t, cmd)
}

func TestModel_Update_ignoresPressWhenDisabled(t *testing.T) {
	m := New("OK", "pressed")
	_ = m.Focus()
	_ = m.Disable()

	_, cmd := m.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyEnter}))

	require.Nil(t, cmd)
}

func TestModel_Update_usesCustomPressKey(t *testing.T) {
	m := New("OK", "pressed", WithPressKeys("x"))
	_ = m.Focus()

	_, cmd := m.Update(tea.KeyPressMsg(tea.Key{Code: 'x'}))

	require.NotNil(t, cmd)
	require.Equal(t, "pressed", cmd())
}

func TestModel_SetMsg_panicsWhenNil(t *testing.T) {
	m := New("OK", "msg")

	require.Panics(t, func() {
		m.SetMsg(nil)
	})
}
