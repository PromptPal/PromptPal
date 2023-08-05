package schema

import (
	"context"
	"errors"
	"net/http"
	"time"

	cache "github.com/Code-Hex/go-generics-cache"
	"github.com/PromptPal/PromptPal/ent"
	"github.com/PromptPal/PromptPal/ent/prompt"
	"github.com/PromptPal/PromptPal/ent/schema"
	dbSchema "github.com/PromptPal/PromptPal/ent/schema"
	"github.com/PromptPal/PromptPal/service"
)

type createPromptData struct {
	ProjectID   int
	Name        string
	Description string
	TokenCount  int
	Debug       *bool
	Enabled     *bool
	Prompts     []dbSchema.PromptRow
	Variables   []dbSchema.PromptVariable
	PublicLevel prompt.PublicLevel
}

type createPromptArgs struct {
	Data createPromptData
}

type promptResponse struct {
	Prompt *ent.Prompt
}

func (q QueryResolver) CreatePrompt(ctx context.Context, args createPromptArgs) (promptResponse, error) {
	payload := args.Data
	ctxValue := ctx.Value(service.GinGraphQLContextKey).(service.GinGraphQLContextType)
	p, err := service.
		EntClient.
		Prompt.
		Create().
		SetName(payload.Name).
		SetDescription(payload.Description).
		SetCreatorID(ctxValue.UserID).
		SetProjectID(payload.ProjectID).
		SetPrompts(payload.Prompts).
		SetVariables(payload.Variables).
		SetPublicLevel(payload.PublicLevel).
		SetTokenCount(payload.TokenCount).
		SetNillableDebug(payload.Debug).
		SetNillableEnabled(payload.Enabled).
		Save(ctx)

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
		Prompt: p,
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

	updator := tx.Prompt.UpdateOneID(int(args.ID)).
		SetDescription(payload.Description).
		SetTokenCount(payload.TokenCount).
		SetPrompts(payload.Prompts).
		SetVariables(payload.Variables).
		SetPublicLevel(payload.PublicLevel)

	if args.Data.Enabled != nil {
		updator = updator.SetEnabled(*args.Data.Enabled)
	}
	if args.Data.Debug != nil {
		updator = updator.SetNillableDebug(args.Data.Debug)
	}

	updatedPrompt, err := updator.Save(ctx)

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
	result.Prompt = updatedPrompt
	return
}

type deletePromptArgs struct {
	ID int32
}

func (q QueryResolver) DeletePrompt(ctx context.Context, args deletePromptArgs) (bool, error) {
	return false, NewGraphQLHttpError(http.StatusNotImplemented, errors.New("not implemented"))
}

func (p promptResponse) ID() int32 {
	return int32(p.Prompt.ID)
}

func (p promptResponse) HashID() (string, error) {
	hid, err := hashidService.Encode(p.Prompt.ID)
	if err != nil {
		return "", NewGraphQLHttpError(http.StatusInternalServerError, err)
	}
	return hid, nil
}
func (p promptResponse) Name() string {
	return p.Prompt.Name
}

func (p promptResponse) Description() string {
	return p.Prompt.Description
}

func (p promptResponse) TokenCount() int {
	return p.Prompt.TokenCount
}

func (p promptResponse) CreatedAt() string {
	return p.Prompt.CreateTime.Format(time.RFC3339)
}
func (p promptResponse) Enabled() bool {
	return p.Prompt.Enabled
}
func (p promptResponse) Debug() bool {
	return p.Prompt.Debug
}
