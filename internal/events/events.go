package events

// typed string aliases make code self-documenting and safer
type EventType string
type AggregateType string
type PaymentStatus string

type TopicType string

const (
	// Topic types
	PaymentAuthorizationTopic TopicType = "payment_authorized"
	PromotionRewardTopic      TopicType = "promotion_rewards"

	// Event types
	EventPaymentRequired   EventType = "payment_required"
	EventOrderCreated      EventType = "order_created"
	EventPromotionRewarded EventType = "promotion_rewarded"
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
func (t TopicType) String() string     { return string(t) }

// TopicType represents the type of Kafka topic
