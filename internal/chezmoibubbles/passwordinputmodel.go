package chezmoibubbles

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type PasswordInputModel struct {
	textInput textinput.Model
	canceled  bool
}

func NewPasswordInputModel(prompt, placeholder string) PasswordInputModel {
	textInput := textinput.New()
	textInput.Prompt = prompt
	textInput.Placeholder = placeholder
	textInput.EchoMode = textinput.EchoNone
	textInput.Focus()
	return PasswordInputModel{
		textInput: textInput,
	}
}

func (m PasswordInputModel) Canceled() bool {
	return m.canceled
}

func (m PasswordInputModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m PasswordInputModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			m.canceled = true
			return m, tea.Quit
		case tea.KeyEnter:
			return m, tea.Quit
		}
	}
	var cmd tea.Cmd
	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

func (m PasswordInputModel) Value() string {
	return m.textInput.Value()
}

func (m PasswordInputModel) View() string {
	return m.textInput.View()
}
