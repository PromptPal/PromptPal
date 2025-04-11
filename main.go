package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/PromptPal/PromptPal/config"
	"github.com/PromptPal/PromptPal/routes"
	"github.com/PromptPal/PromptPal/schema"
	"github.com/PromptPal/PromptPal/service"
	"github.com/graph-gophers/graphql-go"
	"github.com/sirupsen/logrus"
)

var GitCommit string

func main() {
	config.SetupConfig(false)
	service.InitDB()
	startHTTPServer()
}

func startHTTPServer() {
	publicDomain := config.GetRuntimeConfig().PublicDomain
	w3 := service.NewWeb3Service()
	iai := service.NewIsomorphicAIService()
	hi := service.NewHashIDService()
	if err := service.InitRedis(config.GetRuntimeConfig().RedisURL); err != nil {
		logrus.Panicln("Failed to connect to Redis: ", err)
	}
	var graphqlSchema = graphql.MustParseSchema(
		schema.String(),
		&schema.QueryResolver{},
	)

	schema.Setup(hi, w3)
	h := routes.SetupGinRoutes(GitCommit, w3, iai, hi, graphqlSchema)
	server := &http.Server{
		Addr:    publicDomain,
		Handler: h,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logrus.Fatalf("listen: %s\n", err)
		}
	}()

	logrus.Infoln("PromptPal Server is running on: ", publicDomain)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logrus.Println("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	service.Close()

	if err := server.Shutdown(ctx); err != nil {
		logrus.Fatalln("Server forced to shutdown:", err)
	}
	logrus.Infoln("PromptPal Server exiting")
}
