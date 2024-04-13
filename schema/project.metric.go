package schema

import (
	"context"
	"net/http"
	"time"

	"entgo.io/ent/dialect/sql"
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

type ProjectPromptMetricsByDate struct {
	date  time.Time
	datum []projectPromptMetricsRecentCount
}

func (p ProjectPromptMetricsByDate) Date() string {
	return p.date.Format(time.RFC3339)
}

func (p ProjectPromptMetricsByDate) Prompts() []projectPromptMetricsRecentCount {
	return p.datum
}

func (p projectPromptMetricsResponse) Last7Days(ctx context.Context) (res []ProjectPromptMetricsByDate, err error) {
	var result []struct {
		Date  string `json:"d"`
		Pid   int    `json:"prompt_calls"`
		Count int    `json:"count"`
	}

	err = service.
		EntClient.
		PromptCall.
		Query().
		Where(promptcall.HasProjectWith(project.ID(p.p.ID))).
		Where(promptcall.CreateTimeGT(time.Now().AddDate(0, 0, -7))).
		Select(promptcall.FieldPromptId).
		Aggregate(func(s *sql.Selector) string {
			return sql.As("DATE(create_time)", "d")
		}, ent.Count()).
		Modify(func(s *sql.Selector) {
			s.GroupBy("d", promptcall.FieldPromptId)
			s.OrderBy(sql.Desc("d"))
		}).
		Scan(ctx, &result)

	if err != nil {
		err = NewGraphQLHttpError(http.StatusInternalServerError, err)
		return
	}

	if len(result) == 0 {
		return
	}

	promptIds := make([]int, 0)

	for _, r := range result {
		promptIds = append(promptIds, r.Pid)
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

	psMap := make(map[int]*ent.Prompt)
	for _, p := range ps {
		psMap[p.ID] = p
	}

	reply := map[time.Time][]projectPromptMetricsRecentCount{}

	for _, r := range result {
		dd, _ := time.Parse(time.RFC3339, r.Date)
		reply[dd] = append(reply[dd], projectPromptMetricsRecentCount{
			p:     psMap[r.Pid],
			count: r.Count,
		})
	}

	for d, ps := range reply {
		res = append(res, ProjectPromptMetricsByDate{
			date:  d,
			datum: ps,
		})
	}

	return
}

func (p projectPromptMetricsRecentCount) Prompt() promptResponse {
	return promptResponse{p.p}
}
func (p projectPromptMetricsRecentCount) Count() int32 {
	return int32(p.count)
}
