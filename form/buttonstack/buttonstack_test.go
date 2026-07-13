package buttonstack

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gzigzigzeo/bubbles/form/button"
)

func TestKeysAlwaysIncludesNavigationBindings(t *testing.T) {
	m := New(button.New("First", nil), button.New("Second", nil))
	m.FocusFirst()

	for _, position := range []int{0, 1} {
		if position == 1 {
			m.ShiftBounded(1)
		}

		bindings := m.Keys()
		require.Len(t, bindings, 3)
		require.Equal(t, keyLeft.Keys(), bindings[0].Keys())
		require.Equal(t, keyRight.Keys(), bindings[1].Keys())
		require.Equal(t, []string{"enter", " "}, bindings[2].Keys())
	}
}

func TestNewDoesNotSelectOrFocusAButton(t *testing.T) {
	m := New(button.New("First", nil))

	require.Equal(t, -1, m.Position())
	require.False(t, m.Focused())
}
