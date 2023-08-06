package schema

import (
	"context"
	"net/http"
	"time"

	"github.com/PromptPal/PromptPal/ent"
	"github.com/PromptPal/PromptPal/ent/project"
	"github.com/PromptPal/PromptPal/ent/prompt"
	"github.com/PromptPal/PromptPal/ent/promptcall"
	"github.com/PromptPal/PromptPal/service"
)

type projectPromptMetricsResponse struct {
	p *ent.Project
}

func (p projectResponse) PromptMetrics() projectPromptMetricsResponse {
	return projectPromptMetricsResponse{p: p.p}
}

type projectPromptMetricsRecentCount struct {
	p     *ent.Prompt
	count int
}

func (p projectPromptMetricsResponse) RecentCounts(ctx context.Context) (res []projectPromptMetricsRecentCount, err error) {
	var result []struct {
		PromptId int `json:"prompt_calls"`
		Count    int `json:"count"`
	}

	err = service.
		EntClient.
		PromptCall.
		Query().
		Where(promptcall.HasProjectWith(project.ID(p.p.ID))).
		Where(promptcall.CreateTimeGT(time.Now().AddDate(0, 0, -7))).
		GroupBy(promptcall.FieldPromptId).
		Aggregate(ent.Count()).
		Scan(ctx, &result)

	if err != nil {
		err = NewGraphQLHttpError(http.StatusInternalServerError, err)
		return
	}

	promptIds := make([]int, 0)

	for _, r := range result {
		promptIds = append(promptIds, r.PromptId)
	}

	ps, err := service.
		EntClient.
		Prompt.
		Query().
		Where(prompt.IDIn(promptIds...)).
		All(ctx)

	if err != nil {
		err = NewGraphQLHttpError(http.StatusInternalServerError, err)
		return
	}

	for _, r := range ps {
		count := 0

		for _, pc := range result {
			if pc.PromptId == r.ID {
				count = pc.Count
				break
			}
		}

		res = append(res, projectPromptMetricsRecentCount{p: r, count: count})
	}

	return
}

func (p projectPromptMetricsRecentCount) Prompt() promptResponse {
	return promptResponse{p.p}
}
func (p projectPromptMetricsRecentCount) Count() int32 {
	return int32(p.count)
}
