package routes

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	cache "github.com/Code-Hex/go-generics-cache"
	"github.com/PromptPal/PromptPal/ent"
	"github.com/PromptPal/PromptPal/ent/project"
	"github.com/PromptPal/PromptPal/ent/prompt"
	"github.com/PromptPal/PromptPal/ent/schema"
	"github.com/PromptPal/PromptPal/service"
	"github.com/gin-gonic/gin"
	"github.com/sashabaranov/go-openai"
	"github.com/sirupsen/logrus"
)

type publicPromptItem struct {
	HashID      string                  `json:"id"`
	Name        string                  `json:"name"`
	Description string                  `json:"description"`
	TokenCount  int                     `json:"tokenCount"`
	Variables   []schema.PromptVariable `json:"variables"`
	CreatedAt   time.Time               `json:"createdAt"`
}

func apiListPrompts(c *gin.Context) {
	pid := c.GetInt("pid")
	var query queryPagination
	if err := c.BindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{
			ErrorCode:    http.StatusBadRequest,
			ErrorMessage: err.Error(),
		})
		return
	}

	prompts, err := service.EntClient.
		Prompt.
		Query().
		Where(prompt.HasProjectWith(project.ID(pid))).
		Where(prompt.IDLT(query.Cursor)).
		Limit(query.Limit).
		Order(ent.Desc(prompt.FieldID)).
		All(c)

	if err != nil {
		c.JSON(http.StatusNotFound, errorResponse{
			ErrorCode:    http.StatusNotFound,
			ErrorMessage: err.Error(),
		})
		return
	}

	count, err := service.
		EntClient.
		Prompt.
		Query().
		Where(prompt.HasProjectWith(project.ID(pid))).
		Count(c)

	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse{
			ErrorCode:    http.StatusInternalServerError,
			ErrorMessage: err.Error(),
		})
		return
	}

	result := make([]publicPromptItem, len(prompts))
	for i, prompt := range prompts {
		hid, err := hashidService.Encode(prompt.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, errorResponse{
				ErrorCode:    http.StatusInternalServerError,
				ErrorMessage: err.Error(),
			})
			return
		}

		result[i] = publicPromptItem{
			HashID:      hid,
			Name:        prompt.Name,
			Description: prompt.Description,
			Variables:   prompt.Variables,
			TokenCount:  prompt.TokenCount,
			CreatedAt:   prompt.CreateTime,
		}
	}
	c.JSON(http.StatusOK, ListResponse[publicPromptItem]{
		Count: count,
		Data:  result,
	})
}

type apiRunPromptPayload struct {
	Variables map[string]string `json:"variables" binding:"required"`
	UserId    string            `json:"userId"`
}

type apiRunPromptResponse struct {
	PromptID           string `json:"id"`
	ResponseMessage    string `json:"message"`
	ResponseTokenCount int    `json:"tokenCount"`
}

func apiRunPromptMiddleware(c *gin.Context) {
	hashedValue, ok := c.Params.Get("id")

	if !ok {
		c.AbortWithStatusJSON(http.StatusBadRequest, errorResponse{
			ErrorCode:    http.StatusBadRequest,
			ErrorMessage: "invalid id",
		})
		return
	}

	prompt, ok := service.ApiPromptCache.Get(hashedValue)
	if !ok {
		promptID, err := hashidService.Decode(hashedValue)

		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, errorResponse{
				ErrorCode:    http.StatusInternalServerError,
				ErrorMessage: err.Error(),
			})
			return
		}
		promptData, err := service.EntClient.Prompt.Get(c, promptID)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusNotFound, errorResponse{
				ErrorCode:    http.StatusNotFound,
				ErrorMessage: err.Error(),
			})
			return
		}
		service.ApiPromptCache.Set(hashedValue, *promptData, cache.WithExpiration(24*time.Hour))
		prompt = *promptData
	}

	var payload apiRunPromptPayload
	if err := c.Bind(&payload); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, errorResponse{
			ErrorCode:    http.StatusBadRequest,
			ErrorMessage: err.Error(),
		})
		return
	}

	// check the API token and prompt.projectID is equal
	pid := c.GetInt("pid")
	pj, ok := service.ProjectCache.Get(pid)

	if !ok {
		pjt, err := service.EntClient.Project.Get(c, prompt.ProjectId)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusNotFound, errorResponse{
				ErrorCode:    http.StatusNotFound,
				ErrorMessage: err.Error(),
			})
			return
		}
		service.ProjectCache.Set(pid, *pjt, cache.WithExpiration(24*time.Hour))
		pj = *pjt
	}

	if pj.ID != pid {
		c.AbortWithStatusJSON(http.StatusForbidden, errorResponse{
			ErrorCode:    http.StatusForbidden,
			ErrorMessage: "prompt does not belong to the project",
		})
		return
	}

	c.Set("prompt", prompt)
	c.Set("pj", pj)
	c.Set("payload", payload)
	c.Next()
}

