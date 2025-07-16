package schema

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/PromptPal/PromptPal/ent"
	"github.com/PromptPal/PromptPal/ent/webhookcall"
	"github.com/PromptPal/PromptPal/service"
)

// Webhook calls query arguments
type webhookCallsArgs struct {
	Input webhookCallsInput
}

type webhookCallsInput struct {
	WebhookID  int32
	Pagination paginationInput
}

// Webhook calls response
type webhookCallsResponse struct {
	stat       *ent.WebhookCallQuery
	pagination paginationInput
}

// Query resolver for webhook calls
func (q QueryResolver) WebhookCalls(ctx context.Context, args webhookCallsArgs) (webhookCallsResponse, error) {
	ctxValue := ctx.Value(service.GinGraphQLContextKey).(service.GinGraphQLContextType)

	// Get webhook to verify project ownership and permissions
	webhook, err := service.EntClient.Webhook.Get(ctx, int(args.Input.WebhookID))
	if err != nil {
		return webhookCallsResponse{}, NewGraphQLHttpError(http.StatusNotFound, err)
	}

	// Check RBAC permission for viewing webhook calls
	projectID := webhook.ProjectID
	hasPermission, err := rbacService.HasPermission(ctx, ctxValue.UserID, &projectID, service.PermProjectView)
	if err != nil {
		return webhookCallsResponse{}, NewGraphQLHttpError(http.StatusInternalServerError, err)
	}
	if !hasPermission {
		return webhookCallsResponse{}, NewGraphQLHttpError(http.StatusUnauthorized, errors.New("insufficient permissions to view webhook calls"))
	}

	stat := service.EntClient.WebhookCall.Query().
		Where(webhookcall.WebhookID(int(args.Input.WebhookID))).
		Order(ent.Desc(webhookcall.FieldID))

	return webhookCallsResponse{
		stat:       stat,
		pagination: args.Input.Pagination,
	}, nil
}

// Count method for webhook calls response
func (w webhookCallsResponse) Count(ctx context.Context) (int32, error) {
	count, err := w.stat.Clone().Count(ctx)
	if err != nil {
		return 0, NewGraphQLHttpError(http.StatusInternalServerError, err)
	}
	return int32(count), nil
}

// Edges method for webhook calls response
func (w webhookCallsResponse) Edges(ctx context.Context) (res []webhookCallResponse, err error) {
	calls, err := w.stat.Clone().
		Limit(int(w.pagination.Limit)).
		Offset(int(w.pagination.Offset)).
		All(ctx)

	if err != nil {
		return nil, NewGraphQLHttpError(http.StatusInternalServerError, err)
	}

	for _, call := range calls {
		res = append(res, webhookCallResponse{c: call})
	}
	return
}

// WebhookCall response type
type webhookCallResponse struct {
	c *ent.WebhookCall
}

func (w webhookCallResponse) ID() int32 {
	return int32(w.c.ID)
}

func (w webhookCallResponse) WebhookID() int32 {
	return int32(w.c.WebhookID)
}

func (w webhookCallResponse) TraceID() string {
	return w.c.TraceID
}

func (w webhookCallResponse) URL() string {
	return w.c.URL
}

func (w webhookCallResponse) RequestHeaders() *json.RawMessage {
	if w.c.RequestHeaders == nil {
		return nil
	}
	jsonBytes, err := json.Marshal(w.c.RequestHeaders)
	if err != nil {
		return nil
	}
	rawMsg := json.RawMessage(jsonBytes)
	return &rawMsg
}

func (w webhookCallResponse) RequestBody() string {
	return w.c.RequestBody
}

func (w webhookCallResponse) StatusCode() *int32 {
	if w.c.StatusCode == 0 {
		return nil
	}
	statusCode := int32(w.c.StatusCode)
	return &statusCode
}

func (w webhookCallResponse) ResponseHeaders() *json.RawMessage {
	if w.c.ResponseHeaders == nil {
		return nil
	}
	jsonBytes, err := json.Marshal(w.c.ResponseHeaders)
	if err != nil {
		return nil
	}
	rawMsg := json.RawMessage(jsonBytes)
	return &rawMsg
}

func (w webhookCallResponse) ResponseBody() *string {
	if w.c.ResponseBody == "" {
		return nil
	}
	return &w.c.ResponseBody
}

func (w webhookCallResponse) StartTime() string {
	return w.c.StartTime.Format(time.RFC3339)
}

func (w webhookCallResponse) EndTime() *string {
	if w.c.EndTime.IsZero() {
		return nil
	}
	endTime := w.c.EndTime.Format(time.RFC3339)
	return &endTime
}

func (w webhookCallResponse) IsTimeout() bool {
	return w.c.IsTimeout
}

func (w webhookCallResponse) IsSuccess() bool {
	statusCode := w.c.StatusCode
	return statusCode >= 200 && statusCode < 300
}

func (w webhookCallResponse) ErrorMessage() *string {
	if w.c.ErrorMessage == "" {
		return nil
	}
	return &w.c.ErrorMessage
}

func (w webhookCallResponse) UserAgent() *string {
	if w.c.UserAgent == "" {
		return nil
	}
	return &w.c.UserAgent
}

func (w webhookCallResponse) CreatedAt() string {
	return w.c.CreateTime.Format(time.RFC3339)
}

func (w webhookCallResponse) UpdatedAt() string {
	return w.c.UpdateTime.Format(time.RFC3339)
}

func (w webhookCallResponse) Webhook(ctx context.Context) (webhookResponse, error) {
	webhook, err := w.c.QueryWebhook().Only(ctx)
	if err != nil {
		return webhookResponse{}, NewGraphQLHttpError(http.StatusInternalServerError, err)
	}
	return webhookResponse{w: webhook}, nil
}