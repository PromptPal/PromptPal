package routes

import "github.com/gin-gonic/gin"

func getRequestIP(c *gin.Context) string {
	header := c.Request.Header
	IPAddress := header.Get("X-Real-Ip")
	if IPAddress == "" {
		IPAddress = header.Get("X-Forwarded-For")
	}
	if IPAddress == "" {
		IPAddress = c.RemoteIP()
	}
	return IPAddress
}
