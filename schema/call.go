package schema

import (
	"context"
	"net/http"
	"time"

	"github.com/PromptPal/PromptPal/ent"
	"github.com/PromptPal/PromptPal/ent/prompt"
	"github.com/PromptPal/PromptPal/ent/promptcall"
	"github.com/PromptPal/PromptPal/service"
)

//   calls(promptId: Int!, pagination: PaginationInput!): PromptCallList!

type callsArgs struct {
	PromptID   int32
	Pagination paginationInput
}

type promptCallListResponse struct {
	stat       *ent.PromptCallQuery
	pagination paginationInput
}

func (q QueryResolver) Calls(ctx context.Context, args callsArgs) (res promptCallListResponse) {
	stat := service.EntClient.PromptCall.Query().
		Where(promptcall.HasPromptWith(prompt.ID(int(args.PromptID)))).
		Order(ent.Desc(promptcall.FieldID))
	res.stat = stat
	res.pagination = args.Pagination
	return
}

func (p promptCallListResponse) Count(ctx context.Context) (int32, error) {
	count, err := p.stat.Clone().Count(ctx)
	if err != nil {
		err = NewGraphQLHttpError(http.StatusInternalServerError, err)
		return 0, err
	}
	return int32(count), nil
}

type promptCallResponse struct {
	pc *ent.PromptCall
}

func (p promptCallListResponse) Edges(ctx context.Context) (res []promptCallResponse, err error) {
	ps, err := p.stat.Clone().
		Limit(int(p.pagination.Limit)).
		Offset(int(p.pagination.Offset)).
		All(ctx)

	if err != nil {
		err = NewGraphQLHttpError(http.StatusInternalServerError, err)
		return
	}

	for _, p := range ps {
		res = append(res, promptCallResponse{p})
	}

	return
}

func (p promptCallResponse) ID() int32 {
	return int32(p.pc.ID)
}
func (p promptCallResponse) UserId() string {
	return p.pc.UserId
}
func (p promptCallResponse) ResponseToken() int32 {
	return int32(p.pc.ResponseToken)
}
func (p promptCallResponse) TotalToken() int32 {
	return int32(p.pc.TotalToken)
}
func (p promptCallResponse) Duration() int32 {
	return int32(p.pc.Duration)
}
func (p promptCallResponse) Result() string {
	result := p.pc.Result
	if result == 0 {
		return "success"
	}
	return "fail"
}
func (p promptCallResponse) Message() *string {
	return p.pc.Message
}
func (p promptCallResponse) CreatedAt() string {
	return p.pc.CreateTime.Format(time.RFC3339)
}
