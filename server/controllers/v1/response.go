package v1

type V1SuccessResponse struct {
	Status string `json:"status" example:"success" enums:"success,error"`
	Data   any    `json:"data,omitempty"`
}

type V1ErrorResponse struct {
	Status string `json:"status" example:"error" enums:"success,error"`
	Error  string `json:"error" example:"error message"`
}
