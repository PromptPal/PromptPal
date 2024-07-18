package routes

import (
	"fmt"
	"net/http"
	"time"

	"github.com/PromptPal/PromptPal/ent"
	"github.com/PromptPal/PromptPal/service"
	"github.com/gin-gonic/gin"
	"github.com/sashabaranov/go-openai"
	"github.com/sirupsen/logrus"
)

func promptCacheMiddleware(c *gin.Context) {
	hashedValue, _ := c.Params.Get("id")
	promptData, _ := c.Get("prompt")
	payloadData, _ := c.Get("payload")
	pjData, _ := c.Get("pj")

	prompt := promptData.(ent.Prompt)
	payload := payloadData.(apiRunPromptPayload)
	pj := pjData.(ent.Project)

	if !prompt.CacheEnabled {
		c.Next()
		return
	}

	startTime := time.Now()
	result, ok, err := service.GetPromptResponseCache(hashedValue, payload.Variables)

	if err != nil {
		logrus.Warnln("promptCache", err)
	}
	if !ok {
		c.Next()
		return
	}

	endTime := time.Now()
	defer savePromptCall(
		c.Request.Context(),
		prompt,
		1,
		openai.ChatCompletionResponse{
			Usage:   openai.Usage{TotalTokens: result.ResponseTokenCount, CompletionTokens: 0},
			Choices: []openai.ChatCompletionChoice{{Message: openai.ChatCompletionMessage{Content: result.ResponseMessage}}},
		},
		pj,
		payload,
		endTime,
		startTime,
		c.Request.UserAgent(),
		true,
	)
	c.Header("Server-Timing", fmt.Sprintf("prompt;dur=%d", endTime.Sub(startTime).Milliseconds()))
	c.AbortWithStatusJSON(http.StatusOK, result)
}
