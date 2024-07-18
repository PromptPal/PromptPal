package service

type APIRunPromptResponse struct {
	PromptID           string `json:"id"`
	ResponseMessage    string `json:"message"`
	ResponseTokenCount int    `json:"tokenCount"`
}
