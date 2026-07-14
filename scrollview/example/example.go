// example demonstrates scrollview: a scrollable block of text with a
// 1-column scrollbar showing ▲/▼ indicators and a ▒ thumb.
//
// Run: go run ./scrollview/example/
package main

import (
	"fmt"
	"os"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/gzigzigzeo/bubbles/scrollview"
)

var (
	titleStyle = lipgloss.NewStyle().Bold(true).MarginBottom(1)
	hintStyle  = lipgloss.NewStyle().Faint(true).MarginTop(1)
)

// lines is the body text shown in the viewport.
var lines = []string{
	"The quick brown fox jumps over the lazy dog.",
	"Pack my box with five dozen liquor jugs.",
	"How vexingly quick daft zebras jump!",
	"The five boxing wizards jump quickly.",
	"Sphinx of black quartz, judge my vow.",
	"Two driven jocks help fax my big quiz.",
	"Five quacking zephyrs jolt my wax bed.",
	"The jay, pig, fox, zebra and my wolves quack.",
	"Blowzy red vixens fight for a quick jump.",
	"Joaquin Phoenix was gazed by MTV for luck.",
	"A wizard's job is to vex chumps quickly in fog.",
	"Watch Jeopardy! Alex Trebek's fun TV quiz game.",
	"Jump by vow of quick lazy strength in Oxford.",
	"Crazy Fredericka bought many very exquisite opal jewels.",
	"Sixty zippers were quickly picked from the woven jute bag.",
	"A large fawn jumped quickly over white zinc boxes.",
	"The public was amazed to view the quickness of the juggler.",
	"Six big juicy steaks sizzled in a pan as five workmen left.",
	"We promptly judged antique ivory buckles for the next prize.",
	"Few black taxis drive up major roads on quiet hazy nights.",
}

type model struct {
	vp scrollview.Model
}

// newModel creates the demo model with a fixed initial size.
func newModel() model {
	vp := scrollview.New()
	vp.SetWidth(62)
	vp.SetHeight(10)
	vp.SetContent(strings.Join(lines, "\n"))

	return model{vp: vp}
}

// Init satisfies tea.Model.
func (m model) Init() tea.Cmd {
	return m.vp.Init()
}

// Update handles window resize, quit, and forwards other messages to the viewport.
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if km, ok := msg.(tea.KeyMsg); ok && km.String() == "q" {
		return m, tea.Quit
	}

	if ws, ok := msg.(tea.WindowSizeMsg); ok {
		m.vp.SetWidth(ws.Width)
		m.vp.SetHeight(max(1, ws.Height-4))

		return m, nil
	}

	vp, cmd := m.vp.Update(msg)
	m.vp = vp

	return m, cmd
}

// View renders the title, viewport, and hint.
func (m model) View() tea.View {
	content := titleStyle.Render("Scrollview Demo")
	content += "\n" + m.vp.View()
	content += "\n" + hintStyle.Render("↑/↓/k/j  PgUp/PgDn: scroll   q: quit")

	return tea.NewView(content)
}

func main() {
	if _, err := tea.NewProgram(newModel()).Run(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
