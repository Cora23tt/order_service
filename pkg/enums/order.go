package enums

type OrderStatus string

const (
	StatusPendingPayment OrderStatus = "pending_payment"
	StatusPaid           OrderStatus = "paid"
	StatusProcessing     OrderStatus = "processing"
	StatusShipped        OrderStatus = "shipped"
	StatusDelivered      OrderStatus = "delivered"
	StatusCancelled      OrderStatus = "cancelled"
)

func (s OrderStatus) IsValid() bool {
	switch s {
	case StatusPendingPayment,
		StatusPaid,
		StatusProcessing,
		StatusShipped,
		StatusDelivered,
		StatusCancelled:
		return true
	default:
		return false
	}
}
