package events

// typed string aliases make code self-documenting and safer
type EventType string
type AggregateType string

const (
	// Event types
	EventPaymentRequired EventType = "payment_required"
	EventOrderCreated    EventType = "order_created"
	// add more event types here...

	// Aggregate types
	AggregateOrder AggregateType = "order"
	AggregateUser  AggregateType = "user"
	// add more aggregate types here...
)

// Helpers (optional)
func (e EventType) String() string     { return string(e) }
func (a AggregateType) String() string { return string(a) }
