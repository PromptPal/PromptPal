package schema

import "fmt"

type GraphQLHttpError struct {
	code      int
	originErr error
}

func NewGraphQLHttpError(code int, err error) error {
	if err == nil {
		return nil
	}
	return GraphQLHttpError{
		code:      code,
		originErr: err,
	}
}

func (e GraphQLHttpError) Error() string {
	return fmt.Sprintf("[%d]: %s", e.code, e.originErr.Error())
}

func (e GraphQLHttpError) Unwrap() error {
	return e.originErr
}

func (e *GraphQLHttpError) Is(target error) bool {
	t, ok := target.(*GraphQLHttpError)
	if !ok {
		return false
	}

	return e.code == t.code
}

func (e GraphQLHttpError) Extensions() map[string]interface{} {
	return map[string]interface{}{
		"code":    e.code,
		"message": e.originErr.Error(),
	}
}
