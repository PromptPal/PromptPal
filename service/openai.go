package service

import (
	"context"
	"errors"
	"log"
	"net/url"
	"strings"

	"github.com/PromptPal/PromptPal/ent"
	"github.com/PromptPal/PromptPal/ent/schema"
	"github.com/google/generative-ai-go/genai"
	openai "github.com/sashabaranov/go-openai"
	"github.com/sirupsen/logrus"
	"google.golang.org/api/option"
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

	// gemini support
	if strings.HasPrefix(project.OpenAIModel, "gemini") {
		client, err := genai.NewClient(
			ctx,
			option.WithAPIKey(project.GeminiToken),
			option.WithEndpoint(project.GeminiBaseURL),
		)
		if err != nil {
			log.Fatal(err)
			return reply, err
		}
		defer client.Close()
		genModel := client.GenerativeModel(project.OpenAIModel)
		genModel.SetTemperature(float32(project.OpenAITemperature))
		genModel.SetTopP(float32(project.OpenAITopP))
		genModel.SetTopK(20)
		if project.OpenAIMaxTokens > 0 {
			genModel.SetMaxOutputTokens(int32(project.OpenAIMaxTokens))
		}
		pts := []genai.Part{}

		for _, prompt := range prompts {
			content := replacePlaceholders(prompt.Prompt, variables)
			txt := genai.Text(content)
			pts = append(pts, txt)
		}
		resp, err := genModel.GenerateContent(ctx, pts...)
		if err != nil {
			return reply, err
		}

		result := []openai.ChatCompletionChoice{}

		completionTokenCount := int32(0)
		for _, cand := range resp.Candidates {
			if cand.Content == nil {
				continue
			}
			completionTokenCount += cand.TokenCount
			for _, part := range cand.Content.Parts {
				content, ok := part.(genai.Text)
				if !ok {
					logrus.Warnln("not a text part in gemini api")
					continue
				}
				// genai.Text(part.ContentType)
				result = append(result, openai.ChatCompletionChoice{
					Index: int(cand.Index),
					Message: openai.ChatCompletionMessage{
						Role:    cand.Content.Role,
						Content: string(content),
					},
					FinishReason: openai.FinishReasonStop,
				})
			}
		}
		return openai.ChatCompletionResponse{
			Choices: result,
			Usage: openai.Usage{
				CompletionTokens: int(completionTokenCount),
				PromptTokens:     -1,
				TotalTokens:      int(completionTokenCount),
			},
		}, nil
	}

	// openai support
	cfg := openai.DefaultConfig(project.OpenAIToken)

	// add `/v1` if the base url is `api.openai.com`
	// for the client sdk reason
	if project.OpenAIBaseURL != "" {
		baseUrl, err := url.Parse(project.OpenAIBaseURL)
		if err != nil {
			logrus.Errorln(err)
			return reply, err
		}
		if baseUrl.Path != "/v1" {
			baseUrl.Path = "/v1"
		}
		cfg.BaseURL = baseUrl.String()
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
		pt := openai.ChatCompletionMessage{
			Role:    prompt.Role,
			Content: content,
		}
		req.Messages = append(req.Messages, pt)
	}

	return client.CreateChatCompletion(ctx, req)
}
