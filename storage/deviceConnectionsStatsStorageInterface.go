package storage

import (
	"github.com/nnset/iot-cloud-connector/connections"
)

/*
DeviceConnectionsStatsStorageInterface DeviceConnections status information storage interface
*/
type DeviceConnectionsStatsStorageInterface interface {
	Add(connectionID, deviceID, deviceType, userAgent, remoteAddress string) error

	Delete(connectionID string) error

	Get(connectionID string) (connections.DeviceConnectionStats, error)
	GetByDeviceID(deviceID string) (connections.DeviceConnectionStats, error)

	IncomingMessageReceived(connectionID string) error // Updates incoming messages count
	OutgoingMessageSent(connectionID string) error     // Updates outgoing messages count

	IncomingMessages(connectionID string) uint
	OutgoingMessages(connectionID string) uint

	TotalIncomingMessages() uint
	TotalOutgoingMessages() uint

	IsDeviceConnected(deviceID string) bool
	OpenConnections() uint
}
