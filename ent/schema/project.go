package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"
)

// Project holds the schema definition for the Project entity.
type Project struct {
	ent.Schema
}

// Fields of the Project.
// TODO: add steaming control
func (Project) Fields() []ent.Field {
	return []ent.Field{
		field.String("name"),
		field.Bool("enabled").Default(true),
		field.Int("creator_id").StorageKey("user_projects").Optional(),

		// for OpenAI
		field.String("openAIBaseURL").Default("https://api.openai.com"),
		field.String("openAIToken").Default("").Sensitive(),

		// for Google Gemini
		field.String("geminiBaseURL").Default("https://generativelanguage.googleapis.com"),
		field.String("geminiToken").Default("").Sensitive(),

		// the 4 below are for common use. the name just because of legacy design
		field.String("openAIModel").Default("gpt-3.5-turbo"),
		field.Float("openAITemperature").Default(1),
		field.Float("openAITopP").Default(0.9),
		field.Int("openAIMaxTokens").Default(0),
		field.Int("providerId").Optional().Nillable().StorageKey("project_provider"),
	}
}

// Edges of the Project.
func (Project) Edges() []ent.Edge {
	return []ent.Edge{
		edge.
			From("creator", User.Type).
			Ref("projects").
			Unique().
			Field("creator_id"),
		edge.To("prompts", Prompt.Type),
		edge.To("activities", Activity.Type),
		edge.To("openTokens", OpenToken.Type),
		edge.To("calls", PromptCall.Type),
		edge.To("provider", Provider.Type).
			Unique().
			Field("providerId"),
		edge.To("userProjectRoles", UserProjectRole.Type),
		edge.To("webhooks", Webhook.Type),
	}
}

func (Project) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Time{},
	}
}
