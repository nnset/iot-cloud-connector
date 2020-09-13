package connections

type DeviceConnectionDTO struct {
	ConnectionID                 string `json:"connection_id"`
	DeviceID                     string `json:"device_id"`
	DeviceName                   string `json:"device_name"`
	DeviceType                   string `json:"device_type"`
	UserAgent                    string `json:"user_agent"`
	RemoteAddress                string `json:"remote_address"`
	CreatedAt                    int64  `json:"created_at"`
	LastReceivedMessageTimeStamp int64  `json:"last_received_message"`
	LastSentMessageTimeStamp     int64  `json:"last_sent_message"`
	ReceivedMessages             uint   `json:"received_messages"`
	SentMessages                 uint   `json:"sent_messages"`
}

func NewDeviceConnectionDTOFromDeviceConnection(conn *DeviceConnection) *DeviceConnectionDTO {
	return &DeviceConnectionDTO{
		ConnectionID:                 conn.ConnectionID(),
		DeviceID:                     conn.DeviceID(),
		DeviceName:                   conn.DeviceName(),
		DeviceType:                   conn.DeviceType(),
		UserAgent:                    conn.UserAgent(),
		RemoteAddress:                conn.RemoteAddress(),
		CreatedAt:                    conn.CreatedAt(),
		LastReceivedMessageTimeStamp: conn.LastReceivedMessageTimeStamp(),
		LastSentMessageTimeStamp:     conn.LastSentMessageTimeStamp(),
		ReceivedMessages:             conn.ReceivedMessages(),
		SentMessages:                 conn.SentMessages(),
	}
}
