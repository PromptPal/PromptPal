package service

import (
	"context"

	"github.com/PromptPal/PromptPal/ent"
	"github.com/PromptPal/PromptPal/ent/schema"
	openai "github.com/sashabaranov/go-openai"
)

type ChatStreamResponse struct {
	Message chan []openai.ChatCompletionChoice
	Done    chan bool
	Err     chan error
	Info    chan openai.Usage
}

//go:generate mockery --name BaseAIService
type BaseAIService interface {
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
	) (reply *ChatStreamResponse, err error)
}
