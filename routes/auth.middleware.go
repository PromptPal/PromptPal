package routes

import (
	"net/http"
	"strings"

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

func apiMiddleware(c *gin.Context) {
	// TODO
	c.Next()
}
