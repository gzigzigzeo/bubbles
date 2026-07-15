package prompt_test

import (
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gzigzigzeo/bubbles/prompt"
)

// keyPress builds a KeyPressMsg for a single printable character.
func keyPress(ch string) tea.KeyPressMsg {
	r := rune(ch[0])

	return tea.KeyPressMsg{
		Text:        ch,
		Code:        r,
		Mod:         0,
		ShiftedCode: 0,
		BaseCode:    0,
		IsRepeat:    false,
	}
}

// enterPress returns a KeyPressMsg for the Enter key.
func enterPress() tea.KeyPressMsg {
	return tea.KeyPressMsg{
		Text:        "",
		Code:        tea.KeyEnter,
		Mod:         0,
		ShiftedCode: 0,
		BaseCode:    0,
		IsRepeat:    false,
	}
}

// runCmd executes cmd and returns the produced message, or nil.
func runCmd(cmd tea.Cmd) tea.Msg {
	if cmd == nil {
		return nil
	}

	return cmd()
}

// subCmds runs cmd and returns the individual sub-commands of the
// tea.BatchMsg it must produce. This lets a test run each sub-command
// independently and at its own pace, e.g. to defer running a Tick timer.
func subCmds(t *testing.T, cmd tea.Cmd) []tea.Cmd {
	t.Helper()
	require.NotNil(t, cmd)
	batch, ok := cmd().(tea.BatchMsg)
	require.True(t, ok)

	return batch
}

// findInvalidKeyMsg runs every cmd and returns the InvalidKeyMsg among their
// results, or nil if none produced one.
func findInvalidKeyMsg(cmds []tea.Cmd) *prompt.InvalidKeyMsg {
	for _, c := range cmds {
		if ik, ok := runCmd(c).(prompt.InvalidKeyMsg); ok {
			return &ik
		}
	}

	return nil
}

// testMsg is a stand-in Msg for tests that don't care which message a key
// emits, just that pressing it does something structurally sound.
type testMsg struct{ key rune }

// newPrompt builds a PromptModel accepting the given keys, each emitting
// testMsg, failing the test if construction errors (e.g. duplicate keys).
func newPrompt(t *testing.T, question string, keys ...rune) *prompt.Model {
	t.Helper()

	opts := make([]prompt.Option, len(keys))
	for i, k := range keys {
		opts[i] = prompt.WithOption(k, testMsg{key: k})
	}

	promptModel, err := prompt.New(question, opts...)
	require.NoError(t, err)

	return promptModel
}

func TestPrompt_UnfocusedIgnoresKeys(t *testing.T) {
	promptModel := newPrompt(t, "Continue?", 'y', 'n')
	_, cmd := promptModel.Update(keyPress("y"))

	assert.Nil(t, cmd, "unfocused prompt should produce no command")
	assert.Nil(t, promptModel.Value(), "unfocused prompt should have no answer")
}

func TestPrompt_FocusResetsAnswer(t *testing.T) {
	promptModel := newPrompt(t, "Continue?", 'y', 'n')
	promptModel.Focus()
	promptModel.Update(keyPress("y"))

	// Re-focus clears the answer.
	promptModel.Focus()

	assert.Nil(t, promptModel.Value(), "Focus should reset the answer")
}

func TestPrompt_UnregisteredKeyProducesNoAnswer(t *testing.T) {
	promptModel := newPrompt(t, "Continue?", 'y', 'n')
	promptModel.Focus()

	_, cmd := promptModel.Update(keyPress("x"))
	// cmd now carries the invalid-key flash (see TestPrompt_UnregisteredKeyEmitsInvalidKeyMsg),
	// but it must never be one of the registered keys' messages.
	if cmd != nil {
		msg := runCmd(cmd)
		_, isAnswer := msg.(testMsg)
		assert.False(t, isAnswer, "unregistered key must not be accepted")
	}

	assert.Nil(t, promptModel.Value())
}

func TestPrompt_EnterWithDefaultEmitsRegisteredMsg(t *testing.T) {
	promptModel, err := prompt.New("Continue?", prompt.WithYesNoDefaultYes())
	require.NoError(t, err)
	promptModel.Focus()

	_, cmd := promptModel.Update(enterPress())
	require.NotNil(t, cmd)

	msg := runCmd(cmd)
	_, ok := msg.(prompt.YesMsg)
	require.True(t, ok)
	require.NotNil(t, promptModel.Value())
	assert.Equal(t, 'Y', *promptModel.Value())
}

func TestPrompt_EnterWithoutDefaultIsInvalid(t *testing.T) {
	promptModel, err := prompt.New("Continue?",
		prompt.WithOption('y', testMsg{key: 'y'}), prompt.WithOption('n', testMsg{key: 'n'}),
		prompt.WithInvalidKeyDuration(time.Millisecond),
	)
	require.NoError(t, err)
	promptModel.Focus()

	_, cmd := promptModel.Update(enterPress())
	ik := findInvalidKeyMsg(subCmds(t, cmd))
	require.NotNil(t, ik)
	assert.Equal(t, "enter", ik.Key)
	assert.Nil(t, promptModel.Value())
}

