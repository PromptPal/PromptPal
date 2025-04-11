package routes

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/PromptPal/PromptPal/ent"
	"github.com/PromptPal/PromptPal/ent/opentoken"
	"github.com/PromptPal/PromptPal/service"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/cache/v9"
)

func authMiddleware(c *gin.Context) {
	authKey := strings.Split(c.GetHeader("Authorization"), " ")
	if len(authKey) != 2 || authKey[0] != "Bearer" {
		// c.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse{
		// 	ErrorCode:    http.StatusUnauthorized,
		// 	ErrorMessage: "invalid token",
		// })
		// return
		c.Next()
		return
	}

	tk := authKey[1]

	claims, err := service.ParseJWT(tk)
	if err != nil {
		// c.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse{
		// 	ErrorCode:    http.StatusUnauthorized,
		// 	ErrorMessage: err.Error(),
		// })
		c.Next()
		return
	}
	c.Set("uid", claims.Uid)
}

// the header must be like this: `Authorization: API <token>`
func apiMiddleware(c *gin.Context) {
	authKey := strings.Split(c.GetHeader("Authorization"), " ")
	if len(authKey) != 2 || authKey[0] != "API" {
		c.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse{
			ErrorCode:    http.StatusUnauthorized,
			ErrorMessage: "the API token is required.",
		})
		return
	}

	ctx := c.Request.Context()

	tk := authKey[1]
	pid := 0
	var ot ent.OpenToken
	err := service.Cache.Get(ctx, fmt.Sprintf("openToken:%s", tk), &ot)
	if err != nil {
		if !errors.Is(err, cache.ErrCacheMiss) {
			c.AbortWithStatusJSON(http.StatusInternalServerError, errorResponse{
				ErrorCode:    http.StatusInternalServerError,
				ErrorMessage: err.Error(),
			})
			return
		}
		err = nil
		dot, err := service.
			EntClient.
			OpenToken.
			Query().
			Where(opentoken.Token(tk)).
			Only(ctx)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusForbidden, errorResponse{
				ErrorCode:    http.StatusForbidden,
				ErrorMessage: err.Error(),
			})
			return
		}

		ot = *dot

		pj, err := dot.QueryProject().Only(ctx)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, errorResponse{
				ErrorCode:    http.StatusInternalServerError,
				ErrorMessage: err.Error(),
			})
			return
		}
		pid = pj.ID
		service.Cache.Set(&cache.Item{
			Ctx:   ctx,
			Key:   fmt.Sprintf("openToken:%s", tk),
			Value: ot,
			TTL:   time.Hour,
		})
	}

	if pid == 0 {
		pid = ot.ProjectOpenTokens
	}

	c.Set("openToken", ot)
	c.Set("pid", pid)
	c.Next()
}
