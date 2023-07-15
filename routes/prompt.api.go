package routes

import (
	"net/http"
	"time"

	cache "github.com/Code-Hex/go-generics-cache"
	"github.com/PromptPal/PromptPal/ent"
	"github.com/PromptPal/PromptPal/ent/project"
	"github.com/PromptPal/PromptPal/ent/prompt"
	"github.com/PromptPal/PromptPal/ent/schema"
	"github.com/PromptPal/PromptPal/service"
	"github.com/gin-gonic/gin"
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
	Variables map[string]string `json:"variables"`
	UserId    string            `json:"userId"`
}

type apiRunPromptResponse struct {
	PromptID           string `json:"id"`
	ResponseMessage    string `json:"message"`
	ResponseTokenCount int    `json:"tokenCount"`
}

var apiPromptCache *cache.Cache[string, ent.Prompt]
var projectCache *cache.Cache[int, ent.Project]

func init() {
	apiPromptCache = cache.New[string, ent.Prompt]()
	projectCache = cache.New[int, ent.Project]()
}

func apiRunPrompt(c *gin.Context) {
	hashedValue, ok := c.Params.Get("id")

	if !ok {
		c.JSON(http.StatusBadRequest, errorResponse{
			ErrorCode:    http.StatusBadRequest,
			ErrorMessage: "invalid id",
		})
		return
	}

	prompt, ok := apiPromptCache.Get(hashedValue)
	if !ok {
		promptID, err := hashidService.Decode(hashedValue)

		if err != nil {
			c.JSON(http.StatusInternalServerError, errorResponse{
				ErrorCode:    http.StatusInternalServerError,
				ErrorMessage: err.Error(),
			})
			return
		}
		promptData, err := service.EntClient.Prompt.Get(c, promptID)
		if err != nil {
			c.JSON(http.StatusNotFound, errorResponse{
				ErrorCode:    http.StatusNotFound,
				ErrorMessage: err.Error(),
			})
			return
		}
		apiPromptCache.Set(hashedValue, *promptData, cache.WithExpiration(24*time.Hour))
		prompt = *promptData
	}

	var payload apiRunPromptPayload
	if err := c.Bind(&payload); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{
			ErrorCode:    http.StatusBadRequest,
			ErrorMessage: err.Error(),
		})
		return
	}

	// check the API token and prompt.projectID is equal
	pid := c.GetInt("pid")
	pj, ok := projectCache.Get(pid)

	if !ok {
		pjt, err := prompt.QueryProject().Only(c)
		if err != nil {
			c.JSON(http.StatusNotFound, errorResponse{
				ErrorCode:    http.StatusNotFound,
				ErrorMessage: err.Error(),
			})
			return
		}
		projectCache.Set(pid, *pjt, cache.WithExpiration(24*time.Hour))
		pj = *pjt
	}

	if pj.ID != pid {
		c.JSON(http.StatusForbidden, errorResponse{
			ErrorCode:    http.StatusForbidden,
			ErrorMessage: "prompt does not belong to the project",
		})
		return
	}

	startTime := time.Now()
	responseResult := 0
	res, err := openAIService.Chat(c, pj, prompt.Prompts, payload.Variables, payload.UserId)
	endTime := time.Now()

	defer func() {
		message := ""
		if len(res.Choices) > 0 {
			message = res.Choices[0].Message.Content
		}
		exp := service.EntClient.
			PromptCall.
			Create().
			SetPromptID(prompt.ID).
			SetResult(responseResult).
			SetResponseToken(res.Usage.CompletionTokens).
			SetTotalToken(res.Usage.TotalTokens).
			SetMessage(message).
			SetUserId(payload.UserId).
			SetDuration(endTime.Sub(startTime).Milliseconds()).
			SetProjectID(pj.ID).
			Exec(c)
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

	c.JSON(http.StatusOK, result)
}
