package routes

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/PromptPal/PromptPal/ent"
	"github.com/PromptPal/PromptPal/ent/webhook"
	"github.com/PromptPal/PromptPal/service"
	"github.com/PromptPal/PromptPal/utils"
	"github.com/sashabaranov/go-openai"
	"github.com/sirupsen/logrus"
)

// Shared HTTP client for webhook requests to avoid creating new clients for each request
var webhookHTTPClient = &http.Client{
	Timeout: 10 * time.Second,
}

// WebhookPayload represents the data sent to webhook endpoints
type WebhookPayload struct {
	Event     string `json:"event"`
	ProjectID int    `json:"projectId"`
	PromptID  int    `json:"promptId"`
	UserID    string `json:"userId"`
	Result    int    `json:"result"` // 0 for success, 1 for failure
	Timestamp string `json:"timestamp"`
	Duration  int64  `json:"duration"`
	Tokens    struct {
		Prompt     int `json:"prompt"`
		Completion int `json:"completion"`
		Total      int `json:"total"`
	} `json:"tokens"`
	Cached               bool    `json:"cached"`
	IP                   string  `json:"ip"`
	UserAgent            string  `json:"userAgent"`
	ProviderID           *int    `json:"providerId,omitempty"`
	ProviderDefaultModel *string `json:"providerDefaultModel,omitempty"`
}

// WebhookCallData holds all the information needed to record a webhook call
type WebhookCallData struct {
	WebhookID       int
	TraceID         string
	URL             string
	RequestHeaders  map[string]string
	RequestBody     string
	StatusCode      int
	ResponseHeaders map[string]string
	ResponseBody    string
	StartTime       time.Time
	EndTime         time.Time
	IsTimeout       bool
	ErrorMessage    string
	UserAgent       string
	IP              string
	ProviderID      *int
}

// triggerWebhooks sends webhook notifications for onPromptFinished events
func triggerWebhooks(
	ctx context.Context,
	pj ent.Project,
	prompt ent.Prompt,
	responseResult int,
	res openai.ChatCompletionResponse,
	payload apiRunPromptPayload,
	endTime, startTime time.Time,
	ua string,
	clientIP string,
	isCachedResponse bool,
	providerID *int,
) {
	// Use background context to avoid cancellation when request completes
	backgroundCtx := context.Background()

	// Get all enabled webhooks for this project with onPromptFinished event
	webhooks, err := service.EntClient.Webhook.Query().
		Where(
			webhook.ProjectID(pj.ID),
			webhook.Event("onPromptFinished"),
			webhook.Enabled(true),
		).
		All(backgroundCtx)
	if err != nil {
		logrus.WithError(err).Error("Failed to query webhooks")
		return
	}

	if len(webhooks) == 0 {
		return
	}

	// Prepare webhook payload
	webhookPayload := WebhookPayload{
		Event:      "onPromptFinished",
		ProjectID:  pj.ID,
		PromptID:   prompt.ID,
		UserID:     payload.UserId,
		Result:     responseResult,
		Timestamp:  endTime.Format(time.RFC3339),
		Duration:   endTime.Sub(startTime).Milliseconds(),
		Cached:     isCachedResponse,
		IP:         clientIP,
		UserAgent:  ua,
		ProviderID: providerID,
	}

	// Fetch provider default model if provider ID is available
	if providerID != nil {
		provider, err := service.EntClient.Provider.Get(backgroundCtx, *providerID)
		if err != nil {
			logrus.WithError(err).WithField("provider_id", *providerID).Error("Failed to fetch provider for webhook payload")
		} else {
			if provider.DefaultModel != "" {
				webhookPayload.ProviderDefaultModel = &provider.DefaultModel
			}
		}
	}

	webhookPayload.Tokens.Prompt = res.Usage.PromptTokens
	webhookPayload.Tokens.Completion = res.Usage.CompletionTokens
	webhookPayload.Tokens.Total = res.Usage.TotalTokens

	payloadBytes, err := json.Marshal(webhookPayload)
	if err != nil {
		logrus.WithError(err).Error("Failed to marshal webhook payload")
		return
	}

	// Generate trace ID for this webhook trigger
	traceID := utils.RandStringRunes(16)

	// Send webhook requests
	for _, webhook := range webhooks {
		go sendWebhookRequest(backgroundCtx, webhook, payloadBytes, traceID, clientIP, providerID)
	}
}

