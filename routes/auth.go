package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func authHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{})
}