func TestPrompt_AcceptByEnterFalseDisablesEnter(t *testing.T) {
	promptModel, err := prompt.New("Continue?",
		prompt.WithYesNoDefaultYes(),
		prompt.WithAcceptByEnter(false),
		prompt.WithInvalidKeyDuration(time.Millisecond),
	)
	require.NoError(t, err)
	promptModel.Focus()

	_, cmd := promptModel.Update(enterPress())
	ik := findInvalidKeyMsg(subCmds(t, cmd))
	require.NotNil(t, ik)
	assert.Equal(t, "enter", ik.Key)
	assert.Nil(t, promptModel.Value())
}

func TestPrompt_ViewContainsDefaultHint(t *testing.T) {
	promptModel, err := prompt.New("Continue?", prompt.WithYesNoDefaultYes())
	require.NoError(t, err)
	promptModel.Focus()

	view := promptModel.View().Content
	// default key 'Y' is shown uppercase in the choice hint: [Y/n]
	assert.Contains(t, view, "[Y/n]", "View should show the default key hint when focused")
}

func TestPrompt_ViewShowsEchoAfterAnswer(t *testing.T) {
	promptModel := newPrompt(t, "Continue?", 'y', 'n')
	promptModel.Focus()
	promptModel.Update(keyPress("n"))

	view := promptModel.View().Content
	assert.Contains(t, view, "n", "View should echo the answer")
}

func TestPrompt_Focused(t *testing.T) {
	promptModel := newPrompt(t, "Continue?", 'y', 'n')
	assert.False(t, promptModel.Focused())
	promptModel.Focus()
	assert.True(t, promptModel.Focused())
	promptModel.Blur()
	assert.False(t, promptModel.Focused())
}

func TestPrompt_UnregisteredKeyEmitsInvalidKeyMsg(t *testing.T) {
	promptModel, err := prompt.New("Continue?",
		prompt.WithOption('y', testMsg{key: 'y'}), prompt.WithOption('n', testMsg{key: 'n'}),
		prompt.WithInvalidKeyDuration(time.Millisecond),
	)
	require.NoError(t, err)
	promptModel.Focus()

	_, cmd := promptModel.Update(keyPress("x"))
	ik := findInvalidKeyMsg(subCmds(t, cmd))
	require.NotNil(t, ik)
	assert.Equal(t, "x", ik.Key)
	assert.Equal(t, promptModel, ik.Source)
	assert.Nil(t, promptModel.Value(), "invalid key must not set an answer")
}

func TestPrompt_ViewShowsInvalidKeyWithoutHidingHint(t *testing.T) {
	promptModel := newPrompt(t, "Continue?", 'y', 'n')
	promptModel.Focus()
	promptModel.Update(keyPress("x"))

	view := promptModel.View().Content
	assert.Contains(t, view, "[y/n]", "the choice hint must stay visible")
	assert.Contains(t, view, "x", "the invalid key should be shown in place of the cursor")
}

func TestPrompt_InvalidKeyFlashAutoClears(t *testing.T) {
	promptModel, err := prompt.New("Continue?",
		prompt.WithOption('y', testMsg{key: 'y'}), prompt.WithOption('n', testMsg{key: 'n'}),
		prompt.WithInvalidKeyDuration(2*time.Millisecond),
	)
	require.NoError(t, err)
	promptModel.Focus()

	_, cmd := promptModel.Update(keyPress("x"))
	require.Contains(t, promptModel.View().Content, "x")

	cmds := subCmds(t, cmd)
	clearMsg := runCmd(cmds[len(cmds)-1]) // the tick timer is always last
	promptModel.Update(clearMsg)

	assert.NotContains(t, promptModel.View().Content, "x", "flash should auto-clear")
	assert.Contains(t, promptModel.View().Content, "[y/n]", "the choice hint must still be visible")
}

func TestPrompt_InvalidKeyGenerationGuardsStaleTimer(t *testing.T) {
	promptModel, err := prompt.New("Continue?",
		prompt.WithOption('y', testMsg{key: 'y'}), prompt.WithOption('n', testMsg{key: 'n'}),
		prompt.WithInvalidKeyDuration(2*time.Millisecond),
	)
	require.NoError(t, err)
	promptModel.Focus()

	_, cmdA := promptModel.Update(keyPress("x"))
	cmdsA := subCmds(t, cmdA)

	_, cmdB := promptModel.Update(keyPress("z"))
	require.NotNil(t, cmdB)
	require.Contains(t, promptModel.View().Content, "z")

	staleClear := runCmd(cmdsA[len(cmdsA)-1]) // A's stale tick timer
	promptModel.Update(staleClear)

	assert.Contains(t, promptModel.View().Content, "z", "stale timer must not clear a newer flash")
}