// sendWebhookRequest sends a single webhook request and records the call details
func sendWebhookRequest(ctx context.Context, webhook *ent.Webhook, payloadBytes []byte, traceID string, clientIP string, providerID *int) {
	startTime := time.Now()

	// Prepare request headers
	requestHeaders := map[string]string{
		"Content-Type": "application/json",
		"User-Agent":   fmt.Sprintf("PromptPal@%s", versionCommit),
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", webhook.URL, bytes.NewBuffer(payloadBytes))
	if err != nil {
		logrus.WithError(err).WithField("webhook_id", webhook.ID).Error("Failed to create webhook request")
		recordWebhookCall(ctx, WebhookCallData{
			WebhookID:       webhook.ID,
			TraceID:         traceID,
			URL:             webhook.URL,
			RequestHeaders:  requestHeaders,
			RequestBody:     string(payloadBytes),
			StatusCode:      0,
			ResponseHeaders: nil,
			ResponseBody:    "",
			StartTime:       startTime,
			EndTime:         time.Now(),
			IsTimeout:       true,
			ErrorMessage:    err.Error(),
			UserAgent:       requestHeaders["User-Agent"],
			IP:              clientIP,
			ProviderID:      providerID,
		})
		return
	}

	// Set request headers
	for key, value := range requestHeaders {
		req.Header.Set(key, value)
	}

	// Make the HTTP request
	resp, err := webhookHTTPClient.Do(req)
	endTime := time.Now()

	var statusCode int
	var responseHeaders map[string]string
	var responseBody string
	var isTimeout bool
	var errorMessage string

	if err != nil {
		logrus.WithError(err).WithField("webhook_id", webhook.ID).Error("Failed to send webhook request")
		// Check if it's a timeout error
		isTimeout = isTimeoutError(err)
		errorMessage = err.Error()
	} else {
		defer resp.Body.Close()
		statusCode = resp.StatusCode

		// Read response headers
		responseHeaders = make(map[string]string)
		for key, values := range resp.Header {
			if len(values) > 0 {
				responseHeaders[key] = values[0]
			}
		}

		// Read response body
		bodyBytes, bodyErr := io.ReadAll(resp.Body)
		if bodyErr != nil {
			logrus.WithError(bodyErr).WithField("webhook_id", webhook.ID).Error("Failed to read webhook response body")
			errorMessage = fmt.Sprintf("Failed to read response body: %v", bodyErr)
		} else {
			responseBody = string(bodyBytes)
		}

		if !(statusCode >= 200 && statusCode < 300) {
			logrus.WithFields(logrus.Fields{
				"webhook_id":  webhook.ID,
				"status_code": resp.StatusCode,
				"url":         webhook.URL,
			}).Error("Webhook request failed")
			if errorMessage == "" {
				errorMessage = fmt.Sprintf("HTTP %d response", statusCode)
			}
		} else {
			logrus.WithFields(logrus.Fields{
				"webhook_id":  webhook.ID,
				"status_code": resp.StatusCode,
				"url":         webhook.URL,
			}).Info("Webhook request sent successfully")
		}
	}

	// Record the webhook call in database
	recordWebhookCall(ctx, WebhookCallData{
		WebhookID:       webhook.ID,
		TraceID:         traceID,
		URL:             webhook.URL,
		RequestHeaders:  requestHeaders,
		RequestBody:     string(payloadBytes),
		StatusCode:      statusCode,
		ResponseHeaders: responseHeaders,
		ResponseBody:    responseBody,
		StartTime:       startTime,
		EndTime:         endTime,
		IsTimeout:       isTimeout,
		ErrorMessage:    errorMessage,
		UserAgent:       requestHeaders["User-Agent"],
		IP:              clientIP,
		ProviderID:      providerID,
	})
}

// recordWebhookCall saves webhook call details to the database
func recordWebhookCall(ctx context.Context, data WebhookCallData) {
	call := service.EntClient.WebhookCall.Create().
		SetWebhookID(data.WebhookID).
		SetTraceID(data.TraceID).
		SetURL(data.URL).
		SetRequestHeaders(data.RequestHeaders).
		SetRequestBody(data.RequestBody).
		SetStartTime(data.StartTime).
		SetIsTimeout(data.IsTimeout)

	if data.StatusCode > 0 {
		call = call.SetStatusCode(data.StatusCode)
	}
	if data.ResponseHeaders != nil {
		call = call.SetResponseHeaders(data.ResponseHeaders)
	}
	if data.ResponseBody != "" {
		call = call.SetResponseBody(data.ResponseBody)
	}
	if !data.EndTime.IsZero() {
		call = call.SetEndTime(data.EndTime)
	}
	if data.ErrorMessage != "" {
		call = call.SetErrorMessage(data.ErrorMessage)
	}
	if data.UserAgent != "" {
		call = call.SetUserAgent(data.UserAgent)
	}
	if data.IP != "" {
		call = call.SetIP(data.IP)
	}
	if data.ProviderID != nil {
		call = call.SetProviderID(*data.ProviderID)
	}

	_, err := call.Save(ctx)
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"webhook_id": data.WebhookID,
			"trace_id":   data.TraceID,
		}).Error("Failed to record webhook call")
	}
}

// isTimeoutError checks if the error is a timeout error
func isTimeoutError(err error) bool {
	if err == nil {
		return false
	}
	// Check for common timeout error patterns
	errStr := err.Error()
	return bytes.Contains([]byte(errStr), []byte("timeout")) ||
		bytes.Contains([]byte(errStr), []byte("deadline exceeded")) ||
		bytes.Contains([]byte(errStr), []byte("context deadline exceeded"))
}
