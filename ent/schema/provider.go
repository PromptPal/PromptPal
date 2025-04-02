package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"
)

// Provider holds the schema definition for the Provider entity.
type Provider struct {
	ent.Schema
}

// Fields of the Provider.
func (Provider) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").NotEmpty(),
		field.String("description").Default(""),
		field.Bool("enabled").Default(true),

		// Source type of the provider (openai, gemini, claude, deepseek, etc.)
		field.String("source").NotEmpty(),

		// Base endpoint URL for API calls
		field.String("endpoint").Default(""),

		// API key or token for authentication
		field.String("apiKey").Sensitive(),

		// Optional organization ID (for services like OpenAI)
		field.String("organizationId").Optional(),

		// Default model to use
		field.String("defaultModel").Default(""),

		// Default parameters
		field.Float("temperature").Default(1.0),
		field.Float("topP").Default(0.9),
		field.Int("maxTokens").Default(0),

		field.JSON("headers", map[string]string{}).Default(map[string]string{}).Optional(),

		// Additional configuration stored as JSON
		field.JSON("config", map[string]interface{}{}).Optional(),
	}
}

// Edges of the Provider.
func (Provider) Edges() []ent.Edge {
	return []ent.Edge{
		// A provider can be associated with many projects
		edge.From("project", Project.Type).
			Ref("provider"),

		// A provider can be associated with multiple prompts
		edge.To("prompts", Prompt.Type),

		edge.
			From("creator", User.Type).
			Ref("providers").
			Unique(),
	}
}

func (Provider) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Time{},
	}
}
