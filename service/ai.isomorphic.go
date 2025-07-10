package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/PromptPal/PromptPal/ent"
	openai "github.com/sashabaranov/go-openai"
	"github.com/sirupsen/logrus"
)

type isomorphicAIService struct {
}

type loggingTransport struct{}

func (s *loggingTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	bytes, _ := httputil.DumpRequestOut(r, true)

	resp, err := http.DefaultTransport.RoundTrip(r)
	// err is returned after dumping the response

	respBytes, _ := httputil.DumpResponse(resp, true)
	bytes = append(bytes, respBytes...)

	fmt.Printf("%s\n", bytes)

	return resp, err
}

//go:generate mockery --name IsomorphicAIService
type IsomorphicAIService interface {
	GetProvider(ctx context.Context, prompt ent.Prompt) (provider *ent.Provider, err error)
	Chat(
		ctx context.Context,
		provider *ent.Provider,
		prompt ent.Prompt,
		variables map[string]string,
		userId string,
	) (reply openai.ChatCompletionResponse, err error)
	ChatStream(
		ctx context.Context,
		provider *ent.Provider,
		prompt ent.Prompt,
		variables map[string]string,
		userId string,
	) (reply *ChatStreamResponse, err error)
}

func NewIsomorphicAIService() IsomorphicAIService {
	return &isomorphicAIService{}
}

func (o isomorphicAIService) getIsomorphicClient(ctx context.Context, provider *ent.Provider) (*openai.Client, error) {
	cfg := openai.DefaultConfig(provider.ApiKey)
	if provider.Endpoint != "" {
		baseUrl, err := url.Parse(provider.Endpoint)
		if err != nil {
			logrus.Errorln(err)
			return nil, err
		}
		cfg.BaseURL = baseUrl.String()
	}
	client := openai.NewClientWithConfig(cfg)
	return client, nil
}

func (o isomorphicAIService) GetProvider(ctx context.Context, prompt ent.Prompt) (provider *ent.Provider, err error) {
	promptProvider, err := EntClient.Provider.Get(ctx, prompt.ProviderId)

	if err != nil && !ent.IsNotFound(err) {
		return
	}

	if err == nil && promptProvider != nil {
		provider = promptProvider
		return
	}

	err = nil

	pj, err := EntClient.Project.Get(ctx, prompt.ProjectId)
	if err != nil {
		return
	}

	projectProvider, err := EntClient.Provider.Get(ctx, pj.ProviderId)
	if err != nil && !ent.IsNotFound(err) {
		return
	}
	if err == nil && projectProvider != nil {
		provider = projectProvider
		return
	}
	err = nil

	dummyProvider := &ent.Provider{
		Endpoint:     pj.OpenAIBaseURL,
		Source:       "openai",
		ApiKey:       pj.OpenAIToken,
		DefaultModel: pj.OpenAIModel,
		Temperature:  pj.OpenAITemperature,
		TopP:         pj.OpenAITopP,
		MaxTokens:    pj.OpenAIMaxTokens,
	}

	if pj.GeminiToken != "" {
		dummyProvider.Source = "gemini"
		dummyProvider.ApiKey = pj.GeminiToken
	}
	provider = dummyProvider
	return
}

// just for mock
func (o isomorphicAIService) Chat(
	ctx context.Context,
	provider *ent.Provider,
	prompt ent.Prompt,
	variables map[string]string,
	userId string,
) (reply openai.ChatCompletionResponse, err error) {
	client, err := o.getIsomorphicClient(ctx, provider)
	if err != nil {
		return
	}

	req := openai.ChatCompletionRequest{
		Model:       provider.DefaultModel,
		Temperature: float32(provider.Temperature),
		TopP:        float32(provider.TopP),
	}
	if userId != "" {
		req.User = userId
	}
	if provider.MaxTokens > 0 {
		req.MaxTokens = provider.MaxTokens
	}

	prompts := prompt.Prompts

	logrus.Debugln("openai:chat: prompts need to send", prompts)
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

func (o isomorphicAIService) ChatStream(
	ctx context.Context,
	provider *ent.Provider,
	prompt ent.Prompt,
	variables map[string]string,
	userId string,
) (reply *ChatStreamResponse, err error) {
	reply = &ChatStreamResponse{
		Done:    make(chan bool),
		Err:     make(chan error),
		Info:    make(chan openai.Usage),
		Message: make(chan []openai.ChatCompletionChoice),
	}

	client, err := o.getIsomorphicClient(ctx, provider)

	if err != nil {
		return reply, err
	}

	req := openai.ChatCompletionRequest{
		Model:       provider.DefaultModel,
		Temperature: float32(provider.Temperature),
		TopP:        float32(provider.TopP),
	}
	if userId != "" {
		req.User = userId
	}
	if provider.MaxTokens > 0 {
		req.MaxTokens = provider.MaxTokens
	}

	prompts := prompt.Prompts

	logrus.Debugln("openai:stream: prompts need to send", prompts, variables)
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
