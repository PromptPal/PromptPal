package schema

import (
	"context"
	"errors"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/PromptPal/PromptPal/ent"
	"github.com/PromptPal/PromptPal/ent/webhook"
	"github.com/PromptPal/PromptPal/service"
)

// Event constants
const (
	EventOnPromptFinished = "onPromptFinished"
)

// validateWebhookURL validates webhook URL and prevents SSRF attacks
func validateWebhookURL(urlStr string) error {
	if urlStr == "" {
		return errors.New("URL cannot be empty")
	}

	parsed, err := url.Parse(urlStr)
	if err != nil {
		return errors.New("invalid URL format")
	}

	// Only allow HTTP and HTTPS schemes
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return errors.New("only HTTP and HTTPS schemes are allowed")
	}

	// Check for localhost, private IPs, and loopback addresses
	if parsed.Hostname() != "" {
		if parsed.Hostname() == "localhost" {
			return errors.New("localhost URLs are not allowed")
		}

		if ip := net.ParseIP(parsed.Hostname()); ip != nil {
			if ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() {
				return errors.New("private, loopback, and link-local IPs are not allowed")
			}
		}
	}

	return nil
}

type createWebhookData struct {
	Name        string
	Description *string
	URL         string
	Event       string
	Enabled     *bool
	ProjectID   int32
}

type createWebhookArgs struct {
	Data createWebhookData
}

func (q QueryResolver) CreateWebhook(ctx context.Context, args createWebhookArgs) (webhookResponse, error) {
	data := args.Data
	ctxValue := ctx.Value(service.GinGraphQLContextKey).(service.GinGraphQLContextType)

	// Check RBAC permission for webhook creation
	projectID := int(data.ProjectID)
	hasPermission, err := rbacService.HasPermission(ctx, ctxValue.UserID, &projectID, service.PermProjectEdit)
	if err != nil {
		return webhookResponse{}, NewGraphQLHttpError(http.StatusInternalServerError, err)
	}
	if !hasPermission {
		return webhookResponse{}, NewGraphQLHttpError(http.StatusUnauthorized, errors.New("insufficient permissions to create webhook"))
	}

	// Validate event type
	if data.Event != EventOnPromptFinished {
		return webhookResponse{}, NewGraphQLHttpError(http.StatusBadRequest, errors.New("only onPromptFinished event is supported"))
	}

	// Validate URL
	if err := validateWebhookURL(data.URL); err != nil {
		return webhookResponse{}, NewGraphQLHttpError(http.StatusBadRequest, err)
	}

	stat := service.EntClient.Webhook.Create().
		SetName(data.Name).
		SetURL(data.URL).
		SetEvent(data.Event).
		SetCreatorID(ctxValue.UserID).
		SetProjectID(int(data.ProjectID))

	if data.Description != nil {
		stat = stat.SetDescription(*data.Description)
	}
	if data.Enabled != nil {
		stat = stat.SetEnabled(*data.Enabled)
	}

	webhook, err := stat.Save(ctx)
	if err != nil {
		return webhookResponse{}, NewGraphQLHttpError(http.StatusInternalServerError, err)
	}

	return webhookResponse{w: webhook}, nil
}

type updateWebhookData struct {
	Name        *string
	Description *string
	URL         *string
	Event       *string
	Enabled     *bool
}

type updateWebhookArgs struct {
	ID   int32
	Data updateWebhookData
}

func (q QueryResolver) UpdateWebhook(ctx context.Context, args updateWebhookArgs) (webhookResponse, error) {
	ctxValue := ctx.Value(service.GinGraphQLContextKey).(service.GinGraphQLContextType)

	// Get webhook to check project ownership
	webhook, err := service.EntClient.Webhook.Get(ctx, int(args.ID))
	if err != nil {
		return webhookResponse{}, NewGraphQLHttpError(http.StatusNotFound, err)
	}

	// Check RBAC permission for webhook update
	projectID := webhook.ProjectID
	hasPermission, err := rbacService.HasPermission(ctx, ctxValue.UserID, &projectID, service.PermProjectEdit)
	if err != nil {
		return webhookResponse{}, NewGraphQLHttpError(http.StatusInternalServerError, err)
	}
	if !hasPermission {
		return webhookResponse{}, NewGraphQLHttpError(http.StatusUnauthorized, errors.New("insufficient permissions to update webhook"))
	}

	// Validate event type if provided
	if args.Data.Event != nil && *args.Data.Event != EventOnPromptFinished {
		return webhookResponse{}, NewGraphQLHttpError(http.StatusBadRequest, errors.New("only onPromptFinished event is supported"))
	}

	// Validate URL if provided
	if args.Data.URL != nil {
		if err := validateWebhookURL(*args.Data.URL); err != nil {
			return webhookResponse{}, NewGraphQLHttpError(http.StatusBadRequest, err)
		}
	}

	updater := service.EntClient.Webhook.UpdateOneID(int(args.ID))

	if args.Data.Name != nil {
		updater = updater.SetName(*args.Data.Name)
	}
	if args.Data.Description != nil {
		updater = updater.SetDescription(*args.Data.Description)
	}
	if args.Data.URL != nil {
		updater = updater.SetURL(*args.Data.URL)
	}
	if args.Data.Event != nil {
		updater = updater.SetEvent(*args.Data.Event)
	}
	if args.Data.Enabled != nil {
		updater = updater.SetEnabled(*args.Data.Enabled)
	}

	webhook, err = updater.Save(ctx)
	if err != nil {
		return webhookResponse{}, NewGraphQLHttpError(http.StatusInternalServerError, err)
	}

	return webhookResponse{w: webhook}, nil
}

