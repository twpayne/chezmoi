package chezmoibubbles

import (
	"testing"

	"github.com/alecthomas/assert/v2"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/twpayne/chezmoi/v2/internal/chezmoiset"
)

var keyTypes = chezmoiset.New(
	tea.KeyCtrlC,
	tea.KeyEnter,
	tea.KeyEsc,
)

func makeKeyMsg(r rune) tea.Msg { //nolint:ireturn,nolintlint
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

func makeKeyMsgs(s string) []tea.Msg { //nolint:ireturn,nolintlint
	msgs := make([]tea.Msg, 0, len(s))
	for _, r := range s {
		msgs = append(msgs, makeKeyMsg(r))
	}
	return msgs
}

func testRunModelWithInput[M tea.Model]( //nolint:ireturn,nolintlint
	t *testing.T,
	model M,
	input string,
) M {
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
