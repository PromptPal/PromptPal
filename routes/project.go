package routes

import (
	"net/http"
	"strconv"

	"github.com/PromptPal/PromptPal/ent"
	"github.com/PromptPal/PromptPal/ent/project"
	"github.com/PromptPal/PromptPal/service"
	"github.com/gin-gonic/gin"
)

type listProjectsResponse struct {
	Projects   []*ent.Project `json:"projects"`
	Pagination Pagination     `json:"pagination"`
}

func listProjects(c *gin.Context) {
	pjs, err := service.
		EntClient.
		Project.
		Query().
		Order(ent.Desc(project.FieldID)).
		All(c)
	if err != nil {
		c.JSON(http.StatusNotFound, errorResponse{
			ErrorCode:    http.StatusNotFound,
			ErrorMessage: err.Error(),
		})
		return
	}

	cursor := 0

	if len(pjs) > 0 {
		cursor = pjs[len(pjs)-1].ID
	}

	c.JSON(http.StatusOK, listProjectsResponse{
		Projects: pjs,
		Pagination: Pagination{
			Count:  len(pjs),
			Cursor: cursor,
		},
	})
}

func getProject(c *gin.Context) {
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

	pj, err := service.EntClient.Project.Get(c, id)

	if err != nil {
		c.JSON(http.StatusNotFound, errorResponse{
			ErrorCode:    http.StatusNotFound,
			ErrorMessage: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, pj)
}

type createProjectPayload struct {
	Name        string `json:"name"`
	OpenAIToken string `json:"openaiToken"`
}

func createProject(c *gin.Context) {
	payload := createProjectPayload{}
	if err := c.Bind(&payload); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{
			ErrorCode:    http.StatusBadRequest,
			ErrorMessage: err.Error(),
		})
		return
	}

	pj, err := service.
		EntClient.
		Project.
		Create().
		SetName(payload.Name).
		SetOpenAIToken(payload.OpenAIToken).
		SetCreatorID(c.GetInt("uid")).
		Save(c)

	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse{
			ErrorCode:    http.StatusInternalServerError,
			ErrorMessage: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, pj)
}

// TODO: update project
func updateProject(c *gin.Context) {
}
