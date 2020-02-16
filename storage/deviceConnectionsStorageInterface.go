package storage

import(
    "github.com/nnset/iot-cloud-connector/connections"
)

/*
DeviceConnectionsStorageInterface DeviceConnections status information storage interface
*/
type DeviceConnectionsStorageInterface interface {
    Add(connectionID, deviceID, deviceType, userAgent, remoteAddress string) error // Stores a new connection

    Delete(connectionID string) error  // Deletes a stored connection

    Get(connectionID string) (connections.DeviceConnectionStatus, error)  // Gets an stored connection copy
    GetByDeviceID(deviceID string) (connections.DeviceConnectionStatus, error)  // Gets an stored connection copy

    IncomingMessageReceived(connectionID string) error  // Updates incoming messages count
    OutgoingMessageSent(connectionID string) error  // Updates outgoing messages count

    IncomingMessages(connectionID string) uint
    OutgoingMessages(connectionID string) uint

    TotalIncomingMessages() uint
    TotalOutgoingMessages() uint

    IsDeviceConnected(deviceID string) bool
    OpenConnections() uint
}
