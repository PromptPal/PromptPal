package schema

import (
	"context"
	"errors"
	"net/http"
	"time"

	cache "github.com/Code-Hex/go-generics-cache"
	"github.com/PromptPal/PromptPal/ent/prompt"
	"github.com/PromptPal/PromptPal/ent/schema"
	dbSchema "github.com/PromptPal/PromptPal/ent/schema"
	"github.com/PromptPal/PromptPal/service"
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
	service.ApiPromptCache.Set(hid, *p, cache.WithExpiration(time.Hour*24))
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
	tx, err := service.EntClient.Tx(ctx)

	if err != nil {
		err = NewGraphQLHttpError(http.StatusInternalServerError, err)
		return
	}

	oldPrompt, err := tx.Prompt.Get(ctx, int(args.ID))
	if err != nil {
		tx.Rollback()
		err = NewGraphQLHttpError(http.StatusNotFound, err)
		return
	}

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

	updater = updater.SetProviderID(int(args.Data.ProviderId))

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
	service.ApiPromptCache.Set(hid, *updatedPrompt, cache.WithExpiration(time.Hour*24))
	result.prompt = updatedPrompt
	return
}

type deletePromptArgs struct {
	ID int32
}

func (q QueryResolver) DeletePrompt(ctx context.Context, args deletePromptArgs) (bool, error) {
	return false, NewGraphQLHttpError(http.StatusNotImplemented, errors.New("not implemented"))
}
