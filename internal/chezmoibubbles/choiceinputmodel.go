package chezmoibubbles

import (
	"errors"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/twpayne/chezmoi/v2/internal/chezmoi"
)

type ChoiceInputModel struct {
	uniqueAbbreviations map[string]string
	defaultValue        *string
	textInput           textinput.Model
	canceled            bool
}

func NewChoiceInputModel(prompt string, choices []string, defaultValue *string) ChoiceInputModel {
	textInput := textinput.New()
	textInput.Prompt = prompt + "? "
	textInput.Placeholder = strings.Join(choices, "/")
	if defaultValue != nil {
		textInput.Placeholder += ", default " + *defaultValue
	}
	allAbbreviations := make(map[string]struct{})
	for _, choice := range choices {
		for i := range choice {
			allAbbreviations[choice[:i+1]] = struct{}{}
		}
	}
	textInput.Validate = func(s string) error {
		if s == "" && defaultValue != nil {
			return nil
		}
		if _, ok := allAbbreviations[s]; ok {
			return nil
		}
		return errors.New("unknown or ambiguous choice")
	}
	textInput.Focus()
	return ChoiceInputModel{
		textInput:           textInput,
		uniqueAbbreviations: chezmoi.UniqueAbbreviations(choices),
		defaultValue:        defaultValue,
	}
}

func (m ChoiceInputModel) Canceled() bool {
	return m.canceled
}

func (m ChoiceInputModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m ChoiceInputModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			m.canceled = true
			return m, tea.Quit
		case tea.KeyEnter:
			value := m.textInput.Value()
			if value == "" && m.defaultValue != nil {
				m.textInput.SetValue(*m.defaultValue)
				return m, tea.Quit
			} else if value, ok := m.uniqueAbbreviations[value]; ok {
				m.textInput.SetValue(value)
				return m, tea.Quit
			}
		}
	}

	var cmd tea.Cmd
	m.textInput, cmd = m.textInput.Update(msg)
	if _, ok := m.uniqueAbbreviations[m.textInput.Value()]; ok {
		return m, tea.Quit
	}
	return m, cmd
}

func (m ChoiceInputModel) Value() string {
	value := m.textInput.Value()
	if value == "" && m.defaultValue != nil {
		return *m.defaultValue
	}
	return m.uniqueAbbreviations[value]
}

func (m ChoiceInputModel) View() string {
	return m.textInput.View()
}
