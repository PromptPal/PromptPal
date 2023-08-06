package routes

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	cache "github.com/Code-Hex/go-generics-cache"
	"github.com/PromptPal/PromptPal/ent"
	"github.com/PromptPal/PromptPal/ent/opentoken"
	"github.com/PromptPal/PromptPal/ent/project"
	"github.com/PromptPal/PromptPal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func listOpenToken(c *gin.Context) {
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

	var query queryPagination
	if err := c.BindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{
			ErrorCode:    http.StatusBadRequest,
			ErrorMessage: err.Error(),
		})
		return
	}

	openTokenList, err := service.EntClient.
		OpenToken.
		Query().
		Where(opentoken.HasProjectWith(project.ID(pid))).
		Limit(query.Limit).
		Order(ent.Desc(opentoken.FieldID)).
		All(c)

	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse{
			ErrorCode:    http.StatusInternalServerError,
			ErrorMessage: err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, ListResponse[*ent.OpenToken]{
		Count: len(openTokenList),
		Data:  openTokenList,
	})
}

type createOpenTokenRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	// in seconds
	TTL int `json:"ttl"`
}

func createOpenToken(c *gin.Context) {
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

	var payload createOpenTokenRequest
	if err := c.Bind(&payload); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{
			ErrorCode:    http.StatusBadRequest,
			ErrorMessage: err.Error(),
		})
		return
	}

	// TODO: put int tx
	previousCount, err := service.
		EntClient.
		OpenToken.
		Query().
		Where(opentoken.HasProjectWith(project.ID(pid))).
		Count(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse{
			ErrorCode:    http.StatusInternalServerError,
			ErrorMessage: err.Error(),
		})
		return
	}

	if previousCount > 20 {
		c.JSON(http.StatusTooManyRequests, errorResponse{
			ErrorCode:    http.StatusTooManyRequests,
			ErrorMessage: "too many tokens, please remove old tokens first. allow up to 20 tokens",
		})
		return
	}

	tk := strings.Replace(uuid.New().String(), "-", "", -1)
	expireAt := time.Now().Add(time.Second * time.Duration(payload.TTL))

	err = service.
		EntClient.
		OpenToken.
		Create().
		SetName(payload.Name).
		SetDescription(payload.Description).
		SetToken(tk).
		SetUserID(c.GetInt("uid")).
		SetProjectID(pid).
		SetExpireAt(expireAt).
		Exec(c)

	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse{
			ErrorCode:    http.StatusInternalServerError,
			ErrorMessage: err.Error(),
		})
		return
	}

	service.PublicAPIAuthCache.Set(tk, pid, cache.WithExpiration(24*time.Hour))

	c.JSON(http.StatusOK, gin.H{
		"token": tk,
	})
}

func deleteOpenToken(c *gin.Context) {
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
		c.JSON(http.StatusBadRequest, errorResponse{
			ErrorCode:    http.StatusBadRequest,
			ErrorMessage: err.Error(),
		})
		return
	}

	// TODO:
	// check permission
	tx, err := service.EntClient.Tx(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse{
			ErrorCode:    http.StatusInternalServerError,
			ErrorMessage: err.Error(),
		})
		return
	}
	ot, err := tx.OpenToken.Get(c, id)
	if err != nil {
		tx.Rollback()
		c.JSON(http.StatusNotFound, errorResponse{
			ErrorCode:    http.StatusNotFound,
			ErrorMessage: err.Error(),
		})
		return
	}

	err = tx.OpenToken.DeleteOne(ot).Exec(c)
	if err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, errorResponse{
			ErrorCode:    http.StatusInternalServerError,
			ErrorMessage: err.Error(),
		})
		return
	}

	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse{
			ErrorCode:    http.StatusInternalServerError,
			ErrorMessage: err.Error(),
		})
		return
	}

	service.PublicAPIAuthCache.Delete(ot.Token)
	c.JSON(http.StatusOK, gin.H{})
}
