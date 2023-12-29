package main

import (
	"fmt"
	"os"
	"reflect"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	initialInputs = 2
	charLimit     = 0
	maxInputs     = 6
	minInputs     = 1
	helpHeight    = 5
)

var (
	cursorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("212"))

	cursorLineStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("57")).
			Foreground(lipgloss.Color("230"))

	placeholderStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("238"))

	endOfBufferStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("235"))

	focusedPlaceholderStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("99"))

	focusedBorderStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("238"))

	blurredBorderStyle = lipgloss.NewStyle().
				Border(lipgloss.HiddenBorder())
)

func newTextarea() textarea.Model {
	t := textarea.New()
	t.Prompt = ""
	t.Placeholder = "Type something"
	t.ShowLineNumbers = true
	t.Cursor.Style = cursorStyle
	t.FocusedStyle.Placeholder = focusedPlaceholderStyle
	t.BlurredStyle.Placeholder = placeholderStyle
	t.FocusedStyle.CursorLine = cursorLineStyle
	t.FocusedStyle.Base = focusedBorderStyle
	t.BlurredStyle.Base = blurredBorderStyle
	t.FocusedStyle.EndOfBuffer = endOfBufferStyle
	t.BlurredStyle.EndOfBuffer = endOfBufferStyle
	t.KeyMap.DeleteWordBackward.SetEnabled(false)
	t.KeyMap.LineNext = key.NewBinding(key.WithKeys("down"))
	t.KeyMap.LinePrevious = key.NewBinding(key.WithKeys("up"))
	t.Blur()

	t.CharLimit = charLimit

	return t
}

// Next, Prev, Add, Remove
type keymap = struct{ Reset, Quit key.Binding }

type model struct {
	width  int
	height int
	keymap keymap
	help   help.Model
	inputs []textarea.Model
	focus  int
}

func newModel() model {
	m := model{
		inputs: make([]textarea.Model, initialInputs),
		help:   help.New(),
		keymap: keymap{
			//Next: key.NewBinding(
			//	key.WithKeys("tab"),
			//	key.WithHelp("tab", "next"),
			//),
			//Prev: key.NewBinding(
			//	key.WithKeys("shift+tab"),
			//	key.WithHelp("shift+tab", "prev"),
			//),
			//Add: key.NewBinding(
			//	key.WithKeys("ctrl+n"),
			//	key.WithHelp("ctrl+n", "add an editor"),
			//),
			//Remove: key.NewBinding(
			//	key.WithKeys("ctrl+w"),
			//	key.WithHelp("ctrl+w", "remove an editor"),
			//),
			Reset: key.NewBinding(
				key.WithKeys("ctrl+r"),
				key.WithHelp("ctrl+r", "reset"),
			),
			Quit: key.NewBinding(
				key.WithKeys("esc", "ctrl+c"),
				key.WithHelp("esc", "quit"),
			),
		},
	}
	for i := 0; i < initialInputs; i++ {
		m.inputs[i] = newTextarea()
	}
	m.inputs[m.focus].Focus()
	m.updateKeybindings()
	return m
}

func (m model) Init() tea.Cmd {
	return textarea.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keymap.Quit):
			for i := range m.inputs {
				m.inputs[i].Blur()
			}
			return m, tea.Quit
		//case key.Matches(msg, m.keymap.Next):
		//	m.inputs[m.focus].Blur()
		//	m.focus++
		//	if m.focus > len(m.inputs)-1 {
		//		m.focus = 0
		//	}
		//	cmd := m.inputs[m.focus].Focus()
		//	cmds = append(cmds, cmd)
		//case key.Matches(msg, m.keymap.Prev):
		//	m.inputs[m.focus].Blur()
		//	m.focus--
		//	if m.focus < 0 {
		//		m.focus = len(m.inputs) - 1
		//	}
		//	cmd := m.inputs[m.focus].Focus()
		//	cmds = append(cmds, cmd)
		//case key.Matches(msg, m.keymap.Add):
		//	m.inputs = append(m.inputs, newTextarea())
		//case key.Matches(msg, m.keymap.Remove):
		//	m.inputs = m.inputs[:len(m.inputs)-1]
		//	if m.focus > len(m.inputs)-1 {
		//		m.focus = len(m.inputs) - 1
		//	}
		case key.Matches(msg, m.keymap.Reset):
			for i := 0; i < initialInputs; i++ {
				m.inputs[i] = newTextarea()
			}
		}
	case tea.WindowSizeMsg:
		m.height = msg.Height
		m.width = msg.Width
	}

	//m.updateKeybindings()
	m.sizeInputs()

	// Update all textareas
	for i := range m.inputs {
		update, cmd := m.inputs[i].Update(msg)
		m.inputs[i] = update
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m *model) sizeInputs() {
	for i := range m.inputs {
		m.inputs[i].SetWidth(m.width / len(m.inputs))
		m.inputs[i].SetHeight(m.height - helpHeight)
	}
}

func (m *model) updateKeybindings() {
	//m.keymap.Add.SetEnabled(len(m.inputs) < maxInputs)
	//m.keymap.Remove.SetEnabled(len(m.inputs) > minInputs)
}

func (m model) View() string {
	var keyBindings []key.Binding
	keyMap := reflect.ValueOf(m.keymap)
	for i := 0; i < keyMap.NumField(); i++ {
		keyBindings = append(keyBindings, keyMap.Field(i).Interface().(key.Binding))
	}

	helpView := m.help.ShortHelpView(keyBindings)

	var views []string
	for i := range m.inputs {
		views = append(views, m.inputs[i].View())
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, views...) + "\n\n" + helpView
}

func main() {
	if _, err := tea.NewProgram(newModel(), tea.WithAltScreen()).Run(); err != nil {
		fmt.Println("Error while running program:", err)
		os.Exit(1)
	}
}
