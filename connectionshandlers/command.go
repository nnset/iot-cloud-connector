package connectionshandlers

type Command struct {
	DeviceID string `json:"device_id"`
	Payload  string `json:"payload"`
}

func NewCommand(deviceID, payload string) Command {
	return Command{deviceID, payload}
}
