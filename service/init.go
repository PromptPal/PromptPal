package service

const (
	GinGraphQLContextKey = "gin-gql-ctx-key"
)

type GinGraphQLContextType struct {
	UserID int
}
