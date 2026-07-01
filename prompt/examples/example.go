// example runs a sequential tour of all four Prompt style presets.
// Each prompt has a long question that soft-wraps to three lines.
// Run: go run ./ui/prompt/examples/
package main

import (
	"fmt"
	"os"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/gzigzigzeo/bubbles/prompt"
)

const containerWidth = 60

type model struct {
	steps   []*prompt.Prompt
	current int
}

// makeSteps returns the four standard style prompts in order.
func makeSteps() []*prompt.Prompt {
	container := lipgloss.NewStyle().Width(containerWidth).MarginBottom(1)

	build := func(q string, s prompt.Styles, def string, keys ...string) *prompt.Prompt {
		s.Container = container
		p := prompt.New(q, keys...)
		s.Container = s.Container.Margin(1, 0)
		p.SetStyles(s)
		if def != "" {
			p.SetDefault(def)
		}
		return p
	}

	errorPrompt := build(
		"This is an error prompt. Do you like it?",
		prompt.NewErrorStyles(), "", "y", "n",
	)
	errorPrompt.SetAcceptByEnter(false)

	return []*prompt.Prompt{
		build(
			"This is a warning prompt. Do you like it?",
			prompt.NewWarnStyles(), "", "y", "n",
		),
		errorPrompt,
		build(
			"This is a long success prompt. The green color indicates a safe or "+
				"completed action where confirming will trigger a positive outcome "+
				"with no irreversible side effects. Do you like it?",
			prompt.NewSuccessStyles(), "Y", "Y", "n",
		),
		build(
			"This is an info prompt. The neutral color scheme suits questions that "+
				"are neither dangerous nor particularly positive, simply requiring "+
				"a decision from you.",
			prompt.NewInfoStyles(), "N", "y", "N",
		),
	}
}

func newModel() model {
	return model{steps: makeSteps()}
}

// Init focuses the first prompt and starts its cursor.
func (m model) Init() tea.Cmd {
	return m.steps[0].Focus()
}

// Update handles messages for the active step, advancing on answer. Once all
// steps are answered, any further keypress exits.
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.current >= len(m.steps) {
		if _, ok := msg.(tea.KeyPressMsg); ok {
			return m, tea.Quit
		}
		return m, nil
	}
	p := m.steps[m.current]
	if _, ok := p.IsMyAnswer(msg); ok {
		m.current++
		if m.current >= len(m.steps) {
			return m, nil
		}
		return m, m.steps[m.current].Focus()
	}
	_, cmd := p.Update(msg)
	return m, cmd
}

var styleHeader = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#53d1ff")).MarginBottom(1)

// View renders completed steps above the current active prompt.
func (m model) View() tea.View {
	var b strings.Builder
	b.WriteString(styleHeader.Render("    prompt — style tour"))
	b.WriteByte('\n')

	for i, s := range m.steps {
		if i > m.current {
			break
		}
		b.WriteString(s.View().Content)
		b.WriteByte('\n')
	}

	if m.current >= len(m.steps) {
		b.WriteByte('\n')
		b.WriteString("    All done!")
		b.WriteByte('\n')
	}

	return tea.NewView(b.String())
}

func main() {
	prog := tea.NewProgram(newModel())
	if _, err := prog.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
