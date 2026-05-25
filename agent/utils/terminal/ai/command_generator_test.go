package ai

import (
	"context"
	"errors"
	"testing"
)

type stubChatClient struct {
	resp *ChatCompletionResponse
	err  error
}

func (c *stubChatClient) ChatCompletion(context.Context, ChatCompletionRequest) (*ChatCompletionResponse, error) {
	return c.resp, c.err
}

func TestGenerateReturnsSafeCommand(t *testing.T) {
	generator, err := NewCommandGenerator(&stubChatClient{
		resp: &ChatCompletionResponse{
			Model:   "test-model",
			Content: "```bash\nls -la /tmp\n```",
		},
	})
	if err != nil {
		t.Fatalf("NewCommandGenerator() error = %v", err)
	}

	resp, err := generator.Generate(context.Background(), CommandGenerateRequest{Input: "list files"})
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	if resp.Command != "ls -la /tmp" {
		t.Fatalf("Generate() command = %q, want %q", resp.Command, "ls -la /tmp")
	}
}

func TestGenerateAllowsNormalShellSyntax(t *testing.T) {
	generator, err := NewCommandGenerator(&stubChatClient{
		resp: &ChatCompletionResponse{
			Model:   "test-model",
			Content: "printf 'a' | grep a && echo done",
		},
	})
	if err != nil {
		t.Fatalf("NewCommandGenerator() error = %v", err)
	}

	resp, err := generator.Generate(context.Background(), CommandGenerateRequest{Input: "format text"})
	if err != nil {
		t.Fatalf("Generate() error = %v, want normal shell syntax to be accepted", err)
	}
	if resp == nil || resp.Command != "printf 'a' | grep a && echo done" {
		t.Fatalf("Generate() command = %v, want normal shell syntax to pass unchanged", resp)
	}
}

func TestGenerateTruncatesMultiLineOutputToFirstCommand(t *testing.T) {
	generator, err := NewCommandGenerator(&stubChatClient{
		resp: &ChatCompletionResponse{
			Model:   "test-model",
			Content: "echo safe\nrm -rf /",
		},
	})
	if err != nil {
		t.Fatalf("NewCommandGenerator() error = %v", err)
	}

	resp, err := generator.Generate(context.Background(), CommandGenerateRequest{Input: "run a command"})
	if err != nil {
		t.Fatalf("Generate() error = %v, want multi-line output to be sanitized", err)
	}
	if resp == nil || resp.Command != "echo safe" {
		t.Fatalf("Generate() command = %v, want first line only", resp)
	}
}

func TestGenerateStripsANSIColorAndKeepsCommandText(t *testing.T) {
	generator, err := NewCommandGenerator(&stubChatClient{
		resp: &ChatCompletionResponse{
			Model:   "test-model",
			Content: "\x1b[31mls -la /tmp\x1b[0m",
		},
	})
	if err != nil {
		t.Fatalf("NewCommandGenerator() error = %v", err)
	}

	resp, err := generator.Generate(context.Background(), CommandGenerateRequest{Input: "list files"})
	if err != nil {
		t.Fatalf("Generate() error = %v, want ANSI decoration to be stripped", err)
	}
	if resp == nil || resp.Command != "ls -la /tmp" {
		t.Fatalf("Generate() command = %v, want ANSI control sequences removed", resp)
	}
}

func TestGeneratePropagatesClientError(t *testing.T) {
	wantErr := errors.New("boom")
	generator, err := NewCommandGenerator(&stubChatClient{err: wantErr})
	if err != nil {
		t.Fatalf("NewCommandGenerator() error = %v", err)
	}

	_, err = generator.Generate(context.Background(), CommandGenerateRequest{Input: "list files"})
	if !errors.Is(err, wantErr) {
		t.Fatalf("Generate() error = %v, want %v", err, wantErr)
	}
}
