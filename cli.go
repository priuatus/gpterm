package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/priuatus/gpterm/internal/stdin"
	"golang.org/x/term"

	tea "github.com/charmbracelet/bubbletea"
	gogpt "github.com/sashabaranov/go-gpt3"
)

type completionResponse struct {
	*gogpt.CompletionStream
	gogpt.CompletionResponse
	streaming bool
}

type completion struct {
	client *gogpt.Client
	req    gogpt.CompletionRequest
}

func (c completion) Create(ctx context.Context) (resp completionResponse, err error) {
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
	if c.req.Stream {
		resp.streaming = true
		resp.CompletionStream, err = c.client.CreateCompletionStream(ctx, c.req)
		return
	}

	resp.CompletionResponse, err = c.client.CreateCompletion(ctx, c.req)

	if resp.Choices[0].FinishReason == "length" {
		fmt.Fprintf(os.Stderr, "%s: --max-tokens %d reached consider increasing the limit\n", os.Args[0], resp.Usage.CompletionTokens)
	}
	return resp, err
}

type CLI struct {
	APIKey    string   `short:"k" help:"OpenAI API Token." env:"OPENAI_API_KEY"`
	Model     string   `short:"m" default:"text-davinci-003" help:"The model which will generate the completion."`
	Temp      float32  `short:"t" default:"0.0" help:"Generation creativity. Higher is crazier."`
	MaxTokens int      `short:"n" default:"100" help:"Max number of tokens to generate."`
	Stream    bool     `short:"S" default:"true" help:"Whether to stream back partial progress."`
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

	if term.IsTerminal(int(os.Stdout.Fd())) && !t.Stream {
		model := initialModel(cmpltn)
		model.quiet = t.Quiet
		_, err := tea.NewProgram(model).Run()
		return err
	}

	// Non interactive use
	resp, err := cmpltn.Create(context.Background())
	if err != nil {
		return err
	}
	if resp.streaming {
		defer resp.Close()
		for {
			response, err := resp.Recv()
			if response.Choices != nil {
				fmt.Printf("%s", response.Choices[0].Text)
			}
			if errors.Is(err, io.EOF) {
				break
			}
			if err != nil {
				return fmt.Errorf("stream error: %v", err)
			}
		}
		return nil
	}

	if resp.Choices != nil {
		fmt.Printf("%s", strings.TrimLeft(resp.Choices[0].Text, "\n"))
	}

	return nil
}
