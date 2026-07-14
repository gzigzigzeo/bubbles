package button

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/gzigzigzeo/bubbles/navigator/focus"
	"github.com/gzigzigzeo/bubbles/navigator/rows/row"
)

// StackStyles holds the visual pieces of a button stack.
type StackStyles struct {
	Wrapper lipgloss.Style
}

// Stack is a row that contains multiple buttons arranged horizontally. Focus
// enters the stack and moves left/right between buttons. The outer vertical
// navigator handles up/down keys directly, so the stack is a plain
// [row.Focusable] and does not implement [row.FocusReceiver].
type Stack struct {
	styles     StackStyles
	state      row.FocusedState
	Controller *focus.Controller
}

// NewStack creates a horizontal button stack over the given buttons.
func NewStack(buttons ...tea.Model) *Stack {
	ctrl := focus.New(buttons...)
	ctrl.SetPrevKeys("left", "h")
	ctrl.SetNextKeys("right", "l")

	return &Stack{
		Controller: ctrl,
	}
}

// SetStyles replaces the stack styles.
func (s *Stack) SetStyles(styles StackStyles) {
	s.styles = styles
}

// Init initializes all buttons in the stack.
func (s *Stack) Init() tea.Cmd {
	return s.Controller.Init()
}

// Update delegates to the internal focus controller. Left/right move focus
// between buttons; the press key activates the focused button. Up/down are not
// consumed here and are handled by the outer navigator for row navigation.
func (s *Stack) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	cmd := s.Controller.Update(msg)

	return s, cmd
}

// View renders the buttons horizontally and wraps them with the stack wrapper
// style.
func (s *Stack) View() tea.View {
	items := s.Controller.Items()
	parts := make([]string, len(items))

	for i, it := range items {
		parts[i] = it.View().Content
	}

	content := lipgloss.JoinHorizontal(lipgloss.Top, parts...)

	return tea.NewView(s.styles.Wrapper.Render(content))
}

// Focus focuses the first button in the stack. Implements [row.Focusable].
func (s *Stack) Focus() tea.Cmd {
	s.state.Focus()

	return s.Controller.Focus()
}

// Blur removes focus from the stack and its current button. Implements
// [row.Focusable].
func (s *Stack) Blur() tea.Cmd {
	s.state.Blur()

	return s.Controller.Blur()
}

// Focused reports whether the stack holds focus. Implements [row.Focusable].
func (s *Stack) Focused() bool {
	return s.state.Focused()
}

// IsAtFirstFocusable reports whether the currently focused button is the first
// focusable button. Implements [row.BoundaryAware].
func (s *Stack) IsAtFirstFocusable() bool {
	return s.Controller.IsAtFirstFocusable()
}

// IsAtLastFocusable reports whether the currently focused button is the last
// focusable button. Implements [row.BoundaryAware].
func (s *Stack) IsAtLastFocusable() bool {
	return s.Controller.IsAtLastFocusable()
}

// CursorLine returns 0 because the stack renders on a single line. Implements
// [row.CursorAware].
func (s *Stack) CursorLine() int {
	return 0
}
