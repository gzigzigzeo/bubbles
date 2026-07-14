// Run: go run ./examples
package main

import (
	"fmt"
	"os"

	"charm.land/bubbles/v2/help"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/gzigzigzeo/bubbles/form"
	"github.com/gzigzigzeo/bubbles/form/button"
	"github.com/gzigzigzeo/bubbles/form/buttonstack"
	"github.com/gzigzigzeo/bubbles/form/field"
	"github.com/gzigzigzeo/bubbles/form/numberfield"
	"github.com/gzigzigzeo/bubbles/form/selectfield"
	"github.com/gzigzigzeo/bubbles/form/textfield"
	"github.com/gzigzigzeo/bubbles/form/textinputfield"
	"github.com/gzigzigzeo/bubbles/form/togglefield"
	"github.com/gzigzigzeo/bubbles/form/validate"
)

// styles for the example form, its headings, and the key help.
var (
	headingStyle        = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#53d1ff")).MarginTop(1).MarginBottom(1)
	valuesStyle         = lipgloss.NewStyle().Faint(true).MarginTop(1)
	shortKeyStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("#53d1ff"))
	shortDescStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("#a09caa"))
	shortSeparatorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#a09caa"))
)

// static is a tea.Model that renders a static string
type static string

func (static) Init() tea.Cmd                         { return nil }
func (s static) Update(tea.Msg) (tea.Model, tea.Cmd) { return s, nil }
func (s static) View() tea.View                      { return tea.NewView(string(s)) }

// fields holds the form's input fields and the dependent controls that are enabled/disabled by the "Enabled" toggle.
type fields struct {
	name, company            *textfield.Model
	age, teamSize            *numberfield.Model
	region, role             *selectfield.Model[string]
	updates, enabled, remote *togglefield.Model
	dependent                []field.Disableable
}

// model is a the full page form model
type model struct {
	*form.Model
	fields fields
	status string
	help   help.Model
}

func positive(value int) string {
	if value <= 0 {
		return "must be greater than zero"
	}
	return ""
}

func mustBeTrue(value bool) string {
	if !value {
		return "required for this demo"
	}
	return ""
}

