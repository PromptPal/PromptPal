package routes

import (
	"net/http"
	"strconv"
	"time"

	cache "github.com/Code-Hex/go-generics-cache"
	"github.com/PromptPal/PromptPal/ent"
	"github.com/PromptPal/PromptPal/ent/project"
	"github.com/PromptPal/PromptPal/service"
	"github.com/gin-gonic/gin"
)

func listProjects(c *gin.Context) {
	var query queryPagination
	if err := c.BindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{
			ErrorCode:    http.StatusBadRequest,
			ErrorMessage: err.Error(),
		})
		return
	}

	pjs, err := service.
		EntClient.
		Project.
		Query().
		Where(project.IDLT(query.Cursor)).
		Limit(query.Limit).
		Order(ent.Desc(project.FieldID)).
		All(c)
	if err != nil {
		c.JSON(http.StatusNotFound, errorResponse{
			ErrorCode:    http.StatusNotFound,
			ErrorMessage: err.Error(),
		})
		return
	}

	count, err := service.EntClient.Project.Query().Count(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse{
			ErrorCode:    http.StatusInternalServerError,
			ErrorMessage: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, ListResponse[*ent.Project]{
		Count: count,
		Data:  pjs,
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

type updateProjectPayload struct {
	Enabled           *bool    `json:"enabled"`
	OpenAIBaseURL     *string  `json:"openAIBaseURL"`
	OpenAIModel       *string  `json:"openAIModel"`
	OpenAIToken       *string  `json:"openAIToken"`
	OpenAITemperature *float64 `json:"openAITemperature"`
	OpenAITopP        *float64 `json:"openAITopP"`
	OpenAIMaxTokens   *int     `json:"openAIMaxTokens"`
}

func updateProject(c *gin.Context) {
	var payload updateProjectPayload
	if err := c.Bind(&payload); err != nil {
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
		c.JSON(http.StatusBadRequest, errorResponse{
			ErrorCode:    http.StatusBadRequest,
			ErrorMessage: err.Error(),
		})
		return
	}

	// TODO: check permission

	updater := service.EntClient.Project.UpdateOneID(pid)

	if payload.Enabled != nil {
		updater = updater.SetEnabled(*payload.Enabled)
	}
	if payload.OpenAIBaseURL != nil {
		updater = updater.SetOpenAIBaseURL(*payload.OpenAIBaseURL)
	}
	if payload.OpenAIModel != nil {
		updater = updater.SetOpenAIModel(*payload.OpenAIModel)
	}
	if payload.OpenAIToken != nil {
		updater = updater.SetOpenAIToken(*payload.OpenAIToken)
	}
	if payload.OpenAITemperature != nil {
		updater = updater.SetOpenAITemperature(*payload.OpenAITemperature)
	}
	if payload.OpenAITopP != nil {
		updater = updater.SetOpenAITopP(*payload.OpenAITopP)
	}
	if payload.OpenAIMaxTokens != nil {
		updater = updater.SetOpenAIMaxTokens(*payload.OpenAIMaxTokens)
	}

	pj, err := updater.Save(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse{
			ErrorCode:    http.StatusInternalServerError,
			ErrorMessage: err.Error(),
		})
		return
	}

	projectCache.Set(pj.ID, *pj, cache.WithExpiration(time.Hour*24))
	c.JSON(http.StatusOK, pj)
}
