package ia

import (
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

var (
	openaiClient     openai.Client
	openaiClientOnce sync.Once
)

// ensureOpenAIClient lazily builds the OpenAI client on first use, for the
// same reason as ensureClaudeClient: it must read OPENAI_API_KEY only
// after main() has loaded it from a .env file.
func ensureOpenAIClient() {
	openaiClientOnce.Do(func() {
		openaiClient = openai.NewClient(
			option.WithAPIKey(os.Getenv("OPENAI_API_KEY")),
		)
	})
}

// askOpenAI sends prompt to OpenAI and returns its raw text response. It
// is only used as a fallback when the Claude call fails.
func askOpenAI(ctx context.Context, prompt string, maxTokens int64) (string, error) {
	if os.Getenv("OPENAI_API_KEY") == "" {
		return "", fmt.Errorf("openai: OPENAI_API_KEY not set")
	}

	ensureOpenAIClient()

	resp, err := openaiClient.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Model: openai.ChatModelGPT4oMini,
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.UserMessage(prompt),
		},
		Temperature: openai.Float(0.2),
		MaxTokens:   openai.Int(maxTokens),
	})
	if err != nil {
		return "", err
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("openai: no choices returned")
	}

	return resp.Choices[0].Message.Content, nil
}
