// example demonstrates Navigator with menu, text input, number input, toggle,
// and button rows. Navigate with ↑/↓ or k/j, move between buttons with ←/→ or
// h/l, press buttons with enter, and quit with q.
//
// Run: go run ./navigator/example/
package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/gzigzigzeo/bubbles/navigator"
	"github.com/gzigzigzeo/bubbles/navigator/rows/button"
	"github.com/gzigzigzeo/bubbles/navigator/rows/menu"
	"github.com/gzigzigzeo/bubbles/navigator/rows/selectfield"
	"github.com/gzigzigzeo/bubbles/navigator/rows/textinput"
	"github.com/gzigzigzeo/bubbles/navigator/rows/toggle"
	"github.com/gzigzigzeo/bubbles/scrollview"
)

const (
	defaultWidth  = 43
	defaultHeight = 7
	heightPadding = 4
)

// label is a non-focusable, non-disableable display row.
type label struct {
	text  string
	style lipgloss.Style
}

// newLabel creates a label row rendered with the given style.
func newLabel(text string, style lipgloss.Style) label {
	return label{
		text:  text,
		style: style,
	}
}

// Init satisfies tea.Model.
func (l label) Init() tea.Cmd {
	return nil
}

// Update satisfies tea.Model.
func (l label) Update(_ tea.Msg) (tea.Model, tea.Cmd) {
	return l, nil
}

// View renders the label text with the configured style.
func (l label) View() tea.View {
	return tea.NewView(l.style.Render(" " + l.text))
}

// selectMsg is emitted when a menu row is selected.
type selectMsg struct {
	Value string
}

// pressMsg is emitted when a button is pressed.
type pressMsg struct {
	Label string
}

// model is the root Bubble Tea model.
type model struct {
	nav         *navigator.Model
	viewport    *scrollview.Model
	maxHeight   int
	titleStyle  lipgloss.Style
	hintStyle   lipgloss.Style
	cursorStyle lipgloss.Style
	statusStyle lipgloss.Style

	menuRows   []*menu.Model[string]
	textRows   []*textinput.Model
	ageRow     *textinput.Model
	toggleRows []*toggle.Model
	selectRow  *selectfield.Model[string]
	lastAction string
}

// newModel builds the demo.
//
// Outer navigator — open (focus stays at boundaries, viewport scrolls):
//   - menu rows
//   - text input rows
//   - number input rows
//   - toggle field rows
//   - horizontal button stack
//
// A single viewport (height=7) is attached to the navigator's internal
// ViewportController. The controller keeps the viewport offset in sync on
// every update.
func newModel() *model {
	sectionStyle := lipgloss.NewStyle().Faint(true)

	menuRows := []*menu.Model[string]{
		menu.New("Alpha", "alpha", "First option", selectMsg{Value: "alpha"}),
		menu.New("Beta", "beta", "Second option", selectMsg{Value: "beta"}),
		menu.New("Gamma", "gamma", "Third option", selectMsg{Value: "gamma"}),
		menu.New("Delta", "delta", "Fourth option", selectMsg{Value: "delta"}),
		menu.New("Epsilon", "epsilon", "Last option", selectMsg{Value: "epsilon"}),
	}
	_ = menu.NewController(menuRows, menu.WithMode[string](menu.ModeMultiSelect))

	textRows := []*textinput.Model{
		textinput.New(
			textinput.WithLabel("Name"),
			textinput.WithPlaceholder("Ada Lovelace"),
			textinput.WithWidth(30),
		),
		textinput.New(
			textinput.WithLabel("Email"),
			textinput.WithPlaceholder("ada@example.com"),
			textinput.WithWidth(30),
			textinput.WithValidator(func(value string) error {
				if !strings.Contains(value, "@") {
					return fmt.Errorf("Email must contain @")
				}

				return nil
			}),
		),
	}

	ageRow := textinput.New(
		textinput.WithLabel("Age"),
		textinput.WithPlaceholder("30"),
		textinput.WithWidth(10),
		textinput.WithFilter(textinput.NumberFilter),
		textinput.WithValidator(func(value string) error {
			n, err := strconv.Atoi(value)
			if err != nil || n < 18 {
				return fmt.Errorf("Must be at least 18")
			}

			return nil
		}),
	)

	toggleRows := []*toggle.Model{
		toggle.New("Notifications"),
		toggle.New("Dark mode", toggle.WithValue(true)),
		toggle.New("Auto-save"),
	}

	selectRow := selectfield.New([]selectfield.Option[string]{
		{Value: "go", Label: "Go"},
		{Value: "rust", Label: "Rust"},
		{Value: "ts", Label: "TypeScript"},
		{Value: "zig", Label: "Zig"},
	},
		selectfield.WithLabel[string]("Language"),
		selectfield.WithValidator[string](func(value string) error {
			if value == "zig" {
				return fmt.Errorf("Zig is not allowed")
			}

			return nil
		}),
	)

	buttonStack := button.NewStack(
		button.New("Save", pressMsg{Label: "Save"}),
		button.New("Cancel", pressMsg{Label: "Cancel"}),
		button.New("Help", pressMsg{Label: "Help"}),
	)
	buttonStack.SetStyles(button.StackStyles{
		Wrapper: lipgloss.NewStyle().MarginTop(1),
	})

	teaRows := make([]tea.Model, 0,
		1+len(menuRows)+1+len(textRows)+1+1+1+len(toggleRows)+2,
	)

	teaRows = append(teaRows, newLabel("─ Menu ─", sectionStyle))
	for _, r := range menuRows {
		teaRows = append(teaRows, r)
	}

	teaRows = append(teaRows, newLabel("─ Text ─", sectionStyle))
	for _, r := range textRows {
		teaRows = append(teaRows, r)
	}

	teaRows = append(teaRows, newLabel("─ Number ─", sectionStyle))
	teaRows = append(teaRows, ageRow)

	teaRows = append(teaRows, newLabel("─ Toggles ─", sectionStyle))
	for _, r := range toggleRows {
		teaRows = append(teaRows, r)
	}

	teaRows = append(teaRows, newLabel("─ Select ─", sectionStyle))
	teaRows = append(teaRows, selectRow)

	teaRows = append(teaRows, buttonStack)

	outer := navigator.New(teaRows...)

	viewport := scrollview.New()
	viewport.SetWidth(defaultWidth)

	outer.ViewportController().SetHeight(defaultHeight)
	outer.ViewportController().SetViewport(viewport)

	return &model{
		nav:         outer,
		viewport:    viewport,
		maxHeight:   defaultHeight,
		titleStyle:  lipgloss.NewStyle().Bold(true).MarginBottom(1),
		hintStyle:   lipgloss.NewStyle().Faint(true).MarginTop(1),
		cursorStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("3")),
		statusStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("#53d1ff")),
		menuRows:    menuRows,
		textRows:    textRows,
		ageRow:      ageRow,
		toggleRows:  toggleRows,
		selectRow:   selectRow,
	}
}

