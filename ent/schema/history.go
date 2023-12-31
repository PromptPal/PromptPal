package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"
)

// History holds the schema definition for the History entity.
type History struct {
	ent.Schema
}

type PromptComplete struct {
	Name        string
	Description string
	Enabled     bool
	Debug       bool
	Prompts     []PromptRow
	TokenCount  int
	Variables   []PromptVariable
	PublicLevel string
	Version     int
}

// Fields of the History.
func (History) Fields() []ent.Field {
	return []ent.Field{
		field.Int("modifierId"),
		field.Int("promptId"),
		field.JSON("snapshot", PromptComplete{}),
	}
}

// Edges of the History.
func (History) Edges() []ent.Edge {
	return []ent.Edge{
		edge.
			From("modifier", User.Type).
			Ref("histories").
			Unique().
			Required().
			Field("modifierId"),
		edge.
			From("prompt", Prompt.Type).
			Ref("histories").
			Unique().
			Required().
			Field("promptId"),
	}
}

func (History) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Time{},
	}
}
