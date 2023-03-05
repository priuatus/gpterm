package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	gogpt "github.com/sashabaranov/go-openai"
)

type errMsg struct{ err error }

func (e errMsg) Error() string { return e.err.Error() }

type completionMsg struct {
	res gogpt.CompletionResponse
}

type model struct {
	cmp completion
	res *gogpt.CompletionResponse

	spinner  spinner.Model
	quitting bool
	quiet    bool
	err      error
}

var quitKeys = key.NewBinding(
	key.WithKeys("q", "esc", "ctrl+c"),
	key.WithHelp("", "press q to quit"),
)

func initialModel(cmp completion) model {
	m := model{
		cmp: cmp,
		spinner: spinner.New(spinner.WithSpinner(spinner.Meter),
			spinner.WithStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("204")))),
	}
	return m
}

func (m model) createCompletion() tea.Msg {
	resp, err := m.cmp.Create(context.Background())
	if err != nil {
		return errMsg{err}
	}

	return completionMsg{resp.CompletionResponse}
}

func (m model) Init() tea.Cmd {
	// Check for a terminal output and run in interactive mode.
	if !m.quiet {
		if fileInfo, _ := os.Stdout.Stat(); (fileInfo.Mode() & os.ModeCharDevice) != 0 {
			return tea.Batch(m.spinner.Tick, m.createCompletion)
		}
	}
	return m.createCompletion
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.KeyMsg:
		if key.Matches(msg, quitKeys) {
			m.quitting = true
			return m, tea.Quit
		}
		return m, nil
	case errMsg:
		m.err = msg
		return m, tea.Quit
	case completionMsg:
		m.res = &msg.res
		return m, tea.Quit
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m model) View() string {
	if m.err != nil {
		return fmt.Sprintf("%s: %v\n\n", os.Args[0], m.err)
	}

	if m.res != nil {
		return fmt.Sprintf("\n%s\n", strings.TrimLeft(m.res.Choices[0].Text, "\n"))
	}

	var str string
	if !m.quiet {
		str = fmt.Sprintf("\n %s thinking... %s\n\n", m.spinner.View(), quitKeys.Help().Desc)
	}
	if m.quitting {
		str += "\n"
	}
	return str
}
