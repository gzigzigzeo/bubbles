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
	steps   []*prompt.PromptModel
	current int
}

// buildPrompt constructs one step: yesNoOpt picks the key/default shape
// (WithYesNo, WithYesNoDefaultYes, or WithYesNoDefaultNo), s supplies the
// color preset, and extra carries anything else (e.g. WithAcceptByEnter).
func buildPrompt(q string, s prompt.Styles, yesNoOpt prompt.Option, extra ...prompt.Option) *prompt.PromptModel {
	s.Container = lipgloss.NewStyle().Width(containerWidth).MarginBottom(1).Margin(1, 0)
	opts := append([]prompt.Option{yesNoOpt, prompt.WithStyles(s)}, extra...)

	p, err := prompt.New(q, opts...)
	if err != nil {
		panic(err)
	}
	return p
}

const successQuestion = "This is a long success prompt. The green color indicates a safe or " +
	"completed action where confirming will trigger a positive outcome " +
	"with no irreversible side effects. Do you like it?"

const infoQuestion = "This is an info prompt. The neutral color scheme suits questions that " +
	"are neither dangerous nor particularly positive, simply requiring " +
	"a decision from you."

// makeSteps returns the four standard style prompts in order.
func makeSteps() []*prompt.PromptModel {
	errorPrompt := buildPrompt(
		"This is an error prompt. Do you like it?",
		prompt.NewErrorStyles(), prompt.WithYesNo(), prompt.WithAcceptByEnter(false),
	)

	return []*prompt.PromptModel{
		buildPrompt("This is a warning prompt. Do you like it?", prompt.NewWarnStyles(), prompt.WithYesNo()),
		errorPrompt,
		buildPrompt(successQuestion, prompt.NewSuccessStyles(), prompt.WithYesNoDefaultYes()),
		buildPrompt(infoQuestion, prompt.NewInfoStyles(), prompt.WithYesNoDefaultNo()),
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
	switch msg.(type) {
	case prompt.YesMsg, prompt.NoMsg:
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
