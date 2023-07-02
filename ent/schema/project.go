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
		field.String("openAIBaseURL").Default("https://api.openai.com/v1"),
		field.String("openAIModel").Default("gpt-3.5-turbo"),
		field.String("openAIToken").Default("").Sensitive(),
		field.Float("openAITemperature").Default(1),
		field.Float("openAITopP").Default(0.9),
		field.Int("openAIMaxTokens").Default(0),
	}
}

// Edges of the Project.
func (Project) Edges() []ent.Edge {
	return []ent.Edge{
		edge.
			From("creator", User.Type).
			Ref("projects").
			Unique(),
		edge.To("prompts", Prompt.Type),
		edge.To("activities", Activity.Type),
		edge.To("openTokens", OpenToken.Type),
	}
}

func (Project) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Time{},
	}
}
