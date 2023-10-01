package routes

import (
	"context"
	_ "embed"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/PromptPal/PromptPal/service"
	"github.com/gin-gonic/gin"
	"github.com/graph-gophers/graphql-go"
)

var s *graphql.Schema

// var s = graphql.MustParseSchema(
// 	schema.String(),
// 	&schema.QueryResolver{},
// )

//go:embed graphql.html
var graphqliTemplate []byte

// GetByClipping will return comments that belongs to this clipping
func graphqlPlaygroundHandler(c *gin.Context) {
	c.Writer.Header().Set("content-type", "text/html")
	c.Writer.Write(graphqliTemplate)
}

type graphqlRequestPayload struct {
	Query         string                 `json:"query"`
	OperationName string                 `json:"operationName"`
	Variables     map[string]interface{} `json:"variables"`
}

func graphqlExecuteHandler(c *gin.Context) {
	uid := c.GetInt("uid")

	ctx := context.WithValue(
		c.Request.Context(),
		service.GinGraphQLContextKey,
		service.GinGraphQLContextType{
			UserID: uid,
		},
	)
	var params graphqlRequestPayload

	err := c.Bind(&params)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, errorResponse{
			ErrorCode:    http.StatusBadRequest,
			ErrorMessage: err.Error(),
		})
		return
	}

	// skip auth check if is calling auth api
	if !strings.Contains(params.Query, "auth") && uid == 0 {
		c.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse{
			ErrorCode:    http.StatusUnauthorized,
			ErrorMessage: "unauthorized",
		})
		return
	}

	response := s.Exec(ctx, params.Query, params.OperationName, params.Variables)

	// set error code from resolver error
	if len(response.Errors) > 0 {
		firstErr := response.Errors[0]
		code := firstErr.Extensions["code"]
		if code != nil {
			cd, ok := code.(int)
			if ok {
				c.Status(cd)
			}
		}
	}
	responseJSON, err := json.Marshal(response)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, errorResponse{
			ErrorCode:    http.StatusInternalServerError,
			ErrorMessage: err.Error(),
		})
		return
	}

	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.Write(responseJSON)
}
