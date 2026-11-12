package ia

import (
	"context"
	"fmt"
	"sync"

	"github.com/anthropics/anthropic-sdk-go"
)

// claudeModel is Claude Opus 5.5. Its thinking is always on; depth is
// controlled with the effort level passed to askClaude.
const claudeModel = "claude-opus-5-5"

var (
	claudeClient     anthropic.Client
	claudeClientOnce sync.Once
)

// ensureClaudeClient lazily builds the Anthropic client on first use, so
// it reads ANTHROPIC_API_KEY only after main() has had a chance to load it
// from a .env file - constructing it eagerly as a package-level var would
// capture the key before that load happens.
func ensureClaudeClient() {
	claudeClientOnce.Do(func() {
		claudeClient = anthropic.NewClient()
	})
}

// askClaude sends prompt to Claude and returns its raw text response.
//
// It opts into server-side refusal fallbacks ("default"): if the model
// declines for policy reasons, the API re-serves the request on a
// substitute model within the same call.
func askClaude(ctx context.Context, prompt string, maxTokens int64, effort anthropic.BetaOutputConfigEffort) (string, error) {
	ensureClaudeClient()

	resp, err := claudeClient.Beta.Messages.New(ctx, anthropic.BetaMessageNewParams{
		Model:        claudeModel,
		MaxTokens:    maxTokens,
		Betas:        []anthropic.AnthropicBeta{anthropic.AnthropicBetaServerSideFallback2026_07_01},
		Fallbacks:    anthropic.BetaFallbacksParamOfDefault(),
		OutputConfig: anthropic.BetaOutputConfigParam{Effort: effort},
		Messages: []anthropic.BetaMessageParam{
			anthropic.NewBetaUserMessage(anthropic.NewBetaTextBlock(prompt)),
		},
	})
	if err != nil {
		return "", err
	}

	switch resp.StopReason {
	case anthropic.BetaStopReasonRefusal:
		return "", fmt.Errorf("claude: request refused")
	case anthropic.BetaStopReasonMaxTokens:
		return "", fmt.Errorf("claude: response truncated (max_tokens=%d)", maxTokens)
	}

	var raw string
	for _, block := range resp.Content {
		if text, ok := block.AsAny().(anthropic.BetaTextBlock); ok {
			raw += text.Text
		}
	}

	return raw, nil
}
