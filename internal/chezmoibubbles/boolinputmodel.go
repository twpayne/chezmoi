package chezmoibubbles

import (
	"strconv"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/twpayne/chezmoi/internal/chezmoi"
)

type BoolInputModel struct {
	textInput    textinput.Model
	defaultValue *bool
	canceled     bool
}

func NewBoolInputModel(prompt string, defaultValue *bool) BoolInputModel {
	textInput := textinput.New()
	textInput.Prompt = prompt + "? "
	textInput.Placeholder = "bool"
	if defaultValue != nil {
		textInput.Placeholder += ", default " + strconv.FormatBool(*defaultValue)
	}
	textInput.Validate = func(value string) error {
		if value == "" && defaultValue != nil {
			return nil
		}
		_, err := chezmoi.ParseBool(value)
		return err
	}
	textInput.Focus()
	return BoolInputModel{
		textInput:    textInput,
		defaultValue: defaultValue,
	}
}

func (m BoolInputModel) Canceled() bool {
	return m.canceled
}

func (m BoolInputModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m BoolInputModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			m.canceled = true
			return m, tea.Quit
		case tea.KeyEnter:
			if m.defaultValue != nil {
				m.textInput.SetValue(strconv.FormatBool(*m.defaultValue))
				return m, tea.Quit
			}
		}
	}
	var cmd tea.Cmd
	m.textInput, cmd = m.textInput.Update(msg)
	if _, err := chezmoi.ParseBool(m.textInput.Value()); err == nil {
		return m, tea.Quit
	}
	return m, cmd
}

func (m BoolInputModel) Value() bool {
	valueStr := m.textInput.Value()
	if valueStr == "" && m.defaultValue != nil {
		return *m.defaultValue
	}
	value, _ := chezmoi.ParseBool(valueStr)
	return value
}

func (m BoolInputModel) View() string {
	return m.textInput.View()
}
