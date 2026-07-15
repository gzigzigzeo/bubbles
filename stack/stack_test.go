package stack_test

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/stretchr/testify/require"

	"github.com/gzigzigzeo/bubbles/stack"
)

type replaceMsg struct {
	model tea.Model
}

type testModel struct {
	label   string
	updated []tea.Msg
	cmd     tea.Cmd
}

func (m *testModel) Init() tea.Cmd {
	return m.cmd
}

func (m *testModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if r, ok := msg.(replaceMsg); ok {
		return r.model, nil
	}

	m.updated = append(m.updated, msg)

	return m, m.cmd
}

func (m *testModel) View() tea.View {
	return tea.NewView(m.label)
}

func TestNewVertical_createsVerticalStack(t *testing.T) {
	s := stack.NewVertical(&testModel{label: "A"})

	require.Len(t, s.Items(), 1)
}

func TestNewHorizontal_createsHorizontalStack(t *testing.T) {
	s := stack.NewHorizontal(&testModel{label: "A"})

	require.Len(t, s.Items(), 1)
}

func TestInit_returnsSequenceOfCommands(t *testing.T) {
	a := &testModel{label: "A", cmd: func() tea.Msg { return "a-init" }}
	b := &testModel{label: "B", cmd: func() tea.Msg { return "b-init" }}

	s := stack.NewVertical(a, b)
	cmd := s.Init()

	require.NotNil(t, cmd)
	require.NotNil(t, cmd())
}

func TestUpdate_forwardsMessageToAllItems(t *testing.T) {
	a := &testModel{label: "A"}
	b := &testModel{label: "B"}

	s := stack.NewVertical(a, b)
	msg := "hello"
	_, _ = s.Update(msg)

	require.Len(t, a.updated, 1)
	require.Equal(t, msg, a.updated[0])
	require.Len(t, b.updated, 1)
	require.Equal(t, msg, b.updated[0])
}

func TestUpdate_replacesItems(t *testing.T) {
	original := &testModel{label: "A"}
	replacement := &testModel{label: "B"}

	s := stack.NewVertical(original)
	_, _ = s.Update(replaceMsg{replacement})

	require.Equal(t, replacement, s.Items()[0])
}

func TestUpdate_returnsSequenceOfCommands(t *testing.T) {
	a := &testModel{label: "A", cmd: func() tea.Msg { return "a-update" }}
	b := &testModel{label: "B", cmd: func() tea.Msg { return "b-update" }}

	s := stack.NewVertical(a, b)
	_, cmd := s.Update("msg")

	require.NotNil(t, cmd)
	require.NotNil(t, cmd())
}

func TestForEach_visitsAndReplacesItems(t *testing.T) {
	original := &testModel{label: "A"}
	replacement := &testModel{label: "B", cmd: func() tea.Msg { return "b-cmd" }}

	items := []tea.Model{original}
	updated, cmd := stack.ForEach(items, func(it tea.Model) (tea.Model, tea.Cmd) {
		return replacement, replacement.cmd
	})

	require.Equal(t, replacement, updated[0])
	require.NotNil(t, cmd)
	require.NotNil(t, cmd())
}

func TestView_verticalJoinsItems(t *testing.T) {
	s := stack.NewVertical(
		&testModel{label: "A"},
		&testModel{label: "B"},
	)

	view := s.View().Content

	require.Contains(t, view, "A")
	require.Contains(t, view, "B")
	require.Greater(t, lipgloss.Height(view), 1)
}

func TestView_horizontalJoinsItems(t *testing.T) {
	s := stack.NewHorizontal(
		&testModel{label: "A"},
		&testModel{label: "B"},
	)

	view := s.View().Content

	require.Contains(t, view, "A")
	require.Contains(t, view, "B")
	require.Equal(t, 1, lipgloss.Height(view))
}

func TestView_skipsNilItems(t *testing.T) {
	s := stack.NewVertical(&testModel{label: "A"}, nil, &testModel{label: "B"})

	view := s.View().Content

	require.Contains(t, view, "A")
	require.Contains(t, view, "B")
}
