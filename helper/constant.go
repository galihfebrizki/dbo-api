package helper

const (
	RequestIDContextKey = "request_id"
	XRequestIDHeaderKey = "X-Request-Id"
)

// topic consumer
const (
	PaymentProccess = "payment_proccess"
)

// status order
const (
	StatusCreate     = 1
	StatusReadyToPay = 2
	StatusPaid       = 3
	StatusSuccess    = 4
	StatusFailed     = 10
)
