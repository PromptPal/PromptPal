package schema

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/PromptPal/PromptPal/ent/prompt"
	"github.com/PromptPal/PromptPal/ent/schema"
	dbSchema "github.com/PromptPal/PromptPal/ent/schema"
	"github.com/PromptPal/PromptPal/service"
	"github.com/go-redis/cache/v9"
)

type createPromptData struct {
	ProjectID   int32
	Name        string
	Description string
	TokenCount  int32
	Debug       *bool
	Enabled     *bool
	Prompts     []dbSchema.PromptRow
	Variables   []dbSchema.PromptVariable
	PublicLevel prompt.PublicLevel

	ProviderId int32
}

type createPromptArgs struct {
	Data createPromptData
}

func (q QueryResolver) CreatePrompt(ctx context.Context, args createPromptArgs) (promptResponse, error) {
	payload := args.Data
	ctxValue := ctx.Value(service.GinGraphQLContextKey).(service.GinGraphQLContextType)
	
	// Check RBAC permission for prompt creation
	projectID := int(payload.ProjectID)
	hasPermission, err := service.RBACServiceInstance.HasPermission(ctx, ctxValue.UserID, &projectID, service.PermPromptCreate)
	if err != nil {
		return promptResponse{}, NewGraphQLHttpError(http.StatusInternalServerError, err)
	}
	if !hasPermission {
		return promptResponse{}, NewGraphQLHttpError(http.StatusUnauthorized, errors.New("insufficient permissions to create prompt"))
	}

	stat := service.
		EntClient.
		Prompt.
		Create().
		SetName(payload.Name).
		SetDescription(payload.Description).
		SetCreatorID(ctxValue.UserID).
		SetProjectID(int(payload.ProjectID)).
		SetPrompts(payload.Prompts).
		SetVariables(payload.Variables).
		SetPublicLevel(payload.PublicLevel).
		SetTokenCount(int(payload.TokenCount)).
		SetNillableDebug(payload.Debug).
		SetNillableEnabled(payload.Enabled)

	stat.SetProviderID(int(payload.ProviderId))

	p, err := stat.Save(ctx)

	if err != nil {
		return promptResponse{}, NewGraphQLHttpError(http.StatusInternalServerError, err)
	}

	hid, err := hashidService.Encode(p.ID)
	if err != nil {
		return promptResponse{}, NewGraphQLHttpError(http.StatusInternalServerError, err)
	}
	// set cache
	service.Cache.Set(&cache.Item{
		Ctx:   ctx,
		Key:   fmt.Sprintf("prompt:%s", hid),
		Value: *p,
		TTL:   time.Hour * 24,
	})
	return promptResponse{
		prompt: p,
	}, nil
}

type updatePromptArgs struct {
	ID   int32
	Data createPromptData
}

func (q QueryResolver) UpdatePrompt(ctx context.Context, args updatePromptArgs) (result promptResponse, err error) {
	ctxValue := ctx.Value(service.GinGraphQLContextKey).(service.GinGraphQLContextType)
	payload := args.Data
	
	// First get the prompt to check project permissions
	oldPrompt, err := service.EntClient.Prompt.Get(ctx, int(args.ID))
	if err != nil {
		err = NewGraphQLHttpError(http.StatusNotFound, err)
		return
	}
	
	// Check RBAC permission for prompt update
	projectID := oldPrompt.ProjectId
	hasPermission, err := service.RBACServiceInstance.HasPermission(ctx, ctxValue.UserID, &projectID, service.PermPromptEdit)
	if err != nil {
		err = NewGraphQLHttpError(http.StatusInternalServerError, err)
		return
	}
	if !hasPermission {
		err = NewGraphQLHttpError(http.StatusUnauthorized, errors.New("insufficient permissions to update prompt"))
		return
	}
	
	tx, err := service.EntClient.Tx(ctx)

	if err != nil {
		err = NewGraphQLHttpError(http.StatusInternalServerError, err)
		return
	}

	// We already have oldPrompt, so no need to get it again

	snapshotData := schema.PromptComplete{
		Name:        oldPrompt.Name,
		Enabled:     oldPrompt.Enabled,
		Debug:       oldPrompt.Debug,
		Description: oldPrompt.Description,
		TokenCount:  oldPrompt.TokenCount,
		Prompts:     oldPrompt.Prompts,
		Variables:   oldPrompt.Variables,
		PublicLevel: oldPrompt.PublicLevel.String(),
	}

	err = tx.History.
		Create().
		SetModifierID(ctxValue.UserID).
		SetPromptID(int(args.ID)).
		SetSnapshot(snapshotData).
		Exec(ctx)

	if err != nil {
		tx.Rollback()
		err = NewGraphQLHttpError(http.StatusInternalServerError, err)
		return
	}

	updater := tx.Prompt.UpdateOneID(int(args.ID)).
		SetDescription(payload.Description).
		SetTokenCount(int(payload.TokenCount)).
		SetPrompts(payload.Prompts).
		SetVariables(payload.Variables).
		SetPublicLevel(payload.PublicLevel)

	providerId := int(args.Data.ProviderId)

	if providerId > 0 {
		updater = updater.SetProviderID(int(args.Data.ProviderId))
	}

	if args.Data.Enabled != nil {
		updater = updater.SetEnabled(*args.Data.Enabled)
	}
	if args.Data.Debug != nil {
		updater = updater.SetNillableDebug(args.Data.Debug)
	}

	updatedPrompt, err := updater.Save(ctx)

	if err != nil {
		tx.Rollback()
		err = NewGraphQLHttpError(http.StatusInternalServerError, err)
		return
	}

	if exp := tx.Commit(); exp != nil {
		tx.Rollback()
		err = NewGraphQLHttpError(http.StatusInternalServerError, exp)
		return
	}

	hid, err := hashidService.Encode(int(args.ID))
	if err != nil {
		err = NewGraphQLHttpError(http.StatusInternalServerError, err)
		return
	}

	// refresh cache
	service.Cache.Set(&cache.Item{
		Ctx:   ctx,
		Key:   fmt.Sprintf("prompt:%s", hid),
		Value: *updatedPrompt,
		TTL:   time.Hour * 24,
	})
	result.prompt = updatedPrompt
	return
}

type deletePromptArgs struct {
	ID int32
}

func (q QueryResolver) DeletePrompt(ctx context.Context, args deletePromptArgs) (bool, error) {
	ctxValue := ctx.Value(service.GinGraphQLContextKey).(service.GinGraphQLContextType)
	
	// First get the prompt to check project permissions
	prompt, err := service.EntClient.Prompt.Get(ctx, int(args.ID))
	if err != nil {
		return false, NewGraphQLHttpError(http.StatusNotFound, err)
	}
	
	// Check RBAC permission for prompt deletion
	projectID := prompt.ProjectId
	hasPermission, err := service.RBACServiceInstance.HasPermission(ctx, ctxValue.UserID, &projectID, service.PermPromptDelete)
	if err != nil {
		return false, NewGraphQLHttpError(http.StatusInternalServerError, err)
	}
	if !hasPermission {
		return false, NewGraphQLHttpError(http.StatusUnauthorized, errors.New("insufficient permissions to delete prompt"))
	}
	
	// Delete the prompt
	err = service.EntClient.Prompt.DeleteOneID(int(args.ID)).Exec(ctx)
	if err != nil {
		return false, NewGraphQLHttpError(http.StatusInternalServerError, err)
	}
	
	return true, nil
}
