package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"
)

// Prompt holds the schema definition for the Prompt entity.
type Prompt struct {
	ent.Schema
}

type PromptRow struct {
	Prompt string `json:"prompt"`
	Role   string `json:"role"`
}

type PromptVariable struct {
	Name string `json:"name"`
	// string, number, bool
	Type string `json:"type"`
}

// Fields of the Prompt.
func (Prompt) Fields() []ent.Field {
	return []ent.Field{
		field.String("name"),
		field.String("description").Default(""),
		field.Bool("enabled").Default(true),
		field.JSON("prompts", []PromptRow{}),
		field.Int("tokenCount").Default(0),
		field.JSON("variables", []PromptVariable{}),
		field.Enum("publicLevel").Values("public", "protected", "private").Default("protected"),
	}
}

// Edges of the Prompt.
func (Prompt) Edges() []ent.Edge {
	// creator, project
	return []ent.Edge{
		edge.
			From("creator", User.Type).
			Ref("prompts").
			Unique(),
		edge.
			From("project", Project.Type).
			Ref("prompts").
			Unique(),
		edge.To("histories", Prompt.Type),
	}
}

func (Prompt) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Time{},
	}
}
