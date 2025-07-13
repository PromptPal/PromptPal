package schema

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/PromptPal/PromptPal/ent/project"
	"github.com/PromptPal/PromptPal/ent/prompt"
	"github.com/PromptPal/PromptPal/ent/provider"
	"github.com/PromptPal/PromptPal/service"
	"github.com/go-redis/cache/v9"
)

type createProviderData struct {
	Name           string
	Description    *string
	Enabled        *bool
	Source         string
	Endpoint       string
	ApiKey         string
	OrganizationId *string
	DefaultModel   *string
	Temperature    *float64
	TopP           *float64
	MaxTokens      *int32
	Config         string
	Headers        string
}

type createProviderArgs struct {
	Data createProviderData
}

func (q QueryResolver) CreateProvider(ctx context.Context, args createProviderArgs) (providerResponse, error) {
	ctxValue := ctx.Value(service.GinGraphQLContextKey).(service.GinGraphQLContextType)
	
	// Check RBAC permission for provider creation (system admin required due to sensitive API keys)
	hasPermission, err := service.RBACServiceInstance.HasPermission(ctx, ctxValue.UserID, nil, service.PermSystemAdmin)
	if err != nil {
		return providerResponse{}, NewGraphQLHttpError(http.StatusInternalServerError, err)
	}
	if !hasPermission {
		return providerResponse{}, NewGraphQLHttpError(http.StatusUnauthorized, errors.New("insufficient permissions to create provider"))
	}
	
	data := args.Data
	// Start building the provider
	stat := service.
		EntClient.
		Provider.
		Create().
		SetName(data.Name).
		SetSource(data.Source).
		SetCreatorID(ctxValue.UserID)

	// Set optional fields if provided
	if data.Description != nil {
		stat = stat.SetDescription(*data.Description)
	}
	if data.Enabled != nil {
		stat = stat.SetEnabled(*data.Enabled)
	}
	if data.Endpoint != "" {
		stat = stat.SetEndpoint(data.Endpoint)
	}
	if data.ApiKey != "" {
		stat = stat.SetApiKey(data.ApiKey)
	}
	if data.OrganizationId != nil {
		stat = stat.SetOrganizationId(*data.OrganizationId)
	}
	if data.DefaultModel != nil {
		stat = stat.SetDefaultModel(*data.DefaultModel)
	}
	if data.Temperature != nil {
		stat = stat.SetTemperature(*data.Temperature)
	}
	if data.TopP != nil {
		stat = stat.SetTopP(*data.TopP)
	}
	if data.MaxTokens != nil {
		stat = stat.SetMaxTokens(int(*data.MaxTokens))
	}
	if data.Config != "" {
		var providerConfig map[string]interface{}
		err := json.Unmarshal([]byte(data.Config), &providerConfig)
		if err != nil {
			return providerResponse{}, NewGraphQLHttpError(http.StatusBadRequest, err)
		}
		stat = stat.SetConfig(providerConfig)
	}
	if data.Headers != "" {
		var providerHeaders map[string]string
		err := json.Unmarshal([]byte(data.Headers), &providerHeaders)
		if err != nil {
			return providerResponse{}, NewGraphQLHttpError(http.StatusBadRequest, err)
		}
		stat = stat.SetHeaders(providerHeaders)
	}

	provider, err := stat.Save(ctx)

	if err != nil {
		return providerResponse{}, NewGraphQLHttpError(http.StatusInternalServerError, err)
	}

	return providerResponse{p: provider}, nil
}

type updateProviderData struct {
	Name           *string
	Description    *string
	Enabled        *bool
	Source         *string
	Endpoint       *string
	ApiKey         *string
	OrganizationId *string
	DefaultModel   *string
	Temperature    *float64
	TopP           *float64
	MaxTokens      *int32
	Config         *string
	Headers        *string
}

type updateProviderArgs struct {
	ID   int32
	Data updateProviderData
}

