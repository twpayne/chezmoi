package chezmoibubbles

import (
	"strconv"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type IntInputModel struct {
	defaultValue *int64
	textInput    textinput.Model
	canceled     bool
}

func NewIntInputModel(prompt string, defaultValue *int64) IntInputModel {
	textInput := textinput.New()
	textInput.Prompt = prompt + "? "
	textInput.Placeholder = "int"
	if defaultValue != nil {
		textInput.Placeholder += ", default " + strconv.FormatInt(*defaultValue, 10)
	}
	textInput.Validate = func(value string) error {
		if value == "" && defaultValue != nil {
			return nil
		}
		if value == "-" {
			return nil
		}
		_, err := strconv.ParseInt(value, 10, 64)
		return err
	}
	textInput.Focus()
	return IntInputModel{
		textInput:    textInput,
		defaultValue: defaultValue,
	}
}

func (m IntInputModel) Canceled() bool {
	return m.canceled
}

func (m IntInputModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m IntInputModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			m.canceled = true
			return m, tea.Quit
		case tea.KeyEnter:
			if m.textInput.Value() == "" && m.defaultValue != nil {
				m.textInput.SetValue(strconv.FormatInt(*m.defaultValue, 10))
			}
			return m, tea.Quit
		}
	}
	var cmd tea.Cmd
	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

func (m IntInputModel) Value() int64 {
	valueStr := m.textInput.Value()
	if valueStr == "" && m.defaultValue != nil {
		return *m.defaultValue
	}
	value, _ := strconv.ParseInt(valueStr, 10, 64)
	return value
}

func (m IntInputModel) View() string {
	return m.textInput.View()
}
