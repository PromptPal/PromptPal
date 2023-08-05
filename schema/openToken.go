package schema

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	cache "github.com/Code-Hex/go-generics-cache"
	"github.com/PromptPal/PromptPal/ent"
	"github.com/PromptPal/PromptPal/ent/opentoken"
	"github.com/PromptPal/PromptPal/ent/project"
	"github.com/PromptPal/PromptPal/service"
	"github.com/google/uuid"
)

type createOpenTokenData struct {
	ProjectId   int
	Name        string
	Description string
	TTL         int // in seconds
}

type createOpenTokenArgs struct {
	Data createOpenTokenData
}

type createOpenTokenResponse struct {
	token     string
	openToken *ent.OpenToken
}

type openTokenResponse struct {
	openToken *ent.OpenToken
}

func (q QueryResolver) CreateOpenToken(ctx context.Context, args createOpenTokenArgs) (result createOpenTokenResponse, err error) {
	pid := args.Data.ProjectId
	// TODO: put int tx
	previousCount, err := service.
		EntClient.
		OpenToken.
		Query().
		Where(opentoken.HasProjectWith(project.ID(pid))).
		Count(ctx)
	if err != nil {
		err = NewGraphQLHttpError(http.StatusInternalServerError, err)
		return
	}

	if previousCount > 20 {
		err = NewGraphQLHttpError(http.StatusTooManyRequests, errors.New("too many tokens"))
		return
	}

	payload := args.Data
	ctxValue := ctx.Value(service.GinGraphQLContextKey).(service.GinGraphQLContextType)

	tk := strings.Replace(uuid.New().String(), "-", "", -1)
	expireAt := time.Now().Add(time.Second * time.Duration(payload.TTL))

	ot, err := service.
		EntClient.
		OpenToken.
		Create().
		SetName(payload.Name).
		SetDescription(payload.Description).
		SetToken(tk).
		SetUserID(ctxValue.UserID).
		SetProjectID(pid).
		SetExpireAt(expireAt).
		Save(ctx)

	if err != nil {
		err = NewGraphQLHttpError(http.StatusInternalServerError, err)
		return
	}

	service.PublicAPIAuthCache.Set(tk, pid, cache.WithExpiration(24*time.Hour))
	result.openToken = ot
	result.token = tk
	return
}

type deleteOpenTokenArgs struct {
	ID int32
}

func (q QueryResolver) DeleteOpenToken(ctx context.Context, args deleteOpenTokenArgs) (result openTokenResponse, err error) {
	err = service.
		EntClient.
		OpenToken.
		DeleteOneID(int(args.ID)).
		Exec(ctx)
	if err != nil {
		err = NewGraphQLHttpError(http.StatusInternalServerError, err)
	}
	return
}
