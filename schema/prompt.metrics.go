package schema

import (
	"context"
	"net/http"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/PromptPal/PromptPal/ent/prompt"
	"github.com/PromptPal/PromptPal/ent/promptcall"
	"github.com/PromptPal/PromptPal/service"
)

type promptCallMetricBySQL struct {
	P50 float32 `json:"p50"`
	P90 float32 `json:"p90"`
	P99 float32 `json:"p99"`
}

type promptMetrics struct {
	data *promptCallMetricBySQL
}

func (p promptResponse) Metrics(ctx context.Context) (res promptMetrics, err error) {

	var temp []promptCallMetricBySQL

	err = service.
		EntClient.
		PromptCall.
		Query().
		Where(promptcall.HasPromptWith(prompt.ID(int(p.prompt.ID)))).
		Where(promptcall.CreateTimeGT(time.Now().AddDate(-1, 0, 0))).
		Aggregate(func(s *sql.Selector) string {
			return sql.As("percentile_cont(0.5) within group (order by duration asc)", "p50")
		}, func(s *sql.Selector) string {
			return sql.As("percentile_cont(0.9) within group (order by duration asc)", "p90")
		}, func(s *sql.Selector) string {
			return sql.As("percentile_cont(0.99) within group (order by duration asc)", "p99")
		}).
		Scan(ctx, &temp)

	if err != nil {
		err = NewGraphQLHttpError(http.StatusInternalServerError, err)
		return
	}

	if len(temp) > 0 {
		res.data = &temp[0]
	}
	return
}

func (p promptMetrics) P50() float64 {
	return float64(p.data.P50)
}
func (p promptMetrics) P90() float64 {
	return float64(p.data.P90)
}
func (p promptMetrics) P99() float64 {
	return float64(p.data.P99)
}
