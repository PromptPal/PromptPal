package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"entgo.io/ent/schema/mixin"
)

// CostReport holds aggregated cost data by month for performance optimization
type CostReport struct {
	ent.Schema
}

// Fields of the CostReport.
func (CostReport) Fields() []ent.Field {
	return []ent.Field{
		// Month in YYYY-MM format
		field.String("month").
			Comment("Month in YYYY-MM format (e.g., '2024-01')"),
		// User ID to scope the report
		field.String("userId").
			Comment("User ID this report belongs to"),
		// Total costs for the month
		field.Float("totalCostCents").
			Default(0).
			Comment("Total cost in cents for the month"),
		// Aggregated costs by different dimensions as JSON
		field.JSON("costsByProvider", map[string]float64{}).
			Default(map[string]float64{}).
			Comment("Costs grouped by provider ID -> cost in cents"),
		field.JSON("costsByProject", map[string]float64{}).
			Default(map[string]float64{}).
			Comment("Costs grouped by project ID -> cost in cents"),
		field.JSON("costsByPrompt", map[string]float64{}).
			Default(map[string]float64{}).
			Comment("Costs grouped by prompt ID -> cost in cents"),
		field.JSON("costsByDay", map[string]float64{}).
			Default(map[string]float64{}).
			Comment("Costs grouped by day (YYYY-MM-DD) -> cost in cents"),
		// Additional metrics
		field.Int("totalCalls").
			Default(0).
			Comment("Total number of prompt calls in the month"),
		field.Int("totalTokens").
			Default(0).
			Comment("Total tokens consumed in the month"),
		field.Int("successfulCalls").
			Default(0).
			Comment("Number of successful calls (result = 0)"),
		field.Int("cachedCalls").
			Default(0).
			Comment("Number of cached calls"),
	}
}

// Edges of the CostReport.
func (CostReport) Edges() []ent.Edge {
	return []ent.Edge{}
}

// Indexes of the CostReport.
func (CostReport) Indexes() []ent.Index {
	return []ent.Index{
		// Unique index on userId + month combination
		index.Fields("userId", "month").Unique(),
		// Index for efficient querying by month range
		index.Fields("month"),
		// Index for efficient querying by user
		index.Fields("userId"),
	}
}

func (CostReport) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Time{},
	}
}