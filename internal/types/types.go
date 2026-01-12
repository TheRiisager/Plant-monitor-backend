package types

import (
	"time"
)

type DeviceInfo struct {
	Device string `json:"device" validate:"required"`
}

type Reading struct {
	Time         time.Time `db:"time"`
	DeviceName   string    `db:"device_name"`
	Temperature  float32   `db:"temperature"`
	SoilMoisture float32   `db:"soil_moisture"`
}
