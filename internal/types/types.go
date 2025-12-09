package types

type SubscriptionInfo struct {
	Topic string `json:"topic" validate:"required"`
}

type Reading struct {
	Temperature float32
	DeviceName  string
}
