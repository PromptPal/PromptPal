package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"
)

// User holds the schema definition for the User entity.
type User struct {
	ent.Schema
}

// Fields of the User.
func (User) Fields() []ent.Field {
	return []ent.Field{
		field.String("name"),
		field.String("addr").Unique(),
		field.String("avatar").Default(""),
		field.String("email"),
		field.String("phone"),
		field.String("lang"),
		field.Uint8("level"), // 255: admin
		field.String("source").Default("web3"),
		field.Text("ssoInfo").Default("{}"),
	}
}

// Edges of the User.
func (User) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("projects", Project.Type),
		edge.To("prompts", Prompt.Type),
		edge.To("activities", Activity.Type),
		edge.To("histories", History.Type),
		edge.To("openTokens", OpenToken.Type),
		edge.To("providers", Provider.Type),
	}
}

func (User) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Time{},
	}
}
