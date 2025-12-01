package types

type SubscriptionInfo struct {
	Topic string `json:"topic" validate:"required"`
}
