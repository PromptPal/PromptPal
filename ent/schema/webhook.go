package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"
)

// Webhook holds the schema definition for the Webhook entity.
type Webhook struct {
	ent.Schema
}

// Fields of the Webhook.
func (Webhook) Fields() []ent.Field {
	return []ent.Field{
		field.String("name"),
		field.String("description").Default(""),
		field.String("url"),
		field.String("event").Default("onPromptFinished"),
		field.Bool("enabled").Default(true),
		field.Int("creator_id").StorageKey("webhook_creator"),
		field.Int("project_id").StorageKey("webhook_project"),
	}
}

// Edges of the Webhook.
func (Webhook) Edges() []ent.Edge {
	return []ent.Edge{
		edge.
			From("creator", User.Type).
			Ref("webhooks").
			Unique().
			Field("creator_id").
			Required(),
		edge.
			From("project", Project.Type).
			Ref("webhooks").
			Unique().
			Field("project_id").
			Required(),
		edge.
			To("calls", WebhookCall.Type),
	}
}

func (Webhook) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Time{},
	}
}