func TestNew_DuplicateKeyReturnsError(t *testing.T) {
	promptModel, err := prompt.New(
		"Continue?",
		prompt.WithOption('y', testMsg{key: 'y'}),
		prompt.WithOption('y', testMsg{key: 'y'}),
	)
	require.Error(t, err)
	assert.Nil(t, promptModel)
}

func TestNew_DefaultKeyNotRegisteredReturnsError(t *testing.T) {
	promptModel, err := prompt.New(
		"Continue?",
		prompt.WithOption('n', testMsg{key: 'n'}),
		prompt.WithDefault('y'),
	)
	require.Error(t, err)
	assert.Nil(t, promptModel)
}

func TestNew_WithDefaultBeforeWithOption(t *testing.T) {
	promptModel, err := prompt.New("Continue?",
		prompt.WithDefault('y'),
		prompt.WithOption('y', testMsg{key: 'y'}), prompt.WithOption('n', testMsg{key: 'n'}),
	)
	require.NoError(t, err)
	assert.NotNil(t, promptModel)
}

func TestNew_NilMsgReturnsError(t *testing.T) {
	promptModel, err := prompt.New("Continue?", prompt.WithOption('y', nil))
	require.Error(t, err)
	assert.Nil(t, promptModel)
}

type testCustomMsg struct{ note string }

func TestPrompt_KeyWithCustomMsgEmitsIt(t *testing.T) {
	promptModel, err := prompt.New("Continue?",
		prompt.WithOption('y', testCustomMsg{note: "yes"}),
		prompt.WithOption('n', testMsg{key: 'n'}),
	)
	require.NoError(t, err)
	promptModel.Focus()

	_, cmd := promptModel.Update(keyPress("y"))
	msg := runCmd(cmd)
	cm, ok := msg.(testCustomMsg)
	require.True(t, ok)
	assert.Equal(t, "yes", cm.note)
	require.NotNil(t, promptModel.Value())
	assert.Equal(t, 'y', *promptModel.Value())
}

func TestPrompt_EnterDefaultEmitsKeysOwnMsg(t *testing.T) {
	promptModel, err := prompt.New("Continue?",
		prompt.WithOption('y', testCustomMsg{note: "yes"}),
		prompt.WithOption('n', testMsg{key: 'n'}),
		prompt.WithDefault('y'),
	)
	require.NoError(t, err)
	promptModel.Focus()

	_, cmd := promptModel.Update(enterPress())
	msg := runCmd(cmd)
	cm, ok := msg.(testCustomMsg)
	require.True(t, ok)
	assert.Equal(t, "yes", cm.note)
}

func TestPrompt_WithYesNoEmitsYesOrNo(t *testing.T) {
	py, err := prompt.New("Continue?", prompt.WithYesNo())
	require.NoError(t, err)
	py.Focus()
	_, cmd := py.Update(keyPress("y"))
	_, matched := runCmd(cmd).(prompt.YesMsg)
	assert.True(t, matched, "expected YesMsg")

	pn, err := prompt.New("Continue?", prompt.WithYesNo())
	require.NoError(t, err)
	pn.Focus()
	_, cmd = pn.Update(keyPress("n"))
	_, matched = runCmd(cmd).(prompt.NoMsg)
	assert.True(t, matched, "expected NoMsg")
}

func TestPrompt_WithYesNoDefaultNoEmitsNoMsg(t *testing.T) {
	promptModel, err := prompt.New("Continue?", prompt.WithYesNoDefaultNo())
	require.NoError(t, err)
	promptModel.Focus()

	_, cmd := promptModel.Update(enterPress())
	_, ok := runCmd(cmd).(prompt.NoMsg)
	require.True(t, ok)
	require.NotNil(t, promptModel.Value())
	assert.Equal(t, 'N', *promptModel.Value())
}

func TestPrompt_WithYesNoDefaultYesUsesUppercaseY(t *testing.T) {
	promptModel, err := prompt.New("Continue?", prompt.WithYesNoDefaultYes())
	require.NoError(t, err)
	promptModel.Focus()

	assert.Contains(t, promptModel.View().Content, "[Y/n]", "choice hint should show uppercase Y as default")

	_, cmd := promptModel.Update(keyPress("Y"))
	_, matched := runCmd(cmd).(prompt.YesMsg)
	assert.True(t, matched, "pressing uppercase Y should emit YesMsg")
}

func TestPrompt_WithSuccessStylesAppliesPreset(t *testing.T) {
	promptModel, err := prompt.New("Continue?", prompt.WithYesNo(), prompt.WithSuccessStyles())
	require.NoError(t, err)
	promptModel.Focus()

	assert.Contains(t, promptModel.View().Content, "✓", "WithSuccessStyles should apply NewSuccessStyles's icon")
}
