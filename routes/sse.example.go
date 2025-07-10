package routes

import (
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// SSEExample demonstrates a simple SSE endpoint with dummy data
func SSEExample(c *gin.Context) {
	// Set SSE headers
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")

	// Create a channel for sending events
	messageChan := make(chan string)

	// Start a goroutine to send dummy data
	go func() {
		defer close(messageChan)

		// Send initial message
		messageChan <- "Connected to SSE endpoint"

		// Wait 2 seconds
		time.Sleep(2 * time.Second)

		// Send some dummy data points
		for i := 1; i <= 5; i++ {
			messageChan <- fmt.Sprintf("Data point %d: Random value = %d", i, time.Now().Unix()%100)
			time.Sleep(2 * time.Second)
		}

		// Send final message after 10+ seconds total
		messageChan <- "SSE stream completed"
	}()

	// Stream events to client
	c.Stream(func(w io.Writer) bool {
		select {
		case msg, ok := <-messageChan:
			if !ok {
				// Channel closed, end stream
				return false
			}

			logrus.Println("running....", time.Now())
			// Format as SSE event
			c.SSEvent("message", msg)
			c.Writer.Flush()
			return true

		case <-c.Request.Context().Done():
			// Client disconnected
			return false
		}
	})
}

// SSEExample2 demonstrates a simple SSE endpoint using native http module
func SSEExample2(w http.ResponseWriter, r *http.Request) {
	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	// Create flusher
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	// Write initial connection message
	fmt.Fprintf(w, "data: Connected to SSE endpoint\n\n")
	flusher.Flush()

	// Send dummy data with delays
	for i := 1; i <= 5; i++ {
		select {
		case <-r.Context().Done():
			// Client disconnected
			logrus.Println("Client disconnected")
			return
		case <-time.After(2 * time.Second):
			// Send data point
			msg := fmt.Sprintf("Data point %d: Random value = %d", i, time.Now().Unix()%100)
			fmt.Fprintf(w, "data: %s\n\n", msg)
			flusher.Flush()
			logrus.Printf("Sent: %s", msg)
		}
	}

	// Send final message
	fmt.Fprintf(w, "data: SSE stream completed\n\n")
	flusher.Flush()
	logrus.Println("SSE stream completed")
}
