package routes

type paginationPayload struct {
	Pagination Pagination `json:"pagination"`
}

type queryPagination struct {
	Limit  int `json:"limit" form:"limit" query:"limit" binding:"lte=100,gt=0"`
	Cursor int `json:"cursor" form:"cursor" query:"cursor" binding:"gte=0"`
}

type Pagination struct {
	Count  int `json:"count"`
	Limit  int `json:"limit"`
	Cursor int `json:"cursor"`
}

type ListResponse[T any] struct {
	Count int `json:"count"`
	Data  []T `json:"data"`
}
