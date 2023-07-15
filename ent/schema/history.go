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
}

// Fields of the History.
func (History) Fields() []ent.Field {
	return []ent.Field{
		field.JSON("snapshot", PromptComplete{}),
	}
}

// Edges of the History.
func (History) Edges() []ent.Edge {
	return []ent.Edge{
		edge.
			From("modifier", User.Type).
			Ref("histories"),
		edge.
			From("prompt", User.Type).
			Ref("histories").
			Unique(),
	}
}

func (History) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Time{},
	}
}
