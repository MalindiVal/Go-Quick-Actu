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

	"goquickactu/models"
	"goquickactu/topics"
)

var client = openai.NewClient(
	option.WithAPIKey(os.Getenv("OPENAI_API_KEY")),
)

// chatJSON sends prompt to the model and unmarshals its response into out,
// which must be a pointer. It tolerates the model wrapping its JSON in a
// ```json ... ``` markdown fence, which happens intermittently even when
// the prompt asks for raw JSON.
func chatJSON(ctx context.Context, prompt string, out interface{}) error {
	resp, err := client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Model: openai.ChatModelGPT4oMini,
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.UserMessage(prompt),
		},
		Temperature: openai.Float(0.2),
	})
	if err != nil {
		return err
	}

	log.Printf("OpenAI response: %+v", resp)

	if len(resp.Choices) == 0 {
		return fmt.Errorf("openai: no choices returned")
	}

	raw := strings.TrimSpace(resp.Choices[0].Message.Content)
	raw = strings.TrimPrefix(raw, "```json")
	raw = strings.TrimPrefix(raw, "```")
	raw = strings.TrimSuffix(raw, "```")
	raw = strings.TrimSpace(raw)

	return json.Unmarshal([]byte(raw), out)
}

type analysisResponse struct {
	Keywords []string `json:"keywords"`
	Topic    string   `json:"topic"`
}

// AnalyzeContent extracts 5 keywords from content and classifies it into
// one of the known news topics (topics.Topics), in a single OpenAI call.
// If the model returns a topic outside that fixed list, topic comes back
// empty rather than erroring, since the caller treats "uncategorized" as
// a valid, tolerated outcome.
func AnalyzeContent(content string) (keywords []string, topic string, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	prompt := `
Tu es un assistant qui analyse un article de presse.
Extrait exactement 5 mots-clés pertinents et classe l'article dans EXACTEMENT
un des sujets suivants : ` + strings.Join(topics.Topics, ", ") + `

Réponds UNIQUEMENT en JSON valide au format :
{
  "keywords": ["mot1", "mot2", "mot3", "mot4", "mot5"],
  "topic": "un des sujets ci-dessus"
}

Article :
` + content

	var parsed analysisResponse
	if err := chatJSON(ctx, prompt, &parsed); err != nil {
		return nil, "", err
	}

	for i := range parsed.Keywords {
		parsed.Keywords[i] = strings.TrimSpace(parsed.Keywords[i])
	}

	parsed.Topic = strings.TrimSpace(parsed.Topic)
	if !topics.IsValid(parsed.Topic) {
		log.Printf("model returned unknown topic %q, leaving article uncategorized", parsed.Topic)
		parsed.Topic = ""
	}

	return parsed.Keywords, parsed.Topic, nil
}

// SourceViewpoint describes how one source frames a news story, as part of
// a Synthesis.
type SourceViewpoint struct {
	Source           string `json:"source"`
	PoliticalLeaning string `json:"political_leaning,omitempty"`
	Viewpoint        string `json:"viewpoint"`
}

// Synthesis is a cross-source summary of a news story: the facts the
// sources agree on, plus how each of them individually frames it.
type Synthesis struct {
	Summary      string            `json:"summary"`
	Perspectives []SourceViewpoint `json:"perspectives"`
}

type synthesisResponse struct {
	Summary      string `json:"summary"`
	Perspectives []struct {
		Source    string `json:"source"`
		Viewpoint string `json:"viewpoint"`
	} `json:"perspectives"`
}

// SynthesizeArticles takes several articles covering the same news story
// (typically gathered by a shared keyword) and asks the model for a
// neutral factual summary plus an explicit, per-source breakdown of how
// each outlet frames the story - the angle taken, what is emphasized or
// left out - rather than just restating each source's known leaning.
func SynthesizeArticles(articles []models.Article) (Synthesis, error) {
	if len(articles) == 0 {
		return Synthesis{}, fmt.Errorf("ia: no articles to synthesize")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()

	leaningBySource := make(map[string]string, len(articles))
	var b strings.Builder
	b.WriteString(`
Tu es un assistant de synthèse d'actualité. Voici plusieurs articles de presse
traitant du même sujet, chacun précédé de sa source et de son orientation
politique indicative.

Ta tâche :
1. Rédige une synthèse neutre et factuelle des événements communs à ces
   articles (5 à 8 phrases), sans prendre parti.
2. Pour chaque source distincte, explique en 1 à 2 phrases comment ELLE
   traite l'information : angle choisi, éléments mis en avant ou passés
   sous silence, vocabulaire employé. Appuie-toi sur le contenu réel des
   articles, ne te contente pas de répéter l'orientation politique indiquée.

Réponds UNIQUEMENT en JSON valide au format :
{
  "summary": "synthèse neutre des faits",
  "perspectives": [
    {"source": "Nom de la source", "viewpoint": "comment cette source traite l'information"}
  ]
}

Articles :
`)

	for i, article := range articles {
		leaningBySource[article.Source] = article.PoliticalLeaning
		fmt.Fprintf(&b, "%d. [%s - %s] Titre : %s\n   Description : %s\n\n",
			i+1, article.Source, article.PoliticalLeaning, article.Title, article.Description)
	}

	var parsed synthesisResponse
	if err := chatJSON(ctx, b.String(), &parsed); err != nil {
		return Synthesis{}, err
	}

	result := Synthesis{
		Summary:      strings.TrimSpace(parsed.Summary),
		Perspectives: make([]SourceViewpoint, 0, len(parsed.Perspectives)),
	}
	for _, p := range parsed.Perspectives {
		result.Perspectives = append(result.Perspectives, SourceViewpoint{
			Source:           p.Source,
			PoliticalLeaning: leaningBySource[p.Source],
			Viewpoint:        strings.TrimSpace(p.Viewpoint),
		})
	}

	return result, nil
}
