package types

import (
	"slices"
	"sync"
)

type GlobalStore struct {
	Devices Devices
	Mutex   sync.RWMutex
}

func (gs *GlobalStore) DeviceExists(deviceName string) bool {
	gs.Mutex.RLock()
	defer gs.Mutex.RUnlock()
	index := slices.IndexFunc(gs.Devices, func(dev DeviceInfo) bool {
		return dev.Device == deviceName
	})

	if index < 0 {
		return false
	}

	return true
}
