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
