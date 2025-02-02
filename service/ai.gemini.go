package service

import (
	"context"
	"errors"
	"log"
	"strings"

	"github.com/PromptPal/PromptPal/ent"
	"github.com/PromptPal/PromptPal/ent/schema"
	"github.com/google/generative-ai-go/genai"
	openai "github.com/sashabaranov/go-openai"
	"github.com/sirupsen/logrus"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

type geminiService struct {
}

func NewGeminiService() BaseAIService {
	return &geminiService{}
}

func (o geminiService) getGeminiClient(ctx context.Context, project ent.Project) (*genai.Client, *genai.GenerativeModel, error) {
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

func (o geminiService) Chat(
	ctx context.Context,
	project ent.Project,
	prompts []schema.PromptRow,
	variables map[string]string,
	userId string,
) (reply openai.ChatCompletionResponse, err error) {
	if !strings.HasPrefix(project.OpenAIModel, "gemini") {
		return reply, errors.New("not gemini model")
	}

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

func (o geminiService) ChatStream(
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

	if !strings.HasPrefix(project.OpenAIModel, "gemini") {
		return reply, errors.New("not gemini model")
	}

	client, genModel, err := o.getGeminiClient(ctx, project)
	if err != nil {
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
	go func() {
		for {
			resp, err := iter.Next()
			if err == iterator.Done {
				usage := openai.Usage{
					CompletionTokens: int(completionTokenCount),
					PromptTokens:     int(promptTokenCount.TotalTokens),
					TotalTokens:      int(completionTokenCount) + int(promptTokenCount.TotalTokens),
				}
				reply.Info <- usage
				reply.Done <- true
				break
			}
			if err != nil {
				reply.Err <- err
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
			reply.Message <- result
		}
	}()

	return reply, nil
}
