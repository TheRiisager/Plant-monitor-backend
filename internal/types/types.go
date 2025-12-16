package types

import (
	"time"
)

// todo thread safety?
// todo consider moving pgx pool here, maybe with a wrapper struct containing the needed functions?
type GlobalStore struct {
	Devices Devices
}

type SubscriptionInfo struct {
	Device string `json:"device" validate:"required"`
}

type Devices []SubscriptionInfo

func (d *Devices) Add(item SubscriptionInfo) {
	*d = append(*d, item)
}

type Reading struct {
	Time         time.Time `db:"time"`
	DeviceName   string    `db:"device_name"`
	Temperature  float32   `db:"temperature"`
	SoilMoisture float32   `db:"soil_moisture"`
}
