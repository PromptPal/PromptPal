package routes

import (
	"net/http"

	"github.com/PromptPal/PromptPal/ent"
	"github.com/PromptPal/PromptPal/ent/schema"
	"github.com/PromptPal/PromptPal/service"
	"github.com/gin-gonic/gin"
)

type testPromptPayload struct {
	ProjectID  int                `json:"projectId" binding:"required"`
	ProviderID int                `json:"providerId" binding:"required"`
	Name       string             `json:"name"`
	Prompts    []schema.PromptRow `json:"prompts"`
	Variables  map[string]string  `json:"variables"`
}

func testPrompt(c *gin.Context) {
	uid := c.GetInt("uid")
	if uid == 0 {
		c.JSON(http.StatusUnauthorized, errorResponse{
			ErrorCode:    http.StatusUnauthorized,
			ErrorMessage: "invalid uid",
		})
		return
	}

	var payload testPromptPayload
	if err := c.Bind(&payload); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{
			ErrorCode:    http.StatusBadRequest,
			ErrorMessage: err.Error(),
		})
		return
	}

	provider, err := service.EntClient.Provider.Get(c, payload.ProviderID)
	if err != nil {
		c.JSON(http.StatusNotFound, errorResponse{
			ErrorCode:    http.StatusNotFound,
			ErrorMessage: err.Error(),
		})
		return
	}

	prompt := ent.Prompt{
		Prompts: payload.Prompts,
	}

	res, err := isomorphicAIService.Chat(c.Request.Context(), provider, prompt, payload.Variables, "")

	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse{
			ErrorCode:    http.StatusInternalServerError,
			ErrorMessage: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, res)
}
