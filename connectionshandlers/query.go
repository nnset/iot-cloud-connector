package connectionshandlers

type Query struct {
	DeviceID string `json:"device_id"`
	Payload  string `json:"payload"`
}

func NewQuery(deviceID, payload string) Query {
	return Query{deviceID, payload}
}
