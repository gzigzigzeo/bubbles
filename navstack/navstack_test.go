package navstack_test

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gzigzigzeo/bubbles/navstack"
)

// updateFunc lets a test customize how a fakeScreen responds to Update.
type updateFunc func(msg tea.Msg) (tea.Model, tea.Cmd)

// fakeScreen is a minimal tea.Model test double that records calls and lets
// tests control what Init/Update return.
type fakeScreen struct {
	name string

	initCmd    tea.Cmd
	updateFunc updateFunc
	content    string

	initCalls   int
	updateCalls int
}

func newFakeScreen(name string) *fakeScreen {
	return &fakeScreen{
		name:        name,
		content:     name,
		initCmd:     nil,
		updateFunc:  nil,
		initCalls:   0,
		updateCalls: 0,
	}
}

func (f *fakeScreen) Init() tea.Cmd {
	f.initCalls++

	return f.initCmd
}

func (f *fakeScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	f.updateCalls++

	if f.updateFunc != nil {
		return f.updateFunc(msg)
	}

	return f, nil
}

func (f *fakeScreen) View() tea.View {
	v := tea.NewView(f.content)
	v.AltScreen = true

	return v
}

// runCmd executes cmd and returns the produced message, or nil.
func runCmd(cmd tea.Cmd) tea.Msg {
	if cmd == nil {
		return nil
	}

	return cmd()
}

func TestNew_StartsWithSingleScreen(t *testing.T) {
	initial := newFakeScreen("root")
	stack := navstack.New[navstack.TailView](initial)

	assert.Equal(t, 1, stack.Len())
	assert.Same(t, initial, stack.Top())
}

func TestPush_AddsScreenAndReturnsInitCmd(t *testing.T) {
	root := newFakeScreen("root")
	child := newFakeScreen("child")
	child.initCmd = func() tea.Msg { return "child-init" }

	stack := navstack.New[navstack.TailView](root)
	cmd := stack.Push(child)

	assert.Equal(t, 2, stack.Len())
	assert.Same(t, child, stack.Top())
	assert.Equal(t, 1, child.initCalls)
	require.NotNil(t, cmd)
	assert.Equal(t, "child-init", runCmd(cmd))
}

func TestReplace_SwapsTopScreenWithoutChangingLength(t *testing.T) {
	root := newFakeScreen("root")
	replacement := newFakeScreen("replacement")
	replacement.initCmd = func() tea.Msg { return "replacement-init" }

	stack := navstack.New[navstack.TailView](root)
	cmd := stack.Replace(replacement)

	assert.Equal(t, 1, stack.Len())
	assert.Same(t, replacement, stack.Top())
	require.NotNil(t, cmd)
	assert.Equal(t, "replacement-init", runCmd(cmd))
}

func TestPop_RemovesTopScreen(t *testing.T) {
	root := newFakeScreen("root")
	child := newFakeScreen("child")

	stack := navstack.New[navstack.TailView](root)
	stack.Push(child)
	stack.Pop()

	assert.Equal(t, 1, stack.Len())
	assert.Same(t, root, stack.Top())
}

func TestPop_NoopAtRoot(t *testing.T) {
	root := newFakeScreen("root")
	stack := navstack.New[navstack.TailView](root)

	stack.Pop()

	assert.Equal(t, 1, stack.Len())
	assert.Same(t, root, stack.Top())
}

func TestInit_DelegatesToTopScreen(t *testing.T) {
	root := newFakeScreen("root")
	child := newFakeScreen("child")
	child.initCmd = func() tea.Msg { return "child-init" }

	stack := navstack.New[navstack.TailView](root)
	stack.Push(child)
	child.initCalls = 0 // reset the call Push already made

	cmd := stack.Init()

	assert.Equal(t, 1, child.initCalls)
	require.NotNil(t, cmd)
	assert.Equal(t, "child-init", runCmd(cmd))
}

type customMsg struct{}

func TestUpdate_DelegatesNonBackMsgToTopScreen(t *testing.T) {
	root := newFakeScreen("root")
	replacementFromUpdate := newFakeScreen("updated")
	child := newFakeScreen("child")
	child.updateFunc = func(msg tea.Msg) (tea.Model, tea.Cmd) {
		return replacementFromUpdate, func() tea.Msg { return "handled" }
	}

	stack := navstack.New[navstack.TailView](root)
	stack.Push(child)

	model, cmd := stack.Update(customMsg{})

	assert.Same(t, stack, model)
	assert.Equal(t, 1, child.updateCalls)
	assert.Same(t, replacementFromUpdate, stack.Top(), "Update must write the returned model back into the stack")
	require.NotNil(t, cmd)
	assert.Equal(t, "handled", runCmd(cmd))
}

