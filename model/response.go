package model

type response struct {
	Status  int         `json:"status"`
	Message string      `json:"message"`
	Result  interface{} `json:"result"`
}

func NewResponse(status int, message string, result interface{}) *response {
	return &response{status, message, result}
}
