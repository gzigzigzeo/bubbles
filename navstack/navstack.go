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

// NavStack is a reusable navigation stack parameterized on V, the strategy used to
// compose View() from the current screen stack. Embed *NavStack[V] in App and
// per-target flows to get a consistent push/replace/pop stack with back-navigation
// built in. The zero value is not usable; construct with New.
type NavStack[V StackView] struct {
	stack []tea.Model
	view  V
}

// New creates a NavStack with initial as the first (bottom) screen. view is left at
// V's zero value — both TailView and SequenceView are stateless, so this needs no
// explicit initialization.
func New[V StackView](initial tea.Model) *NavStack[V] {
	return &NavStack[V]{stack: []tea.Model{initial}}
}

// Push adds a new screen to the top of the stack and returns its Init() command.
func (b *NavStack[V]) Push(screen tea.Model) tea.Cmd {
	b.stack = append(b.stack, screen)
	return screen.Init()
}

// Replace replaces the top screen with a new one and returns its Init() command.
func (b *NavStack[V]) Replace(screen tea.Model) tea.Cmd {
	b.stack[len(b.stack)-1] = screen
	return screen.Init()
}

// Pop removes the top screen. Does nothing when only one screen remains.
func (b *NavStack[V]) Pop() {
	if len(b.stack) < 2 {
		return
	}

	b.stack[len(b.stack)-1] = nil
	b.stack = b.stack[:len(b.stack)-1]
}

// Len returns the number of screens in the stack.
func (b *NavStack[V]) Len() int {
	return len(b.stack)
}

// Top returns the top screen in the stack.
func (b *NavStack[V]) Top() tea.Model {
	return b.stack[len(b.stack)-1]
}

// Init delegates to the top screen's Init().
func (b *NavStack[V]) Init() tea.Cmd {
	return b.Top().Init()
}

// View delegates to V's composition strategy over the full screen stack.
func (b *NavStack[V]) View() tea.View {
	return b.view.View(b.stack)
}

// Strategy returns the StackView backing View(). Most callers only need
// TailView or SequenceView's View() method, satisfied through NavStack
// itself, but a caller with its own StackView implementation carrying extra
// state or methods (e.g. sizing) can use this to reach it directly. Use
// WithStrategy to inject a configured instance.
func (b *NavStack[V]) Strategy() V {
	return b.view
}

// WithStrategy sets the StackView instance backing View() and returns b for
// chaining. Use it when V carries its own state (e.g. sizing); TailView and
// SequenceView are stateless and don't need this.
func (b *NavStack[V]) WithStrategy(view V) *NavStack[V] {
	b.view = view
	return b
}

// noop is batched alongside a revealed screen's Init() cmd after a BackMsg
// pop so the result is always a non-nil tea.Cmd, even when that Init()
// itself returns nil. Update's own cmd == nil check is how a NavStack
// embedded as a screen inside another NavStack tells its parent "I already
// handled this BackMsg" — tea.Batch propagates non-nil-ness based on the cmd
// function values it's given, not on the messages they produce when called,
// so it stays non-nil regardless of what noop or Init() yield when invoked.
// Do not replace this with a literal nil: that would make cmd == nil true
// again after a pop, and a parent NavStack would double-pop in response.
var noop tea.Cmd = func() tea.Msg { return nil }

// Update delegates BackMsg to the top screen first. If the top screen
// returns a non-nil command, Update forwards it unchanged — BackMsg is
// considered handled, but any real work the screen requested still runs. If
// it returns nil and this stack has more than one screen, this stack pops
// itself, calls Init() on the screen it reveals (so it can reclaim focus,
// e.g. a text field), and returns noop batched with that Init() cmd so its
// parent knows not to pop again even if the revealed screen's own Init()
// returns nil. All other messages are delegated to the top screen and their
// command is returned as-is.
func (b *NavStack[V]) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	updated, cmd := b.Top().Update(msg)
	b.stack[len(b.stack)-1] = updated

	if _, ok := msg.(BackMsg); ok && cmd == nil && len(b.stack) > 1 {
		b.Pop()
		return b, tea.Batch(noop, b.Top().Init())
	}

	return b, cmd
}