func TestUpdate_BackMsgNotPoppedWhenTopHandlesIt(t *testing.T) {
	root := newFakeScreen("root")
	child := newFakeScreen("child")
	child.updateFunc = func(msg tea.Msg) (tea.Model, tea.Cmd) {
		return child, func() tea.Msg { return "child-handled-back" }
	}

	stack := navstack.New[navstack.TailView](root)
	stack.Push(child)

	_, cmd := stack.Update(navstack.BackMsg{})

	assert.Equal(t, 2, stack.Len(), "stack must not pop when the top screen handles BackMsg itself")
	assert.Same(t, child, stack.Top())
	require.NotNil(t, cmd)
	assert.Equal(t, "child-handled-back", runCmd(cmd))
}

func TestUpdate_BackMsgPopsAndInitsRevealedScreenWhenTopCannotGoBack(t *testing.T) {
	root := newFakeScreen("root")
	root.initCmd = func() tea.Msg { return "root-init" }
	child := newFakeScreen("child")
	child.updateFunc = func(msg tea.Msg) (tea.Model, tea.Cmd) {
		return child, nil // child is at its own root, can't go back further
	}

	stack := navstack.New[navstack.TailView](root)
	stack.Push(child)

	root.initCalls = 0 // reset the call New/Push may or may not have made

	_, cmd := stack.Update(navstack.BackMsg{})

	assert.Equal(t, 1, stack.Len(), "stack should pop itself once the top screen can't handle BackMsg")
	assert.Same(t, root, stack.Top())
	assert.Equal(t, 1, root.initCalls, "the revealed screen's Init() must be called so it can reclaim focus")

	require.NotNil(t, cmd)
	batch, ok := runCmd(cmd).(tea.BatchMsg)
	require.True(t, ok)
	require.Len(t, batch, 2)

	var sawRootInit bool

	for _, sub := range batch {
		if msg := runCmd(sub); msg == "root-init" {
			sawRootInit = true
		}
	}

	assert.True(t, sawRootInit, "the batch must include the revealed screen's Init() command")
}

func TestUpdate_BackMsgAtRootDoesNothing(t *testing.T) {
	root := newFakeScreen("root")
	root.updateFunc = func(msg tea.Msg) (tea.Model, tea.Cmd) {
		return root, nil // already at root, can't go back further
	}

	stack := navstack.New[navstack.TailView](root)

	_, cmd := stack.Update(navstack.BackMsg{})

	assert.Equal(t, 1, stack.Len())
	assert.Same(t, root, stack.Top())
	assert.Nil(t, cmd)
}

func TestTailView_ShowsOnlyTopScreen(t *testing.T) {
	root := newFakeScreen("root")
	child := newFakeScreen("child")

	stack := navstack.New[navstack.TailView](root)
	stack.Push(child)

	view := stack.View()
	assert.Equal(t, "child", view.Content)
}

func TestSequenceView_JoinsAllScreensInStackOrder(t *testing.T) {
	root := newFakeScreen("root")
	child := newFakeScreen("child")

	stack := navstack.New[navstack.SequenceView](root)
	stack.Push(child)

	// lipgloss.JoinVertical left-pads narrower lines, so compare trimmed
	// lines rather than the raw joined string.
	lines := strings.Split(stack.View().Content, "\n")
	require.Len(t, lines, 2)
	assert.Equal(t, "root", strings.TrimRight(lines[0], " "))
	assert.Equal(t, "child", strings.TrimRight(lines[1], " "))
}

func TestSequenceView_PreservesTopScreenOtherViewFields(t *testing.T) {
	root := newFakeScreen("root")
	child := newFakeScreen("child")

	stack := navstack.New[navstack.SequenceView](root)
	stack.Push(child)

	view := stack.View()
	assert.True(t, view.AltScreen, "SequenceView must preserve the top screen's non-Content View fields")
}

func TestStrategy_ReturnsStackView(t *testing.T) {
	stack := navstack.New[navstack.TailView](newFakeScreen("root"))
	assert.Equal(t, navstack.TailView{}, stack.Strategy())
}

func TestWithStrategy_SetsStrategy(t *testing.T) {
	stack := navstack.New[navstack.SequenceView](newFakeScreen("root"))
	returned := stack.WithStrategy(navstack.SequenceView{})

	assert.Same(t, stack, returned, "WithStrategy must return the receiver for chaining")
	assert.Equal(t, navstack.SequenceView{}, stack.Strategy())
}

func TestBack_ReturnsBackMsg(t *testing.T) {
	assert.Equal(t, navstack.BackMsg{}, navstack.Back())
}
