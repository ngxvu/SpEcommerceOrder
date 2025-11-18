package events

// typed string aliases make code self-documenting and safer
type EventType string
type AggregateType string

type PaymentStatus string

const (
	// Event types
	EventPaymentRequired EventType = "payment_required"
	EventOrderCreated    EventType = "order_created"
	// add more event types here...

	// Aggregate types
	AggregateOrder AggregateType = "order"
	AggregateUser  AggregateType = "user"
	// add more aggregate types here...

	// PaymentPending describes a payment that is created but not yet processed
	PaymentPending PaymentStatus = "PENDING"
	// PaymentAuthorized describes a payment that has been successfully authorized
	PaymentAuthorized PaymentStatus = "AUTHORIZED"
	// PaymentDeclined describes a payment that has been declined
	PaymentDeclined PaymentStatus = "DECLINED"
)

// Helpers (optional)
func (e EventType) String() string     { return string(e) }
func (a AggregateType) String() string { return string(a) }
