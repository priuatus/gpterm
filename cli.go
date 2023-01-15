package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/priuatus/gpterm/internal/stdin"
	"golang.org/x/term"

	tea "github.com/charmbracelet/bubbletea"
	gogpt "github.com/sashabaranov/go-gpt3"
)

type completion struct {
	client *gogpt.Client
	req    gogpt.CompletionRequest
}

func (c completion) Create() (resp gogpt.CompletionResponse, err error) {
	var stdIn string
	stdIn, err = stdin.Read()
	if err != nil && err != stdin.ErrEmpty {
		return resp, err
	}
	if err == stdin.ErrEmpty {
		stdIn = ""
	}
	c.req.Prompt += stdIn
	if c.req.Prompt == "" {
		return resp, fmt.Errorf("missing prompt")
	}

	return c.client.CreateCompletion(context.Background(), c.req)
}

type CLI struct {
	APIKey    string   `short:"k" help:"OpenAI API Token." env:"OPENAI_API_KEY"`
	Model     string   `short:"m" default:"text-davinci-003" help:"The model which will generate the completion."`
	Temp      float32  `short:"t" default:"0.0" help:"Generation creativity. Higher is crazier."`
	MaxTokens int      `short:"n" default:"100" help:"Max number of tokens to generate."`
	Quiet     bool     `short:"q" default:"false" help:"Print only the model response."`
	Prompt    []string `arg:"" optional:"" help:"text prompt"`
}

func (t CLI) Run() error {
	cmpltn := completion{
		client: gogpt.NewClient(t.APIKey),
		req: gogpt.CompletionRequest{
			Model:       t.Model,
			Temperature: t.Temp,
			MaxTokens:   t.MaxTokens,
			TopP:        1.0,
			Prompt:      strings.Trim(fmt.Sprint(t.Prompt), "[]"),
		},
	}

	if term.IsTerminal(int(os.Stdout.Fd())) {
		model := initialModel(cmpltn)
		model.quiet = t.Quiet
		_, err := tea.NewProgram(model).Run()
		return err
	}

	// Non interactive use
	resp, err := cmpltn.Create()
	if err != nil {
		return err
	}

	if resp.Choices != nil {
		fmt.Printf("%s", strings.TrimLeft(resp.Choices[0].Text, "\n"))
	}
	return nil
}
