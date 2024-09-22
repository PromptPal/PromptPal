package schema

import (
	"context"
	"net/http"

	"github.com/PromptPal/PromptPal/ent"
	"github.com/PromptPal/PromptPal/ent/prompt"
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
