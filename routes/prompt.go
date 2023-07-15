package routes

import (
	"net/http"
	"strconv"
	"time"

	cache "github.com/Code-Hex/go-generics-cache"
	"github.com/PromptPal/PromptPal/ent"
	"github.com/PromptPal/PromptPal/ent/project"
	"github.com/PromptPal/PromptPal/ent/prompt"
	"github.com/PromptPal/PromptPal/ent/schema"
	"github.com/PromptPal/PromptPal/service"
	"github.com/gin-gonic/gin"
)

type internalPromptItem struct {
	ent.Prompt
	HashID string `json:"hid"`
}

func listProjectPrompts(c *gin.Context) {
	var payload queryPagination
	if err := c.BindQuery(&payload); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{
			ErrorCode:    http.StatusBadRequest,
			ErrorMessage: err.Error(),
		})
		return
	}

	idStr, ok := c.Params.Get("id")
	if !ok {
		c.JSON(http.StatusBadRequest, errorResponse{
			ErrorCode:    http.StatusBadRequest,
			ErrorMessage: "invalid project id",
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
		Where(prompt.IDLT(payload.Cursor)).
		Limit(payload.Limit).
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

	result := make([]internalPromptItem, len(prompts))

	for i, p := range prompts {
		hid, err := hashidService.Encode(p.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, errorResponse{
				ErrorCode:    http.StatusInternalServerError,
				ErrorMessage: err.Error(),
			})
			return
		}
		result[i] = internalPromptItem{
			Prompt: *p,
			HashID: hid,
		}
	}

	c.JSON(http.StatusOK, ListResponse[internalPromptItem]{
		Count: count,
		Data:  result,
	})
}

func listPrompts(c *gin.Context) {
	// TODO: only admin can list all prompts across projects

	var payload queryPagination
	if err := c.BindQuery(&payload); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{
			ErrorCode:    http.StatusBadRequest,
			ErrorMessage: err.Error(),
		})
		return
	}

	prompts, err := service.
		EntClient.
		Prompt.
		Query().
		Where(prompt.IDLT(payload.Cursor)).
		Limit(payload.Limit).
		Order(ent.Desc(prompt.FieldID)).
		All(c)

	if err != nil {
		c.JSON(http.StatusNotFound, errorResponse{
			ErrorCode:    http.StatusNotFound,
			ErrorMessage: err.Error(),
		})
		return
	}

	count, err := service.EntClient.Prompt.Query().Count(c)

	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse{
			ErrorCode:    http.StatusInternalServerError,
			ErrorMessage: err.Error(),
		})
		return
	}

	result := make([]internalPromptItem, len(prompts))

	for i, p := range prompts {
		hid, err := hashidService.Encode(p.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, errorResponse{
				ErrorCode:    http.StatusInternalServerError,
				ErrorMessage: err.Error(),
			})
			return
		}
		result[i] = internalPromptItem{
			Prompt: *p,
			HashID: hid,
		}
	}

	c.JSON(http.StatusOK, ListResponse[internalPromptItem]{
		Count: count,
		Data:  result,
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

	hid, err := hashidService.Encode(prompt.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse{
			ErrorCode:    http.StatusInternalServerError,
			ErrorMessage: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, internalPromptItem{
		Prompt: *prompt,
		HashID: hid,
	})
}

type createPromptPayload struct {
	ProjectID   int                     `json:"projectId"`
	Name        string                  `json:"name"`
	Description string                  `json:"description"`
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
		SetName(payload.Name).
		SetDescription(payload.Description).
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

	hid, err := hashidService.Encode(p.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse{
			ErrorCode:    http.StatusInternalServerError,
			ErrorMessage: err.Error(),
		})
		return
	}
	// set cache
	apiPromptCache.Set(hid, *p, cache.WithExpiration(time.Hour*24))
	c.JSON(http.StatusOK, internalPromptItem{
		Prompt: *p,
		HashID: hid,
	})
}

type updatePromptPayload struct {
	Debug       bool                    `json:"debug"`
	Description string                  `json:"description"`
	TokenCount  int                     `json:"tokenCount"`
	Prompts     []schema.PromptRow      `json:"prompts"`
	Variables   []schema.PromptVariable `json:"variables"`
	PublicLevel prompt.PublicLevel      `json:"publicLevel"`
}

func updatePrompt(c *gin.Context) {
	pidStr, ok := c.Params.Get("id")
	if !ok {
		c.JSON(http.StatusBadRequest, errorResponse{
			ErrorCode:    http.StatusBadRequest,
			ErrorMessage: "invalid prompt id",
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

	var payload updatePromptPayload
	if err := c.Bind(&payload); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{
			ErrorCode:    http.StatusBadRequest,
			ErrorMessage: err.Error(),
		})
		return
	}

	tx, err := service.EntClient.Tx(c)

	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse{
			ErrorCode:    http.StatusInternalServerError,
			ErrorMessage: err.Error(),
		})
		return
	}

	oldPrompt, err := tx.Prompt.Get(c, pid)
	if err != nil {
		tx.Rollback()
		c.JSON(http.StatusNotFound, errorResponse{
			ErrorCode:    http.StatusNotFound,
			ErrorMessage: err.Error(),
		})
		return
	}

	snapshotData := schema.PromptComplete{
		Name:        oldPrompt.Name,
		Enabled:     oldPrompt.Enabled,
		Debug:       oldPrompt.Debug,
		Description: oldPrompt.Description,
		TokenCount:  oldPrompt.TokenCount,
		Prompts:     oldPrompt.Prompts,
		Variables:   oldPrompt.Variables,
		PublicLevel: oldPrompt.PublicLevel.String(),
	}

	err = tx.History.
		Create().
		AddModifierIDs(c.GetInt("uid")).
		SetPromptID(pid).
		SetSnapshot(snapshotData).
		Exec(c)

	if err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, errorResponse{
			ErrorCode:    http.StatusInternalServerError,
			ErrorMessage: err.Error(),
		})
		return
	}

	updatedPrompt, err := tx.Prompt.
		UpdateOneID(pid).
		SetDescription(payload.Description).
		SetTokenCount(payload.TokenCount).
		SetPrompts(payload.Prompts).
		SetDebug(payload.Debug).
		SetPublicLevel(payload.PublicLevel).
		Save(c)

	if err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, errorResponse{
			ErrorCode:    http.StatusInternalServerError,
			ErrorMessage: err.Error(),
		})
		return
	}

	if err := tx.Commit(); err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, errorResponse{
			ErrorCode:    http.StatusInternalServerError,
			ErrorMessage: err.Error(),
		})
		return
	}

	hid, err := hashidService.Encode(pid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse{
			ErrorCode:    http.StatusInternalServerError,
			ErrorMessage: err.Error(),
		})
		return
	}

	// refresh cache
	apiPromptCache.Set(hid, *updatedPrompt, cache.WithExpiration(time.Hour*24))

	c.JSON(http.StatusOK, internalPromptItem{
		Prompt: *updatedPrompt,
		HashID: hid,
	})
}

type testPromptPayload struct {
	ProjectID int                `json:"projectId"`
	Name      string             `json:"name"`
	Prompts   []schema.PromptRow `json:"prompts"`
	Variables map[string]string  `json:"variables"`
}

func testPrompt(c *gin.Context) {
	var payload testPromptPayload
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

	res, err := openAIService.Chat(c, *pj, payload.Prompts, payload.Variables, "")

	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse{
			ErrorCode:    http.StatusInternalServerError,
			ErrorMessage: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, res)
}
