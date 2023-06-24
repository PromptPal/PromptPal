package routes

import (
	"net/http"
	"strconv"

	"github.com/PromptPal/PromptPal/ent"
	"github.com/PromptPal/PromptPal/ent/project"
	"github.com/PromptPal/PromptPal/ent/prompt"
	"github.com/PromptPal/PromptPal/ent/schema"
	"github.com/PromptPal/PromptPal/service"
	"github.com/gin-gonic/gin"
)

type listPromptsResponse struct {
	Prompts    []*ent.Prompt `json:"prompts"`
	Pagination Pagination    `json:"pagination"`
}

func listProjectPrompts(c *gin.Context) {

	var payload paginationPayload
	if err := c.Bind(&payload); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{
			ErrorCode:    http.StatusBadRequest,
			ErrorMessage: err.Error(),
		})
		return
	}

	if payload.Pagination.Limit > 50 {
		c.JSON(http.StatusBadRequest, errorResponse{
			ErrorCode:    http.StatusBadRequest,
			ErrorMessage: "limit must be less than 50",
		})
		return
	}

	idStr, ok := c.Params.Get("id")
	if !ok {
		c.JSON(http.StatusBadRequest, errorResponse{
			ErrorCode:    http.StatusBadRequest,
			ErrorMessage: "invalid id",
		})
		return
	}

	pid, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse{
			ErrorCode:    http.StatusInternalServerError,
			ErrorMessage: err.Error(),
		})
		return
	}

	prompts, err := service.
		EntClient.
		Prompt.
		Query().
		Where(prompt.HasProjectWith(project.ID(pid))).
		Where(prompt.IDLT(payload.Pagination.Cursor)).
		Limit(payload.Pagination.Limit).
		Order(ent.Desc(prompt.FieldID)).
		All(c)

	if err != nil {
		c.JSON(http.StatusNotFound, errorResponse{
			ErrorCode:    http.StatusNotFound,
			ErrorMessage: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, listPromptsResponse{
		Prompts: prompts,
		Pagination: Pagination{
			Count:  len(prompts),
			Cursor: 0,
		},
	})
}

func listPrompts(c *gin.Context) {
	// TODO: only admin can do this

	var payload paginationPayload
	if err := c.Bind(&payload); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{
			ErrorCode:    http.StatusBadRequest,
			ErrorMessage: err.Error(),
		})
		return
	}

	if payload.Pagination.Limit > 50 {
		c.JSON(http.StatusBadRequest, errorResponse{
			ErrorCode:    http.StatusBadRequest,
			ErrorMessage: "limit must be less than 50",
		})
		return
	}

	prompts, err := service.
		EntClient.
		Prompt.
		Query().
		Where(prompt.IDLT(payload.Pagination.Cursor)).
		Limit(payload.Pagination.Limit).
		Order(ent.Desc(prompt.FieldID)).
		All(c)

	if err != nil {
		c.JSON(http.StatusNotFound, errorResponse{
			ErrorCode:    http.StatusNotFound,
			ErrorMessage: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, listPromptsResponse{
		Prompts: prompts,
		Pagination: Pagination{
			Count:  len(prompts),
			Cursor: 0,
		},
	})
}

func getPrompt(c *gin.Context) {
	idStr, ok := c.Params.Get("id")
	if !ok {
		c.JSON(http.StatusBadRequest, errorResponse{
			ErrorCode:    http.StatusBadRequest,
			ErrorMessage: "invalid id",
		})
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse{
			ErrorCode:    http.StatusInternalServerError,
			ErrorMessage: err.Error(),
		})
		return
	}

	prompt, err := service.EntClient.Prompt.Get(c, id)
	if err != nil {
		c.JSON(http.StatusNotFound, errorResponse{
			ErrorCode:    http.StatusNotFound,
			ErrorMessage: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, prompt)
}

type createPromptPayload struct {
	ProjectID   int                     `json:"projectId"`
	TokenCount  int                     `json:"tokenCount"`
	Prompts     []schema.PromptRow      `json:"prompts"`
	Variables   []schema.PromptVariable `json:"variables"`
	PublicLevel prompt.PublicLevel      `json:"publicLevel"`
}

func createPrompt(c *gin.Context) {
	var payload createPromptPayload
	if err := c.Bind(&payload); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{
			ErrorCode:    http.StatusBadRequest,
			ErrorMessage: err.Error(),
		})
		return
	}

	p, err := service.
		EntClient.
		Prompt.
		Create().
		SetCreatorID(c.GetInt("uid")).
		SetProjectID(payload.ProjectID).
		SetPrompts(payload.Prompts).
		SetVariables(payload.Variables).
		SetPublicLevel(payload.PublicLevel).
		SetTokenCount(payload.TokenCount).
		Save(c)

	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse{
			ErrorCode:    http.StatusInternalServerError,
			ErrorMessage: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, p)
}

func updatePrompt(c *gin.Context) {
}

func testPrompt(c *gin.Context) {
	var payload createPromptPayload
	if err := c.Bind(&payload); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{
			ErrorCode:    http.StatusBadRequest,
			ErrorMessage: err.Error(),
		})
		return
	}

	pj, err := service.EntClient.Project.Get(c, payload.ProjectID)
	if err != nil {
		c.JSON(http.StatusNotFound, errorResponse{
			ErrorCode:    http.StatusNotFound,
			ErrorMessage: err.Error(),
		})
		return
	}

	res, err := openAIService.Chat(c, pj, payload.Prompts, payload.Variables)

	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse{
			ErrorCode:    http.StatusInternalServerError,
			ErrorMessage: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, res)
}
