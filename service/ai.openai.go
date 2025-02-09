package service

import (
	"context"
	"errors"
	"io"
	"net/url"

	"github.com/PromptPal/PromptPal/ent"
	"github.com/PromptPal/PromptPal/ent/schema"
	openai "github.com/sashabaranov/go-openai"
	"github.com/sirupsen/logrus"
)

type openAIService struct {
}

func NewOpenAIService() BaseAIService {
	return &openAIService{}
}

func (o openAIService) getOpenAIClient(ctx context.Context, project ent.Project) (*openai.Client, error) {
	cfg := openai.DefaultConfig(project.OpenAIToken)
	if project.OpenAIBaseURL != "" {
		baseUrl, err := url.Parse(project.OpenAIBaseURL)
		if err != nil {
			logrus.Errorln(err)
			return nil, err
		}
		cfg.BaseURL = baseUrl.String()
	}
	client := openai.NewClientWithConfig(cfg)
	return client, nil
}

// just for mock
func (o openAIService) Chat(
	ctx context.Context,
	project ent.Project,
	prompts []schema.PromptRow,
	variables map[string]string,
	userId string,
) (reply openai.ChatCompletionResponse, err error) {
	if project.OpenAIToken == "" {
		return reply, errors.New("token is empty")
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

	logrus.Infoln("openai:chat: prompts need to send", prompts)
	for _, prompt := range prompts {
		content := replacePlaceholdersLegacy(prompt.Prompt, variables)
		pt := openai.ChatCompletionMessage{
			Role:    prompt.Role,
			Content: content,
		}
		req.Messages = append(req.Messages, pt)
	}

	return client.CreateChatCompletion(ctx, req)
}

func (o openAIService) ChatStream(
	ctx context.Context,
	project ent.Project,
	prompts []schema.PromptRow,
	variables map[string]string,
	userId string,
) (reply *ChatStreamResponse, err error) {
	reply = &ChatStreamResponse{
		Done:    make(chan bool),
		Err:     make(chan error),
		Info:    make(chan openai.Usage),
		Message: make(chan []openai.ChatCompletionChoice),
	}

	if project.OpenAIToken == "" {
		return reply, errors.New("token is empty")
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

	logrus.Infoln("openai:stream: prompts need to send", prompts, variables)
	for _, prompt := range prompts {
		content := replacePlaceholdersLegacy(prompt.Prompt, variables)
		pt := openai.ChatCompletionMessage{
			Role:    prompt.Role,
			Content: content,
		}
		req.Messages = append(req.Messages, pt)
	}
	req.StreamOptions = &openai.StreamOptions{
		IncludeUsage: true,
	}

	stream, err := client.CreateChatCompletionStream(ctx, req)

	if err != nil {
		return reply, err
	}

	go func() {
		for {
			resp, err := stream.Recv()
			if errors.Is(err, io.EOF) {
				err = nil
				reply.Done <- true
				stream.Close()
				break
			}
			if err != nil {
				reply.Err <- err
				stream.Close()
				break
			}

			if resp.Usage != nil {
				reply.Info <- *resp.Usage
			}

			if len(resp.Choices) == 0 {
				continue
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

			reply.Message <- temp
		}
	}()
	return reply, nil
}
