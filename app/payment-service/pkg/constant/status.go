package constant

type PayStatus string

const (
	InitPayStatus    = "wait_pay"
	SuceessPayStatus = "success_pay"
	FailPayStatus    = "failed_pay"
	ClosePayStatus   = "closed_pay"
)
