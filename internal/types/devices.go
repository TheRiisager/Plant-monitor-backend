package types

type Devices []DeviceInfo

func (d *Devices) Add(item DeviceInfo) {
	*d = append(*d, item)
}
