package routes

import (
	"net/http"
	"strconv"

	"github.com/PromptPal/PromptPal/ent"
	"github.com/PromptPal/PromptPal/ent/project"
	"github.com/PromptPal/PromptPal/ent/prompt"
	"github.com/PromptPal/PromptPal/service"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func apiListPrompts(c *gin.Context) {
	pidStr, ok := c.Params.Get("id")

	if !ok {
		c.JSON(http.StatusBadRequest, errorResponse{
			ErrorCode:    http.StatusBadRequest,
			ErrorMessage: "invalid id",
		})
		return
	}

	pid, err := strconv.Atoi(pidStr)

	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{
			ErrorCode:    http.StatusBadRequest,
			ErrorMessage: err.Error(),
		})
		return
	}

	var payload paginationPayload
	if err := c.Bind(&payload); err != nil {
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

func apiRunPrompt(c *gin.Context) {

	hashedValue, ok := c.Params.Get("id")

	if !ok {
		c.JSON(http.StatusBadRequest, errorResponse{
			ErrorCode:    http.StatusBadRequest,
			ErrorMessage: "invalid id",
		})
		return
	}

	promptID, err := hashidService.Decode(hashedValue)

	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse{
			ErrorCode:    http.StatusInternalServerError,
			ErrorMessage: err.Error(),
		})
		return
	}

	prompt, err := service.EntClient.Prompt.Get(c, promptID)
	if err != nil {
		c.JSON(http.StatusNotFound, errorResponse{
			ErrorCode:    http.StatusNotFound,
			ErrorMessage: err.Error(),
		})
		return
	}

	pj := prompt.Edges.Project

	logrus.Println("executing prompt:", prompt, pj)

	res, err := openAIService.Chat(c, pj, prompt.Prompts, prompt.Variables)

	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse{
			ErrorCode:    http.StatusInternalServerError,
			ErrorMessage: err.Error(),
		})
		return
	}

	// TODO: update response to a very simple one
	c.JSON(http.StatusOK, res)
}