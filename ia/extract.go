package ia

import (
	"context"
	"encoding/json"
	"os"
	"strings"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

var client = openai.NewClient(
	option.WithAPIKey(os.Getenv("OPENAI_API_KEY")),
)

type KeywordsResponse struct {
	Keywords []string `json:"keywords"`
}

func ExtractKeywords(content string) ([]string, error) {

	ctx := context.Background()

	prompt := `
Tu es un assistant qui extrait exactement 5 mots-clés pertinents.
Réponds UNIQUEMENT en JSON valide au format :
{
  "keywords": ["mot1", "mot2", "mot3", "mot4", "mot5"]
}

Article :
` + content

	resp, err := client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Model: openai.ChatModelGPT4oMini,
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.UserMessage(prompt),
		},
		Temperature: openai.Float(0.2),
	})

	if err != nil {
		return nil, err
	}

	raw := resp.Choices[0].Message.Content

	var parsed KeywordsResponse
	err = json.Unmarshal([]byte(raw), &parsed)
	if err != nil {
		return nil, err
	}

	// Nettoyage sécurité
	for i := range parsed.Keywords {
		parsed.Keywords[i] = strings.TrimSpace(parsed.Keywords[i])
	}

	return parsed.Keywords, nil
}
