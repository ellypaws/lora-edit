package main

import (
	"errors"
	"github.com/charmbracelet/bubbles/cursor"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"strconv"
	"strings"
)

var (
	focusedStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	blurredStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	noStyle             = lipgloss.NewStyle()
	helpStyle           = blurredStyle.Copy()
	cursorModeHelpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
)

type settings struct {
	focusIndex int
	inputs     []textinput.Model
	cursorMode cursor.Mode
}

const (
	LoraToKeep = iota
	KeepWeight
	LoseWeight
)

func initialModel() settings {
	m := settings{
		focusIndex: 3,
		inputs:     make([]textinput.Model, 3),
	}

	var t textinput.Model
	for i := range m.inputs {
		t = textinput.New()
		t.Cursor.Style = cursorStyle
		t.CharLimit = 32

		isFloat := func(s string) error {
			if f, e := strconv.ParseFloat(s, 64); e != nil || f < 0 || f >= 10 {
				return errors.New("invalid number")
			}
			return nil
		}

		switch i {
		case LoraToKeep:
			t.Placeholder = "Lora to keep"
			t.PromptStyle = focusedStyle
			t.TextStyle = focusedStyle
		case KeepWeight:
			t.Placeholder = "0.75"
			t.SetValue(t.Placeholder)
			t.Validate = isFloat
			t.Prompt = "Keep: "
			t.CharLimit = 4
		case LoseWeight:
			t.Placeholder = "0.15"
			t.SetValue(t.Placeholder)
			t.Validate = isFloat
			t.Prompt = "Lose: "
			t.CharLimit = 4
		}

		m.inputs[i] = t
	}

	return m
}

func (m settings) Init() tea.Cmd {
	return textinput.Blink
}

func (m settings) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return m, tea.Quit

		// Change cursor mode
		case "ctrl+r":
			m.cursorMode++
			if m.cursorMode > cursor.CursorHide {
				m.cursorMode = cursor.CursorBlink
			}
			cmds := make([]tea.Cmd, len(m.inputs))
			for i := range m.inputs {
				cmds[i] = m.inputs[i].Cursor.SetMode(m.cursorMode)
			}
			return m, tea.Batch(cmds...)

		// Set focus to next input
		case "tab", "shift+tab", "enter", "up", "down":
			s := msg.String()

			if s == "enter" && m.focusIndex == len(m.inputs) {
				return m, nil
			}

			// Cycle indexes
			if s == "up" || s == "shift+tab" {
				m.focusIndex--
			} else {
				m.focusIndex++
			}

			m.focusIndex = max(0, min(m.focusIndex, len(m.inputs)))

			cmds := make([]tea.Cmd, len(m.inputs))
			for i := 0; i < len(m.inputs); i++ {
				if i == m.focusIndex {
					// Set focused state
					cmds[i] = m.inputs[i].Focus()
					m.inputs[i].PromptStyle = focusedStyle
					m.inputs[i].TextStyle = focusedStyle
					continue
				}
				// Remove focused state
				m.inputs[i].Blur()
				m.inputs[i].PromptStyle = noStyle
				m.inputs[i].TextStyle = noStyle
			}

			return m, tea.Batch(cmds...)

		case "left", "right":
			if m.focusIndex != KeepWeight && m.focusIndex != LoseWeight {
				break
			}

			if msg.String() == "left" {
				float, err := strconv.ParseFloat(m.inputs[m.focusIndex].Value(), 64)
				if err != nil {
					break
				}
				float -= 0.05
				m.inputs[m.focusIndex].SetValue(strconv.FormatFloat(float, 'f', 2, 64))
			}

			if msg.String() == "right" {
				float, err := strconv.ParseFloat(m.inputs[m.focusIndex].Value(), 64)
				if err != nil {
					break
				}
				float += 0.05
				m.inputs[m.focusIndex].SetValue(strconv.FormatFloat(float, 'f', 2, 64))
			}
			return m, nil
		}
	}

	// Handle character input and blinking
	cmd := m.updateInputs(msg)

	return m, cmd
}

func (m *settings) updateInputs(msg tea.Msg) tea.Cmd {
	cmds := make([]tea.Cmd, len(m.inputs))

	// Only text inputs with Focus() set will respond, so it's safe to simply
	// update all of them here without any further logic.
	for i := range m.inputs {
		m.inputs[i], cmds[i] = m.inputs[i].Update(msg)
	}

	return tea.Batch(cmds...)
}

func (m settings) View() string {
	var b strings.Builder

	for i := range m.inputs {
		b.WriteString(m.inputs[i].View())
		if i < len(m.inputs)-1 {
			b.WriteRune('\n')
		}
	}

	return b.String()
}
