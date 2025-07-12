package schema

import (
	"context"
	"net/http"

	"github.com/PromptPal/PromptPal/ent"
	"github.com/PromptPal/PromptPal/ent/user"
	"github.com/PromptPal/PromptPal/service"
)

type userArgs struct {
	ID *int32
}

func (q QueryResolver) User(ctx context.Context, args userArgs) (result userResponse, err error) {
	var uid int
	if args.ID == nil {
		ctxValue := ctx.Value(service.GinGraphQLContextKey).(service.GinGraphQLContextType)
		uid = ctxValue.UserID
	} else {
		uid = int(*args.ID)
	}
	u, err := service.
		EntClient.
		User.
		Query().
		Where(user.ID(uid)).
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

func (u userResponse) Avatar() string {
	return u.u.Avatar
}

func (u userResponse) Email() string {
	return u.u.Email
}

func (u userResponse) Phone() string {
	return u.u.Phone
}

func (u userResponse) Lang() string {
	return u.u.Lang
}

func (u userResponse) Level() int32 {
	return int32(u.u.Level)
}

func (u userResponse) Source() string {
	return u.u.Source
}