func newModel() *model {
	h := help.New()

	h.Styles.ShortKey = shortKeyStyle
	h.Styles.ShortDesc = shortDescStyle
	h.Styles.ShortSeparator = shortSeparatorStyle

	m := &model{help: h}
	f := &m.fields

	f.name = textfield.New(
		textfield.WithPlaceholder("Ada Lovelace"),
		field.WithStyles[*textfield.Model](textinputfield.DefaultStyles()),
	)
	f.age = numberfield.New(
		field.WithStyles[*numberfield.Model](textinputfield.DefaultStyles()),
	)
	f.region = selectfield.NewFromStrings(regions(),
		field.WithStyles[*selectfield.Model[string]](selectfield.DefaultStyles()),
	)
	f.updates = togglefield.New(
		field.WithStyles[*togglefield.Model](togglefield.DefaultStyles("yes", "no")),
	)
	f.enabled = togglefield.New(
		field.WithStyles[*togglefield.Model](togglefield.DefaultStyles("enabled", "disabled")),
	)
	f.enabled.Set(true)
	f.company = textfield.New(
		textfield.WithPlaceholder("Analytical Engines"),
		field.WithStyles[*textfield.Model](textinputfield.DefaultStyles()),
	)
	f.teamSize = numberfield.New(
		field.WithStyles[*numberfield.Model](textinputfield.DefaultStyles()),
	)
	f.role = selectfield.NewFromStrings([]string{"Engineer", "Designer", "Product manager"},
		field.WithStyles[*selectfield.Model[string]](selectfield.DefaultStyles()),
	)
	f.remote = togglefield.New(
		field.WithStyles[*togglefield.Model](togglefield.DefaultStyles("remote", "on-site")),
	)

	// validate := func() tea.Msg {
	// 	valid := m.Validate()
	// 	if !valid {
	// 		m.FocusFirstError()
	// 	}
	// 	return valid
	// }
	btnStyle := button.Styles{
		Focused:  lipgloss.NewStyle().Background(lipgloss.Color("#53d1ff")).Foreground(lipgloss.Color("#000000")).Bold(true).PaddingLeft(1).PaddingRight(1).MarginRight(1),
		Blurred:  lipgloss.NewStyle().Foreground(lipgloss.Color("#d3d3d9")).PaddingLeft(1).PaddingRight(1).MarginRight(1),
		Disabled: lipgloss.NewStyle().Foreground(lipgloss.Color("#5C5C5C")).PaddingLeft(1).PaddingRight(1).MarginRight(1),
	}
	stackStyle := buttonstack.Styles{
		Block: lipgloss.NewStyle().PaddingLeft(1).PaddingTop(1).PaddingBottom(1),
	}

	accept := button.New("Accept", tea.Quit, field.WithStyles[*button.Model](btnStyle))
	cancel := button.New("Cancel", tea.Quit, field.WithStyles[*button.Model](btnStyle))
	accept.HintText = "Enter · submit the form"
	cancel.HintText = "Enter · quit"
	acceptAgain := button.New("Accept", tea.Quit, field.WithStyles[*button.Model](btnStyle))
	cancelAgain := button.New("Cancel", tea.Quit, field.WithStyles[*button.Model](btnStyle))
	acceptAgain.HintText = "Enter · submit the form"
	cancelAgain.HintText = "Enter · quit"
	firstButtons := buttonstack.New(accept, cancel)
	firstButtons.SetStyles(stackStyle)
	secondButtons := buttonstack.New(acceptAgain, cancelAgain)
	secondButtons.SetStyles(stackStyle)
	f.dependent = []field.Disableable{f.company, f.teamSize, f.role, f.remote, secondButtons}

	formStyles := form.DefaultStyles()
	//formStyles.Gutter = lipgloss.NewStyle().SetString("  ")

	m.Model = form.New(
		form.WithStyles(formStyles),
		form.WithEntry(static(headingStyle.Render("Account details"))),
		form.WithEntry(form.NewField("Name", f.name, form.WithValidator(validate.NotEmptyString()), form.WithHint[string]("Your full name"))),
		form.WithEntry(form.NewField("Age", f.age, form.WithValidator(positive), form.WithHint[int]("Your age in years"))),
		form.WithEntry(form.NewField("Region", f.region, form.WithValidator(validate.NotEmptyString()), form.WithHint[string]("Your geographic region"))),
		form.WithEntry(form.NewField("Updates", f.updates, form.WithValidator(mustBeTrue), form.WithHint[bool]("Receive product updates by email"))),
		form.WithEntry(firstButtons),
		form.WithEntry(static(headingStyle.Render("Optional workplace details"))),
		form.WithEntry(form.NewField("Enabled", f.enabled, form.WithValidator(mustBeTrue), form.WithHint[bool]("Enable optional workplace details"))),
		form.WithEntry(form.NewField("Company", f.company, form.WithValidator(validate.NotEmptyString()), form.WithHint[string]("Name of your current employer"))),
		form.WithEntry(form.NewField("Team size", f.teamSize, form.WithValidator(positive), form.WithHint[int]("Number of people on your team"))),
		form.WithEntry(form.NewField("Role", f.role, form.WithValidator(validate.NotEmptyString()), form.WithHint[string]("Your primary job function"))),
		form.WithEntry(form.NewField("Remote", f.remote, form.WithValidator(mustBeTrue), form.WithHint[bool]("Whether you work remotely"))),
		form.WithEntry(secondButtons),
	)
	m.syncDependent()
	return m
}

func regions() []string {
	return []string{"Africa", "Antarctica", "Asia", "Australia", "Europe", "North America", "South America", "Central America", "Middle East", "Caribbean", "Oceania"}
}

func (m *model) syncDependent() {
	for _, control := range m.fields.dependent {
		if m.fields.enabled.Get() {
			control.Enable()
		} else {
			control.Disable()
		}
	}
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		m.SetWidth(max(1, msg.Width-4))
		m.SetHeight(max(1, msg.Height-8))
		m.help.SetWidth(max(1, msg.Width-4))

	case bool:
		m.status = "Fix the highlighted field."
		if msg {
			m.status = "Form is valid."
		}
		return m, nil
	}

	cmd := m.Model.Update(msg)
	m.syncDependent()
	return m, cmd
}

func (m *model) View() tea.View {
	f := m.fields
	values := fmt.Sprintf("Values: name=%q age=%d region=%q updates=%t | enabled=%t company=%q team=%d role=%q remote=%t", f.name.Get(), f.age.Get(), f.region.Get(), f.updates.Get(), f.enabled.Get(), f.company.Get(), f.teamSize.Get(), f.role.Get(), f.remote.Get())
	content := m.Model.View()

	if keys := m.help.ShortHelpView(m.Model.Keys()); keys != "" {
		content += "\n" + keys
	}

	if hint := m.Model.RenderHint(); hint != "" {
		content += "\n" + hint
	}

	content += "\n" + valuesStyle.Render(values)

	if m.status != "" {
		content += "\n" + valuesStyle.Render(m.status)
	}

	return tea.NewView(content)
}

func main() {
	if _, err := tea.NewProgram(newModel()).Run(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
