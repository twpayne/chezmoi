package chezmoibubbles

import (
	"slices"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/paginator"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// This is adapted from github.com/charmbracelet/gum/blob/main/choose/... for
// chezmoi, as of gum 0.15.2. Many things which are configurable in gum are not
// configurable in this module.
//
// Specific configuration options:
// - Provided options (`choices`) are also the labels.
// - No support for automatic select with one entry in the choice list.
// - Any number of options may be selected.
// - Default values are pre-selected. There is no support for wildcard
// 	 selection.
// - Options are shown in the order provided, selections are preserved in the
// 	 order selected.

type item struct {
	text     string
	selected bool
}

type keymap struct {
	Down,
	Up,
	Right,
	Left,
	Home,
	End,
	ToggleAll,
	Toggle,
	Abort,
	Quit,
	Submit key.Binding
}

type MultichoiceInputModel struct {
	items        []item
	defaultValue *[]string

	header      string
	quitting    bool
	submitted   bool
	index       int
	numSelected int

	paginator paginator.Model
	help      help.Model
	keymap    keymap
}

const (
	height           = 10
	cursor           = "> "
	selectedPrefix   = "✓ "
	unselectedPrefix = "• "
	cursorPrefix     = "• "
)

var (
	cursorStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("212"))
	headerStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("99"))
	itemStyle         = lipgloss.NewStyle()
	selectedItemStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("212"))
)

func NewMultichoiceInputModel(prompt string, choices []string, defaultValue *[]string) MultichoiceInputModel {
	var (
		subduedStyle     = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#847A85", Dark: "#979797"})
		verySubduedStyle = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#DDDADA", Dark: "#3C3C3C"})
	)

	currentSelected := 0
	hasSelectedItems := defaultValue != nil && len(*defaultValue) > 0
	startingIndex := 0

	items := make([]item, len(choices))
	for i, option := range choices {
		// Check if the option should be selected
		isSelected := hasSelectedItems && slices.Contains(*defaultValue, option)
		if isSelected {
			currentSelected++
		}
		items[i] = item{text: option, selected: isSelected}
	}

	// Use the pagination model to display the current and total number of
	// pages.
	pager := paginator.New()
	pager.SetTotalPages((len(items) + height - 1) / height)
	pager.PerPage = height
	pager.Type = paginator.Dots
	pager.ActiveDot = subduedStyle.Render("•")
	pager.InactiveDot = verySubduedStyle.Render("•")
	pager.KeyMap = paginator.KeyMap{}
	pager.Page = startingIndex / height

	km := keymap{
		Down:  key.NewBinding(key.WithKeys("down", "j", "ctrl+j", "ctrl+n")),
		Up:    key.NewBinding(key.WithKeys("up", "k", "ctrl+k", "ctrl+p")),
		Right: key.NewBinding(key.WithKeys("right", "l", "ctrl+f")),
		Left:  key.NewBinding(key.WithKeys("left", "h", "ctrl+b")),
		Home:  key.NewBinding(key.WithKeys("g", "home")),
		End:   key.NewBinding(key.WithKeys("G", "end")),
		ToggleAll: key.NewBinding(
			key.WithKeys("a", "A", "ctrl+a"),
			key.WithHelp("ctrl+a", "select all"),
		),
		Toggle: key.NewBinding(
			key.WithKeys(" ", "tab", "x", "ctrl+@"),
			key.WithHelp("x", "toggle"),
		),
		Abort: key.NewBinding(
			key.WithKeys("ctrl+c"),
			key.WithHelp("ctrl+c", "abort"),
		),
		Quit: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "quit"),
		),
		Submit: key.NewBinding(
			key.WithKeys("enter", "ctrl+q"),
			key.WithHelp("enter", "submit"),
		),
	}

	m := MultichoiceInputModel{
		items:        items,
		defaultValue: defaultValue,

		header: prompt,
		index:  startingIndex,

		paginator:   pager,
		numSelected: currentSelected,
		help:        help.New(),
		keymap:      km,
	}

	return m
}

func (m MultichoiceInputModel) Init() tea.Cmd {
	return nil
}

