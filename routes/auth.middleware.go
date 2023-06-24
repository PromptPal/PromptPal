package routes

import "github.com/gin-gonic/gin"

func authMiddleware(c *gin.Context) {
	// TODO
	// do jwt token check
	c.Next()
}

func apiMiddleware(c *gin.Context) {
	// TODO
	c.Next()
}
