package responses

type Message struct {
	ID string `json:"id"`
	EN string `json:"en"`
}

type ErrorContext struct {
	ErrorID int     `json:"error_id"`
	Message Message `json:"message"`
}

type GenericResponse struct {
	Success bool         `json:"success"`
	Error   ErrorContext `json:"error"`
	Data    interface{}  `json:"data,omitempty"`
}

type DataPaginationResponse struct {
	DataPage interface{} `json:"data_page,omitempty"`
	Count    int         `json:"count,omitempty"`
}

func NewGenericResponse(errorId int, data interface{}) *GenericResponse {
	messageEn := GetErrorCodeEN(errorId)
	messageId := GetErrorCodeID(errorId)
	status := false
	if errorId == 0 {
		status = true
	}

	return &GenericResponse{
		Success: status,
		Error: ErrorContext{
			ErrorID: errorId,
			Message: Message{
				ID: messageId,
				EN: messageEn,
			},
		},
		Data: data,
	}
}

func CustomGenericResponse(success bool, errorId int, errorMessageID string, errorMessageEND string) *GenericResponse {
	return &GenericResponse{
		Success: success,
		Error: ErrorContext{
			ErrorID: errorId,
			Message: Message{
				ID: errorMessageID,
				EN: errorMessageEND,
			},
		},
		Data: nil,
	}
}

func NonAuthorizationWebhookGenaricResponse() *GenericResponse {
	messageEn := GetErrorCodeEN(1001)
	messageId := GetErrorCodeID(1001)
	return &GenericResponse{
		Success: false,
		Error: ErrorContext{
			ErrorID: 1001,
			Message: Message{
				ID: messageId,
				EN: messageEn,
			},
		},
	}
}
