package storage

import (
	"github.com/nnset/iot-cloud-connector/connections"
)

/*
DeviceConnectionsStatsStorageInterface DeviceConnections status information storage interface
*/
type DeviceConnectionsStatsStorageInterface interface {
	Add(connectionID, deviceID, deviceType, userAgent, remoteAddress string) error

	Delete(deviceID string) error

	Get(deviceID string) (connections.DeviceConnectionStats, error)

	IncomingMessageReceived(deviceID string) error // Updates incoming messages count
	OutgoingMessageSent(deviceID string) error     // Updates outgoing messages count

	IncomingMessages(deviceID string) uint
	OutgoingMessages(deviceID string) uint

	TotalIncomingMessages() uint
	TotalOutgoingMessages() uint

	IsDeviceConnected(deviceID string) bool
	OpenConnections() uint
}
