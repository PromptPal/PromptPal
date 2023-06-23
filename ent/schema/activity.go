package schema

import "entgo.io/ent"

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
	return nil
}
