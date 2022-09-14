package chezmoibubbles

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/require"
)

var keyTypes = map[tea.KeyType]struct{}{
	tea.KeyCtrlC: {},
	tea.KeyEnter: {},
	tea.KeyEsc:   {},
}

//nolint:ireturn,nolintlint
func makeKeyMsg(r rune) tea.Msg {
	key := tea.Key{
		Type:  tea.KeyRunes,
		Runes: []rune{r},
	}
	if _, ok := keyTypes[tea.KeyType(r)]; ok {
		key = tea.Key{
			Type: tea.KeyType(r),
		}
	}
	return tea.KeyMsg(key)
}

//nolint:ireturn,nolintlint
func makeKeyMsgs(s string) []tea.Msg {
	msgs := make([]tea.Msg, 0, len(s))
	for _, r := range s {
		msgs = append(msgs, makeKeyMsg(r))
	}
	return msgs
}

//nolint:ireturn,nolintlint
func testRunModelWithInput[M tea.Model](t *testing.T, model M, input string) M {
	t.Helper()
	for _, msg := range makeKeyMsgs(input) {
		m, _ := model.Update(msg)
		var ok bool
		model, ok = m.(M)
		require.True(t, ok)
	}
	return model
}

func newBool(b bool) *bool       { return &b }
func newInt64(i int64) *int64    { return &i }
func newString(s string) *string { return &s }
