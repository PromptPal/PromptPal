package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type authPayload struct {
	Message string `json:"message"`
}

func authHandler(c *gin.Context) {
	payload := authPayload{}
	if err := c.Bind(&payload); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{
			ErrorCode:    http.StatusBadRequest,
			ErrorMessage: err.Error(),
		})
		return
	}

	// do web3 check

	// sign web3 token to client

	c.JSON(http.StatusOK, gin.H{})
}

func listUsers(c *gin.Context) {
	// check signed data
}

func createUsers(c *gin.Context) {
	// check signed data
}

func removeUsers(c *gin.Context) {
	// check signed data
}
