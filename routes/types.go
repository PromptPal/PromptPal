package routes

type paginationPayload struct {
	Pagination Pagination `json:"pagination"`
}

type Pagination struct {
	Count  int `json:"count"`
	Limit  int `json:"limit"`
	Cursor int `json:"cursor"`
}

type PromptRow struct {
	Prompt string `json:"prompt"`
	Role   string `json:"role"`
}

type PromptVariable struct {
	Name string `json:"name"`
	// string, number, bool
	Type string `json:"type"`
}
