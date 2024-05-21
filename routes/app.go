package routes

import (
	"context"
	"net/http"
	"time"

	"github.com/PromptPal/PromptPal/config"
	"github.com/PromptPal/PromptPal/service"
	brotli "github.com/anargu/gin-brotli"
	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/graph-gophers/graphql-go"
	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
)

type errorResponse struct {
	ErrorCode    int    `json:"code"`
	ErrorMessage string `json:"error"`
}

// var s
var web3Service service.Web3Service
var aiService service.OpenAIService
var hashidService service.HashIDService

var oidcProvider *oidc.Provider
var ssoGoogle *oauth2.Config

func SetupGinRoutes(
	commitSha string,
	w3 service.Web3Service,
	o service.OpenAIService,
	hi service.HashIDService,
	graphqlSchema *graphql.Schema,
) *gin.Engine {
	web3Service = w3
	aiService = o
	hashidService = hi

	rc := config.GetRuntimeConfig()
	if rc.SSOGoogleCallbackURL != "" && rc.SSOGoogleClientID != "" && rc.SSOGoogleClientSecret != "" {
		provider, err := oidc.NewProvider(context.Background(), "https://accounts.google.com")
		if err != nil {
			logrus.Panicln(err)
		}
		oidcProvider = provider
		ssoGoogle = &oauth2.Config{
			ClientID:     rc.SSOGoogleClientID,
			ClientSecret: rc.SSOGoogleClientSecret,
			RedirectURL:  rc.SSOGoogleCallbackURL,
			Endpoint:     provider.Endpoint(),
			Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
		}
	}
	s = graphqlSchema

	h := gin.Default()

	store := cookie.NewStore(rc.JwtTokenKey)
	h.Use(sessions.Sessions("pp-sess", store))
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

	h.POST("/api/v1/admin/prompts/test", authMiddleware, testPrompt)

	apiRoutes := h.Group("/api/v1/public")
	apiRoutes.Use(apiMiddleware)
	{
		apiRoutes.GET("/prompts", apiListPrompts)
		apiRoutes.POST("/prompts/run/:id", apiRunPromptMiddleware, apiRunPrompt)
		// FIXME: PLEASE DO NOT USE IT!!!
		apiRoutes.POST("/prompts/run/:id/stream", apiRunPromptMiddleware, apiRunPromptStream)
	}

	// !!! IMPORTANT !!!
	// this feature should only available for enterprise
	sso := h.Group("/api/v1/sso")
	sso.GET("/settings", authSettings)
	sso.Use(ssoProviderCheck)
	{
		sso.GET("/login/:provider", ssoLogin)
		sso.GET("/callback/:provider", ssoCallback)
	}

	return h
}
