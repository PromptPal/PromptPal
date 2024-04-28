package schema

import (
	"context"
	"net/http"
	"time"

	"github.com/PromptPal/PromptPal/ent"
	"github.com/PromptPal/PromptPal/ent/history"
	"github.com/PromptPal/PromptPal/ent/prompt"
	"github.com/PromptPal/PromptPal/ent/promptcall"
	"github.com/PromptPal/PromptPal/service"
)

func (p promptResponse) Histories(ctx context.Context) (res promptHistoryResp, err error) {
	stat := service.
		EntClient.
		History.
		Query().
		Where(history.PromptId(int(p.prompt.ID))).
		Order(ent.Desc(history.FieldID))
	res.promptID = int(p.prompt.ID)
	res.pagination = paginationInput{
		Offset: 0,
		Limit:  10,
	}
	res.stat = stat
	return
}

type promptHistoryResp struct {
	promptID   int
	stat       *ent.HistoryQuery
	pagination paginationInput
}

func (p promptHistoryResp) Edges(ctx context.Context) (res []promptHistory, err error) {
	histories, err := p.stat.
		Clone().
		Limit(int(p.pagination.Limit)).
		Offset(int(p.pagination.Offset)).
		All(ctx)

	if err != nil {
		err = NewGraphQLHttpError(http.StatusInternalServerError, err)
		return
	}

	for _, v := range histories {
		res = append(res, promptHistory{
			snapshot: v,
		})
	}
	return
}

func (p promptHistoryResp) Count(ctx context.Context) (int32, error) {
	count, err := p.stat.Clone().Count(ctx)
	if err != nil {
		err = NewGraphQLHttpError(http.StatusInternalServerError, err)
		return 0, err
	}
	return int32(count), nil
}

type promptHistory struct {
	snapshot *ent.History
}

func (p promptHistory) ID() int32 {
	return int32(p.snapshot.ID)
}

func (p promptHistory) Name() string {
	return p.snapshot.Snapshot.Name
}

func (p promptHistory) Description() string {
	return p.snapshot.Snapshot.Description
}

func (p promptHistory) Prompts() []promptRowResponse {
	result := make([]promptRowResponse, len(p.snapshot.Snapshot.Prompts))
	for i, v := range p.snapshot.Snapshot.Prompts {
		result[i] = promptRowResponse{
			p: v,
		}
	}
	return result
}

func (p promptHistory) Variables() []promptVariableResponse {
	result := make([]promptVariableResponse, len(p.snapshot.Snapshot.Variables))
	for i, v := range p.snapshot.Snapshot.Variables {
		result[i] = promptVariableResponse{
			p: v,
		}
	}
	return result
}

func (p promptHistory) ModifiedBy(ctx context.Context) (userResponse, error) {
	uid := p.snapshot.ModifierId
	u, err := service.EntClient.User.Get(ctx, uid)
	if err != nil {
		return userResponse{}, err
	}
	return userResponse{
		u: u,
	}, nil
}

func (p promptHistory) CreatedAt() string {
	return p.snapshot.CreateTime.Format(time.RFC3339)
}

func (p promptHistory) UpdatedAt() string {
	return p.snapshot.UpdateTime.Format(time.RFC3339)
}

func (p promptHistory) LatestCalls(ctx context.Context) (res promptCallListResponse, err error) {
	pid := p.snapshot.PromptId

	// find the previous snapshot, use the created time of the previous snapshot as the start time of this snapshot
	previousSnapshot, err := service.EntClient.History.Query().
		Where(history.PromptId(pid)).
		Where(history.IDLT(int(p.ID()))).
		Order(ent.Desc(history.FieldID)).
		First(ctx)
	startDateOfThisSnapshot := time.Time{}
	if err != nil && !ent.IsNotFound(err) {
		err = NewGraphQLHttpError(http.StatusInternalServerError, err)
		return
	}
	if previousSnapshot != nil {
		startDateOfThisSnapshot = previousSnapshot.CreateTime
	}
	// if no previous snapshot, use the Time.Zero
	if ent.IsNotFound(err) {
		err = nil
	}

	stat := service.
		EntClient.
		PromptCall.
		Query().
		Where(promptcall.HasPromptWith(prompt.ID(pid))).
		Where(promptcall.CreateTimeLTE(p.snapshot.CreateTime)).
		Where(promptcall.CreateTimeGTE(startDateOfThisSnapshot)).
		Order(ent.Desc(promptcall.FieldID))
	res.stat = stat
	res.pagination = paginationInput{
		// only show 10 latest calls on each snapshot
		Limit:  10,
		Offset: 0,
	}
	return
}