func (q QueryResolver) UpdateProvider(ctx context.Context, args updateProviderArgs) (providerResponse, error) {
	ctxValue := ctx.Value(service.GinGraphQLContextKey).(service.GinGraphQLContextType)
	
	// Check RBAC permission for provider update (system admin required due to sensitive API keys)
	hasPermission, err := service.RBACServiceInstance.HasPermission(ctx, ctxValue.UserID, nil, service.PermSystemAdmin)
	if err != nil {
		return providerResponse{}, NewGraphQLHttpError(http.StatusInternalServerError, err)
	}
	if !hasPermission {
		return providerResponse{}, NewGraphQLHttpError(http.StatusUnauthorized, errors.New("insufficient permissions to update provider"))
	}
	
	updater := service.EntClient.Provider.UpdateOneID(int(args.ID))

	if args.Data.Name != nil {
		updater = updater.SetName(*args.Data.Name)
	}
	if args.Data.Description != nil {
		updater = updater.SetDescription(*args.Data.Description)
	}
	if args.Data.Enabled != nil {
		updater = updater.SetEnabled(*args.Data.Enabled)
	}
	if args.Data.Source != nil {
		updater = updater.SetSource(*args.Data.Source)
	}
	if args.Data.Endpoint != nil {
		updater = updater.SetEndpoint(*args.Data.Endpoint)
	}
	if args.Data.ApiKey != nil {
		updater = updater.SetApiKey(*args.Data.ApiKey)
	}
	if args.Data.OrganizationId != nil {
		updater = updater.SetOrganizationId(*args.Data.OrganizationId)
	}
	if args.Data.DefaultModel != nil {
		updater = updater.SetDefaultModel(*args.Data.DefaultModel)
	}
	if args.Data.Temperature != nil {
		updater = updater.SetTemperature(*args.Data.Temperature)
	}
	if args.Data.TopP != nil {
		updater = updater.SetTopP(*args.Data.TopP)
	}
	if args.Data.MaxTokens != nil {
		updater = updater.SetMaxTokens(int(*args.Data.MaxTokens))
	}
	if args.Data.Config != nil {
		var providerConfig map[string]interface{}
		err := json.Unmarshal([]byte(*args.Data.Config), &providerConfig)
		if err != nil {
			return providerResponse{}, NewGraphQLHttpError(http.StatusBadRequest, err)
		}
		updater = updater.SetConfig(providerConfig)
	}
	if args.Data.Headers != nil {
		var providerHeaders map[string]string
		err := json.Unmarshal([]byte(*args.Data.Headers), &providerHeaders)
		if err != nil {
			return providerResponse{}, NewGraphQLHttpError(http.StatusBadRequest, err)
		}
		updater = updater.SetHeaders(providerHeaders)
	}

	provider, err := updater.Save(ctx)
	if err != nil {
		return providerResponse{}, NewGraphQLHttpError(http.StatusInternalServerError, err)
	}

	// Cache the provider for future queries
	service.Cache.Set(&cache.Item{
		Ctx:   ctx,
		Key:   fmt.Sprintf("provider:%d", provider.ID),
		Value: *provider,
		TTL:   time.Hour * 24,
	})

	return providerResponse{
		p: provider,
	}, nil
}

type deleteProviderArgs struct {
	ID int32
}

func (q QueryResolver) DeleteProvider(ctx context.Context, args deleteProviderArgs) (bool, error) {
	ctxValue := ctx.Value(service.GinGraphQLContextKey).(service.GinGraphQLContextType)
	
	// Check RBAC permission for provider deletion (system admin required due to sensitive API keys)
	hasPermission, err := service.RBACServiceInstance.HasPermission(ctx, ctxValue.UserID, nil, service.PermSystemAdmin)
	if err != nil {
		return false, NewGraphQLHttpError(http.StatusInternalServerError, err)
	}
	if !hasPermission {
		return false, NewGraphQLHttpError(http.StatusUnauthorized, errors.New("insufficient permissions to delete provider"))
	}
	
	tx, err := service.EntClient.Tx(ctx)

	if err != nil {
		return false, NewGraphQLHttpError(http.StatusInternalServerError, err)
	}

	err = tx.Provider.DeleteOneID(int(args.ID)).Exec(ctx)
	if err != nil {
		tx.Rollback()
		return false, NewGraphQLHttpError(http.StatusInternalServerError, err)
	}

	err = tx.Commit()
	if err != nil {
		return false, NewGraphQLHttpError(http.StatusInternalServerError, err)
	}

	// Remove from cache if it exists
	service.Cache.Delete(ctx, fmt.Sprintf("provider:%d", args.ID))

	return true, nil
}