func (m MultichoiceInputModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		return m, nil

	case tea.KeyMsg:
		start, end := m.paginator.GetSliceBounds(len(m.items))
		km := m.keymap
		switch {
		case key.Matches(msg, km.Down):
			m.index++
			if m.index >= len(m.items) {
				m.index = 0
				m.paginator.Page = 0
			}
			if m.index >= end {
				m.paginator.NextPage()
			}
		case key.Matches(msg, km.Up):
			m.index--
			if m.index < 0 {
				m.index = len(m.items) - 1
				m.paginator.Page = m.paginator.TotalPages - 1
			}
			if m.index < start {
				m.paginator.PrevPage()
			}
		case key.Matches(msg, km.Right):
			m.index = clamp(m.index+height, 0, len(m.items)-1)
			m.paginator.NextPage()
		case key.Matches(msg, km.Left):
			m.index = clamp(m.index-height, 0, len(m.items)-1)
			m.paginator.PrevPage()
		case key.Matches(msg, km.End):
			m.index = len(m.items) - 1
			m.paginator.Page = m.paginator.TotalPages - 1
		case key.Matches(msg, km.Home):
			m.index = 0
			m.paginator.Page = 0
		case key.Matches(msg, km.ToggleAll):
			if m.numSelected < len(m.items) {
				m = m.selectAll()
			} else {
				m = m.deselectAll()
			}
		case key.Matches(msg, km.Quit):
			if m.numSelected < 1 && m.defaultValue != nil {
				for i := range m.items {
					if slices.Contains(*m.defaultValue, m.items[i].text) {
						m.items[i].selected = true
						m.numSelected++
					}
				}
			}
			m.quitting = true
			return m, tea.Quit
		case key.Matches(msg, km.Abort):
			m.quitting = true
			return m, tea.Interrupt
		case key.Matches(msg, km.Toggle):
			if m.items[m.index].selected {
				m.items[m.index].selected = false
				m.numSelected--
			} else {
				m.items[m.index].selected = true
				m.numSelected++
			}
		case key.Matches(msg, km.Submit):
			m.quitting = true
			if m.numSelected < 1 {
				m.items[m.index].selected = true
			}
			m.submitted = true
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.paginator, cmd = m.paginator.Update(msg)
	return m, cmd
}

func (m MultichoiceInputModel) selectAll() MultichoiceInputModel {
	for i := range m.items {
		if m.items[i].selected {
			continue
		}
		m.items[i].selected = true
		m.numSelected++
	}
	return m
}

func (m MultichoiceInputModel) deselectAll() MultichoiceInputModel {
	for i := range m.items {
		m.items[i].selected = false
	}
	m.numSelected = 0
	return m
}

func (m MultichoiceInputModel) View() string {
	if m.quitting {
		return ""
	}

	var s strings.Builder

	start, end := m.paginator.GetSliceBounds(len(m.items))
	for i, item := range m.items[start:end] {
		if i == m.index%height {
			s.WriteString(cursorStyle.Render(cursor))
		} else {
			s.WriteString(strings.Repeat(" ", lipgloss.Width(cursor)))
		}

		switch {
		case item.selected:
			s.WriteString(selectedItemStyle.Render(selectedPrefix + item.text))
		case i == m.index%height:
			s.WriteString(cursorStyle.Render(cursorPrefix + item.text))
		default:
			s.WriteString(itemStyle.Render(unselectedPrefix + item.text))
		}
		if i != height {
			s.WriteRune('\n')
		}
	}

	if m.paginator.TotalPages > 1 {
		s.WriteString(strings.Repeat("\n", height-m.paginator.ItemsOnPage(len(m.items))+1))
		s.WriteString("  " + m.paginator.View())
	}

	var parts []string

	if m.header != "" {
		parts = append(parts, headerStyle.Render(m.header))
	}
	parts = append(parts, s.String(), m.help.View(m.keymap))

	return lipgloss.JoinVertical(lipgloss.Left, parts...)
}

func (m MultichoiceInputModel) Value() []string {
	if (m.numSelected == 0 || !m.submitted) && m.defaultValue != nil {
		return *m.defaultValue
	}

	var out []string

	for _, item := range m.items {
		if item.selected {
			out = append(out, item.text)
		}
	}

	return out
}

func clamp(x, low, high int) int {
	if x < low {
		return low
	}
	if x > high {
		return high
	}
	return x
}

// FullHelp implements help.KeyMap.
func (k keymap) FullHelp() [][]key.Binding { return nil }

// ShortHelp implements help.KeyMap.
func (k keymap) ShortHelp() []key.Binding {
	return []key.Binding{
		k.Toggle,
		key.NewBinding(
			key.WithKeys("up", "down", "right", "left"),
			key.WithHelp("←↓↑→", "navigate"),
		),
		k.Submit,
		k.ToggleAll,
	}
}
