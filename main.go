package main

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"text/template"

	"github.com/go-semantic-release/semantic-release/v2/pkg/generator"
	"github.com/go-semantic-release/semantic-release/v2/pkg/plugin"
	"github.com/sashabaranov/go-openai"
)

var version = "dev"

type LLMChangelogGenerator struct {
	openaiAPIKey string
}

func (g *LLMChangelogGenerator) Init(m map[string]string) error {
	if m["openai_api_key"] != "" {
		g.openaiAPIKey = m["openai_api_key"]
	}
	if g.openaiAPIKey == "" {
		g.openaiAPIKey = os.Getenv("OPENAI_API_KEY")
	}
	return nil
}

func (g *LLMChangelogGenerator) Name() string {
	return "llm"
}

func (g *LLMChangelogGenerator) Version() string {
	return version
}

var promptTemplate = template.Must(template.New("prompt").Funcs(template.FuncMap{"join": strings.Join}).Parse(`
Generate a meaningful changelog summary for the newest release ({{.NewVersion}}) based on the following list of commits, which contains changes made since the last release.
Please provide a concise summary of the updates, features, and bug fixes included in this release.
Keep the summary short and only respond with the changelog summary in the markdown format.
Additionally highlight important changes in the summary.
Do not use markdown backticks to format the summary and use "Release {{.NewVersion}}" as heading.
The tone of voice of the summary must be technical and emotionless.
Commits:
{{range .Commits}}
{{join .Raw ", "}}
{{- end}}
`))

func (g *LLMChangelogGenerator) sendRequestToOpenAI(prompt string) (string, error) {
	client := openai.NewClient(g.openaiAPIKey)
	resp, err := client.CreateChatCompletion(context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT4TurboPreview,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: prompt,
				},
			},
			MaxTokens: 256,
			User:      "changelog-generator-llm",
		})
	if err != nil {
		return "", err
	}
	return resp.Choices[0].Message.Content, nil
}

func (g *LLMChangelogGenerator) Generate(changelogConfig *generator.ChangelogGeneratorConfig) string {
	buf := &bytes.Buffer{}
	err := promptTemplate.Execute(buf, changelogConfig)
	if err != nil {
		log.Fatal(fmt.Errorf("could not template prompt: %w", err))
	}
	changelog, err := g.sendRequestToOpenAI(buf.String())
	if err != nil {
		log.Fatal(fmt.Errorf("could not generate changelog: %w", err))
	}
	return changelog
}

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ChangelogGenerator: func() generator.ChangelogGenerator {
			return &LLMChangelogGenerator{}
		},
	})
}
