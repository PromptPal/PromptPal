package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"
)

// OpenToken holds the schema definition for the OpenToken entity.
type OpenToken struct {
	ent.Schema
}

// Fields of the OpenToken.
func (OpenToken) Fields() []ent.Field {
	return []ent.Field{
		field.String("name"),
		field.String("description").Default(""),
		field.String("token").Sensitive(),
		field.Bool("apiValidateEnabled").Default(false),
		field.String("apiValidatePath").Default("/api/v1/validate"),
		field.Time("expireAt"),
	}
}

// Edges of the OpenToken.
func (OpenToken) Edges() []ent.Edge {
	return []ent.Edge{
		edge.
			From("project", Project.Type).
			Ref("openTokens").
			Unique(),
		edge.From("user", User.Type).Ref("openTokens").Unique(),
	}
}

func (OpenToken) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Time{},
	}
}
