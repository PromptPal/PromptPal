package schema

import "entgo.io/ent"

// History holds the schema definition for the History entity.
type History struct {
	ent.Schema
}

// Fields of the History.
func (History) Fields() []ent.Field {
	return nil
}

// Edges of the History.
func (History) Edges() []ent.Edge {
	return nil
}