// Association mutations
type assignProviderToProjectArgs struct {
	ProviderId int32
	ProjectId  int32
}

func (q QueryResolver) AssignProviderToProject(ctx context.Context, args assignProviderToProjectArgs) (bool, error) {
	// Get the provider
	provider, err := service.EntClient.Provider.Get(ctx, int(args.ProviderId))
	if err != nil {
		return false, NewGraphQLHttpError(http.StatusInternalServerError, err)
	}

	// Get the project
	project, err := service.EntClient.Project.Get(ctx, int(args.ProjectId))
	if err != nil {
		return false, NewGraphQLHttpError(http.StatusInternalServerError, err)
	}

	// Update the provider to associate with the project
	_, err = provider.Update().AddProject(project).Save(ctx)
	if err != nil {
		return false, NewGraphQLHttpError(http.StatusInternalServerError, err)
	}

	return true, nil
}

type removeProviderFromProjectArgs struct {
	ProjectId int32
}

func (q QueryResolver) RemoveProviderFromProject(ctx context.Context, args removeProviderFromProjectArgs) (bool, error) {
	// Get the project
	pj, err := service.EntClient.Project.Get(ctx, int(args.ProjectId))
	if err != nil {
		return false, NewGraphQLHttpError(http.StatusInternalServerError, err)
	}
	// Find providers associated with this project
	providers, err := service.EntClient.Provider.Query().
		Where(
			provider.HasProjectWith(
				project.ID(int(args.ProjectId)),
			),
		).All(ctx)

	if err != nil {
		return false, NewGraphQLHttpError(http.StatusInternalServerError, err)
	}

	// If no providers found, return success
	if len(providers) == 0 {
		return true, nil
	}

	// Remove the association for each provider
	for _, provider := range providers {
		_, err = provider.Update().RemoveProject(pj).Save(ctx)
		if err != nil {
			return false, NewGraphQLHttpError(http.StatusInternalServerError, err)
		}
	}

	return true, nil
}

type assignProviderToPromptArgs struct {
	ProviderId int32
	PromptId   int32
}

func (q QueryResolver) AssignProviderToPrompt(ctx context.Context, args assignProviderToPromptArgs) (bool, error) {
	// Get the provider
	provider, err := service.EntClient.Provider.Get(ctx, int(args.ProviderId))
	if err != nil {
		return false, NewGraphQLHttpError(http.StatusInternalServerError, err)
	}

	// Get the prompt
	prompt, err := service.EntClient.Prompt.Get(ctx, int(args.PromptId))
	if err != nil {
		return false, NewGraphQLHttpError(http.StatusInternalServerError, err)
	}

	// Update the provider to associate with the prompt
	_, err = provider.Update().AddPrompt(prompt).Save(ctx)
	if err != nil {
		return false, NewGraphQLHttpError(http.StatusInternalServerError, err)
	}

	return true, nil
}

type removeProviderFromPromptArgs struct {
	PromptId int32
}

func (q QueryResolver) RemoveProviderFromPrompt(ctx context.Context, args removeProviderFromPromptArgs) (bool, error) {
	// Get the prompt
	pt, err := service.EntClient.Prompt.Get(ctx, int(args.PromptId))
	if err != nil {
		return false, NewGraphQLHttpError(http.StatusInternalServerError, err)
	}

	// Find providers associated with this prompt
	providers, err := service.EntClient.Provider.Query().
		Where(
			provider.HasPromptWith(
				prompt.ID(int(args.PromptId)),
			),
		).All(ctx)

	if err != nil {
		return false, NewGraphQLHttpError(http.StatusInternalServerError, err)
	}

	// If no providers found, return success
	if len(providers) == 0 {
		return true, nil
	}

	// Remove the association for each provider
	for _, provider := range providers {
		_, err = provider.Update().RemovePrompt(pt).Save(ctx)
		if err != nil {
			return false, NewGraphQLHttpError(http.StatusInternalServerError, err)
		}
	}

	return true, nil
}
