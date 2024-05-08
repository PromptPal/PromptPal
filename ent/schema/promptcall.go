package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"
)

// metric data for prompt activity
type PromptCall struct {
	ent.Schema
}

// Fields of the Metric.
func (PromptCall) Fields() []ent.Field {
	return []ent.Field{
		field.Int("promptId").StorageKey("prompt_calls"),
		field.String("userId").Optional(),
		field.Int("responseToken"),
		field.Int("totalToken"),
		// how long the prompt executed.
		field.Int64("duration"),
		// 0: success, 1: fail
		field.Int("result"),
		field.JSON("payload", map[string]string{}).Optional(),
		field.Float("cost_cents").Default(0),
		field.String("ua").Default(""),
		// only available when prompt.debug is true
		field.String("message").Optional().Nillable(),
	}
}

// Edges of the Metric.
func (PromptCall) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("prompt", Prompt.Type).
			Ref("calls").
			Unique().
			Field("promptId").
			Required(),
		edge.From("project", Project.Type).Ref("calls").Unique(),
	}
}

func (PromptCall) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Time{},
	}
}
