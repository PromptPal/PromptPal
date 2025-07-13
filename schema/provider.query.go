package schema

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/PromptPal/PromptPal/ent"
	"github.com/PromptPal/PromptPal/ent/project"
	"github.com/PromptPal/PromptPal/ent/prompt"
	"github.com/PromptPal/PromptPal/ent/provider"
	"github.com/PromptPal/PromptPal/service"
	"github.com/go-redis/cache/v9"
)

type providerArgs struct {
	ID int32
}

type providerResponse struct {
	p *ent.Provider
}

func (q QueryResolver) Provider(ctx context.Context, args providerArgs) (res providerResponse, err error) {
	ctxValue := ctx.Value(service.GinGraphQLContextKey).(service.GinGraphQLContextType)
	
	// Check RBAC permission for provider view (system admin required due to sensitive API keys)
	hasPermission, err := service.RBACServiceInstance.HasPermission(ctx, ctxValue.UserID, nil, service.PermSystemAdmin)
	if err != nil {
		err = NewGraphQLHttpError(http.StatusInternalServerError, err)
		return
	}
	if !hasPermission {
		err = NewGraphQLHttpError(http.StatusUnauthorized, errors.New("insufficient permissions to view provider"))
		return
	}
	
	// Check if provider exists in cache
	err = service.Cache.Get(ctx, fmt.Sprintf("provider:%d", args.ID), &res.p)

	if err != nil {
		if !errors.Is(err, cache.ErrCacheMiss) {
			err = NewGraphQLHttpError(http.StatusInternalServerError, err)
			return
		}
		err = nil
		pjt, err := service.EntClient.Provider.Get(ctx, int(args.ID))
		if err != nil {
			err = NewGraphQLHttpError(http.StatusNotFound, err)
			return res, err
		}
		service.Cache.Set(&cache.Item{
			Ctx:   ctx,
			Key:   fmt.Sprintf("provider:%d", args.ID),
			Value: *pjt,
			TTL:   time.Hour * 24,
		})
		res.p = pjt
	}

	return
}

type providersArgs struct {
	Pagination paginationInput
}

type providersResponse struct {
	providers []*ent.Provider
}

func (q QueryResolver) Providers(ctx context.Context, args providersArgs) (res providersResponse, err error) {
	ctxValue := ctx.Value(service.GinGraphQLContextKey).(service.GinGraphQLContextType)
	
	// Check RBAC permission for providers list (system admin required due to sensitive API keys)
	hasPermission, err := service.RBACServiceInstance.HasPermission(ctx, ctxValue.UserID, nil, service.PermSystemAdmin)
	if err != nil {
		err = NewGraphQLHttpError(http.StatusInternalServerError, err)
		return
	}
	if !hasPermission {
		err = NewGraphQLHttpError(http.StatusUnauthorized, errors.New("insufficient permissions to list providers"))
		return
	}
	
	providers, err := service.
		EntClient.
		Provider.
		Query().
		Limit(int(args.Pagination.Limit)).
		Offset(int(args.Pagination.Offset)).
		Order(ent.Desc(provider.FieldID)).
		All(ctx)

	if err != nil {
		err = NewGraphQLHttpError(http.StatusInternalServerError, err)
		return
	}
	res.providers = providers
	return
}

// Query provider by project ID
type projectProviderArgs struct {
	ProjectId int32
}

func (q QueryResolver) ProjectProvider(ctx context.Context, args projectProviderArgs) (res providerResponse, err error) {
	ctxValue := ctx.Value(service.GinGraphQLContextKey).(service.GinGraphQLContextType)
	
	// Check RBAC permission for project provider view
	projectID := int(args.ProjectId)
	hasPermission, err := service.RBACServiceInstance.HasPermission(ctx, ctxValue.UserID, &projectID, service.PermProjectView)
	if err != nil {
		err = NewGraphQLHttpError(http.StatusInternalServerError, err)
		return
	}
	if !hasPermission {
		err = NewGraphQLHttpError(http.StatusUnauthorized, errors.New("insufficient permissions to view project provider"))
		return
	}
	
	p, err := service.
		EntClient.
		Provider.
		Query().
		Where(
			provider.HasProjectWith(
				project.ID(int(args.ProjectId)),
			),
		).
		Only(ctx)

	// If no provider is found, return empty response without error
	if ent.IsNotFound(err) {
		return providerResponse{}, nil
	}

	if err != nil {
		err = NewGraphQLHttpError(http.StatusInternalServerError, err)
		return
	}

	res.p = p
	return
}

