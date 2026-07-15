package mass_test

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/require"

	"github.com/gzigzigzeo/bubbles/mass"
)

type modelA struct {
	label string
	cmd   tea.Cmd
}

func (m *modelA) Init() tea.Cmd {
	return m.cmd
}

func (m *modelA) Update(_ tea.Msg) (tea.Model, tea.Cmd) {
	return m, m.cmd
}

func (m *modelA) View() tea.View {
	return tea.NewView(m.label)
}

type modelB struct {
	label string
	cmd   tea.Cmd
}

func (m *modelB) Init() tea.Cmd {
	return m.cmd
}

func (m *modelB) Update(_ tea.Msg) (tea.Model, tea.Cmd) {
	return m, m.cmd
}

func (m *modelB) View() tea.View {
	return tea.NewView(m.label)
}

func TestUpdate_replacesItemsAndCollectsCommands(t *testing.T) {
	original := &modelA{label: "original", cmd: nil}
	replacement := &modelA{label: "replacement", cmd: func() tea.Msg { return "replaced" }}

	items := []tea.Model{original}
	updated, cmds := mass.Update(items, func(it tea.Model) (tea.Model, tea.Cmd) {
		return replacement, replacement.cmd
	})

	require.Equal(t, replacement, updated[0])
	require.Len(t, cmds, 1)
	require.NotNil(t, cmds[0])
}

func TestUpdate_skipsNilItems(t *testing.T) {
	called := 0

	items := []tea.Model{nil, &modelA{label: "a", cmd: nil}}
	_, cmds := mass.Update(items, func(it tea.Model) (tea.Model, tea.Cmd) {
		called++

		return it, nil
	})

	require.Equal(t, 1, called)
	require.Empty(t, cmds)
}

func TestUpdate_omitsNilCommands(t *testing.T) {
	items := []tea.Model{&modelA{label: "a", cmd: nil}, &modelA{label: "b", cmd: nil}}
	_, cmds := mass.Update(items, func(it tea.Model) (tea.Model, tea.Cmd) {
		return it, nil
	})

	require.Empty(t, cmds)
}

func TestUpdate_returnsSameSlice(t *testing.T) {
	items := []tea.Model{&modelA{label: "a", cmd: nil}}
	updated, _ := mass.Update(items, func(it tea.Model) (tea.Model, tea.Cmd) {
		return it, nil
	})

	require.Equal(t, &items[0], &updated[0])
}

func TestPropagate_callsFunctionForMatchingType(t *testing.T) {
	a := &modelA{label: "a", cmd: func() tea.Msg { return "a-cmd" }}
	b := &modelB{label: "b", cmd: func() tea.Msg { return "b-cmd" }}

	items := []tea.Model{a, b}

	var seen []string

	cmds := mass.Propagate[*modelA](items, func(m *modelA) tea.Cmd {
		seen = append(seen, m.label)

		return m.cmd
	})

	require.Equal(t, []string{"a"}, seen)
	require.Len(t, cmds, 1)
	require.NotNil(t, cmds[0])
}

func TestPropagate_skipsNilItems(t *testing.T) {
	called := 0

	items := []tea.Model{nil, &modelA{label: "a", cmd: nil}}
	cmds := mass.Propagate[*modelA](items, func(m *modelA) tea.Cmd {
		called++

		return m.cmd
	})

	require.Equal(t, 1, called)
	require.Empty(t, cmds)
}

func TestPropagate_omitsNilCommands(t *testing.T) {
	items := []tea.Model{&modelA{label: "a", cmd: nil}}
	cmds := mass.Propagate[*modelA](items, func(m *modelA) tea.Cmd {
		_ = m

		return nil
	})

	require.Empty(t, cmds)
}
