package routes

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/PromptPal/PromptPal/ent"
	"github.com/PromptPal/PromptPal/ent/webhook"
	"github.com/PromptPal/PromptPal/service"
	"github.com/sashabaranov/go-openai"
	"github.com/sirupsen/logrus"
)

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
	Cached bool `json:"cached"`
	IP     string `json:"ip"`
	UserAgent string `json:"userAgent"`
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
) {
	// Get all enabled webhooks for this project with onPromptFinished event
	webhooks, err := service.EntClient.Webhook.Query().
		Where(
			webhook.ProjectID(pj.ID),
			webhook.Event("onPromptFinished"),
			webhook.Enabled(true),
		).
		All(ctx)
	if err != nil {
		logrus.WithError(err).Error("Failed to query webhooks")
		return
	}

	if len(webhooks) == 0 {
		return
	}

	// Prepare webhook payload
	webhookPayload := WebhookPayload{
		Event:     "onPromptFinished",
		ProjectID: pj.ID,
		PromptID:  prompt.ID,
		UserID:    payload.UserId,
		Result:    responseResult,
		Timestamp: endTime.Format(time.RFC3339),
		Duration:  endTime.Sub(startTime).Milliseconds(),
		Cached:    isCachedResponse,
		IP:        clientIP,
		UserAgent: ua,
	}

	webhookPayload.Tokens.Prompt = res.Usage.PromptTokens
	webhookPayload.Tokens.Completion = res.Usage.CompletionTokens
	webhookPayload.Tokens.Total = res.Usage.TotalTokens

	payloadBytes, err := json.Marshal(webhookPayload)
	if err != nil {
		logrus.WithError(err).Error("Failed to marshal webhook payload")
		return
	}

	// Send webhook requests
	for _, webhook := range webhooks {
		go sendWebhookRequest(webhook, payloadBytes)
	}
}

// sendWebhookRequest sends a single webhook request
func sendWebhookRequest(webhook *ent.Webhook, payloadBytes []byte) {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	req, err := http.NewRequest("POST", webhook.URL, bytes.NewBuffer(payloadBytes))
	if err != nil {
		logrus.WithError(err).WithField("webhook_id", webhook.ID).Error("Failed to create webhook request")
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", fmt.Sprintf("PromptPal-Webhook/%s", versionCommit))

	resp, err := client.Do(req)
	if err != nil {
		logrus.WithError(err).WithField("webhook_id", webhook.ID).Error("Failed to send webhook request")
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		logrus.WithFields(logrus.Fields{
			"webhook_id": webhook.ID,
			"status_code": resp.StatusCode,
			"url": webhook.URL,
		}).Error("Webhook request failed")
		return
	}

	logrus.WithFields(logrus.Fields{
		"webhook_id": webhook.ID,
		"status_code": resp.StatusCode,
		"url": webhook.URL,
	}).Info("Webhook request sent successfully")
}