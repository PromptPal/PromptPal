package schema

import (
	"context"

	"github.com/PromptPal/PromptPal/ent"
)

type projectResponse struct {
	p *ent.Project
}

type projectArgs struct {
	ID int32
}

func (q QueryResolver) Project(ctx context.Context, args projectArgs) (projectResponse, error) {
	return projectResponse{}, GraphQLHttpError{}
}

func (p projectResponse) ID() int32 {
	return int32(p.p.ID)
}

func (q QueryResolver) CreateProject() (projectResponse, error) {
	return projectResponse{}, GraphQLHttpError{}
}
