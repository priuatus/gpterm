package main

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	gogpt "github.com/sashabaranov/go-gpt3"
)

func isTerm() bool {
	if fileInfo, _ := os.Stdout.Stat(); (fileInfo.Mode() & os.ModeCharDevice) != 0 {
		return true
	}
	return false
}

type CLI struct {
	APIKey    string   `short:"k" help:"OpenAI API Token." env:"OPENAI_API_KEY"`
	Model     string   `short:"m" default:"text-davinci-003" help:"The model which will generate the completion."`
	Temp      float32  `short:"t" default:"0.0" help:"Generation creativity. Higher is crazier."`
	MaxTokens int      `short:"n" default:"100" help:"Max number of tokens to generate."`
	Prompt    []string `arg:"" optional:"" help:"text prompt"`
}

func (t CLI) Run() error {
	gpt := gogpt.NewClient(t.APIKey)

	req := gogpt.CompletionRequest{
		Model:       t.Model,
		Temperature: t.Temp,
		MaxTokens:   t.MaxTokens,
		TopP:        1.0,
		Prompt:      strings.Trim(fmt.Sprint(t.Prompt), "[]"),
	}

	model := initialModel(gpt, req)
	teaProg := tea.NewProgram(model)
	_, err := teaProg.Run()
	return err
}