// Init focuses the first row and initialises all rows.
func (m *model) Init() tea.Cmd {
	return tea.Batch(m.nav.FocusFirst(), m.nav.Init(), m.viewport.Init())
}

// Update routes messages to the navigator and records user actions.
func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case selectMsg:
		m.lastAction = fmt.Sprintf("Selected: %s", msg.Value)

		return m, nil

	case pressMsg:
		if msg.Label == "Save" {
			if err := m.validate(); err != nil {
				m.lastAction = err.Error()
			} else {
				m.lastAction = "Saved"
			}
		} else {
			m.lastAction = fmt.Sprintf("Pressed: %s", msg.Label)
		}

		return m, nil

	case tea.KeyMsg:
		quitKey := key.NewBinding(key.WithKeys("q"))
		if key.Matches(msg, quitKey) {
			return m, tea.Quit
		}

	case tea.WindowSizeMsg:
		m.maxHeight = max(1, msg.Height-heightPadding)
		m.viewport.SetWidth(msg.Width)
		m.syncViewportHeight()

		return m, nil
	}

	updated, cmd := m.nav.Update(msg)

	updatedNav, ok := updated.(*navigator.Model)
	if !ok {
		return m, cmd
	}

	m.nav = updatedNav
	m.syncViewportHeight()

	return m, cmd
}

// View composes the final screen from the navigator's clipped viewport output.
func (m *model) View() tea.View {
	content := lipgloss.JoinVertical(lipgloss.Left,
		m.titleStyle.Render("Navigator Demo"),
		m.nav.ViewportController().View(),
		"",
		m.cursorStyle.Render(fmt.Sprintf("cursor line: %d", m.nav.CursorLine())),
		m.hintStyle.Render("↑/↓/k/j: navigate   enter: select   ←/→/h/l: buttons   q: quit"),
	)

	status := m.status()
	if status != "" {
		content = lipgloss.JoinVertical(
			lipgloss.Left,
			content,
			m.statusStyle.Render(status),
		)
	}

	return tea.NewView(content)
}

// validate runs validation on every validated row and returns the first error.
func (m *model) validate() error {
	for _, r := range m.textRows {
		if err := r.Validate(); err != nil {
			return err
		}
	}

	if err := m.ageRow.Validate(); err != nil {
		return err
	}

	if err := m.selectRow.Validate(); err != nil {
		return err
	}

	return nil
}

// status builds the status line from the last action, current values, and any
// validation error.
func (m *model) status() string {
	if m.lastAction != "" {
		return m.lastAction
	}

	if err := m.selectRow.Err(); err != nil {
		return err.Error()
	}

	for _, r := range m.textRows {
		if err := r.Err(); err != nil {
			return err.Error()
		}
	}

	if err := m.ageRow.Err(); err != nil {
		return err.Error()
	}

	parts := make([]string, 0,
		len(m.textRows)+len(m.toggleRows)+2,
	)

	parts = append(parts, fmt.Sprintf("%s=%s", m.selectRow.Label(), m.selectRow.Get()))

	for _, r := range m.textRows {
		parts = append(parts, fmt.Sprintf("%s=%s", r.Label(), r.Get()))
	}

	parts = append(parts, fmt.Sprintf("%s=%s", m.ageRow.Label(), m.ageRow.Get()))

	for _, r := range m.toggleRows {
		parts = append(parts, fmt.Sprintf("%s=%v", r.Label(), r.Get()))
	}

	if len(parts) == 0 {
		return ""
	}

	return fmt.Sprintf("Values: %s", parts)
}

// syncViewportHeight sets the viewport and controller heights to the smaller
// of the available screen space and the navigator's current content height.
func (m *model) syncViewportHeight() {
	contentHeight := m.nav.Height()
	h := max(1, min(m.maxHeight, contentHeight))
	m.nav.ViewportController().SetHeight(h)
}

func main() {
	_, err := tea.NewProgram(newModel()).Run()
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
