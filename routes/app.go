package routes

import (
	"net/http"
	"time"

	"github.com/PromptPal/PromptPal/service"
	brotli "github.com/anargu/gin-brotli"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/graph-gophers/graphql-go"
)

type errorResponse struct {
	ErrorCode    int    `json:"code"`
	ErrorMessage string `json:"error"`
}

// var s
var web3Service service.Web3Service
var openAIService service.OpenAIService
var hashidService service.HashIDService

func SetupGinRoutes(
	commitSha string,
	w3 service.Web3Service,
	o service.OpenAIService,
	hi service.HashIDService,
	graphqlSchema *graphql.Schema,
) *gin.Engine {
	web3Service = w3
	openAIService = o
	hashidService = hi
	s = graphqlSchema

	h := gin.Default()

	// h.Use(brotli.Brotli(brotli.DefaultCompression))

	// with version
	h.Use(func(c *gin.Context) {
		c.Writer.Header().Add("X-PP-VER", commitSha)
		c.Next()
	})
	h.Use(brotli.Brotli(brotli.DefaultCompression))

	h.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:8080", "http://*.annatarhe.com", "http://*.annatarhe.cn"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "Content-Encoding", "Date", "X-RSA-Auth", "X-RSA-Nonce"},
		ExposeHeaders:    []string{"Content-Length", "Content-Encoding", "Date"},
		AllowCredentials: true,
		AllowOriginFunc: func(origin string) bool {
			return true
		},
		MaxAge: 12 * time.Hour,
	}))

	authRoutes := h.Group("/api/v1/auth")
	authRoutes.POST("/login", authHandler)

	adminRoutes := h.Group("/api/v1/admin")
	adminRoutes.Use(authMiddleware)
	{
		adminRoutes.GET("/users", listUsers)
		adminRoutes.GET("/users/:id", getUser)
		adminRoutes.POST("/users", createUsers)
		adminRoutes.DELETE("/users/:id", removeUsers)

		// TODO: add this API
		// adminRoutes.GET("/projects/:id/calls", listProjectCalls)
		adminRoutes.GET("/projects/:id/top-prompts", getTopPromptsMetricOfProject)
		adminRoutes.GET("/prompts/:id/calls", getPromptCalls)
	}

	if graphqlSchema != nil {
		if true {
			h.GET("/api/v2/graphql", graphqlPlaygroundHandler)
		}
		h.POST("/api/v2/graphql", authMiddleware, graphqlExecuteHandler)
	}

	h.LoadHTMLFiles("./public/index.html")
	h.Static("/public", "./public")
	h.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", nil)
	})
	h.NoRoute(func(c *gin.Context) {
		c.HTML(http.StatusNotFound, "index.html", nil)
	})

	apiRoutes := h.Group("/api/v1/public")
	apiRoutes.Use(apiMiddleware)
	{
		apiRoutes.GET("/prompts", apiListPrompts)
		apiRoutes.POST("/prompts/run/:id", apiRunPrompt)
	}

	return h
}
