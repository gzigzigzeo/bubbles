package prompt

import (
	"strings"

	"charm.land/bubbles/v2/cursor"
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

var enterKey = key.NewBinding(key.WithKeys("enter"))

// AnsweredMsg is emitted by Update when the user presses one of the accepted
// keys. Source identifies which Prompt sent the answer so a parent with
// multiple prompts can dispatch by identity.
type AnsweredMsg struct {
	Source *Prompt
	Answer rune // the key rune that was pressed, e.g. 'y' or 'n'
}

// Prompt is a configurable inline prompt that accepts one of the provided
// single-character keys, then echoes the chosen key. The choice hint (e.g.
// [y/n]) is rendered automatically — do not include it in the question string.
// Use SetDefault to designate a default key and SetAcceptByEnter to control
// whether Enter triggers that default.
type Prompt struct {
	question    string
	keys        []string
	bindings    []key.Binding
	defaultKey  rune // zero means no default
	rejectEnter bool // when true Enter is ignored even if defaultKey is set
	answer      *rune
	focused     bool
	styles      Styles

	// Cursor is the blinking cursor model. It is exported so callers can
	// configure Cursor.SetChar or other cursor options.
	Cursor cursor.Model
}

// New creates a Prompt that accepts any of the given single-character keys.
// Example: New("Continue?", "y", "n")
func New(question string, keys ...string) *Prompt {
	bindings := make([]key.Binding, len(keys))
	for i, k := range keys {
		bindings[i] = key.NewBinding(key.WithKeys(k))
	}

	return &Prompt{
		question: question,
		keys:     keys,
		bindings: bindings,
		Cursor:   cursor.New(),
	}
}

// SetStyles injects styles and applies the cursor appearance.
func (p *Prompt) SetStyles(s Styles) {
	p.styles = s
	p.Cursor.Style = s.CursorStyle
	p.Cursor.TextStyle = s.CursorTextStyle
}

// SetDefault marks the first rune of k as the default answer. When a default
// is set and Enter acceptance is enabled, pressing Enter emits AnsweredMsg with
// that rune as if the user had pressed it directly. The default key is shown
// in upper case inside the choice hint (e.g. [Y/n]).
func (p *Prompt) SetDefault(k string) {
	runes := []rune(k)
	if len(runes) > 0 {
		p.defaultKey = runes[0]
	}
}

// SetAcceptByEnter controls whether the Enter key triggers the default answer.
// The default is true. Set to false to require an explicit key press.
func (p *Prompt) SetAcceptByEnter(accept bool) {
	p.rejectEnter = !accept
}

// Init starts cursor blinking without changing focus or answer state.
// Call this from your tea.Model Init to satisfy the tea.Model interface.
func (p *Prompt) Init() tea.Cmd {
	return p.Cursor.Focus()
}

// Focus focuses the prompt, resets any previous answer, and starts the cursor.
func (p *Prompt) Focus() tea.Cmd {
	p.focused = true
	p.answer = nil
	p.Cursor.SetChar(" ")
	return p.Cursor.Focus()
}

// Blur removes focus and stops cursor blinking.
func (p *Prompt) Blur() {
	p.focused = false
	p.Cursor.Blur()
}

// Focused reports whether the prompt is focused.
func (p *Prompt) Focused() bool {
	return p.focused
}

// Value returns a pointer to the rune the user answered with, or nil if no
// answer has been given yet.
func (p *Prompt) Value() *rune {
	return p.answer
}

// IsMyAnswer reports whether msg is an AnsweredMsg emitted by this prompt and
// returns the answer rune. It encapsulates the common dispatch pattern:
//
//	if ans, ok := p.IsMyAnswer(msg); ok { ... }
func (p *Prompt) IsMyAnswer(msg tea.Msg) (rune, bool) {
	am, ok := msg.(AnsweredMsg)
	if !ok || am.Source != p {
		return 0, false
	}
	return am.Answer, true
}

// Update handles messages when focused: forwards to the cursor model for blink
// animation and checks whether the key pressed is one of the accepted keys.
func (p *Prompt) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if !p.focused {
		return p, nil
	}

	newCur, curCmd := p.Cursor.Update(msg)
	p.Cursor = newCur

	if km, ok := msg.(tea.KeyMsg); ok {
		for i, b := range p.bindings {
			if key.Matches(km, b) {
				r := []rune(p.keys[i])[0]
				p.answer = &r
				return p, func() tea.Msg { return AnsweredMsg{Source: p, Answer: r} }
			}
		}
		if !p.rejectEnter && p.defaultKey != 0 && key.Matches(km, enterKey) {
			r := p.defaultKey
			p.answer = &r
			return p, func() tea.Msg { return AnsweredMsg{Source: p, Answer: r} }
		}
	}

	return p, curCmd
}

// View renders the prompt: icon, question text with auto-generated choice hint,
// and either a blinking cursor (while waiting) or the echoed answer rune.
func (p *Prompt) View() tea.View {
	var rawQuestion string

	switch {
	case p.answer != nil:
		rawQuestion = p.question + " " + p.styles.Label.Render(p.choiceHint()+":") +
			p.styles.Echo.Render(" "+string(*p.answer))
	case p.focused:
		rawQuestion = p.question + " " + p.styles.Label.Render(p.choiceHint()) +
			" " + p.Cursor.View()
	default:
		rawQuestion = p.question + " " + p.styles.Label.Render(p.choiceHint())
	}

	iconCol := p.styles.Icon.Render()
	iconW := lipgloss.Width(iconCol)
	innerWidth := p.styles.Container.GetWidth() -
		p.styles.Container.GetPaddingLeft() -
		p.styles.Container.GetPaddingRight()
	questionWidth := innerWidth - iconW
	questionCol := p.styles.Label.Width(questionWidth).Render(rawQuestion)
	content := lipgloss.JoinHorizontal(lipgloss.Top, iconCol, questionCol)

	return tea.NewView(p.styles.Container.Render(content))
}

// choiceHint returns a formatted list of accepted keys, e.g. "[Y/n]".
// The default key (if any) is shown in upper case.
func (p *Prompt) choiceHint() string {
	parts := make([]string, len(p.keys))
	for i, k := range p.keys {
		runes := []rune(k)
		if len(runes) > 0 && p.defaultKey != 0 && runes[0] == p.defaultKey {
			parts[i] = strings.ToUpper(k)
		} else {
			parts[i] = k
		}
	}
	return "[" + strings.Join(parts, "/") + "]"
}
