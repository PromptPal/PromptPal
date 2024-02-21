package schema

import (
	"context"
	"errors"
	"net/http"
	"time"

	cache "github.com/Code-Hex/go-generics-cache"
	"github.com/PromptPal/PromptPal/service"
)

type createProjectData struct {
	Name    *string
	Enabled *bool

	// OpenAI
	OpenAIBaseURL *string
	OpenAIToken   *string

	// gemini
	GeminiBaseURL *string
	GeminiToken   *string

	// common
	OpenAIModel       *string
	OpenAITemperature *float64
	OpenAITopP        *float64
	OpenAIMaxTokens   *int32
}

type createProjectArgs struct {
	Data createProjectData
}

func (q QueryResolver) CreateProject(ctx context.Context, args createProjectArgs) (projectResponse, error) {
	data := args.Data
	if data.Name == nil {
		return projectResponse{}, NewGraphQLHttpError(http.StatusBadRequest, errors.New("name is required"))
	}
	if data.OpenAIToken == nil && data.GeminiToken == nil {
		return projectResponse{}, NewGraphQLHttpError(http.StatusBadRequest, errors.New("openAIToken is required"))
	}
	ctxValue := ctx.Value(service.GinGraphQLContextKey).(service.GinGraphQLContextType)
	stat := service.
		EntClient.
		Project.
		Create().
		SetName(*data.Name).
		SetNillableOpenAIBaseURL(data.OpenAIBaseURL).
		SetNillableOpenAIToken(data.OpenAIToken).
		SetNillableGeminiBaseURL(data.GeminiBaseURL).
		SetNillableGeminiToken(data.GeminiToken).
		SetNillableEnabled(data.Enabled).
		SetNillableOpenAIModel(data.OpenAIModel).
		SetNillableOpenAITemperature(data.OpenAITemperature).
		SetNillableOpenAITopP(data.OpenAITopP)

	if data.OpenAIMaxTokens != nil {
		stat = stat.SetOpenAIMaxTokens(int(*data.OpenAIMaxTokens))
	}

	pj, err := stat.
		SetCreatorID(ctxValue.UserID).
		Save(ctx)

	if err != nil {
		return projectResponse{}, NewGraphQLHttpError(http.StatusInternalServerError, err)
	}

	return projectResponse{p: pj}, nil
}

type updateProjectArgs struct {
	ID   int32
	Data createProjectData
}

func (q QueryResolver) UpdateProject(ctx context.Context, args updateProjectArgs) (projectResponse, error) {
	updater := service.EntClient.Project.UpdateOneID(int(args.ID))

	if args.Data.Enabled != nil {
		updater = updater.SetEnabled(*args.Data.Enabled)
	}
	if args.Data.OpenAIBaseURL != nil {
		updater = updater.SetOpenAIBaseURL(*args.Data.OpenAIBaseURL)
	}
	if args.Data.OpenAIToken != nil {
		updater = updater.SetOpenAIToken(*args.Data.OpenAIToken)
	}
	if args.Data.GeminiBaseURL != nil {
		updater = updater.SetGeminiBaseURL(*args.Data.GeminiBaseURL)
	}
	if args.Data.GeminiToken != nil {
		updater = updater.SetGeminiToken(*args.Data.GeminiToken)
	}
	if args.Data.OpenAIModel != nil {
		updater = updater.SetOpenAIModel(*args.Data.OpenAIModel)
	}
	if args.Data.OpenAITemperature != nil {
		updater = updater.SetOpenAITemperature(*args.Data.OpenAITemperature)
	}
	if args.Data.OpenAITopP != nil {
		updater = updater.SetOpenAITopP(*args.Data.OpenAITopP)
	}
	if args.Data.OpenAIMaxTokens != nil {
		updater = updater.SetOpenAIMaxTokens(int(*args.Data.OpenAIMaxTokens))
	}

	pj, err := updater.Save(ctx)
	if err != nil {
		return projectResponse{}, NewGraphQLHttpError(http.StatusInternalServerError, err)
	}
	service.ProjectCache.Set(pj.ID, *pj, cache.WithExpiration(time.Hour*24))
	return projectResponse{
		p: pj,
	}, nil
}

type deleteProjectArgs struct {
	ID int32
}

func (q QueryResolver) DeleteProject(ctx context.Context, args deleteProjectArgs) (bool, error) {
	err := service.EntClient.Project.DeleteOneID(int(args.ID)).Exec(ctx)
	if err != nil {
		return false, NewGraphQLHttpError(http.StatusInternalServerError, err)
	}
	return true, nil
}
