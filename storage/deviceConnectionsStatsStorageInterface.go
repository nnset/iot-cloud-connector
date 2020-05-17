package storage

import (
	"github.com/nnset/iot-cloud-connector/connections"
)

// DeviceConnectionsStatsStorageInterface Defines how DeviceConnectionStats should be stored.
// We provide a thread safe in memory implementation.
type DeviceConnectionsStatsStorageInterface interface {
	Add(connectionID, deviceID, deviceType, userAgent, remoteAddress string) error

	Delete(deviceID string) error

	Get(deviceID string) (connections.DeviceConnectionStats, error)

	MessageWasReceived(deviceID string) error
	MessageWasSent(deviceID string) error

	SentMessages(deviceID string) uint
	ReceivedMessages(deviceID string) uint

	TotalReceivedMessages() uint
	TotalSentMessages() uint

	IsDeviceConnected(deviceID string) bool
	OpenConnections() uint
	ConnectedDevices() []string
}
