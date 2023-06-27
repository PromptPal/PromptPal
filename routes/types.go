package routes

type paginationPayload struct {
	Pagination Pagination `json:"pagination"`
}

type Pagination struct {
	Count  int `json:"count"`
	Limit  int `json:"limit"`
	Cursor int `json:"cursor"`
}
