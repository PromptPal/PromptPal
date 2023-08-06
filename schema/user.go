package schema

import (
	"context"
	"net/http"

	"github.com/PromptPal/PromptPal/ent"
	"github.com/PromptPal/PromptPal/ent/user"
	"github.com/PromptPal/PromptPal/service"
)

type userArgs struct {
	ID int32
}

func (q QueryResolver) User(ctx context.Context, args userArgs) (result userResponse, err error) {
	u, err := service.
		EntClient.
		User.
		Query().
		Where(user.ID(int(args.ID))).
		Only(ctx)
	if err != nil {
		err = NewGraphQLHttpError(http.StatusInternalServerError, err)
		return
	}
	result.u = u
	return
}

type userResponse struct {
	u *ent.User
}

func (u userResponse) ID() int32 {
	return int32(u.u.ID)
}

func (u userResponse) Name() string {
	return u.u.Name
}

func (u userResponse) Addr() string {
	return u.u.Addr
}
