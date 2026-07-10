package navstack

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// TailView renders only the topmost screen in the stack.
type TailView struct{}

// View returns the topmost screen's view unchanged.
func (TailView) View(stack []tea.Model) tea.View {
	return stack[len(stack)-1].View()
}

// SequenceView renders every screen in the stack, oldest first, joined vertically,
// with the topmost screen's content last.
type SequenceView struct{}

// View joins every screen's rendered content in stack order and returns it as the
// topmost screen's view, preserving that view's other fields (e.g. AltScreen,
// BackgroundColor).
func (SequenceView) View(stack []tea.Model) tea.View {
	parts := make([]string, len(stack))
	var top tea.View

	for i, s := range stack {
		parts[i] = s.View().Content
	}

	if len(parts) > 1 {
		top = stack[len(stack)-1].View()
	}

	if len(parts) > 0 {
		top.SetContent(lipgloss.JoinVertical(lipgloss.Left, parts...))
	}

	return top
}
