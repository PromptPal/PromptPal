package schema

import (
	"context"
	"errors"
	"net/http"
	"time"

	cache "github.com/Code-Hex/go-generics-cache"
	"github.com/PromptPal/PromptPal/ent"
	"github.com/PromptPal/PromptPal/ent/prompt"
	"github.com/PromptPal/PromptPal/ent/promptcall"
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
}

type createPromptArgs struct {
	Data createPromptData
}

type promptResponse struct {
	prompt *ent.Prompt
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
		SetProjectID(int(payload.ProjectID)).
		SetPrompts(payload.Prompts).
		SetVariables(payload.Variables).
		SetPublicLevel(payload.PublicLevel).
		SetTokenCount(int(payload.TokenCount)).
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

func (p promptResponse) ID() int32 {
	return int32(p.prompt.ID)
}

func (p promptResponse) HashID() (string, error) {
	hid, err := hashidService.Encode(p.prompt.ID)
	if err != nil {
		return "", NewGraphQLHttpError(http.StatusInternalServerError, err)
	}
	return hid, nil
}
func (p promptResponse) Name() string {
	return p.prompt.Name
}

func (p promptResponse) Description() string {
	return p.prompt.Description
}

func (p promptResponse) TokenCount() int32 {
	return int32(p.prompt.TokenCount)
}

func (p promptResponse) CreatedAt() string {
	return p.prompt.CreateTime.Format(time.RFC3339)
}
func (p promptResponse) UpdatedAt() string {
	return p.prompt.UpdateTime.Format(time.RFC3339)
}
func (p promptResponse) Enabled() bool {
	return p.prompt.Enabled
}
func (p promptResponse) Debug() bool {
	return p.prompt.Debug
}

func (p promptResponse) PublicLevel() prompt.PublicLevel {
	return prompt.PublicLevel(p.prompt.PublicLevel)
}

type promptRowResponse struct {
	p dbSchema.PromptRow
}

func (p promptResponse) Prompts() (result []promptRowResponse) {
	for _, v := range p.prompt.Prompts {
		result = append(result, promptRowResponse{
			p: v,
		})
	}
	return
}

func (p promptRowResponse) Prompt() string {
	return p.p.Prompt
}
func (p promptRowResponse) Role() string {
	r := p.p.Role
	switch r {
	case "system":
		return "system"
	case "assistant":
		return "assistant"
	case "user":
		return "user"
	default:
		return "unknown"
	}
}

type promptVariableResponse struct {
	p dbSchema.PromptVariable
}

func (p promptResponse) Variables() (result []promptVariableResponse) {
	for _, v := range p.prompt.Variables {
		result = append(result, promptVariableResponse{
			p: v,
		})
	}
	return
}

func (p promptVariableResponse) Name() string {
	return p.p.Name
}
func (p promptVariableResponse) Type() string {
	return p.p.Type
}

func (p promptResponse) LatestCalls(ctx context.Context) (res promptCallListResponse) {
	stat := service.EntClient.PromptCall.Query().
		Where(
			promptcall.HasPromptWith(prompt.ID(int(p.prompt.ID))),
		).Order(ent.Desc(promptcall.FieldID))

	res.stat = stat
	res.pagination = paginationInput{
		Limit:  10,
		Offset: 0,
	}
	return
}
