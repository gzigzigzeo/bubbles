package prompt

import (
	"fmt"
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
// WithInvalidKeyDuration.
const defaultInvalidKeyFlash = 600 * time.Millisecond

// invalidKeyMsg is emitted internally by flagInvalid when the user presses a
// key that is not one of the accepted keys (and not Enter with a default
// set). Update uses it to drive the brief inline "invalid" hint.
type invalidKeyMsg struct {
	Source *Model
	Key    string // string representation of the rejected key, e.g. "x", "up"
}

// invalidKeyClearedMsg is sent by a timer started when an invalid key is
// pressed. gen guards against a stale timer clearing a newer flash.
type invalidKeyClearedMsg struct {
	source *Model
	gen    int
}

// Option configures a PromptModel during construction via New.
type Option func(*Model)

// keyOption pairs an accepted key with the message it emits.
type keyOption struct {
	key rune
	msg tea.Msg
}

// WithOption registers an accepted key. msg must be non-nil — New returns an
// error if it isn't. Pressing key emits msg.
func WithOption(key rune, msg tea.Msg) Option {
	return func(p *Model) {
		p.options = append(p.options, keyOption{key: key, msg: msg})
	}
}

// YesMsg is emitted when the user presses 'y' on a prompt built with
// WithYesNo, WithYesNoDefaultYes, or WithYesNoDefaultNo.
type YesMsg struct{}

// NoMsg is emitted when the user presses 'n' on a prompt built with
// WithYesNo, WithYesNoDefaultYes, or WithYesNoDefaultNo.
type NoMsg struct{}

// WithYesNo registers the common y/n confirmation keys in one call: 'y'
// emits YesMsg, 'n' emits NoMsg. It doesn't set a default — use
// WithYesNoDefaultYes or WithYesNoDefaultNo for that.
func WithYesNo() Option {
	return func(p *Model) {
		WithOption('y', YesMsg{})(p)
		WithOption('n', NoMsg{})(p)
	}
}

// WithYesNoDefaultYes registers 'Y' (uppercase, the default answer) and 'n':
// pressing 'Y' or Enter emits YesMsg, pressing 'n' emits NoMsg.
func WithYesNoDefaultYes() Option {
	return func(p *Model) {
		WithOption('Y', YesMsg{})(p)
		WithOption('n', NoMsg{})(p)
		WithDefault('Y')(p)
	}
}

// WithYesNoDefaultNo registers 'y' and 'N' (uppercase, the default answer):
// pressing 'y' emits YesMsg, pressing 'N' or Enter emits NoMsg.
func WithYesNoDefaultNo() Option {
	return func(p *Model) {
		WithOption('y', YesMsg{})(p)
		WithOption('N', NoMsg{})(p)
		WithDefault('N')(p)
	}
}

// WithDefault marks key as the default answer: when Enter acceptance is
// enabled (WithAcceptByEnter, on by default) and no key has been pressed,
// Enter emits that key's registered Msg, exactly as if it had been pressed
// directly. New returns an error if key was never registered via WithOption.
func WithDefault(key rune) Option {
	return func(p *Model) { p.defaultKey = key }
}

// WithStyles sets the visual styles. New re-applies it (via SetStyles) after
// every Option has run, so Cursor.Style/CursorTextStyle end up correct
// regardless of whether WithCursor or WithStyles came first.
func WithStyles(s Styles) Option {
	return func(p *Model) { p.styles = s }
}

// WithAcceptByEnter controls whether the Enter key triggers the default
// answer set via WithDefault. Defaults to true.
func WithAcceptByEnter(accept bool) Option {
	return func(p *Model) { p.rejectEnter = !accept }
}

// WithInvalidKeyDuration overrides how long the invalid-key hint stays
// visible before automatically clearing. Defaults to 600ms.
func WithInvalidKeyDuration(d time.Duration) Option {
	return func(p *Model) { p.invalidFlashTime = d }
}

// WithCursor seeds the initial cursor (e.g. after calling SetChar) instead
// of the default cursor.New().
func WithCursor(c cursor.Model) Option {
	return func(p *Model) { p.Cursor = c }
}

// Model is a configurable inline prompt that accepts one of the
// provided single-character keys, then echoes the chosen key. The choice
// hint (e.g. [y/n]) is rendered automatically — do not include it in the
// question string. Use WithDefault to designate a default key and
// WithAcceptByEnter to control whether Enter triggers that default.
type Model struct {
	question    string
	options     []keyOption
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

// New creates a PromptModel from the given Options (WithOption, WithYesNo,
// WithDefault, WithStyles, WithAcceptByEnter, WithInvalidKeyDuration,
// WithCursor). It returns an error if two options register the same key, if
// any option's Msg is nil, or if WithDefault names a key that wasn't
// registered via WithOption.
func New(question string, opts ...Option) (*Model, error) {
	p := &Model{
		question:         question,
		invalidFlashTime: defaultInvalidKeyFlash,
		Cursor:           cursor.New(),
	}

	for _, opt := range opts {
		opt(p)
	}

	seen := make(map[rune]bool, len(p.options))
	for _, o := range p.options {
		if seen[o.key] {
			return nil, fmt.Errorf("prompt: duplicate key %q", o.key)
		}
		seen[o.key] = true
		if o.msg == nil {
			return nil, fmt.Errorf("prompt: option for key %q has a nil Msg", o.key)
		}
	}

	if p.defaultKey != 0 && !seen[p.defaultKey] {
		return nil, fmt.Errorf("prompt: default key %q is not among the registered options", p.defaultKey)
	}

	p.SetStyles(p.styles)

	return p, nil
}

// SetStyles injects styles and applies the cursor appearance.
func (p *Model) SetStyles(s Styles) {
	p.styles = s
	p.Cursor.Style = s.CursorStyle
	p.Cursor.TextStyle = s.CursorTextStyle
}

// Init starts cursor blinking without changing focus or answer state.
// Call this from your tea.Model Init to satisfy the tea.Model interface.
func (p *Model) Init() tea.Cmd {
	return p.Cursor.Focus()
}

// Focus focuses the prompt, resets any previous answer, and starts the cursor.
func (p *Model) Focus() tea.Cmd {
	p.focused = true
	p.answer = nil
	p.invalidKey = ""
	p.invalidGen++
	p.Cursor.SetChar(" ")
	return p.Cursor.Focus()
}

// Blur removes focus and stops cursor blinking.
func (p *Model) Blur() {
	p.focused = false
	p.invalidKey = ""
	p.invalidGen++
	p.Cursor.Blur()
}

// Focused reports whether the prompt is focused.
func (p *Model) Focused() bool {
	return p.focused
}

// Value returns a pointer to the rune the user answered with, or nil if no
// answer has been given yet.
func (p *Model) Value() *rune {
	return p.answer
}

// Update handles messages when focused: forwards to the cursor model for blink
// animation and checks whether the key pressed is one of the accepted keys.
// Unrecognized keys trigger a brief inline "invalid" flash instead of being
// silently ignored.
func (p *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
func (p *Model) handleInvalidKeyCleared(cm invalidKeyClearedMsg) {
	if cm.source == p && cm.gen == p.invalidGen {
		p.invalidKey = ""
	}
}

// handleKeyPress answers the prompt if km matches an accepted key or the
// default-Enter binding, otherwise flags km as an invalid key. curCmd is the
// cursor's own command for this message, preserved in either outcome.
func (p *Model) handleKeyPress(km tea.KeyMsg, curCmd tea.Cmd) tea.Cmd {
	if i, ok := p.matchedKeyIndex(km); ok {
		return p.accept(p.options[i].key, p.options[i].msg)
	}

	if !p.rejectEnter && p.defaultKey != 0 && key.Matches(km, enterKey) {
		for _, o := range p.options {
			if o.key == p.defaultKey {
				return p.accept(o.key, o.msg)
			}
		}
	}

	return p.flagInvalid(km, curCmd)
}

// matchedKeyIndex reports the index of the accepted key bound to km, if any.
func (p *Model) matchedKeyIndex(km tea.KeyMsg) (int, bool) {
	s := km.String()
	for i, o := range p.options {
		if string(o.key) == s {
			return i, true
		}
	}

	return 0, false
}

// accept records r as the answer, clears any invalid-key flash, and emits msg.
func (p *Model) accept(r rune, msg tea.Msg) tea.Cmd {
	p.answer = &r
	p.invalidKey = ""

	return func() tea.Msg { return msg }
}

// flagInvalid shows km as a brief invalid-key flash in place of the cursor,
// schedules it to clear itself after invalidFlashTime, and emits
// invalidKeyMsg.
func (p *Model) flagInvalid(km tea.KeyMsg, curCmd tea.Cmd) tea.Cmd {
	p.invalidGen++
	gen := p.invalidGen
	p.invalidKey = km.String()

	return tea.Batch(
		curCmd,
		func() tea.Msg { return invalidKeyMsg{Source: p, Key: km.String()} },
		tea.Tick(p.invalidFlashTime, func(time.Time) tea.Msg {
			return invalidKeyClearedMsg{source: p, gen: gen}
		}),
	)
}

// View renders the prompt: icon, question text with auto-generated choice hint,
// and either a blinking cursor (while waiting) or the echoed answer rune.
func (p *Model) View() tea.View {
	rawQuestion := p.question + " " + p.styles.Label.Render(p.choiceHint())

	switch {
	case p.answer != nil:
		rawQuestion = rawQuestion + " " + p.styles.Echo.Render(" "+string(*p.answer))
	case p.focused:
		marker := p.Cursor.View()
		if p.invalidKey != "" {
			marker = p.styles.Invalid.Render(p.invalidKey)
		}
		rawQuestion = rawQuestion + " " + marker
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
func (p *Model) choiceHint() string {
	parts := make([]string, len(p.options))

	for i, o := range p.options {
		if p.defaultKey != 0 && o.key == p.defaultKey {
			parts[i] = strings.ToUpper(string(o.key))
		} else {
			parts[i] = string(o.key)
		}
	}

	return "[" + strings.Join(parts, "/") + "]:"
}