func apiRunPrompt(c *gin.Context) {
	hashedValue, _ := c.Params.Get("id")
	promptData, _ := c.Get("prompt")
	pjData, _ := c.Get("pj")
	payloadData, _ := c.Get("payload")

	prompt := promptData.(ent.Prompt)
	pj := pjData.(ent.Project)
	payload := payloadData.(apiRunPromptPayload)

	startTime := time.Now()
	responseResult := 0
	res, err := aiService.Chat(c, pj, prompt.Prompts, payload.Variables, payload.UserId)
	endTime := time.Now()

	defer func() {
		stat := service.EntClient.
			PromptCall.
			Create().
			SetPromptID(prompt.ID).
			SetResult(responseResult).
			SetResponseToken(res.Usage.CompletionTokens).
			SetTotalToken(res.Usage.TotalTokens).
			SetUserId(payload.UserId).
			SetDuration(endTime.Sub(startTime).Milliseconds()).
			SetProjectID(pj.ID).
			SetUa(c.Request.UserAgent())

		if prompt.Debug {
			stat.SetPayload(payload.Variables)
		}
		if prompt.Debug && len(res.Choices) > 0 {
			stat.SetMessage(res.Choices[0].Message.Content)
		}

		cost, err := service.GetCosts(pj.OpenAIModel, endTime)
		if err != nil {
			logrus.Errorln(err)
			err = nil
		} else {
			inputCosts := cost.InputTokenCostInCents * float64(res.Usage.PromptTokens)
			outputCosts := cost.OutputTokenCostInCents * float64(res.Usage.CompletionTokens)
			stat.SetCostCents(inputCosts + outputCosts)
		}

		exp := stat.Exec(c)
		if exp != nil {
			logrus.Errorln(exp)
		}
	}()

	if err != nil {
		responseResult = 1
		c.JSON(http.StatusInternalServerError, errorResponse{
			ErrorCode:    http.StatusInternalServerError,
			ErrorMessage: err.Error(),
		})
		return
	}

	if len(res.Choices) == 0 {
		c.JSON(http.StatusBadRequest, errorResponse{
			ErrorCode:    http.StatusBadRequest,
			ErrorMessage: "no choices",
		})
		return
	}

	result := apiRunPromptResponse{}
	result.PromptID = hashedValue
	result.ResponseMessage = res.Choices[0].Message.Content
	result.ResponseTokenCount = res.Usage.CompletionTokens

	c.Header("Server-Timing", fmt.Sprintf("prompt;dur=%d", endTime.Sub(startTime).Milliseconds()))
	c.JSON(http.StatusOK, result)
}

func apiRunPromptStream(c *gin.Context) {
	hashedValue, _ := c.Params.Get("id")
	promptData, _ := c.Get("prompt")
	pjData, _ := c.Get("pj")
	payloadData, _ := c.Get("payload")

	prompt := promptData.(ent.Prompt)
	pj := pjData.(ent.Project)
	payload := payloadData.(apiRunPromptPayload)

	replyStream, err := aiService.ChatStream(c, pj, prompt.Prompts, payload.Variables, payload.UserId)

	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, errorResponse{
			ErrorCode:    http.StatusInternalServerError,
			ErrorMessage: err.Error(),
		})
		return
	}

	startTime := time.Now()
	responseResult := 0
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("Transfer-Encoding", "chunked")

	var info openai.Usage
	result := make([]openai.ChatCompletionChoice, 0)

	c.Stream(func(w io.Writer) bool {
		select {
		case _info := <-replyStream.Info:
			info = _info
			return true
		case <-replyStream.Done:
			close(replyStream.Done)
			close(replyStream.Err)
			close(replyStream.Message)
			close(replyStream.Info)
			return false
		case err := <-replyStream.Err:
			c.SSEvent("error", err.Error())
			responseResult = 1
			return false
		case data := <-replyStream.Message:
			result = append(result, data...)
			chunkResponse := apiRunPromptResponse{
				PromptID:           hashedValue,
				ResponseMessage:    data[0].Message.Content,
				ResponseTokenCount: -1,
			}
			b, err := json.Marshal(chunkResponse)
			if err != nil {
				c.SSEvent("error", err.Error())
				return false
			}
			c.SSEvent("message", string(b))
			return true
		}
	})

	endTime := time.Now()

	defer func() {
		stat := service.EntClient.
			PromptCall.
			Create().
			SetPromptID(prompt.ID).
			SetResult(responseResult).
			SetResponseToken(info.CompletionTokens).
			SetTotalToken(info.TotalTokens).
			SetUserId(payload.UserId).
			SetDuration(endTime.Sub(startTime).Milliseconds()).
			SetProjectID(pj.ID).
			SetUa(c.Request.UserAgent())

		if prompt.Debug {
			stat.SetPayload(payload.Variables)
		}

		choices := make([]string, 0)

		for _, choice := range result {
			choices = append(choices, choice.Message.Content)
		}

		if prompt.Debug && len(choices) > 0 {
			stat.SetMessage(choices[len(choices)-1])
		}

		cost, err := service.GetCosts(pj.OpenAIModel, endTime)
		if err != nil {
			logrus.Errorln(err)
			err = nil
		} else {
			inputCosts := cost.InputTokenCostInCents * float64(info.PromptTokens)
			outputCosts := cost.OutputTokenCostInCents * float64(info.CompletionTokens)
			stat.SetCostCents(inputCosts + outputCosts)
		}
		exp := stat.Exec(c)
		if exp != nil {
			logrus.Errorln(exp)
		}
	}()
}
