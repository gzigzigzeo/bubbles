package navstack

import (
	tea "charm.land/bubbletea/v2"
)

// BackMsg is dispatched when a screen wants to return to the previous screen.
type BackMsg struct{}

// Back is a command that emits BackMsg.
func Back() tea.Msg {
	return BackMsg{}
}

// StackView composes a NavStack's []tea.Model stack into a single tea.View.
// TailView and SequenceView are the two provided implementations.
type StackView interface {
	View(stack []tea.Model) tea.View
}

// Model is a reusable navigation stack parameterized on V, the strategy used to
// compose View() from the current screen stack. Embed *Model[V] in App and
// per-target flows to get a consistent push/replace/pop stack with back-navigation
// built in. The zero value is not usable; construct with New.
type Model[V StackView] struct {
	stack []tea.Model
	view  V
}

// New creates a NavStack with initial as the first (bottom) screen. view is left at
// V's zero value — both TailView and SequenceView are stateless, so this needs no
// explicit initialization.
func New[V StackView](initial tea.Model) *Model[V] {
	var view V

	return &Model[V]{
		stack: []tea.Model{initial},
		view:  view,
	}
}

// Push adds a new screen to the top of the stack and returns its Init() command.
func (b *Model[V]) Push(screen tea.Model) tea.Cmd {
	b.stack = append(b.stack, screen)

	return screen.Init()
}

// Replace replaces the top screen with a new one and returns its Init() command.
func (b *Model[V]) Replace(screen tea.Model) tea.Cmd {
	b.stack[len(b.stack)-1] = screen

	return screen.Init()
}

const minStackSizeForPop = 2

// Pop removes the top screen. Does nothing when only one screen remains.
func (b *Model[V]) Pop() {
	if len(b.stack) < minStackSizeForPop {
		return
	}

	b.stack[len(b.stack)-1] = nil
	b.stack = b.stack[:len(b.stack)-1]
}

// Len returns the number of screens in the stack.
func (b *Model[V]) Len() int {
	return len(b.stack)
}

// Top returns the top screen in the stack.
func (b *Model[V]) Top() tea.Model {
	return b.stack[len(b.stack)-1]
}

// Init delegates to the top screen's Init().
func (b *Model[V]) Init() tea.Cmd {
	return b.Top().Init()
}

// View delegates to V's composition strategy over the full screen stack.
func (b *Model[V]) View() tea.View {
	return b.view.View(b.stack)
}

// Strategy returns the StackView backing View(). Most callers only need
// TailView or SequenceView's View() method, satisfied through NavStack
// itself, but a caller with its own StackView implementation carrying extra
// state or methods (e.g. sizing) can use this to reach it directly. Use
// WithStrategy to inject a configured instance.
func (b *Model[V]) Strategy() V {
	return b.view
}

// WithStrategy sets the StackView instance backing View() and returns b for
// chaining. Use it when V carries its own state (e.g. sizing); TailView and
// SequenceView are stateless and don't need this.
func (b *Model[V]) WithStrategy(view V) *Model[V] {
	b.view = view

	return b
}

// Update delegates BackMsg to the top screen first. If the top screen
// returns a non-nil command, Update forwards it unchanged — BackMsg is
// considered handled, but any real work the screen requested still runs. If
// it returns nil and this stack has more than one screen, this stack pops
// itself, calls Init() on the screen it reveals (so it can reclaim focus,
// e.g. a text field), and returns a non-nil command batched with that Init()
// cmd so its parent knows not to pop again even if the revealed screen's own
// Init() returns nil. All other messages are delegated to the top screen and
// their command is returned as-is.
func (b *Model[V]) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	updated, cmd := b.Top().Update(msg)
	b.stack[len(b.stack)-1] = updated

	if _, ok := msg.(BackMsg); ok && cmd == nil && len(b.stack) > 1 {
		b.Pop()

		noop := func() tea.Msg { return nil }

		return b, tea.Batch(noop, b.Top().Init())
	}

	return b, cmd
}
