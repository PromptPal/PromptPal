package schema

import (
	"context"
	"net/http"
	"time"

	"github.com/PromptPal/PromptPal/ent"
	"github.com/PromptPal/PromptPal/ent/prompt"
	"github.com/PromptPal/PromptPal/ent/promptcall"
	dbSchema "github.com/PromptPal/PromptPal/ent/schema"
	"github.com/PromptPal/PromptPal/service"
)

type promptsArgs struct {
	ProjectID  int32
	Pagination paginationInput
}

type promptsResponse struct {
	stat       *ent.PromptQuery
	pagination paginationInput
}

func (q QueryResolver) Prompts(ctx context.Context, args promptsArgs) (res promptsResponse) {
	res.stat = service.EntClient.
		Debug().
		Prompt.Query().
		Where(prompt.ProjectId(int(args.ProjectID))).
		Order(ent.Desc(prompt.FieldID))

	res.pagination = args.Pagination
	return
}

type promptSearchFilters struct {
	UserID *string
}

type promptArgs struct {
	ID      int32
	Filters *promptSearchFilters
}

func (q QueryResolver) Prompt(ctx context.Context, args promptArgs) (res promptResponse, err error) {
	p, err := service.EntClient.Prompt.Get(ctx, int(args.ID))
	if err != nil {
		err = NewGraphQLHttpError(http.StatusInternalServerError, err)
		return
	}
	res.prompt = p
	res.filters = args.Filters
	return
}

func (p promptsResponse) Count(ctx context.Context) (int32, error) {
	count, err := p.stat.Clone().Count(ctx)
	if err != nil {
		err = NewGraphQLHttpError(http.StatusInternalServerError, err)
		return 0, err
	}
	return int32(count), nil
}

func (p promptsResponse) Edges(ctx context.Context) (res []promptResponse, err error) {
	ps, err := p.stat.Clone().
		Limit(int(p.pagination.Limit)).
		Offset(int(p.pagination.Offset)).
		All(ctx)

	if err != nil {
		err = NewGraphQLHttpError(http.StatusInternalServerError, err)
		return
	}
	for _, p := range ps {
		res = append(res, promptResponse{prompt: p, filters: nil})
	}
	return
}

type promptResponse struct {
	prompt  *ent.Prompt
	filters *promptSearchFilters
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

func (p promptResponse) Project(ctx context.Context) (result projectResponse) {
	pj, err := service.EntClient.Project.Get(ctx, int(p.prompt.ProjectId))
	if err != nil {
		err = NewGraphQLHttpError(http.StatusInternalServerError, err)
		return
	}
	result.p = pj
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
func (p promptVariableResponse) Type() dbSchema.PromptVariableTypes {
	t := p.p.Type
	switch t {
	case "number":
		return dbSchema.PromptVariableTypesNumber
	case "boolean":
		return dbSchema.PromptVariableTypesBoolean
	case "video":
		return dbSchema.PromptVariableTypesVideo
	case "audio":
		return dbSchema.PromptVariableTypesAudio
	case "image":
		return dbSchema.PromptVariableTypesImage
	case "string":
	default:
		return dbSchema.PromptVariableTypesString
	}
	return dbSchema.PromptVariableTypesString
}

func (p promptResponse) Creator(ctx context.Context) (userResponse, error) {
	u, err := p.prompt.QueryCreator().Only(ctx)
	// u, err := service.EntClient.User.Get(ctx, uid)
	if err != nil {
		err = NewGraphQLHttpError(http.StatusInternalServerError, err)
		return userResponse{}, err
	}
	return userResponse{u}, nil
}

func (p promptResponse) Provider(ctx context.Context) (res *providerResponse, err error) {
	prompt, err := p.prompt.QueryProvider().Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		err = NewGraphQLHttpError(http.StatusInternalServerError, err)
		return
	}
	res = &providerResponse{p: prompt}
	return
}

func (p promptResponse) LatestCalls(ctx context.Context) (res promptCallListResponse) {
	stat := service.EntClient.PromptCall.Query().
		Where(
			promptcall.HasPromptWith(prompt.ID(int(p.prompt.ID))),
		)
	if p.filters != nil {
		if p.filters.UserID != nil && *p.filters.UserID != "" {
			stat = stat.Where(promptcall.UserIdContains(*p.filters.UserID))
		}
	}

	stat = stat.Order(ent.Desc(promptcall.FieldID))
	res.stat = stat
	res.pagination = paginationInput{
		Limit:  10,
		Offset: 0,
	}
	return
}
