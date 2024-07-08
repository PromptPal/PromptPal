package service

import (
	"context"
	"errors"
	"io"
	"log"
	"net/url"
	"strings"

	"github.com/PromptPal/PromptPal/ent"
	"github.com/PromptPal/PromptPal/ent/schema"
	"github.com/google/generative-ai-go/genai"
	openai "github.com/sashabaranov/go-openai"
	"github.com/sirupsen/logrus"
	"google.golang.org/api/iterator"
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
	ChatStream(
		ctx context.Context,
		project ent.Project,
		prompts []schema.PromptRow,
		variables map[string]string,
		userId string,
		onData func(data []openai.ChatCompletionChoice) error,
	) (reply openai.ChatCompletionResponse, err error)
}

type aiService struct {
}

func NewOpenAIService() OpenAIService {
	return &aiService{}
}

func (o aiService) getOpenAIClient(ctx context.Context, project ent.Project) (*openai.Client, error) {
	cfg := openai.DefaultConfig(project.OpenAIToken)

	// add `/v1` if the base url is `api.openai.com`
	// for the client sdk reason
	if project.OpenAIBaseURL != "" {
		baseUrl, err := url.Parse(project.OpenAIBaseURL)
		if err != nil {
			logrus.Errorln(err)
			return nil, err
		}
		if baseUrl.Path != "/v1" {
			baseUrl.Path = "/v1"
		}
		cfg.BaseURL = baseUrl.String()
	}
	client := openai.NewClientWithConfig(cfg)
	return client, nil
}

func (o aiService) getGeminiClient(ctx context.Context, project ent.Project) (*genai.Client, *genai.GenerativeModel, error) {
	if !strings.HasPrefix(project.OpenAIModel, "gemini") {
		return nil, nil, errors.New("not gemini model")
	}
	client, err := genai.NewClient(
		ctx,
		option.WithAPIKey(project.GeminiToken),
		option.WithEndpoint(project.GeminiBaseURL),
	)
	if err != nil {
		return nil, nil, err
	}
	genModel := client.GenerativeModel(project.OpenAIModel)
	genModel.SetTemperature(float32(project.OpenAITemperature))
	genModel.SetTopP(float32(project.OpenAITopP))
	genModel.SetTopK(20)
	if project.OpenAIMaxTokens > 0 {
		genModel.SetMaxOutputTokens(int32(project.OpenAIMaxTokens))
	}

	return client, genModel, nil
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
		client, genModel, err := o.getGeminiClient(ctx, project)
		if err != nil {
			log.Fatal(err)
			return reply, err
		}
		defer client.Close()

		pts := []genai.Part{}
		for _, prompt := range prompts {
			content := replacePlaceholders(prompt.Prompt, variables)
			txt := genai.Text(content)
			pts = append(pts, txt)
		}

		promptTokenCount, err := genModel.CountTokens(ctx, pts...)

		if err != nil {
			return reply, err
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
				PromptTokens:     int(promptTokenCount.TotalTokens),
				TotalTokens:      int(completionTokenCount),
			},
		}, nil
	}

	client, err := o.getOpenAIClient(ctx, project)

	if err != nil {
		return reply, err
	}

	req := openai.ChatCompletionRequest{
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
		pt := openai.ChatCompletionMessage{
			Role:    prompt.Role,
			Content: content,
		}
		req.Messages = append(req.Messages, pt)
	}

	return client.CreateChatCompletion(ctx, req)
}

func (o aiService) ChatStream(
	ctx context.Context,
	project ent.Project,
	prompts []schema.PromptRow,
	variables map[string]string,
	userId string,
	onData func(data []openai.ChatCompletionChoice) error,
) (reply openai.ChatCompletionResponse, err error) {

	if project.OpenAIToken == "" {
		return reply, errors.New("token is empty")
	}

	// gemini support
	if strings.HasPrefix(project.OpenAIModel, "gemini") {
		client, genModel, err := o.getGeminiClient(ctx, project)
		if err != nil {
			log.Fatal(err)
			return reply, err
		}
		defer client.Close()

		pts := []genai.Part{}
		for _, prompt := range prompts {
			content := replacePlaceholders(prompt.Prompt, variables)
			txt := genai.Text(content)
			pts = append(pts, txt)
		}

		promptTokenCount, err := genModel.CountTokens(ctx, pts...)

		if err != nil {
			return reply, err
		}

		iter := genModel.GenerateContentStream(ctx, pts...)
		completionTokenCount := int32(0)
		for {
			resp, err := iter.Next()
			if err == iterator.Done {
				break
			}
			if err != nil {
				break
			}

			result := []openai.ChatCompletionChoice{}

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
			onData(result)
		}

		reply = openai.ChatCompletionResponse{
			ID:      "",
			Choices: []openai.ChatCompletionChoice{},
			Usage: openai.Usage{
				CompletionTokens: int(completionTokenCount),
				PromptTokens:     int(promptTokenCount.TotalTokens),
				TotalTokens:      int(completionTokenCount) + int(promptTokenCount.TotalTokens),
			},
		}
		return reply, nil
	}

	client, err := o.getOpenAIClient(ctx, project)

	if err != nil {
		return reply, err
	}

	req := openai.ChatCompletionRequest{
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
		pt := openai.ChatCompletionMessage{
			Role:    prompt.Role,
			Content: content,
		}
		req.Messages = append(req.Messages, pt)
	}

	stream, err := client.CreateChatCompletionStream(ctx, req)

	if err != nil {
		return reply, err
	}
	defer stream.Close()

	var info openai.ChatCompletionStreamResponse
	for {
		resp, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			info = resp
			err = nil
			break
		}
		if err != nil {
			break
		}

		temp := make([]openai.ChatCompletionChoice, len(resp.Choices))

		for i, cand := range resp.Choices {
			content := cand.Delta.Content
			chunk := openai.ChatCompletionChoice{
				Index:        cand.Index,
				FinishReason: openai.FinishReasonStop,
				Message: openai.ChatCompletionMessage{
					Role:    cand.Delta.Role,
					Content: content,
				},
			}
			temp[i] = chunk
		}
		onData(temp)
	}

	logrus.Println("xxxx", info)

	reply = openai.ChatCompletionResponse{
		ID:      info.ID,
		Choices: []openai.ChatCompletionChoice{},
		Usage:   openai.Usage{},
	}
	return reply, nil
}
