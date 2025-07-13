package routes

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
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

// rbacService is the RBAC service instance
var rbacService *service.RBACService

// InitRBACMiddleware initializes the RBAC service for middleware use
func InitRBACMiddleware(client *ent.Client) {
	rbacService = service.NewRBACService(client)
}

// RequirePermission creates a middleware that checks if the current user has the required permission
func RequirePermission(permission string) gin.HandlerFunc {
	return RequireProjectPermission(permission, false)
}

// RequireProjectPermission creates a middleware that checks if the current user has the required permission
// If requireProject is true, the middleware expects a project ID in the context
func RequireProjectPermission(permission string, requireProject bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		// Get user ID from context
		userID, exists := c.Get("uid")
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse{
				ErrorCode:    http.StatusUnauthorized,
				ErrorMessage: "authentication required",
			})
			return
		}

		uid, ok := userID.(int)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse{
				ErrorCode:    http.StatusUnauthorized,
				ErrorMessage: "invalid user ID",
			})
			return
		}

		var projectID *int
		if requireProject {
			// Try to get project ID from various sources
			if pid, exists := c.Get("pid"); exists {
				if pidInt, ok := pid.(int); ok {
					projectID = &pidInt
				}
			}

			// If still no project ID and required, check URL parameter
			if projectID == nil {
				if pidStr := c.Param("projectId"); pidStr != "" {
					// Convert string to int
					if pidInt, err := strconv.Atoi(pidStr); err == nil {
						projectID = &pidInt
					}
				}
			}

			if projectID == nil {
				c.AbortWithStatusJSON(http.StatusBadRequest, errorResponse{
					ErrorCode:    http.StatusBadRequest,
					ErrorMessage: "project ID required",
				})
				return
			}
		}

		// Check permission
		if rbacService == nil {
			// Fallback to legacy permission check
			// This should be removed once RBAC is fully implemented
			c.Next()
			return
		}

		hasPermission, err := rbacService.HasPermission(ctx, uid, projectID, permission)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, errorResponse{
				ErrorCode:    http.StatusInternalServerError,
				ErrorMessage: fmt.Sprintf("permission check failed: %v", err),
			})
			return
		}

		if !hasPermission {
			c.AbortWithStatusJSON(http.StatusForbidden, errorResponse{
				ErrorCode:    http.StatusForbidden,
				ErrorMessage: fmt.Sprintf("insufficient permissions: %s required", permission),
			})
			return
		}

		c.Next()
	}
}

// RequireSystemAdmin is a convenience middleware for system admin permissions
func RequireSystemAdmin() gin.HandlerFunc {
	return RequirePermission(service.PermSystemAdmin)
}

// RequireProjectAdmin is a convenience middleware for project admin permissions
func RequireProjectAdmin() gin.HandlerFunc {
	return RequireProjectPermission(service.PermProjectAdmin, true)
}
