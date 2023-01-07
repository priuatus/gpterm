package main

import (
	"context"
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/priuatus/gpterm/internal/stdin"
	gogpt "github.com/sashabaranov/go-gpt3"
)

type errMsg struct{ err error }

func (e errMsg) Error() string { return e.err.Error() }

type completionMsg struct {
	res gogpt.CompletionResponse
}

type model struct {
	gpt *gogpt.Client
	req gogpt.CompletionRequest
	res *gogpt.CompletionResponse

	spinner  spinner.Model
	quitting bool
	err      error
}

var quitKeys = key.NewBinding(
	key.WithKeys("q", "esc", "ctrl+c"),
	key.WithHelp("", "press q to quit"),
)

func initialModel(gpt *gogpt.Client, req gogpt.CompletionRequest) model {
	m := model{
		gpt: gpt,
		req: req,
		spinner: spinner.New(spinner.WithSpinner(spinner.Meter),
			spinner.WithStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("204")))),
	}
	return m
}

func (m model) createCompletion() tea.Msg {
	stdIn, err := stdin.Read()
	if err != nil && err != stdin.ErrEmpty {
		return errMsg{err}
	}
	if stdIn != "" {
		m.req.Prompt += stdIn
	}
	if m.req.Prompt == "" {
		return errMsg{fmt.Errorf("missing prompt")}
	}

	ctx := context.Background()
	resp, err := m.gpt.CreateCompletion(ctx, m.req)
	if err != nil {
		return errMsg{err}
	}

	return completionMsg{resp}
}

func (m model) Init() tea.Cmd {
	// Check for a terminal output
	if isTerm() {
		return tea.Batch(m.spinner.Tick, m.createCompletion)
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
		return fmt.Sprintf("\n%s: %v\n\n", os.Args[0], m.err)
	}

	if m.res != nil {
		return fmt.Sprintf("%s\n\n", m.res.Choices[0].Text)
	}

	str := fmt.Sprintf("\n %s thinking... %s\n\n", m.spinner.View(), quitKeys.Help().Desc)
	if m.quitting {
		return str + "\n"
	}
	return str
}
