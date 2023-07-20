package models

type PaymentResult struct {
	OrderId       string  `json:"order_id"`
	TotalAmount   float64 `json:"total_amount"`
	PaymentMethod string  `json:"payment_method"`
	AcquirementId string  `json:"acquirement_id"`
}

type PaymentOrder struct {
	OrderId string `json:"order_id"`
}