type deleteWebhookArgs struct {
	ID int32
}

func (q QueryResolver) DeleteWebhook(ctx context.Context, args deleteWebhookArgs) (bool, error) {
	ctxValue := ctx.Value(service.GinGraphQLContextKey).(service.GinGraphQLContextType)

	// Get webhook to check project ownership
	webhook, err := service.EntClient.Webhook.Get(ctx, int(args.ID))
	if err != nil {
		return false, NewGraphQLHttpError(http.StatusNotFound, err)
	}

	// Check RBAC permission for webhook deletion
	projectID := webhook.ProjectID
	hasPermission, err := rbacService.HasPermission(ctx, ctxValue.UserID, &projectID, service.PermProjectEdit)
	if err != nil {
		return false, NewGraphQLHttpError(http.StatusInternalServerError, err)
	}
	if !hasPermission {
		return false, NewGraphQLHttpError(http.StatusUnauthorized, errors.New("insufficient permissions to delete webhook"))
	}

	err = service.EntClient.Webhook.DeleteOneID(int(args.ID)).Exec(ctx)
	if err != nil {
		return false, NewGraphQLHttpError(http.StatusInternalServerError, err)
	}

	return true, nil
}

type webhooksArgs struct {
	ProjectID  int32
	Pagination paginationInput
}

type webhooksResponse struct {
	stat       *ent.WebhookQuery
	pagination paginationInput
}

func (q QueryResolver) Webhooks(ctx context.Context, args webhooksArgs) (webhooksResponse, error) {
	ctxValue := ctx.Value(service.GinGraphQLContextKey).(service.GinGraphQLContextType)

	// Check RBAC permission for viewing webhooks in this project
	projectID := int(args.ProjectID)
	hasPermission, err := rbacService.HasPermission(ctx, ctxValue.UserID, &projectID, service.PermProjectView)
	if err != nil {
		return webhooksResponse{}, NewGraphQLHttpError(http.StatusInternalServerError, err)
	}
	if !hasPermission {
		return webhooksResponse{}, NewGraphQLHttpError(http.StatusUnauthorized, errors.New("insufficient permissions to view webhooks"))
	}

	stat := service.EntClient.Webhook.Query().
		Where(webhook.ProjectID(int(args.ProjectID))).
		Order(ent.Desc(webhook.FieldID))

	return webhooksResponse{
		stat:       stat,
		pagination: args.Pagination,
	}, nil
}

func (w webhooksResponse) Count(ctx context.Context) (int32, error) {
	count, err := w.stat.Clone().Count(ctx)
	if err != nil {
		return 0, NewGraphQLHttpError(http.StatusInternalServerError, err)
	}
	return int32(count), nil
}

func (w webhooksResponse) Edges(ctx context.Context) (res []webhookResponse, err error) {
	webhooks, err := w.stat.Clone().
		Limit(int(w.pagination.Limit)).
		Offset(int(w.pagination.Offset)).
		All(ctx)

	if err != nil {
		return nil, NewGraphQLHttpError(http.StatusInternalServerError, err)
	}

	for _, webhook := range webhooks {
		res = append(res, webhookResponse{w: webhook})
	}
	return
}

type webhookResponse struct {
	w *ent.Webhook
}

func (w webhookResponse) ID() int32 {
	return int32(w.w.ID)
}

func (w webhookResponse) Name() string {
	return w.w.Name
}

func (w webhookResponse) Description() string {
	return w.w.Description
}

func (w webhookResponse) URL() string {
	return w.w.URL
}

func (w webhookResponse) Event() string {
	return w.w.Event
}

func (w webhookResponse) Enabled() bool {
	return w.w.Enabled
}

func (w webhookResponse) CreatedAt() string {
	return w.w.CreateTime.Format(time.RFC3339)
}

func (w webhookResponse) UpdatedAt() string {
	return w.w.UpdateTime.Format(time.RFC3339)
}

func (w webhookResponse) Creator(ctx context.Context) (userResponse, error) {
	u, err := w.w.QueryCreator().Only(ctx)
	if err != nil {
		return userResponse{}, NewGraphQLHttpError(http.StatusInternalServerError, err)
	}
	return userResponse{u}, nil
}

func (w webhookResponse) Project(ctx context.Context) (projectResponse, error) {
	p, err := w.w.QueryProject().Only(ctx)
	if err != nil {
		return projectResponse{}, NewGraphQLHttpError(http.StatusInternalServerError, err)
	}
	return projectResponse{p: p}, nil
}