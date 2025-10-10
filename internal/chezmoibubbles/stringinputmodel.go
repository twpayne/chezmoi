package chezmoibubbles

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type StringInputModel struct {
	textInput    textinput.Model
	defaultValue *string
	canceled     bool
}

func NewStringInputModel(prompt string, defaultValue *string) StringInputModel {
	textInput := textinput.New()
	textInput.Prompt = prompt
	textInput.Placeholder = "string"
	if defaultValue != nil {
		textInput.Placeholder += ", default " + *defaultValue
	}
	textInput.Focus()
	return StringInputModel{
		textInput:    textInput,
		defaultValue: defaultValue,
	}
}

func (m StringInputModel) Canceled() bool {
	return m.canceled
}

func (m StringInputModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m StringInputModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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

func (m StringInputModel) Value() string {
	value := m.textInput.Value()
	if value == "" && m.defaultValue != nil {
		return *m.defaultValue
	}
	return value
}

func (m StringInputModel) View() string {
	return m.textInput.View()
}
