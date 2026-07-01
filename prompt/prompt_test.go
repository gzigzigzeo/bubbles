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
	return tea.KeyPressMsg{Text: ch, Code: r}
}

// enterPress returns a KeyPressMsg for the Enter key.
func enterPress() tea.KeyPressMsg {
	return tea.KeyPressMsg{Code: tea.KeyEnter}
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
	require.True(t, ok, "expected a tea.BatchMsg")

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

func TestPrompt_UnfocusedIgnoresKeys(t *testing.T) {
	p := prompt.New("Continue?", "y", "n")
	_, cmd := p.Update(keyPress("y"))
	assert.Nil(t, cmd, "unfocused prompt should produce no command")
	assert.Nil(t, p.Value(), "unfocused prompt should have no answer")
}

func TestPrompt_FocusResetsAnswer(t *testing.T) {
	p := prompt.New("Continue?", "y", "n")
	p.Focus() //nolint:errcheck
	p.Update(keyPress("y"))

	// Re-focus clears the answer.
	p.Focus() //nolint:errcheck
	assert.Nil(t, p.Value(), "Focus should reset the answer")
}

func TestPrompt_RegisteredKeyEmitsAnsweredMsg(t *testing.T) {
	p := prompt.New("Continue?", "y", "n")
	p.Focus() //nolint:errcheck

	_, cmd := p.Update(keyPress("y"))
	require.NotNil(t, cmd)

	msg := runCmd(cmd)
	am, ok := msg.(prompt.AnsweredMsg)
	require.True(t, ok, "expected AnsweredMsg")
	assert.Equal(t, 'y', am.Answer)
	assert.Equal(t, p, am.Source)

	val := p.Value()
	require.NotNil(t, val)
	assert.Equal(t, 'y', *val)
}

func TestPrompt_UnregisteredKeyProducesNoAnsweredMsg(t *testing.T) {
	p := prompt.New("Continue?", "y", "n")
	p.Focus() //nolint:errcheck

	_, cmd := p.Update(keyPress("x"))
	// cmd now carries the invalid-key flash (see TestPrompt_UnregisteredKeyEmitsInvalidKeyMsg),
	// but it must never be an AnsweredMsg.
	if cmd != nil {
		msg := runCmd(cmd)
		_, isAnswer := msg.(prompt.AnsweredMsg)
		assert.False(t, isAnswer, "unregistered key must not emit AnsweredMsg")
	}
	assert.Nil(t, p.Value())
}

func TestPrompt_EnterWithDefaultEmitsAnsweredMsg(t *testing.T) {
	p := prompt.New("Continue?", "y", "n")
	p.SetDefault("y")
	p.Focus() //nolint:errcheck

	_, cmd := p.Update(enterPress())
	require.NotNil(t, cmd)

	msg := runCmd(cmd)
	am, ok := msg.(prompt.AnsweredMsg)
	require.True(t, ok, "Enter with default should emit AnsweredMsg")
	assert.Equal(t, 'y', am.Answer)
}

func TestPrompt_EnterWithoutDefaultIsInvalid(t *testing.T) {
	p := prompt.New("Continue?", "y", "n")
	p.SetInvalidKeyFlashDuration(time.Millisecond)
	p.Focus() //nolint:errcheck

	_, cmd := p.Update(enterPress())
	ik := findInvalidKeyMsg(subCmds(t, cmd))
	require.NotNil(t, ik, "Enter without a default should be treated as an invalid key")
	assert.Equal(t, "enter", ik.Key)
	assert.Nil(t, p.Value())
}

func TestPrompt_IsMyAnswer(t *testing.T) {
	p1 := prompt.New("First?", "y", "n")
	p2 := prompt.New("Second?", "y", "n")
	p1.Focus() //nolint:errcheck

	_, cmd := p1.Update(keyPress("y"))
	msg := runCmd(cmd)

	ans, ok := p1.IsMyAnswer(msg)
	assert.True(t, ok, "IsMyAnswer should match own source")
	assert.Equal(t, 'y', ans)

	_, ok = p2.IsMyAnswer(msg)
	assert.False(t, ok, "IsMyAnswer should not match a different prompt")
}

func TestPrompt_IsMyAnswerReturnsFalseForOtherMsgs(t *testing.T) {
	p := prompt.New("Continue?", "y", "n")
	r, ok := p.IsMyAnswer(keyPress("y"))
	assert.False(t, ok)
	assert.Equal(t, rune(0), r)
}

func TestPrompt_SetAcceptByEnterFalseDisablesEnter(t *testing.T) {
	p := prompt.New("Continue?", "y", "n")
	p.SetDefault("y")
	p.SetAcceptByEnter(false)
	p.SetInvalidKeyFlashDuration(time.Millisecond)
	p.Focus() //nolint:errcheck

	_, cmd := p.Update(enterPress())
	ik := findInvalidKeyMsg(subCmds(t, cmd))
	require.NotNil(t, ik, "Enter should be treated as invalid when SetAcceptByEnter(false)")
	assert.Equal(t, "enter", ik.Key)
	assert.Nil(t, p.Value())
}

func TestPrompt_ViewContainsDefaultHint(t *testing.T) {
	p := prompt.New("Continue?", "y", "n")
	p.SetDefault("y")
	p.Focus() //nolint:errcheck

	view := p.View().Content
	// default key 'y' is shown uppercase in the choice hint: [Y/n]
	assert.Contains(t, view, "[Y/n]", "View should show the default key hint when focused")
}

func TestPrompt_ViewShowsEchoAfterAnswer(t *testing.T) {
	p := prompt.New("Continue?", "y", "n")
	p.Focus() //nolint:errcheck
	p.Update(keyPress("n"))

	view := p.View().Content
	assert.Contains(t, view, "n", "View should echo the answer")
}

func TestPrompt_Focused(t *testing.T) {
	p := prompt.New("Continue?", "y", "n")
	assert.False(t, p.Focused())
	p.Focus() //nolint:errcheck
	assert.True(t, p.Focused())
	p.Blur()
	assert.False(t, p.Focused())
}

func TestPrompt_UnregisteredKeyEmitsInvalidKeyMsg(t *testing.T) {
	p := prompt.New("Continue?", "y", "n")
	p.SetInvalidKeyFlashDuration(time.Millisecond)
	p.Focus() //nolint:errcheck

	_, cmd := p.Update(keyPress("x"))
	ik := findInvalidKeyMsg(subCmds(t, cmd))
	require.NotNil(t, ik, "expected InvalidKeyMsg")
	assert.Equal(t, "x", ik.Key)
	assert.Equal(t, p, ik.Source)
	assert.Nil(t, p.Value(), "invalid key must not set an answer")
}

func TestPrompt_ViewShowsInvalidKeyWithoutHidingHint(t *testing.T) {
	p := prompt.New("Continue?", "y", "n")
	p.Focus() //nolint:errcheck
	p.Update(keyPress("x"))

	view := p.View().Content
	assert.Contains(t, view, "[y/n]", "the choice hint must stay visible")
	assert.Contains(t, view, "x", "the invalid key should be shown in place of the cursor")
}

func TestPrompt_InvalidKeyFlashAutoClears(t *testing.T) {
	p := prompt.New("Continue?", "y", "n")
	p.SetInvalidKeyFlashDuration(2 * time.Millisecond)
	p.Focus() //nolint:errcheck

	_, cmd := p.Update(keyPress("x"))
	require.Contains(t, p.View().Content, "x")

	cmds := subCmds(t, cmd)
	clearMsg := runCmd(cmds[len(cmds)-1]) // the tick timer is always last
	p.Update(clearMsg)

	assert.NotContains(t, p.View().Content, "x", "flash should auto-clear")
	assert.Contains(t, p.View().Content, "[y/n]", "the choice hint must still be visible")
}

func TestPrompt_InvalidKeyGenerationGuardsStaleTimer(t *testing.T) {
	p := prompt.New("Continue?", "y", "n")
	p.SetInvalidKeyFlashDuration(2 * time.Millisecond)
	p.Focus() //nolint:errcheck

	_, cmdA := p.Update(keyPress("x"))
	cmdsA := subCmds(t, cmdA)

	_, cmdB := p.Update(keyPress("z"))
	require.NotNil(t, cmdB)
	require.Contains(t, p.View().Content, "z")

	staleClear := runCmd(cmdsA[len(cmdsA)-1]) // A's stale tick timer
	p.Update(staleClear)

	assert.Contains(t, p.View().Content, "z", "stale timer must not clear a newer flash")
}
