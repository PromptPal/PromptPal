package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
)

// Prompt holds the schema definition for the Prompt entity.
type Prompt struct {
	ent.Schema
}

type PromptRow struct {
	Prompt string `json:"prompt"`
	Role   string `json:"role"`
}

// Fields of the Prompt.
func (Prompt) Fields() []ent.Field {
	return []ent.Field{
		field.JSON("prompts", []PromptRow{}),
		field.Enum("publicLevel").Values("public", "protected", "private"),
	}
}

// Edges of the Prompt.
func (Prompt) Edges() []ent.Edge {
	// creator, project
	return nil
}
