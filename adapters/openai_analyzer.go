package adapters

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/nick5616/holodeck-art-api/models"
	"github.com/sashabaranov/go-openai"
)

type OpenAIAnalyzer struct {
	client *openai.Client
}

func NewOpenAIAnalyzer(apiKey string) *OpenAIAnalyzer {
	return &OpenAIAnalyzer{
		client: openai.NewClient(apiKey),
	}
}

func (o *OpenAIAnalyzer) AnalyzeImage(ctx context.Context, data []byte) (models.AIAnalysis, error) {
	// Encode image to base64
	base64Image := base64.StdEncoding.EncodeToString(data)
	
	resp, err := o.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: openai.GPT4o,
		Messages: []openai.ChatCompletionMessage{
			{
				Role: openai.ChatMessageRoleUser,
				MultiContent: []openai.ChatMessagePart{
					{
						Type: openai.ChatMessagePartTypeText,
						Text: `Analyze this digital artwork. Provide:
1) A creative 2-4 word title
2) Three descriptive tags (single words, comma-separated)

Return ONLY valid JSON in this exact format:
{"title": "...", "tags": "tag1,tag2,tag3"}`,
					},
					{
						Type: openai.ChatMessagePartTypeImageURL,
						ImageURL: &openai.ChatMessageImageURL{
							URL: fmt.Sprintf("data:image/png;base64,%s", base64Image),
						},
					},
				},
			},
		},
		MaxTokens: 100,
	})
	
	if err != nil {
		return models.AIAnalysis{}, fmt.Errorf("OpenAI API error: %w", err)
	}
	
	if len(resp.Choices) == 0 {
		return models.AIAnalysis{}, fmt.Errorf("no response from OpenAI")
	}
	
	// Parse JSON response - use intermediate struct
	var rawResponse struct {
		Title string `json:"title"`
		Tags  string `json:"tags"`
	}

	content := resp.Choices[0].Message.Content

	// Strip markdown code blocks if present
	content = strings.TrimSpace(content)
	content = strings.TrimPrefix(content, "```json")
	content = strings.TrimPrefix(content, "```")
	content = strings.TrimSuffix(content, "```")
	content = strings.TrimSpace(content)

	if err := json.Unmarshal([]byte(content), &rawResponse); err != nil {
		return models.AIAnalysis{}, fmt.Errorf("failed to parse AI response: %w", err)
	}
	
	// Convert comma-separated string to array
	tags := strings.Split(rawResponse.Tags, ",")
	for i := range tags {
		tags[i] = strings.TrimSpace(tags[i])
	}
	
	return models.AIAnalysis{
		Title: rawResponse.Title,
		Tags:  tags,  // Now it's an array
	}, nil
	
}