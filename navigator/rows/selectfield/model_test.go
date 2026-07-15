package selectfield_test

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/require"

	"github.com/gzigzigzeo/bubbles/navigator"
	"github.com/gzigzigzeo/bubbles/navigator/rows/selectfield"
)

var (
	keyDown  = tea.KeyPressMsg(tea.Key{Code: tea.KeyDown})
	keyEnter = tea.KeyPressMsg(tea.Key{Code: tea.KeyEnter})
	keyEsc   = tea.KeyPressMsg(tea.Key{Code: tea.KeyEscape})
)

func TestModel_GetSet_roundtrip(t *testing.T) {
	f := selectfield.NewFromStrings([]string{"a", "b", "c"})

	require.Equal(t, "a", f.Get())

	f.Set("b")
	require.Equal(t, "b", f.Get())

	f.Set("does-not-exist")
	require.Equal(t, "b", f.Get())
}

func TestModel_openNavigateSelect_commits(t *testing.T) {
	f := selectfield.NewFromStrings([]string{"a", "b", "c"})
	nav := navigator.NewBuilder().
		WithItems(f).
		WithControllerItems(f.Controller()).
		Build()

	_ = nav.FocusFirst()

	openAndLock(t, nav, f)
	sendKey(t, nav, keyDown)
	sendKey(t, nav, keyEnter)

	require.Equal(t, "b", f.Get())
}

func TestModel_escapeCancelsWithoutCommitting(t *testing.T) {
	f := selectfield.NewFromStrings([]string{"a", "b", "c"})
	nav := navigator.NewBuilder().
		WithItems(f).
		WithControllerItems(f.Controller()).
		Build()

	_ = nav.FocusFirst()

	openAndLock(t, nav, f)
	sendKey(t, nav, keyDown)
	sendKey(t, nav, keyEsc)

	require.Equal(t, "a", f.Get())
}

func TestModel_WithValidator_rejectsInvalidValue(t *testing.T) {
	f := selectfield.NewFromStrings([]string{"a", "b", "c"},
		selectfield.WithValidator[string](func(value string) error {
			if value == "b" {
				return selectfield.ErrInvalidValue
			}

			return nil
		}),
	)
	nav := navigator.NewBuilder().
		WithItems(f).
		WithControllerItems(f.Controller()).
		Build()

	_ = nav.FocusFirst()

	openAndLock(t, nav, f)
	sendKey(t, nav, keyDown)
	sendKey(t, nav, keyEnter)

	require.Equal(t, "b", f.Get())
	require.ErrorIs(t, f.Err(), selectfield.ErrInvalidValue)
}

func TestModel_WithValidator_clearsErrorOnValidValue(t *testing.T) {
	f := selectfield.NewFromStrings([]string{"a", "b", "c"},
		selectfield.WithValidator[string](func(value string) error {
			if value == "b" {
				return selectfield.ErrInvalidValue
			}

			return nil
		}),
	)
	nav := navigator.NewBuilder().
		WithItems(f).
		WithControllerItems(f.Controller()).
		Build()

	_ = nav.FocusFirst()

	// commit invalid "b"
	openAndLock(t, nav, f)
	sendKey(t, nav, keyDown)
	sendKey(t, nav, keyEnter)
	require.Error(t, f.Err())

	// commit valid "c"
	openAndLock(t, nav, f)
	sendKey(t, nav, keyDown)
	sendKey(t, nav, keyDown)
	sendKey(t, nav, keyEnter)
	require.NoError(t, f.Err())
}

func TestModel_SetErr_roundtrip(t *testing.T) {
	f := selectfield.NewFromStrings([]string{"a", "b", "c"})

	require.NoError(t, f.Err())

	f.SetErr(selectfield.ErrInvalidValue)
	require.ErrorIs(t, f.Err(), selectfield.ErrInvalidValue)
}

// openAndLock opens the select field and processes the resulting LockFocusMsg
// so the navigator locks focus inside the picker.
func openAndLock(t *testing.T, nav *navigator.Model, f *selectfield.Model[string]) {
	t.Helper()

	_, cmd := nav.Update(keyEnter)
	require.NotNil(t, cmd)

	_, _ = nav.Update(cmd())
}

// sendKey forwards a key message to the navigator and processes the first
// follow-up message produced by the returned command.
func sendKey(t *testing.T, nav *navigator.Model, msg tea.Msg) {
	t.Helper()

	updated, cmd := nav.Update(msg)
	if n, ok := updated.(*navigator.Model); ok {
		*nav = *n
	}

	if cmd == nil {
		return
	}

	follow := cmd()
	if follow == nil {
		return
	}

	updated, _ = nav.Update(follow)
	if n, ok := updated.(*navigator.Model); ok {
		*nav = *n
	}
}
