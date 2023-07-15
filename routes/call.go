package routes

import (
	"net/http"
	"strconv"

	"github.com/PromptPal/PromptPal/ent"
	"github.com/PromptPal/PromptPal/ent/prompt"
	"github.com/PromptPal/PromptPal/ent/promptcall"
	"github.com/PromptPal/PromptPal/service"
	"github.com/gin-gonic/gin"
)

func getPromptCalls(c *gin.Context) {
	var payload queryPagination
	if err := c.BindQuery(&payload); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{
			ErrorCode:    http.StatusBadRequest,
			ErrorMessage: err.Error(),
		})
		return
	}

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
		c.JSON(http.StatusInternalServerError, errorResponse{
			ErrorCode:    http.StatusInternalServerError,
			ErrorMessage: err.Error(),
		})
		return
	}

	stat := service.
		EntClient.
		PromptCall.
		Query().
		Where(promptcall.HasPromptWith(prompt.ID(pid))).
		Order(ent.Desc(promptcall.FieldID))

	count, err := stat.Clone().Count(c)

	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse{
			ErrorCode:    http.StatusInternalServerError,
			ErrorMessage: err.Error(),
		})
		return
	}

	promptCalls, err := stat.
		Clone().
		Where(promptcall.IDLT(payload.Cursor)).
		Limit(payload.Limit).
		All(c)

	if err != nil {
		c.JSON(http.StatusNotFound, errorResponse{
			ErrorCode:    http.StatusNotFound,
			ErrorMessage: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, ListResponse[*ent.PromptCall]{
		Count: count,
		Data:  promptCalls,
	})
}
