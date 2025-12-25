package httpServer

type QueryTimeSpanByDeviceRequest struct {
	Timespan string `json:"timespan" validate:"required"`
}
