package routes

import (
	"net/http"
	"time"

	"github.com/PromptPal/PromptPal/service"
	brotli "github.com/anargu/gin-brotli"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

type errorResponse struct {
	ErrorCode    int    `json:"code"`
	ErrorMessage string `json:"error"`
}

var web3Service service.Web3Service
var openAIService service.OpenAIService
var hashidService service.HashIDService

func SetupGinRoutes(
	commitSha string,
	w3 service.Web3Service,
	o service.OpenAIService,
	hi service.HashIDService,
) *gin.Engine {
	web3Service = w3
	openAIService = o
	hashidService = hi

	h := gin.Default()

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

		adminRoutes.GET("/projects", listProjects)
		adminRoutes.GET("/projects/:id", getProject)
		adminRoutes.GET("/projects/:id/open-tokens", listOpenToken)
		adminRoutes.GET("/projects/:id/prompts", listProjectPrompts)

		// TODO: add this API
		// adminRoutes.GET("/projects/:id/calls", listProjectCalls)

		adminRoutes.POST("/projects", createProject)
		adminRoutes.POST("/projects/:id/open-tokens", createOpenToken)
		adminRoutes.PUT("/projects/:id", updateProject)

		adminRoutes.GET("/prompts", listPrompts)
		adminRoutes.GET("/prompts/:id", getPrompt)
		adminRoutes.GET("/prompts/:id/calls", getPromptCalls)
		adminRoutes.POST("/prompts", createPrompt)
		adminRoutes.POST("/prompts/test", testPrompt)
		adminRoutes.PUT("/prompts/:id", updatePrompt)

		adminRoutes.DELETE("/open-tokens", deleteOpenToken)
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
