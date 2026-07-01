package prompt

import (
	"strings"
	"time"

	"charm.land/bubbles/v2/cursor"
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

var enterKey = key.NewBinding(key.WithKeys("enter"))

// defaultInvalidKeyFlash is how long the invalid-key hint stays visible
// before automatically clearing, unless overridden with
// SetInvalidKeyFlashDuration.
const defaultInvalidKeyFlash = 600 * time.Millisecond

// AnsweredMsg is emitted by Update when the user presses one of the accepted
// keys. Source identifies which Prompt sent the answer so a parent with
// multiple prompts can dispatch by identity.
type AnsweredMsg struct {
	Source *Prompt
	Answer rune // the key rune that was pressed, e.g. 'y' or 'n'
}

// InvalidKeyMsg is emitted by Update when the user presses a key that is not
// one of the accepted keys (and not Enter with a default set). Prompt also
// shows a brief inline hint itself, so most callers can ignore this message;
// it exists for host apps that want to react too, e.g. a terminal bell.
type InvalidKeyMsg struct {
	Source *Prompt
	Key    string // string representation of the rejected key, e.g. "x", "up"
}

// invalidKeyClearedMsg is sent by a timer started when an invalid key is
// pressed. gen guards against a stale timer clearing a newer flash.
type invalidKeyClearedMsg struct {
	source *Prompt
	gen    int
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

	invalidKey       string        // non-empty while the invalid-key flash is showing
	invalidGen       int           // guards stale clear-timers after a newer invalid key
	invalidFlashTime time.Duration // how long the invalid-key flash stays visible

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
		question:         question,
		keys:             keys,
		bindings:         bindings,
		invalidFlashTime: defaultInvalidKeyFlash,
		Cursor:           cursor.New(),
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

// SetInvalidKeyFlashDuration overrides how long the invalid-key hint stays
// visible before automatically clearing. The default is 600ms.
func (p *Prompt) SetInvalidKeyFlashDuration(d time.Duration) {
	p.invalidFlashTime = d
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
	p.invalidKey = ""
	p.invalidGen++
	p.Cursor.SetChar(" ")
	return p.Cursor.Focus()
}

// Blur removes focus and stops cursor blinking.
func (p *Prompt) Blur() {
	p.focused = false
	p.invalidKey = ""
	p.invalidGen++
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
// Unrecognized keys trigger a brief inline "invalid" flash and an
// InvalidKeyMsg instead of being silently ignored.
func (p *Prompt) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if cm, ok := msg.(invalidKeyClearedMsg); ok {
		p.handleInvalidKeyCleared(cm)
		return p, nil
	}

	if !p.focused {
		return p, nil
	}

	newCur, curCmd := p.Cursor.Update(msg)
	p.Cursor = newCur

	km, ok := msg.(tea.KeyMsg)
	if !ok {
		return p, curCmd
	}

	return p, p.handleKeyPress(km, curCmd)
}

// handleInvalidKeyCleared clears the invalid-key flash if cm still matches
// the current generation, i.e. no newer invalid key has been pressed since
// this clear timer was started.
func (p *Prompt) handleInvalidKeyCleared(cm invalidKeyClearedMsg) {
	if cm.source == p && cm.gen == p.invalidGen {
		p.invalidKey = ""
	}
}

// handleKeyPress answers the prompt if km matches an accepted key or the
// default-Enter binding, otherwise flags km as an invalid key. curCmd is the
// cursor's own command for this message, preserved in either outcome.
func (p *Prompt) handleKeyPress(km tea.KeyMsg, curCmd tea.Cmd) tea.Cmd {
	if r, ok := p.matchedKey(km); ok {
		return p.accept(r)
	}

	if !p.rejectEnter && p.defaultKey != 0 && key.Matches(km, enterKey) {
		return p.accept(p.defaultKey)
	}

	return p.flagInvalid(km, curCmd)
}

// matchedKey reports the rune bound to km among the accepted keys, if any.
func (p *Prompt) matchedKey(km tea.KeyMsg) (rune, bool) {
	for i, b := range p.bindings {
		if key.Matches(km, b) {
			return []rune(p.keys[i])[0], true
		}
	}

	return 0, false
}

// accept records r as the answer, clears any invalid-key flash, and emits
// AnsweredMsg.
func (p *Prompt) accept(r rune) tea.Cmd {
	p.answer = &r
	p.invalidKey = ""

	return func() tea.Msg { return AnsweredMsg{Source: p, Answer: r} }
}

// flagInvalid shows km as a brief invalid-key flash in place of the cursor,
// schedules it to clear itself after invalidFlashTime, and emits
// InvalidKeyMsg.
func (p *Prompt) flagInvalid(km tea.KeyMsg, curCmd tea.Cmd) tea.Cmd {
	p.invalidGen++
	gen := p.invalidGen
	p.invalidKey = km.String()

	return tea.Batch(
		curCmd,
		func() tea.Msg { return InvalidKeyMsg{Source: p, Key: km.String()} },
		tea.Tick(p.invalidFlashTime, func(time.Time) tea.Msg {
			return invalidKeyClearedMsg{source: p, gen: gen}
		}),
	)
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
		marker := p.Cursor.View()
		if p.invalidKey != "" {
			marker = p.styles.Invalid.Render(p.invalidKey)
		}
		rawQuestion = p.question + " " + p.styles.Label.Render(p.choiceHint()) +
			" " + marker
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
