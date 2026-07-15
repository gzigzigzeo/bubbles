package navigator_test

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/require"

	"github.com/gzigzigzeo/bubbles/navigator"
	"github.com/gzigzigzeo/bubbles/navigator/row"
	"github.com/gzigzigzeo/bubbles/navigator/rows/menu"
	"github.com/gzigzigzeo/bubbles/navigator/rows/toggle"
)

func TestBuilder_WithControllerItems_routesKeysToController(t *testing.T) {
	ctrl := menu.NewControllerBuilder[string]().
		Add("Alpha", "alpha", "", menuMsg{Value: "alpha"}).
		Add("Beta", "beta", "", menuMsg{Value: "beta"}).
		Build()

	label := testLabel("header")
	nav := navigator.NewBuilder().
		WithItems(label).
		WithControllerItems(ctrl).
		Build()

	_ = nav.FocusFirst()
	require.Equal(t, 1, nav.FocusedIndex())

	_, cmd := nav.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyEnter}))
	require.NotNil(t, cmd)
	require.Equal(t, "alpha", cmd().(menuMsg).Value)
}

func TestBuilder_WithController_updatesControllersGlobally(t *testing.T) {
	target := &testDisableable{}
	toggleRow := toggle.New("Feature")
	ctrl := toggle.NewDisableControllerFor(toggleRow, target)

	nav := navigator.NewBuilder().
		WithItems(toggleRow).
		WithController(ctrl).
		Build()

	_ = nav.FocusFirst()
	require.False(t, target.Disabled())

	// Toggle the row on; the disable controller is updated and disables target.
	_, cmd := nav.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeySpace}))
	require.NotNil(t, cmd)

	msg := cmd()
	require.IsType(t, toggle.OnMsg{}, msg)

	_, _ = nav.Update(msg)

	require.True(t, target.Disabled())
}

func TestLockFocus_restrictsMovement(t *testing.T) {
	items := []tea.Model{
		&testItem{text: "Alpha"},
		&testItem{text: "Beta"},
		&testItem{text: "Gamma"},
		&testItem{text: "Delta"},
	}

	nav := navigator.New(items...)
	nav.LockFocus(1, 2)
	_ = nav.FocusIndex(1)

	require.Equal(t, 1, nav.FocusedIndex())

	// Down should stop at Gamma (index 2) because the lock ends there.
	sendKey(t, nav, "down")
	require.Equal(t, 2, nav.FocusedIndex())

	// Another down should stay at Gamma.
	sendKey(t, nav, "down")
	require.Equal(t, 2, nav.FocusedIndex())

	// Up should stop at Beta.
	sendKey(t, nav, "up")
	sendKey(t, nav, "up")
	require.Equal(t, 1, nav.FocusedIndex())

	// Another up should stay at Beta.
	sendKey(t, nav, "up")
	require.Equal(t, 1, nav.FocusedIndex())

	nav.UnlockFocus()
	sendKey(t, nav, "down")

	require.Equal(t, 2, nav.FocusedIndex())
}

type menuMsg struct {
	Value string
}

type testDisableable struct {
	disabled bool
}

func (t *testDisableable) Enable() tea.Cmd {
	t.disabled = false

	return nil
}

func (t *testDisableable) Disable() tea.Cmd {
	t.disabled = true

	return nil
}

func (t *testDisableable) Disabled() bool {
	return t.disabled
}

var _ row.Disableable = (*testDisableable)(nil)
