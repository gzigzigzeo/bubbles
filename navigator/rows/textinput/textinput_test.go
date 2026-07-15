package textinput_test

import (
	"errors"
	"regexp"
	"testing"

	btextinput "charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/require"

	"github.com/gzigzigzeo/bubbles/navigator/rows/textinput"
)

var ansiRe = regexp.MustCompile("\x1b\\[[0-9;]*[a-zA-Z]")

func stripANSI(s string) string {
	return ansiRe.ReplaceAllString(s, "")
}

func keyPress(code rune) tea.Msg {
	return tea.KeyPressMsg(tea.Key{
		Code:        code,
		Text:        string(code),
		Mod:         0,
		ShiftedCode: 0,
		BaseCode:    0,
		IsRepeat:    false,
	})
}

func TestModel_View_rendersLabel(t *testing.T) {
	m := textinput.New(textinput.WithLabel("Name"))

	require.Contains(t, m.View().Content, "Name")
}

func TestModel_View_rendersPlaceholder(t *testing.T) {
	m := textinput.New(textinput.WithPlaceholder("Ada"), textinput.WithWidth(20))

	require.Contains(t, stripANSI(m.View().Content), "Ada")
}

func TestModel_View_faintWhenDisabled(t *testing.T) {
	m := textinput.New(textinput.WithLabel("Name"))
	_ = m.Disable()

	require.Contains(t, m.View().Content, "\u001b[2m")
}

func TestModel_Focus_focusesInput(t *testing.T) {
	m := textinput.New()

	require.False(t, m.Focused())

	_ = m.Focus()

	require.True(t, m.Focused())
}

func TestModel_Blur_blursInput(t *testing.T) {
	m := textinput.New()
	_ = m.Focus()
	_ = m.Blur()

	require.False(t, m.Focused())
}

func TestModel_Update_returnsModel(t *testing.T) {
	m := textinput.New()
	_ = m.Focus()

	updated, _ := m.Update(keyPress('a'))

	require.Equal(t, m, updated)
}

func TestModel_Update_typesWhenFocused(t *testing.T) {
	m := textinput.New()
	_ = m.Focus()

	_, _ = m.Update(keyPress('a'))

	require.Equal(t, "a", m.Get())
}

func TestModel_Update_ignoresKeysWhenBlurred(t *testing.T) {
	m := textinput.New()

	_, _ = m.Update(keyPress('a'))

	require.Equal(t, "", m.Get())
}

func TestModel_Update_ignoresKeysWhenDisabled(t *testing.T) {
	m := textinput.New()
	_ = m.Focus()
	_ = m.Disable()

	_, _ = m.Update(keyPress('a'))

	require.Equal(t, "", m.Get())
}

func TestModel_Update_filterBlocksKeys(t *testing.T) {
	filter := func(k tea.KeyMsg) bool {
		return k.Key().Text != "x"
	}

	m := textinput.New(textinput.WithFilter(filter))
	_ = m.Focus()

	_, _ = m.Update(keyPress('x'))

	require.Equal(t, "", m.Get())

	_, _ = m.Update(keyPress('a'))

	require.Equal(t, "a", m.Get())
}

func TestModel_NumberFilter_blocksLetters(t *testing.T) {
	m := textinput.New(textinput.WithFilter(textinput.NumberFilter))
	_ = m.Focus()

	_, _ = m.Update(keyPress('a'))

	require.Equal(t, "", m.Get())

	_, _ = m.Update(keyPress('1'))
	_, _ = m.Update(keyPress('2'))

	require.Equal(t, "12", m.Get())
}

func TestModel_GetSet(t *testing.T) {
	m := textinput.New()

	m.Set("hello")

	require.Equal(t, "hello", m.Get())
}

func TestModel_Label(t *testing.T) {
	m := textinput.New(textinput.WithLabel("Name"))

	require.Equal(t, "Name", m.Label())

	m.SetLabel("Company")

	require.Equal(t, "Company", m.Label())
}

func TestModel_SetWidth(t *testing.T) {
	m := textinput.New(textinput.WithWidth(20))

	require.Equal(t, 20, m.Width())
}

func TestModel_WithEchoMode(t *testing.T) {
	m := textinput.New(textinput.WithEchoMode(btextinput.EchoPassword))

	require.Equal(t, btextinput.EchoPassword, m.EchoMode())
}

func TestModel_WithValidator_invalidOnBlur(t *testing.T) {
	m := textinput.New(
		textinput.WithValidator(func(value string) error {
			if value == "" {
				return errors.New("required")
			}

			return nil
		}),
	)
	_ = m.Focus()

	_ = m.Blur()

	require.Error(t, m.Err())
}

func TestModel_WithValidator_clearsErrorWhenValid(t *testing.T) {
	m := textinput.New(
		textinput.WithValidator(func(value string) error {
			if value == "" {
				return errors.New("required")
			}

			return nil
		}),
	)
	_ = m.Focus()
	m.Set("hello")
	_ = m.Blur()

	require.NoError(t, m.Err())
}

func TestModel_SetErr_roundtrip(t *testing.T) {
	m := textinput.New()

	require.NoError(t, m.Err())

	m.SetErr(errors.New("bad"))
	require.Error(t, m.Err())
}
