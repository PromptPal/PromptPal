package schema

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/PromptPal/PromptPal/ent"
	"github.com/PromptPal/PromptPal/ent/opentoken"
	"github.com/PromptPal/PromptPal/ent/project"
	"github.com/PromptPal/PromptPal/ent/prompt"
	"github.com/PromptPal/PromptPal/ent/user"
	"github.com/PromptPal/PromptPal/service"
)

type projectArgs struct {
	ID int32
}
type projectResponse struct {
	p *ent.Project
}

func (q QueryResolver) Project(ctx context.Context, args projectArgs) (res projectResponse, err error) {
	ctxValue := ctx.Value(service.GinGraphQLContextKey).(service.GinGraphQLContextType)
	
	// Check RBAC permission for project view
	projectID := int(args.ID)
	hasPermission, err := service.RBACServiceInstance.HasPermission(ctx, ctxValue.UserID, &projectID, service.PermProjectView)
	if err != nil {
		err = NewGraphQLHttpError(http.StatusInternalServerError, err)
		return
	}
	if !hasPermission {
		err = NewGraphQLHttpError(http.StatusUnauthorized, errors.New("insufficient permissions to view project"))
		return
	}
	
	pj, err := service.
		EntClient.
		Project.
		Query().
		Where(project.ID(int(args.ID))).
		Only(ctx)

	if err != nil {
		err = NewGraphQLHttpError(http.StatusInternalServerError, err)
		return
	}
	res.p = pj
	return
}

type projectsArgs struct {
	ProjectID  int32
	Pagination paginationInput
}

type projectsResponse struct {
	projects []*ent.Project
}

func (q QueryResolver) Projects(ctx context.Context, args projectsArgs) (res projectsResponse, err error) {
	ctxValue := ctx.Value(service.GinGraphQLContextKey).(service.GinGraphQLContextType)
	
	// Check RBAC permission for listing projects
	hasPermission, err := service.RBACServiceInstance.HasPermission(ctx, ctxValue.UserID, nil, service.PermProjectView)
	if err != nil {
		err = NewGraphQLHttpError(http.StatusInternalServerError, err)
		return
	}
	if !hasPermission {
		err = NewGraphQLHttpError(http.StatusUnauthorized, errors.New("insufficient permissions to list projects"))
		return
	}
	
	pjs, err := service.
		EntClient.
		Project.
		Query().
		Limit(int(args.Pagination.Limit)).
		Offset(int(args.Pagination.Offset)).
		Order(ent.Desc(project.FieldID)).
		All(ctx)

	if err != nil {
		err = NewGraphQLHttpError(http.StatusInternalServerError, err)
		return
	}
	res.projects = pjs
	return
}

func (p projectsResponse) Count() int32 {
	return int32(len(p.projects))
}

func (p projectsResponse) Edges() (result []projectResponse) {
	for _, pj := range p.projects {
		result = append(result, projectResponse{p: pj})
	}
	return
}

func (p projectResponse) ID() int32 {
	return int32(p.p.ID)
}
func (p projectResponse) Name() string {
	return p.p.Name
}
func (p projectResponse) Enabled() bool {
	return p.p.Enabled
}
func (p projectResponse) OpenAIBaseURL() string {
	return p.p.OpenAIBaseURL
}
func (p projectResponse) OpenAIToken() string {
	return p.p.OpenAIToken
}

func (p projectResponse) GeminiBaseURL() string {
	return p.p.GeminiBaseURL
}
func (p projectResponse) GeminiToken() string {
	return p.p.GeminiToken
}

func (p projectResponse) OpenAIModel() string {
	return p.p.OpenAIModel
}
func (p projectResponse) OpenAITemperature() float64 {
	return p.p.OpenAITemperature
}
func (p projectResponse) OpenAITopP() float64 {
	return p.p.OpenAITopP
}
func (p projectResponse) OpenAIMaxTokens() int32 {
	return int32(p.p.OpenAIMaxTokens)
}

func (p projectResponse) CreatedAt() string {
	return p.p.CreateTime.Format(time.RFC3339)
}
func (p projectResponse) UpdatedAt() string {
	return p.p.UpdateTime.Format(time.RFC3339)
}

func (p projectResponse) Provider(ctx context.Context) (*providerResponse, error) {
	pj, err := p.p.QueryProvider().Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, NewGraphQLHttpError(http.StatusInternalServerError, err)
	}
	return &providerResponse{p: pj}, nil
}

func (p projectResponse) Creator(ctx context.Context) (res userResponse, err error) {
	u, err := service.
		EntClient.
		User.
		Query().
		Where(user.ID(p.p.CreatorID)).
		Only(ctx)
	if err != nil {
		err = NewGraphQLHttpError(http.StatusInternalServerError, err)
		return
	}
	res.u = u
	return
}
func (p projectResponse) OpenTokens(ctx context.Context) (result openTokenListResponse, err error) {
	ots, err := service.
		EntClient.
		OpenToken.
		Query().
		Where(opentoken.HasProjectWith(project.ID(p.p.ID))).
		Order(ent.Desc(opentoken.FieldID)).
		All(ctx)

	if err != nil {
		err = NewGraphQLHttpError(http.StatusInternalServerError, err)
		return
	}

	result.openTokens = ots
	return
}

func (p projectResponse) LatestPrompts(ctx context.Context) (result promptsResponse) {
	stat := service.
		EntClient.
		Prompt.
		Query().
		Where(prompt.HasProjectWith(project.ID(p.p.ID))).
		Order(ent.Desc(prompt.FieldID))
	result.stat = stat
	result.pagination = paginationInput{
		Limit:  10,
		Offset: 0,
	}
	return
}
