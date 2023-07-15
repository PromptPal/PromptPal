package routes

import (
	"net/http"
	"strings"

	cache "github.com/Code-Hex/go-generics-cache"
	"github.com/PromptPal/PromptPal/ent/opentoken"
	"github.com/PromptPal/PromptPal/service"
	"github.com/gin-gonic/gin"
)

func authMiddleware(c *gin.Context) {
	authKey := strings.Split(c.GetHeader("Authorization"), " ")
	if len(authKey) != 2 || authKey[0] != "Bearer" {
		c.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse{
			ErrorCode:    http.StatusUnauthorized,
			ErrorMessage: "invalid token",
		})
		return
	}

	tk := authKey[1]

	claims, err := service.ParseJWT(tk)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse{
			ErrorCode:    http.StatusUnauthorized,
			ErrorMessage: err.Error(),
		})
		return
	}
	c.Set("uid", claims.Uid)
	c.Next()
}

var publicAPIAuthCache *cache.Cache[string, int] = cache.New[string, int]()

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

	tk := authKey[1]
	pid, ok := publicAPIAuthCache.Get(tk)
	if !ok {
		ot, err := service.
			EntClient.
			OpenToken.
			Query().
			Where(opentoken.Token(tk)).
			Only(c)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusForbidden, errorResponse{
				ErrorCode:    http.StatusForbidden,
				ErrorMessage: err.Error(),
			})
			return
		}

		pj, err := ot.QueryProject().Only(c)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, errorResponse{
				ErrorCode:    http.StatusInternalServerError,
				ErrorMessage: err.Error(),
			})
			return
		}
		pid = pj.ID
	}
	c.Set("pid", pid)
	c.Next()
}
