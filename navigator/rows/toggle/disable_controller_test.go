package toggle_test

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/require"

	"github.com/gzigzigzeo/bubbles/navigator/row"
	"github.com/gzigzigzeo/bubbles/navigator/rows/toggle"
)

// fakeTarget is a minimal [row.Disableable] used for controller tests.
type fakeTarget struct {
	disabled bool
}

// Enable marks the target as enabled.
func (f *fakeTarget) Enable() tea.Cmd {
	f.disabled = false

	return nil
}

// Disable marks the target as disabled.
func (f *fakeTarget) Disable() tea.Cmd {
	f.disabled = true

	return nil
}

// Disabled reports whether the target is disabled.
func (f *fakeTarget) Disabled() bool {
	return f.disabled
}

// compile-time check that fakeTarget implements [row.Disableable].
var _ row.Disableable = (*fakeTarget)(nil)

func TestDisableController_Update_disablesTargetsOnOnMsg(t *testing.T) {
	target := &fakeTarget{}
	ctrl := toggle.NewDisableController([]row.Disableable{target})

	_ = ctrl.Update(toggle.OnMsg{})

	require.True(t, target.Disabled())
}

func TestDisableController_Update_enablesTargetsOnOffMsg(t *testing.T) {
	target := &fakeTarget{disabled: true}
	ctrl := toggle.NewDisableController([]row.Disableable{target})

	_ = ctrl.Update(toggle.OffMsg{})

	require.False(t, target.Disabled())
}

func TestDisableController_Update_invertSwapsMapping(t *testing.T) {
	target := &fakeTarget{}
	ctrl := toggle.NewDisableController([]row.Disableable{target}, toggle.WithInvert())

	_ = ctrl.Update(toggle.OnMsg{})

	require.False(t, target.Disabled())

	_ = ctrl.Update(toggle.OffMsg{})

	require.True(t, target.Disabled())
}

func TestDisableController_Update_ignoresUnrelatedMessages(t *testing.T) {
	target := &fakeTarget{}
	ctrl := toggle.NewDisableController([]row.Disableable{target})

	cmd := ctrl.Update(keyPress(tea.KeySpace))

	require.Nil(t, cmd)
	require.False(t, target.Disabled())
}

func TestDisableController_Update_controlsMultipleTargets(t *testing.T) {
	a := &fakeTarget{}
	b := &fakeTarget{}
	ctrl := toggle.NewDisableController([]row.Disableable{a, b})

	_ = ctrl.Update(toggle.OnMsg{})

	require.True(t, a.Disabled())
	require.True(t, b.Disabled())
}

func TestDisableController_WithSource_ignoresOtherToggles(t *testing.T) {
	source := toggle.New("Source")
	other := toggle.New("Other")
	target := &fakeTarget{}
	ctrl := toggle.NewDisableControllerFor(source, target)

	_ = ctrl.Update(toggle.OnMsg{Source: source})

	require.True(t, target.Disabled())

	_ = ctrl.Update(toggle.OffMsg{Source: other})

	require.True(t, target.Disabled())

	_ = ctrl.Update(toggle.OffMsg{Source: source})

	require.False(t, target.Disabled())
}

func TestModel_Update_emitsOnMsgWhenToggledOn(t *testing.T) {
	m := toggle.New("Notifications")
	_ = m.Focus()

	updated, cmd := m.Update(keyPress(tea.KeySpace))

	require.NotNil(t, cmd)

	msg := cmd()
	_, ok := msg.(toggle.OnMsg)
	require.True(t, ok)

	toggledModel, ok := updated.(*toggle.Model)
	require.True(t, ok)
	require.True(t, toggledModel.Get())
}

func TestModel_Update_emitsOffMsgWhenToggledOff(t *testing.T) {
	m := toggle.New("Notifications", toggle.WithValue(true))
	_ = m.Focus()

	updated, cmd := m.Update(keyPress(tea.KeySpace))

	require.NotNil(t, cmd)

	msg := cmd()
	_, ok := msg.(toggle.OffMsg)
	require.True(t, ok)

	toggledModel, ok := updated.(*toggle.Model)
	require.True(t, ok)
	require.False(t, toggledModel.Get())
}

func TestModel_Update_doesNotEmitMessageWhenBlurred(t *testing.T) {
	m := toggle.New("Notifications")

	_, cmd := m.Update(keyPress(tea.KeySpace))

	require.Nil(t, cmd)
}

func TestModel_Update_doesNotEmitMessageWhenDisabled(t *testing.T) {
	m := toggle.New("Notifications")
	_ = m.Focus()
	_ = m.Disable()

	_, cmd := m.Update(keyPress(tea.KeySpace))

	require.Nil(t, cmd)
}
