package types

import "time"

type SubscriptionInfo struct {
	Topic string `json:"topic" validate:"required"`
}

type Reading struct {
	Time         time.Time `db:"time"`
	DeviceName   string    `db:"device_name"`
	Temperature  float32   `db:"temperature"`
	SoilMoisture float32   `db:"soil_moisture"`
}
