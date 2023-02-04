package main

import (
	"context"
	"fmt"
	"io"
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

	resp, err = c.client.CreateCompletion(context.Background(), c.req)

	return resp, err
}

type CLI struct {
	APIKey    string   `short:"k" help:"OpenAI API Token." env:"OPENAI_API_KEY"`
	Model     string   `short:"m" default:"text-davinci-003" help:"The model which will generate the completion."`
	Temp      float32  `short:"t" default:"0.0" help:"Generation creativity. Higher is crazier."`
	MaxTokens int      `short:"n" default:"100" help:"Max number of tokens to generate."`
	Stream    bool     `short:"S" default:"true" help:"Whether to stream back partial progress"`
	Quiet     bool     `short:"q" default:"false" help:"Print only the model response."`
	Stop      []string `short:"s" help:"Up to 4 sequences where the model will stop generating further. The returned text will not contain the stop sequence."`
	Prompt    []string `arg:"" optional:"" help:"text prompt"`
}

func (t CLI) Run() error {
	cmpltn := completion{
		client: gogpt.NewClient(t.APIKey),
		req: gogpt.CompletionRequest{
			Model:       t.Model,
			Prompt:      strings.Join(t.Prompt, " "),
			MaxTokens:   t.MaxTokens,
			Temperature: t.Temp,
			TopP:        1.0,
			Echo:        true,
			Stop:        t.Stop,
			Stream:      t.Stream,
		},
	}
	if t.Quiet {
		cmpltn.req.Echo = false
	}

	if !t.Stream && term.IsTerminal(int(os.Stdout.Fd())) {
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

	if resp.IsStream() {
		var err error
		defer resp.Close()
		fmt.Println("Streaming response start:")
		_, err = io.Copy(os.Stdout, resp)
		if err != nil {
			return err
		}
		fmt.Println("\nStreaming response end.")
		return nil
	}

	if resp.Choices != nil {
		fmt.Printf("%s", strings.TrimLeft(resp.Choices[0].Text, "\n"))
	}
	if resp.Choices[0].FinishReason == "length" {
		fmt.Fprintf(os.Stderr, "%s: --max-tokens %d reached consider increasing the limit\n", os.Args[0], resp.Usage.CompletionTokens)
	}

	return nil
}
