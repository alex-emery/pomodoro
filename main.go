package main

import (
	"fmt"
	"os"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/timer"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var timers = []time.Duration{45 * time.Minute, 5 * time.Minute, 15 * time.Minute}
var labels = []string{"focus", "short break", "long break"}

type TimerMode int

const (
	FocusMode TimerMode = iota
	ShortBreakMode
	LongBreakMode
)

type model struct {
	total     int
	completed int
	mode      TimerMode
	timer     timer.Model
	keymap    keymap
	help      help.Model
	quitting  bool
}

type keymap struct {
	start key.Binding
	stop  key.Binding
	reset key.Binding
	quit  key.Binding
}

func (m model) Init() tea.Cmd {
	return m.timer.Stop()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case timer.TickMsg:
		var cmd tea.Cmd
		m.timer, cmd = m.timer.Update(msg)
		return m, cmd

	case timer.StartStopMsg:
		var cmd tea.Cmd
		m.timer, cmd = m.timer.Update(msg)
		m.keymap.stop.SetEnabled(m.timer.Running())
		m.keymap.start.SetEnabled(!m.timer.Running())
		return m, cmd

	case timer.TimeoutMsg:
		// focus goes to short
		if m.mode == FocusMode {
			m.completed += 1
			m.total += 1
			if m.completed > 3 {
				// long break earnt
				m.mode = LongBreakMode
				m.completed = 0 // reset completed focus bouts
			} else {
				m.mode = ShortBreakMode
			}
		} else {
			m.mode = FocusMode
		}

		m.timer.Timeout = timers[m.mode]
		m.keymap.stop.SetEnabled(m.timer.Running())
		m.keymap.start.SetEnabled(!m.timer.Running())
		return m, m.timer.Stop()

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keymap.quit):
			m.quitting = true
			return m, tea.Quit
		case key.Matches(msg, m.keymap.reset):
			m.timer.Timeout = timers[m.mode]
		case key.Matches(msg, m.keymap.start, m.keymap.stop):
			return m, m.timer.Toggle()
		}
	}

	return m, nil
}

func (m model) helpView() string {
	return "\n" + m.help.ShortHelpView([]key.Binding{
		m.keymap.start,
		m.keymap.stop,
		m.keymap.reset,
		m.keymap.quit,
	})
}

func (m model) View() string {
	s := "Mode: "
	for i := range labels {
		if i == int(m.mode) {
			s += lipgloss.NewStyle().Underline(true).Background(lipgloss.Color("#0000FF")).Render(labels[i])
		} else {
			s += labels[i]
		}
		s += " "
	}
	s += "\n"

	s += "\n"
	if !m.quitting {
		s += "Time left "
		s += fmtDuration(m.timer.Timeout)
		s += "\n"
	}
	s += "Completed: "
	s += fmt.Sprintf("%d", m.total)
	s += "\n"
	s += m.helpView()
	return lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Render(s)
}

func fmtDuration(d time.Duration) string {
	m := d / time.Minute
	d -= m * time.Minute
	s := d / time.Second

	return fmt.Sprintf("%02d:%02d", m, s)
}

func main() {

	m := model{
		timer: timer.NewWithInterval(timers[0], time.Second),
		keymap: keymap{
			start: key.NewBinding(
				key.WithKeys("s"),
				key.WithHelp("s", "start"),
			),
			stop: key.NewBinding(
				key.WithKeys("s"),
				key.WithHelp("s", "stop"),
			),
			reset: key.NewBinding(
				key.WithKeys("r"),
				key.WithHelp("r", "reset"),
			),
			quit: key.NewBinding(
				key.WithKeys("q", "ctrl+c"),
				key.WithHelp("q", "quit"),
			),
		},
		help: help.New(),
	}
	m.keymap.stop.SetEnabled(false)

	if _, err := tea.NewProgram(m).Run(); err != nil {
		fmt.Println("Uh oh, we encountered an error:", err)
		os.Exit(1)
	}
}
