package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"entgo.io/ent/schema/mixin"
)

// UserProjectRole holds the schema definition for the UserProjectRole entity.
// This is a junction table that connects Users, Projects, and Roles.
type UserProjectRole struct {
	ent.Schema
}

// Fields of the UserProjectRole.
func (UserProjectRole) Fields() []ent.Field {
	return []ent.Field{
		field.Int("user_id").StorageKey("user_id"),
		field.Int("project_id").StorageKey("project_id"),
		field.Int("role_id").StorageKey("role_id"),
	}
}

// Edges of the UserProjectRole.
func (UserProjectRole) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).
			Ref("userProjectRoles").
			Unique().
			Required().
			Field("user_id"),
		edge.From("project", Project.Type).
			Ref("userProjectRoles").
			Unique().
			Required().
			Field("project_id"),
		edge.From("role", Role.Type).
			Ref("userProjectRoles").
			Unique().
			Required().
			Field("role_id"),
	}
}

// Indexes of the UserProjectRole.
func (UserProjectRole) Indexes() []ent.Index {
	return []ent.Index{
		// Ensure unique user-project-role combination
		index.Fields("user_id", "project_id", "role_id").Unique(),
		// Index for quick lookups
		index.Fields("user_id", "project_id"),
		index.Fields("project_id"),
		index.Fields("role_id"),
	}
}

func (UserProjectRole) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Time{},
	}
}