package schema

import "entgo.io/ent"

// Project holds the schema definition for the Project entity.
type Project struct {
	ent.Schema
}

// Fields of the Project.
func (Project) Fields() []ent.Field {
	return nil
}

// Edges of the Project.
func (Project) Edges() []ent.Edge {
	return nil
}
