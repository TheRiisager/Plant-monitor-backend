package types

import (
	"sync"
	"time"
)

// todo thread safety?
type GlobalStore struct {
	Devices Devices
	Mutex   sync.RWMutex
}

type DeviceInfo struct {
	Device string `json:"device" validate:"required"`
}

type Devices []DeviceInfo

func (d *Devices) Add(item DeviceInfo) {
	*d = append(*d, item)
}

type Reading struct {
	Time         time.Time `db:"time"`
	DeviceName   string    `db:"device_name"`
	Temperature  float32   `db:"temperature"`
	SoilMoisture float32   `db:"soil_moisture"`
}
