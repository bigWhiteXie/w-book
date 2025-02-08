package constant

type PayStatus string

const (
	InitPayStatus      = "init_pay"
	SuceessPayStatus   = "success_pay"
	FailPayStatus      = "failed_pay"
	ClosePayStatus     = "closed_pay"
	RefundingStatus    = "refunding_pay"
	RefundedStatus     = "refunded_pay"
	RefundFailedStatus = "refund_failed"
)
