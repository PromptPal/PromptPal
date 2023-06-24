package service

import (
	"context"

	"github.com/PromptPal/PromptPal/ent"
	"github.com/PromptPal/PromptPal/ent/schema"
	"github.com/otiai10/openaigo"
)

type OpenAIService interface {
	Chat(
		ctx context.Context,
		project *ent.Project,
		prompts []schema.PromptRow,
		promptsVariables []schema.PromptVariable,
	) (reply openaigo.ChatCompletionResponse, err error)
}

type openAIService struct {
}

func NewOpenAIService() OpenAIService {
	return &openAIService{}
}

// just for mock
func (o openAIService) Chat(
	ctx context.Context,
	project *ent.Project,
	prompts []schema.PromptRow,
	promptsVariables []schema.PromptVariable,
) (reply openaigo.ChatCompletionResponse, err error) {
	client := openaigo.NewClient(project.OpenAIToken)
	client.BaseURL = project.OpenAIBaseURL

	req := openaigo.ChatRequest{
		Model:       project.OpenAIModel,
		Temperature: float32(project.OpenAITemperature),
		TopP:        float32(project.OpenAITopP),
	}
	if project.OpenAIMaxTokens > 0 {
		req.MaxTokens = project.OpenAIMaxTokens
	}

	for _, prompt := range prompts {
		// TODO: update with variables
		pt := openaigo.Message{
			Role:    prompt.Role,
			Content: prompt.Prompt,
		}
		req.Messages = append(req.Messages, pt)
	}

	return client.Chat(ctx, req)
}
