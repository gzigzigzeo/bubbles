package field_test

import (
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/stretchr/testify/require"

	"github.com/gzigzigzeo/bubbles/form/button"
	"github.com/gzigzigzeo/bubbles/form/field"
)

// nonControl is a bare tea.Model that does not implement field.Control, used
// to verify NewFocusState filters it out.
type nonControl struct{}

func (nonControl) Init() tea.Cmd                       { return nil }
func (nonControl) Update(tea.Msg) (tea.Model, tea.Cmd) { return nonControl{}, nil }
func (nonControl) View() tea.View                      { return tea.NewView("") }

func TestNewFocusState_filtersNonControls(t *testing.T) {
	a := button.New("a", nil)
	b := button.New("b", nil)

	g := field.NewFocusState[tea.Model](a, nonControl{}, b)

	require.Equal(t, []field.Control{a, b}, g.Items(), "only Control-implementing models should be kept, in order")
}

func TestFocusState_zeroValueIsSafe(t *testing.T) {
	var g field.FocusState

	// The zero value's position is 0, not the -1 sentinel NewFocusState
	// sets, but Current() must still guard against the empty items slice.
	require.Nil(t, g.Current())
	require.False(t, g.Focused())
	require.NotPanics(t, func() { g.Blur() })
	require.Nil(t, g.Focus())
}

func TestFocusState_selectFirstSkipsDisabledLeadingItems(t *testing.T) {
	a := button.New("a", nil)
	a.Disable()
	b := button.New("b", nil)

	g := field.NewFocusState(a, b)
	g.FocusFirst()

	require.Equal(t, 1, g.Position())
	require.Equal(t, field.Control(b), g.Current())
}

func TestFocusState_selectFirstNoneWhenAllDisabled(t *testing.T) {
	a := button.New("a", nil)
	a.Disable()

	g := field.NewFocusState(a)
	g.FocusFirst()

	require.Equal(t, -1, g.Position())
	require.Nil(t, g.Current())
}

func TestFocusState_selectLastSkipsDisabledTrailingItems(t *testing.T) {
	a := button.New("a", nil)
	b := button.New("b", nil)
	b.Disable()

	g := field.NewFocusState(a, b)
	g.FocusLast()

	require.Equal(t, 0, g.Position())
	require.Equal(t, field.Control(a), g.Current())
}

func TestFocusState_selectLastNoneWhenAllDisabled(t *testing.T) {
	a := button.New("a", nil)
	a.Disable()

	g := field.NewFocusState(a)
	g.FocusLast()

	require.Equal(t, -1, g.Position())
	require.Nil(t, g.Current())
}

func TestFocusState_focusBlurFocusedDelegateToCurrent(t *testing.T) {
	a := button.New("a", nil)
	b := button.New("b", nil)

	g := field.NewFocusState(a, b)
	g.FocusFirst()

	require.False(t, g.Focused())

	g.Focus()
	require.True(t, g.Focused())
	require.True(t, a.Focused())
	require.False(t, b.Focused())

	g.Blur()
	require.False(t, g.Focused())
	require.False(t, a.Focused())
}

func TestFocusState_shiftSkipsDisabledAndWraps(t *testing.T) {
	a := button.New("a", nil)
	b := button.New("b", nil)
	b.Disable()
	c := button.New("c", nil)

	g := field.NewFocusState(a, b, c)
	g.FocusFirst()
	g.Focus()

	g.Shift(1) // a -> c, skipping disabled b
	require.Equal(t, 2, g.Position())
	require.False(t, a.Focused())
	require.True(t, c.Focused())

	g.Shift(1) // c -> a, wrapping around and skipping disabled b
	require.Equal(t, 0, g.Position())
	require.True(t, a.Focused())
	require.False(t, c.Focused())

	g.Shift(-1) // a -> c, wrapping the other way, still skipping b
	require.Equal(t, 2, g.Position())
	require.True(t, c.Focused())
}

func TestFocusState_shiftBoundedStopsInsteadOfWrapping(t *testing.T) {
	a := button.New("a", nil)
	b := button.New("b", nil)
	b.Disable()
	c := button.New("c", nil)

	g := field.NewFocusState(a, b, c)
	g.FocusFirst()
	g.Focus()

	_, ok := g.ShiftBounded(1) // a -> c, skipping disabled b
	require.True(t, ok)
	require.Equal(t, 2, g.Position())
	require.True(t, c.Focused())

	cmd, ok := g.ShiftBounded(1) // already at the end: no wrap, no-op
	require.False(t, ok)
	require.Nil(t, cmd)
	require.Equal(t, 2, g.Position())
	require.True(t, c.Focused())

	_, ok = g.ShiftBounded(-1) // c -> a, skipping disabled b
	require.True(t, ok)
	require.Equal(t, 0, g.Position())

	cmd, ok = g.ShiftBounded(-1) // already at the start: no wrap, no-op
	require.False(t, ok)
	require.Nil(t, cmd)
	require.Equal(t, 0, g.Position())
}

func TestFocusState_setFocusesDisabledItemDirectly(t *testing.T) {
	a := button.New("a", nil)
	b := button.New("b", nil)
	b.Disable()

	g := field.NewFocusState(a, b)
	g.FocusFirst()
	g.Focus()

	g.Set(1)

	require.Equal(t, 1, g.Position())
	require.False(t, a.Focused())
	require.True(t, b.Focused(), "Set must focus the target item directly, even if disabled")
}
