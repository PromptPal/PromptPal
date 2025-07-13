package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"
)

// Permission holds the schema definition for the Permission entity.
type Permission struct {
	ent.Schema
}

// Fields of the Permission.
func (Permission) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").Unique(),
		field.String("description").Optional(),
		field.String("resource").Optional(), // e.g., "project", "prompt", "system"
		field.String("action").Optional(),   // e.g., "read", "write", "delete", "admin"
	}
}

// Edges of the Permission.
func (Permission) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("roles", Role.Type).
			Ref("permissions"),
	}
}

func (Permission) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Time{},
	}
}