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

type PromptVariableTypes string

const (
	PromptVariableTypesString  PromptVariableTypes = "string"
	PromptVariableTypesNumber  PromptVariableTypes = "number"
	PromptVariableTypesBoolean PromptVariableTypes = "boolean"
	PromptVariableTypesVideo   PromptVariableTypes = "video"
	PromptVariableTypesAudio   PromptVariableTypes = "audio"
	PromptVariableTypesImage   PromptVariableTypes = "image"
)

type PromptVariable struct {
	Name string `json:"name"`
	// string, number, bool
	Type PromptVariableTypes `json:"type"`
}

// Fields of the Prompt.
func (Prompt) Fields() []ent.Field {
	return []ent.Field{
		field.String("name"),
		field.String("description").Default(""),
		field.Bool("enabled").Default(true),
		field.Bool("debug").Default(false),
		field.Bool("cacheEnabled").Default(true),
		field.JSON("prompts", []PromptRow{}),
		field.Int("tokenCount").Default(0),
		field.Int("version").Default(0),
		field.JSON("variables", []PromptVariable{}),
		field.Int("projectId").StorageKey("project_prompts"),
		field.Int("providerId").Optional().StorageKey("provider_prompts"),
		field.Enum("publicLevel").
			Values("public", "protected", "private").
			Default("protected"),
	}
}

// Edges of the Prompt.
func (Prompt) Edges() []ent.Edge {
	// creator, project, provider
	return []ent.Edge{
		edge.
			From("creator", User.Type).
			Ref("prompts").
			Unique(),
		edge.
			From("project", Project.Type).
			Ref("prompts").
			Unique().
			Field("projectId").
			Required(),
		// edge.
		// 	From("provider", Provider.Type).
		// 	Ref("prompts").
		// 	Unique(),
		edge.To("provider", Provider.Type).
			Unique().
			Field("providerId"),
		edge.To("calls", PromptCall.Type),
		edge.To("histories", History.Type),
	}
}

func (Prompt) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Time{},
	}
}
