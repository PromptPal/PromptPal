package service

import (
	"context"
	"errors"

	"github.com/PromptPal/PromptPal/ent"
	"github.com/PromptPal/PromptPal/ent/schema"
	"github.com/otiai10/openaigo"
)

//go:generate mockery --name OpenAIService
type OpenAIService interface {
	Chat(
		ctx context.Context,
		project ent.Project,
		prompts []schema.PromptRow,
		variables map[string]string,
		userId string,
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
	project ent.Project,
	prompts []schema.PromptRow,
	variables map[string]string,
	userId string,
) (reply openaigo.ChatCompletionResponse, err error) {
	if project.OpenAIToken == "" {
		return reply, errors.New("token is empty")
	}

	client := openaigo.NewClient(project.OpenAIToken)
	if project.OpenAIBaseURL != "" {
		client.BaseURL = project.OpenAIBaseURL
	}

	req := openaigo.ChatRequest{
		Model:       project.OpenAIModel,
		Temperature: float32(project.OpenAITemperature),
		TopP:        float32(project.OpenAITopP),
	}
	if userId != "" {
		req.User = userId
	}
	if project.OpenAIMaxTokens > 0 {
		req.MaxTokens = project.OpenAIMaxTokens
	}

	for _, prompt := range prompts {
		content := replacePlaceholders(prompt.Prompt, variables)
		// TODO: update with variables
		pt := openaigo.Message{
			Role:    prompt.Role,
			Content: content,
		}
		req.Messages = append(req.Messages, pt)
	}

	return client.Chat(ctx, req)
}
