package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/mixin"
)

// Activity holds the schema definition for the Activity entity.
type Activity struct {
	ent.Schema
}

// Fields of the Activity.
func (Activity) Fields() []ent.Field {
	return nil
}

// Edges of the Activity.
func (Activity) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("project", Project.Type).Ref("activities").Unique(),
		edge.From("user", User.Type).Ref("activities").Unique(),
	}
}

func (Activity) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Time{},
	}
}
