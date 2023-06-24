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
	"github.com/PromptPal/PromptPal/service"
	"github.com/sirupsen/logrus"
)

var GitCommit string

func main() {
	service.InitDB()
	// updateOldImages()
	startHTTPServer()
}

func startHTTPServer() {
	// Send buffered spans and free resources.
	publicDomain := config.GetRuntimeConfig().PublicDomain

	w3 := service.NewWeb3Service()
	o := service.NewOpenAIService()
	hi := service.NewHashIDService()

	h := routes.SetupGinRoutes(GitCommit, w3, o, hi)
	server := &http.Server{
		Addr:    publicDomain,
		Handler: h,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logrus.Fatalf("listen: %s\n", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logrus.Println("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	service.Close()

	if err := server.Shutdown(ctx); err != nil {
		logrus.Fatal("Server forced to shutdown:", err)
	}
	logrus.Println("Server exiting")
}
