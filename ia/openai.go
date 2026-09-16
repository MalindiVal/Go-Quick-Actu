package ia

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

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

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

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

	log.Printf("OpenAI response: %+v", resp)

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("openai: no choices returned")
	}

	raw := strings.TrimSpace(resp.Choices[0].Message.Content)
	raw = strings.TrimPrefix(raw, "```json")
	raw = strings.TrimPrefix(raw, "```")
	raw = strings.TrimSuffix(raw, "```")
	raw = strings.TrimSpace(raw)

	var parsed KeywordsResponse
	if err := json.Unmarshal([]byte(raw), &parsed); err != nil {
		return nil, err
	}

	// Nettoyage sécurité
	for i := range parsed.Keywords {
		parsed.Keywords[i] = strings.TrimSpace(parsed.Keywords[i])
		log.Printf("Extracted keyword: %s", parsed.Keywords[i])
	}

	return parsed.Keywords, nil
}
