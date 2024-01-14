package main

import (
	"fmt"
	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"log"
	"os"
	"reflect"
	"regexp"
	"strconv"
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
type keymap = struct{ Copy, Reset, Quit, Left, Right key.Binding }

type model struct {
	width    int
	height   int
	keymap   keymap
	settings settings
	help     help.Model
	inputs   []textarea.Model
	focus    int
}

func newModel() model {
	m := model{
		inputs:   make([]textarea.Model, initialInputs),
		help:     help.New(),
		settings: initialModel(),
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
			Copy: key.NewBinding(
				key.WithKeys("ctrl+c"),
				key.WithHelp("ctrl+c", "copy"),
			),
			Reset: key.NewBinding(
				key.WithKeys("ctrl+r"),
				key.WithHelp("ctrl+r", "reset"),
			),
			Quit: key.NewBinding(
				key.WithKeys("esc", "ctrl+c"),
				key.WithHelp("esc", "quit"),
			),
			Left: key.NewBinding(
				key.WithKeys("left"),
				key.WithHelp("←", "decrease weight"),
			),
			Right: key.NewBinding(
				key.WithKeys("right"),
				key.WithHelp("→", "add weight"),
			),
		},
	}
	for i := 0; i < initialInputs; i++ {
		m.inputs[i] = newTextarea()
	}
	m.inputs[m.focus].Focus()
	//m.updateKeybindings()
	return m
}

func (m model) Init() tea.Cmd {
	return tea.Batch(textarea.Blink, m.settings.Init())
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keymap.Copy):
			// copy m.inputs[1].Value() to clipboard
			err := clipboard.WriteAll(m.inputs[1].Value())
			if err != nil {
				log.Panicf("Error copying to clipboard: %v", err)
			}
			return m, nil
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
			m.inputs[m.focus].Focus()
		case msg.String() == "up":
			if m.inputs[m.focus].Line() == 0 {
				m.inputs[m.focus].Blur()
				if m.settings.focusIndex > 0 {
					cmds = m.updateSettings(msg, cmds)
				}
			}
		case msg.String() == "down" || msg.String() == "enter":
			cmds = m.updateSettings(msg, cmds)
			if m.settings.focusIndex == len(m.settings.inputs) && !m.inputs[m.focus].Focused() {
				m.inputs[m.focus].Focus()
				return m, tea.Batch(cmds...)
			}
		default:
			cmds = m.updateSettings(msg, cmds)
		}
	case tea.WindowSizeMsg:
		m.height = msg.Height - lipgloss.Height(m.settings.View())
		m.width = msg.Width
	}

	//m.updateKeybindings()
	m.sizeInputs()

	// Update all textareas
	if m.settings.focusIndex == len(m.settings.inputs) {
		//m.inputs[m.focus].Focus()
		for i := range m.inputs {
			update, cmd := m.inputs[i].Update(msg)
			m.inputs[i] = update
			cmds = append(cmds, cmd)
		}
	}

	m.parse()

	return m, tea.Batch(cmds...)
}

func (m *model) updateSettings(msg tea.Msg, cmds []tea.Cmd) []tea.Cmd {
	newSettings, cmd := m.settings.Update(msg)
	m.settings = newSettings.(settings)
	cmds = append(cmds, cmd)
	return cmds
}

var regEx = regexp.MustCompile(`<lora:([\w-]+):([\d.]+)>`)

func (m *model) parse() {
	if m.inputs[0].Value() == "" {
		m.inputs[1].SetValue("")
		return
	}
	var (
		keep       = m.settings.inputs[0].Value()
		keepWeight = m.settings.inputs[1].Value()
		weight     = m.settings.inputs[2].Value()
	)
	weightFloat, err := strconv.ParseFloat(weight, 64)
	if err != nil {
		weightFloat = 0.15
	}
	result := regEx.ReplaceAllStringFunc(m.inputs[0].Value(), func(s string) string {
		matches := regEx.FindStringSubmatch(s)
		float, err := strconv.ParseFloat(matches[2], 64)
		if err != nil {
			return s
		}
		return fmt.Sprintf("<lora:%s:%v>", matches[1], untypedIF(matches[1] == keep, keepWeight, min(float, weightFloat)))
	})
	m.inputs[1].SetValue(result)
}

// IF returns trueVal if condition is true, otherwise falseVal.
func IF[T any](condition bool, trueVal, falseVal T) T {
	if condition {
		return trueVal
	}
	return falseVal
}

func untypedIF(condition bool, trueVal, falseVal any) any {
	if condition {
		return trueVal
	}
	return falseVal
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

	return lipgloss.JoinVertical(lipgloss.Center, m.settings.View(),
		//fmt.Sprintf("focusIndex: %d, m.focus: %d, m.inputs[m.focus].Line(): %d", m.settings.focusIndex, m.focus, m.inputs[m.focus].Line()),
		//fmt.Sprintf("m.settings.focusIndex < len(m.settings.inputs) - 1 = %v < %v = %t\n", m.settings.focusIndex, len(m.settings.inputs)-1, m.settings.focusIndex < len(m.settings.inputs)-1),
		lipgloss.JoinHorizontal(lipgloss.Top, views...)+"\n\n"+helpView)
}

func main() {
	if _, err := tea.NewProgram(newModel(), tea.WithAltScreen()).Run(); err != nil {
		fmt.Println("Error while running program:", err)
		os.Exit(1)
	}
}
