package selectfield_test

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/require"

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
	_ = f.Focus()

	_, _ = f.Update(keyEnter)
	_, _ = f.Update(keyDown)
	_, _ = f.Update(keyEnter)

	require.Equal(t, "b", f.Get())
}

func TestModel_escapeCancelsWithoutCommitting(t *testing.T) {
	f := selectfield.NewFromStrings([]string{"a", "b", "c"})
	_ = f.Focus()

	_, _ = f.Update(keyEnter)
	_, _ = f.Update(keyDown)
	_, _ = f.Update(keyEsc)

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
	_ = f.Focus()

	_, _ = f.Update(keyEnter)
	_, _ = f.Update(keyDown)
	_, _ = f.Update(keyEnter)

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
	_ = f.Focus()

	// commit invalid "b"
	_, _ = f.Update(keyEnter)
	_, _ = f.Update(keyDown)
	_, _ = f.Update(keyEnter)
	require.Error(t, f.Err())

	// commit valid "c"
	_, _ = f.Update(keyEnter)
	_, _ = f.Update(keyDown)
	_, _ = f.Update(keyEnter)
	require.NoError(t, f.Err())
}

func TestModel_SetErr_roundtrip(t *testing.T) {
	f := selectfield.NewFromStrings([]string{"a", "b", "c"})

	require.NoError(t, f.Err())

	f.SetErr(selectfield.ErrInvalidValue)
	require.ErrorIs(t, f.Err(), selectfield.ErrInvalidValue)
}
