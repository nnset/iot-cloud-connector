package storage

import(
	"github.com/nnset/iot-cloud-connector/connections"
)


/*
DeviceConnectionsStorageInterface DeviceConnections status information storage interface
*/
type DeviceConnectionsStorageInterface interface {
	Add(connectionID, deviceID, deviceType, userAgent,  remoteAddress string) error

	Delete(connectionID string) error

	Get(connectionID string) (connections.DeviceConnectionStatus, error)
	GetByDeviceID(deviceID string) (connections.DeviceConnectionStatus, error)
	
	MessageReceived(connectionID string) error
	MessageSent(connectionID string) error

	IsDeviceConnected(deviceID string) bool
	TotalConnections() int
}
