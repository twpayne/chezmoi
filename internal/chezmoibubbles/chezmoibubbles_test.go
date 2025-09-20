package chezmoibubbles

import (
	"testing"

	"github.com/alecthomas/assert/v2"
	tea "github.com/charmbracelet/bubbletea"

	"chezmoi.io/chezmoi/internal/chezmoiset"
)

var keyTypes = chezmoiset.New(
	tea.KeyCtrlC,
	tea.KeyEnter,
	tea.KeyEsc,
)

func makeKeyMsg(r rune) tea.Msg {
	key := tea.Key{
		Type:  tea.KeyRunes,
		Runes: []rune{r},
	}
	if keyTypes.Contains(tea.KeyType(r)) {
		key = tea.Key{
			Type: tea.KeyType(r),
		}
	}
	return tea.KeyMsg(key)
}

func makeKeyMsgs(s string) []tea.Msg {
	msgs := make([]tea.Msg, len(s))
	for i, r := range s {
		msgs[i] = makeKeyMsg(r)
	}
	return msgs
}

func testRunModelWithInput[M tea.Model](t *testing.T, model M, input string) M {
	t.Helper()
	for _, msg := range makeKeyMsgs(input) {
		m, _ := model.Update(msg)
		var ok bool
		model, ok = m.(M)
		assert.True(t, ok)
	}
	return model
}

func newValue[T any](value T) *T {
	return &value
}