// Response methods for ProviderList
func (p providersResponse) Count() int32 {
	return int32(len(p.providers))
}

func (p providersResponse) Edges() (result []providerResponse) {
	for _, provider := range p.providers {
		result = append(result, providerResponse{p: provider})
	}
	return
}

// Response methods for Provider
func (p providerResponse) ID() int32 {
	if p.p == nil {
		return 0
	}
	return int32(p.p.ID)
}

func (p providerResponse) Name() string {
	if p.p == nil {
		return ""
	}
	return p.p.Name
}

func (p providerResponse) Description() string {
	if p.p == nil {
		return ""
	}
	return p.p.Description
}

func (p providerResponse) Enabled() bool {
	if p.p == nil {
		return false
	}
	return p.p.Enabled
}

func (p providerResponse) Source() string {
	if p.p == nil {
		return ""
	}
	return p.p.Source
}

func (p providerResponse) Endpoint() string {
	if p.p == nil {
		return ""
	}
	return p.p.Endpoint
}

func (p providerResponse) OrganizationId() *string {
	if p.p == nil || p.p.OrganizationId == "" {
		return nil
	}
	return &p.p.OrganizationId
}

func (p providerResponse) DefaultModel() string {
	if p.p == nil {
		return ""
	}
	return p.p.DefaultModel
}

func (p providerResponse) Temperature() float64 {
	if p.p == nil {
		return 0
	}
	return p.p.Temperature
}

func (p providerResponse) TopP() float64 {
	if p.p == nil {
		return 0
	}
	return p.p.TopP
}

func (p providerResponse) MaxTokens() int32 {
	if p.p == nil {
		return 0
	}
	return int32(p.p.MaxTokens)
}

func (p providerResponse) Config() string {
	if p.p == nil {
		return ""
	}
	config, err := json.Marshal(p.p.Config)
	if err != nil {
		return ""
	}
	return string(config)
}

func (p providerResponse) Headers() string {
	if p.p == nil {
		return ""
	}
	headers, err := json.Marshal(p.p.Headers)
	if err != nil {
		return ""
	}
	return string(headers)
}

func (p providerResponse) CreatedAt() string {
	if p.p == nil {
		return ""
	}
	return p.p.CreateTime.Format(time.RFC3339)
}

func (p providerResponse) UpdatedAt() string {
	if p.p == nil {
		return ""
	}
	return p.p.UpdateTime.Format(time.RFC3339)
}

// Relationship resolvers
func (p providerResponse) Projects(ctx context.Context) (res projectsResponse, err error) {
	if p.p == nil {
		return
	}

	projects, err := service.EntClient.Project.Query().
		Where(project.HasProviderWith(provider.ID(p.p.ID))).
		Limit(1000).
		All(ctx)

	if err != nil {
		err = NewGraphQLHttpError(http.StatusInternalServerError, err)
		return
	}

	// If no projects are associated, return empty response
	if len(projects) == 0 {
		return
	}

	res.projects = projects
	return
}

func (p providerResponse) Prompts(ctx context.Context) (result promptsResponse, err error) {
	if p.p == nil {
		// Return empty list if provider is nil
		result.stat = service.EntClient.Prompt.Query().Where(prompt.ID(0))
		result.pagination = paginationInput{
			Limit:  10,
			Offset: 0,
		}
		return
	}

	stat := service.
		EntClient.
		Prompt.
		Query().
		Where(prompt.HasProviderWith(provider.ID(p.p.ID))).
		Order(ent.Desc(prompt.FieldID))

	result.stat = stat
	result.pagination = paginationInput{
		Limit:  10,
		Offset: 0,
	}
	return
}
