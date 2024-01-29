package service

import (
	"context"
	"errors"
	"strings"

	"github.com/PromptPal/PromptPal/ent"
	"github.com/PromptPal/PromptPal/ent/schema"
	openai "github.com/sashabaranov/go-openai"
)

//go:generate mockery --name OpenAIService
type OpenAIService interface {
	Chat(
		ctx context.Context,
		project ent.Project,
		prompts []schema.PromptRow,
		variables map[string]string,
		userId string,
	) (reply openai.ChatCompletionResponse, err error)
}

type aiService struct {
}

func NewOpenAIService() OpenAIService {
	return &aiService{}
}

// just for mock
func (o aiService) Chat(
	ctx context.Context,
	project ent.Project,
	prompts []schema.PromptRow,
	variables map[string]string,
	userId string,
) (reply openai.ChatCompletionResponse, err error) {
	if project.OpenAIToken == "" {
		return reply, errors.New("token is empty")
	}

	// if !strings.HasPrefix(project.OpenAIModel, "gpt-") {
	// 	client, err := genai.NewClient(ctx, option.WithAPIKey(project.OpenAIToken), option.WithEndpoint(project.OpenAIBaseURL))
	// 	if err != nil {
	// 		log.Fatal(err)
	// 		return reply, err
	// 	}
	// 	defer client.Close()
	// }

	cfg := openai.DefaultConfig(project.OpenAIToken)

	if project.OpenAIBaseURL != "" {
		cfg.BaseURL = project.OpenAIBaseURL
	}

	client := openai.NewClientWithConfig(cfg)

	req := openai.ChatCompletionRequest{
		Model:       project.OpenAIModel,
		Temperature: float32(project.OpenAITemperature),
		TopP:        float32(project.OpenAITopP),
	}
	if strings.Contains(req.Model, "-1106") {
		req.ResponseFormat = &openai.ChatCompletionResponseFormat{
			Type: openai.ChatCompletionResponseFormatTypeJSONObject,
		}
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
		pt := openai.ChatCompletionMessage{
			Role:    prompt.Role,
			Content: content,
		}
		req.Messages = append(req.Messages, pt)
	}

	return client.CreateChatCompletion(ctx, req)
}
