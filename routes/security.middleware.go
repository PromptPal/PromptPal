package routes

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/PromptPal/PromptPal/ent"
	"github.com/gin-gonic/gin"
)

type temporaryTokenValidationResponse struct {
	Limit     int    `json:"limit"`
	Remaining int    `json:"remaining"`
	UserId    string `json:"userId"`
}

func temporaryTokenValidationMiddleware(c *gin.Context) {
	otd, ok := c.Get("openToken")
	if !ok {
		c.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse{
			ErrorCode:    http.StatusUnauthorized,
			ErrorMessage: "No OpenToken found in context",
		})
		return
	}
	ot := otd.(ent.OpenToken)

	if !ot.ApiValidateEnabled {
		c.Next()
		return
	}

	tempToken := c.Request.Header.Get("X-TEMPORARY-TOKEN")

	if tempToken == "" {
		c.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse{
			ErrorCode:    http.StatusUnauthorized,
			ErrorMessage: "No temporary token found",
		})
		return
	}

	req, err := http.NewRequest("POST", ot.ApiValidatePath, nil)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, errorResponse{
			ErrorCode:    http.StatusInternalServerError,
			ErrorMessage: err.Error(),
		})
		return
	}
	reqUserId := c.Request.Header.Get("X-User-Id")

	req.Header.Add("Authorization", tempToken)
	req.Header.Add("X-User-Id", reqUserId)
	req.Header.Add("User-Agent", fmt.Sprintf("PromptPal@%s", versionCommit))
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, errorResponse{
			ErrorCode:    http.StatusInternalServerError,
			ErrorMessage: fmt.Sprintf("Failed to validate temporary token: %s", err.Error()),
		})
		return
	}

	defer res.Body.Close()
	buf, err := io.ReadAll(res.Body)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, errorResponse{
			ErrorCode:    http.StatusInternalServerError,
			ErrorMessage: fmt.Sprintf("Failed to read response body from validation server : %s", err.Error()),
		})
		return
	}

	var resp temporaryTokenValidationResponse
	if err := json.Unmarshal(buf, &resp); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, errorResponse{
			ErrorCode:    http.StatusInternalServerError,
			ErrorMessage: fmt.Sprintf("Failed to parse response body from validation server : %s", err.Error()),
		})
		return
	}

	c.Writer.Header().Set("X-TEMPORARY-TOKEN-VALIDATED", fmt.Sprintf("%t", resp.Remaining > 0))
	c.Writer.Header().Set("X-TEMPORARY-TOKEN-LIMIT", fmt.Sprintf("%d", resp.Limit))
	c.Writer.Header().Set("X-TEMPORARY-TOKEN-REMAINING", fmt.Sprintf("%d", resp.Remaining))

	if resp.Remaining <= 0 {
		c.AbortWithStatusJSON(http.StatusForbidden, errorResponse{
			ErrorCode:    http.StatusForbidden,
			ErrorMessage: fmt.Sprintf("Temporary token limit reached: %d", resp.Remaining),
		})
		return
	}
	if resp.UserId != "" {
		c.Set("server_uid", resp.UserId)
	}
	c.Next()
}
