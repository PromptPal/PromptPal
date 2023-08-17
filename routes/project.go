package routes

import (
	"net/http"
	"strconv"
	"time"

	"github.com/PromptPal/PromptPal/ent"
	"github.com/PromptPal/PromptPal/ent/project"
	"github.com/PromptPal/PromptPal/ent/prompt"
	"github.com/PromptPal/PromptPal/ent/promptcall"
	"github.com/PromptPal/PromptPal/service"
	"github.com/gin-gonic/gin"
)

type getTopPromptsMetricOfProjectResponse struct {
	Prompt *ent.Prompt `json:"prompt"`
	Count  int         `json:"count"`
	// TODO: add more metrics later
}

func getTopPromptsMetricOfProject(c *gin.Context) {
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

	pc := make([]struct {
		PromptID int `json:"prompt_calls"`
		Count    int `json:"count"`
	}, 0)

	err = service.
		EntClient.
		PromptCall.
		Query().
		Where(promptcall.HasProjectWith(project.ID(pid))).
		Where(promptcall.CreateTimeGT(time.Now().Add(-24*7*time.Hour))).
		Limit(5).
		GroupBy("prompt_calls").
		Aggregate(ent.Count()).
		Scan(c, &pc)

	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse{
			ErrorCode:    http.StatusInternalServerError,
			ErrorMessage: err.Error(),
		})
		return
	}

	pidList := make([]int, 0)
	for _, p := range pc {
		pidList = append(pidList, p.PromptID)
	}

	prompts, err := service.EntClient.
		Prompt.
		Query().
		Where(prompt.IDIn(pidList...)).
		All(c)

	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse{
			ErrorCode:    http.StatusInternalServerError,
			ErrorMessage: err.Error(),
		})
		return
	}

	result := make([]getTopPromptsMetricOfProjectResponse, len(prompts))
	for i, p := range prompts {
		count := 0

		for _, pc := range pc {
			if pc.PromptID == p.ID {
				count = pc.Count
				break
			}
		}

		result[i] = getTopPromptsMetricOfProjectResponse{
			Prompt: p,
			Count:  count,
		}
	}

	c.JSON(http.StatusOK, ListResponse[getTopPromptsMetricOfProjectResponse]{
		Count: len(result),
		Data:  result,
	})
}
