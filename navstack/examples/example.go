// example tours the same two-step wizard twice: once on NavStack[TailView],
// where only the current step is shown, then again on NavStack[SequenceView],
// where completed steps stay visible above the active one. In both phases,
// Enter pushes forward (Replace on the last step, mirroring how a real flow
// swaps in its terminal screen) and Esc goes back.
//
// The two phases are themselves just screens pushed onto an outer NavStack —
// the same pattern a real app uses to nest a flow's own NavStack inside its
// root model's.
// Run: go run ./navstack/examples/
package main

import (
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/gzigzigzeo/bubbles/navstack"
)

var (
	titleStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#53d1ff")).MarginBottom(1)
	hintStyle  = lipgloss.NewStyle().Faint(true).MarginTop(1)
	doneStyle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#2ecc71"))
)

// advanceMsg is emitted by a stepScreen when the user presses Enter. wizard
// decides whether that means Push (more steps left) or Replace (this was the
// last one).
type advanceMsg struct{}

// wizardDoneMsg is emitted by doneScreen when the user presses any key.
// wizard doesn't know what should happen next, so it just reports "done" and
// leaves that decision to whatever embeds it.
type wizardDoneMsg struct{}

// stepScreen is one page of the wizard. It doesn't know about the stack it
// lives in — it just reports "advance" or "back" and lets the parent decide.
type stepScreen struct {
	title string
	body  string
}

func (s stepScreen) Init() tea.Cmd { return nil }

func (s stepScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	key, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return s, nil
	}

	switch key.String() {
	case "enter":
		return s, func() tea.Msg { return advanceMsg{} }
	case "esc":
		return s, navstack.Back
	}

	return s, nil
}

func (s stepScreen) View() tea.View {
	// The trailing newline gives SequenceView a blank line between this step
	// and whatever comes after it once they're joined into the same stack.
	return tea.NewView(titleStyle.Render(s.title) + "\n" + s.body + "\n")
}

// doneScreen is the wizard's terminal step, swapped in with Replace instead
// of Push since there's nothing to come back to it from.
type doneScreen struct{}

func (doneScreen) Init() tea.Cmd { return nil }

func (s doneScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if _, ok := msg.(tea.KeyPressMsg); ok {
		return s, func() tea.Msg { return wizardDoneMsg{} }
	}
	return s, nil
}

func (doneScreen) View() tea.View {
	return tea.NewView(doneStyle.Render("✓ All done!") + "\n" + hintStyle.Render("press any key to continue"))
}

func welcomeStep(note string) stepScreen {
	return stepScreen{title: "Step 1/2 — Welcome", body: "This wizard has two steps before it's done.\n" + note}
}

func detailsStep() stepScreen {
	return stepScreen{title: "Step 2/2 — Details", body: "Enter finishes the wizard; Esc goes back to step 1."}
}

// wizard is the Welcome/Details/Done flow, generic over the render strategy
// so the same steps can be toured under both TailView and SequenceView.
type wizard[V navstack.StackView] struct {
	*navstack.NavStack[V]
	step int
}

func newWizard[V navstack.StackView](note string) *wizard[V] {
	return &wizard[V]{NavStack: navstack.New[V](welcomeStep(note)), step: 1}
}

func (w *wizard[V]) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if _, ok := msg.(advanceMsg); ok {
		switch w.step {
		case 1:
			w.step = 2
			return w, w.Push(detailsStep())
		case 2:
			w.step = 3
			return w, w.Replace(doneScreen{})
		}
		return w, nil
	}

	_, cmd := w.NavStack.Update(msg)
	return w, cmd
}

// View renders the wizard's stack, then appends the step controls hint once
// below everything — even under SequenceView, where every step in the stack
// gets its own line, that hint would otherwise repeat once per step.
func (w *wizard[V]) View() tea.View {
	view := w.NavStack.View()

	if _, ok := w.Top().(stepScreen); ok {
		view.SetContent(view.Content + "\n" + hintStyle.Render("[enter] continue   [esc] back"))
	}

	return view
}

// app tours the wizard twice: once under TailView, once under SequenceView,
// pushing from one to the other on wizardDoneMsg. It's a NavStack itself —
// Len() tells it which phase is active — the same way a real app's root
// model nests a flow's NavStack inside its own.
type app struct {
	*navstack.NavStack[navstack.TailView]
}

func newApp() app {
	return app{NavStack: navstack.New[navstack.TailView](
		newWizard[navstack.TailView]("Only this step is shown — that's TailView."),
	)}
}

func (a app) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		if msg.String() == "ctrl+c" {
			return a, tea.Quit
		}

	case wizardDoneMsg:
		if a.Len() == 1 {
			return a, a.Push(newWizard[navstack.SequenceView](
				"Completed steps stay visible below — that's SequenceView.",
			))
		}
		return a, tea.Quit
	}

	_, cmd := a.NavStack.Update(msg)
	return a, cmd
}

func main() {
	if _, err := tea.NewProgram(newApp()).Run(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
