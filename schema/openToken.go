package schema

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/PromptPal/PromptPal/ent"
	"github.com/PromptPal/PromptPal/ent/opentoken"
	"github.com/PromptPal/PromptPal/ent/project"
	"github.com/PromptPal/PromptPal/service"
	"github.com/go-redis/cache/v9"
	"github.com/google/uuid"
)

type createOpenTokenData struct {
	ProjectID          int32
	Name               string
	Description        string
	TTL                int32 // in seconds
	ApiValidateEnabled bool
	ApiValidatePath    *string
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
	ctxValue := ctx.Value(service.GinGraphQLContextKey).(service.GinGraphQLContextType)
	pid := int(args.Data.ProjectID)
	
	// Check RBAC permission for open token creation
	hasPermission, err := rbacService.HasPermission(ctx, ctxValue.UserID, &pid, service.PermProjectEdit)
	if err != nil {
		err = NewGraphQLHttpError(http.StatusInternalServerError, err)
		return
	}
	if !hasPermission {
		err = NewGraphQLHttpError(http.StatusUnauthorized, errors.New("insufficient permissions to create open token"))
		return
	}
	
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
		err = NewGraphQLHttpError(http.StatusInsufficientStorage, errors.New("too many tokens"))
		return
	}

	payload := args.Data

	tk := strings.Replace(uuid.New().String(), "-", "", -1)
	expireAt := time.Now().Add(time.Second * time.Duration(payload.TTL))

	ot, err := service.
		EntClient.
		OpenToken.
		Create().
		SetName(payload.Name).
		SetDescription(payload.Description).
		SetToken(tk).
		SetApiValidateEnabled(payload.ApiValidateEnabled).
		SetNillableApiValidatePath(payload.ApiValidatePath).
		SetUserID(ctxValue.UserID).
		SetProjectID(pid).
		SetExpireAt(expireAt).
		Save(ctx)

	if err != nil {
		err = NewGraphQLHttpError(http.StatusInternalServerError, err)
		return
	}

	service.Cache.Set(&cache.Item{
		Ctx:   ctx,
		Key:   fmt.Sprintf("openToken:%d", ot.ID),
		Value: *ot,
		TTL:   time.Hour,
	})
	result.openToken = ot
	result.token = tk
	return
}

type openTokenUpdate struct {
	ID   int32
	Data struct {
		Description        *string
		TTL                *int32
		ApiValidateEnabled *bool
		ApiValidatePath    *string
	}
}

func (q QueryResolver) UpdateOpenToken(ctx context.Context, args openTokenUpdate) (openTokenResponse, error) {
	ctxValue := ctx.Value(service.GinGraphQLContextKey).(service.GinGraphQLContextType)
	
	// First get the token to check its project permissions
	existingToken, err := service.EntClient.OpenToken.Get(ctx, int(args.ID))
	if err != nil {
		return openTokenResponse{}, NewGraphQLHttpError(http.StatusNotFound, err)
	}
	
	// Check RBAC permission for open token update
	projectID := existingToken.ProjectOpenTokens
	hasPermission, err := rbacService.HasPermission(ctx, ctxValue.UserID, &projectID, service.PermProjectEdit)
	if err != nil {
		return openTokenResponse{}, NewGraphQLHttpError(http.StatusInternalServerError, err)
	}
	if !hasPermission {
		return openTokenResponse{}, NewGraphQLHttpError(http.StatusUnauthorized, errors.New("insufficient permissions to update open token"))
	}
	
	stat := service.
		EntClient.
		OpenToken.
		UpdateOneID(int(args.ID))

	if args.Data.Description != nil {
		stat = stat.SetDescription(*args.Data.Description)
	}
	if args.Data.TTL != nil {
		stat = stat.SetExpireAt(time.Now().Add(time.Second * time.Duration(*args.Data.TTL)))
	}
	if args.Data.ApiValidateEnabled != nil {
		stat = stat.SetApiValidateEnabled(*args.Data.ApiValidateEnabled)
	}
	if args.Data.ApiValidatePath != nil {
		stat = stat.SetApiValidatePath(*args.Data.ApiValidatePath)
	}
	ot, err := stat.Save(ctx)
	if err != nil {
		return openTokenResponse{}, err
	}
	service.Cache.Set(&cache.Item{
		Ctx:   ctx,
		Key:   fmt.Sprintf("openToken:%d", ot.ID),
		Value: *ot,
		TTL:   time.Hour,
	})
	return openTokenResponse{
		openToken: ot,
	}, nil
}

type deleteOpenTokenArgs struct {
	ID int32
}

func (q QueryResolver) DeleteOpenToken(ctx context.Context, args deleteOpenTokenArgs) (bool, error) {
	ctxValue := ctx.Value(service.GinGraphQLContextKey).(service.GinGraphQLContextType)
	
	// First get the token to check its project permissions
	existingToken, err := service.EntClient.OpenToken.Get(ctx, int(args.ID))
	if err != nil {
		return false, NewGraphQLHttpError(http.StatusNotFound, err)
	}
	
	// Check RBAC permission for open token deletion
	projectID := existingToken.ProjectOpenTokens
	hasPermission, err := rbacService.HasPermission(ctx, ctxValue.UserID, &projectID, service.PermProjectEdit)
	if err != nil {
		return false, NewGraphQLHttpError(http.StatusInternalServerError, err)
	}
	if !hasPermission {
		return false, NewGraphQLHttpError(http.StatusUnauthorized, errors.New("insufficient permissions to delete open token"))
	}

	err = service.
		EntClient.
		OpenToken.
		DeleteOneID(int(args.ID)).
		Exec(ctx)
	if err != nil {
		err = NewGraphQLHttpError(http.StatusInternalServerError, err)
		return false, err
	}
	return true, nil
}

type openTokenListResponse struct {
	openTokens []*ent.OpenToken
}

func (o openTokenListResponse) Count(ctx context.Context) (int32, error) {
	return int32(len(o.openTokens)), nil
}

func (o openTokenListResponse) Edges(ctx context.Context) (result []openTokenResponse, err error) {
	for _, ot := range o.openTokens {
		result = append(result, openTokenResponse{
			openToken: ot,
		})
	}
	return
}

func (o createOpenTokenResponse) Token() string {
	return o.token
}

func (o createOpenTokenResponse) Data() openTokenResponse {
	return openTokenResponse{
		openToken: o.openToken,
	}
}

func (o openTokenResponse) ID() int32 {
	return int32(o.openToken.ID)
}

func (o openTokenResponse) Name() string {
	return o.openToken.Name
}

func (o openTokenResponse) Description() string {
	return o.openToken.Description
}

func (o openTokenResponse) ExpireAt() string {
	return o.openToken.ExpireAt.Format(time.RFC3339)
}

func (o openTokenResponse) ApiValidateEnabled() bool {
	return o.openToken.ApiValidateEnabled
}

func (o openTokenResponse) ApiValidatePath() string {
	return o.openToken.ApiValidatePath
}
