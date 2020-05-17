package storage

import (
	"errors"
	"sync"

	"github.com/google/uuid"
	"github.com/nnset/iot-cloud-connector/connections"
)

// InMemoryDeviceConnectionsStatsStorage Concurrency safe in memory implementation of
// DeviceConnectionsStorageInterface
type InMemoryDeviceConnectionsStatsStorage struct {
	id                    string
	activeConnections     map[string]*connections.DeviceConnectionStats
	dataMutex             sync.Mutex
	totalSentMessages     uint
	totalReceivedMessages uint
}

// NewInMemoryDeviceConnectionsStatsStorage Returns a new instance
func NewInMemoryDeviceConnectionsStatsStorage() *InMemoryDeviceConnectionsStatsStorage {
	return &InMemoryDeviceConnectionsStatsStorage{
		id:                uuid.New().String(),
		activeConnections: make(map[string]*connections.DeviceConnectionStats),
		dataMutex:         sync.Mutex{},
	}
}

// Add Adds a new connection
func (storage *InMemoryDeviceConnectionsStatsStorage) Add(connectionID, deviceID, deviceType, userAgent, remoteAddress string) error {
	storage.dataMutex.Lock()
	defer storage.dataMutex.Unlock()

	_, alreadyConnected := storage.activeConnections[deviceID]

	if alreadyConnected {
		return errors.New("Connection already established")
	}

	storage.activeConnections[deviceID] =
		connections.NewDeviceConnectionStats(connectionID, deviceID, deviceType, userAgent, remoteAddress)

	return nil
}

// Delete Deletes a connection
func (storage *InMemoryDeviceConnectionsStatsStorage) Delete(deviceID string) error {
	storage.dataMutex.Lock()
	defer storage.dataMutex.Unlock()

	_, exists := storage.activeConnections[deviceID]

	if !exists {
		return nil
	}

	delete(storage.activeConnections, deviceID)

	return nil
}

// Get Gets a copy of an existing connection
func (storage *InMemoryDeviceConnectionsStatsStorage) Get(deviceID string) (connections.DeviceConnectionStats, error) {
	connection, exists := storage.activeConnections[deviceID]

	if !exists {
		return connections.DeviceConnectionStats{}, errors.New("Connection not found")
	}

	connectionCopy := *connection

	return connectionCopy, nil
}

// MessageWasReceived Increments incoming message
func (storage *InMemoryDeviceConnectionsStatsStorage) MessageWasReceived(deviceID string) error {
	storage.dataMutex.Lock()
	defer storage.dataMutex.Unlock()

	_, exists := storage.activeConnections[deviceID]

	if !exists {
		return errors.New("Connection nor found")
	}

	storage.activeConnections[deviceID].MessageReceived()
	storage.totalReceivedMessages++

	return nil
}

// MessageWasSent Updates a connection when a message is sent
func (storage *InMemoryDeviceConnectionsStatsStorage) MessageWasSent(deviceID string) error {
	storage.dataMutex.Lock()
	defer storage.dataMutex.Unlock()

	_, exists := storage.activeConnections[deviceID]

	if !exists {
		return errors.New("Connection not found")
	}

	storage.activeConnections[deviceID].MessageSent()
	storage.totalSentMessages++

	return nil
}

// IsDeviceConnected Is there a connection with the given Device?
func (storage *InMemoryDeviceConnectionsStatsStorage) IsDeviceConnected(deviceID string) bool {
	_, exists := storage.activeConnections[deviceID]

	return exists
}

// OpenConnections How many connections are established
func (storage *InMemoryDeviceConnectionsStatsStorage) OpenConnections() uint {
	return uint(len(storage.activeConnections))
}

// ConnectedDevices A list of connected Devices IDs
func (storage *InMemoryDeviceConnectionsStatsStorage) ConnectedDevices() []string {
	storage.dataMutex.Lock()
	defer storage.dataMutex.Unlock()

	devices := make([]string, len(storage.activeConnections))

	i := 0
	for id := range storage.activeConnections {
		devices[i] = id
		i++
	}

	return devices
}

// ReceivedMessages How many messages current server received from a given connection.
// If connection does not exists 0 is returned.
func (storage *InMemoryDeviceConnectionsStatsStorage) ReceivedMessages(deviceID string) uint {
	_, exists := storage.activeConnections[deviceID]

	if !exists {
		return 0
	}

	return storage.activeConnections[deviceID].ReceivedMessages()
}

// SentMessages How many messages current server sent to a given connection.
// If connection does not exists 0 is returned.
func (storage *InMemoryDeviceConnectionsStatsStorage) SentMessages(deviceID string) uint {
	_, exists := storage.activeConnections[deviceID]

	if !exists {
		return 0
	}

	return storage.activeConnections[deviceID].SentMessages()
}

// TotalReceivedMessages How many messages were received from all IoT devices
func (storage *InMemoryDeviceConnectionsStatsStorage) TotalReceivedMessages() uint {
	return storage.totalReceivedMessages
}

// TotalSentMessages How many messages were sent to all IoT devices
func (storage *InMemoryDeviceConnectionsStatsStorage) TotalSentMessages() uint {
	return storage.totalSentMessages
}
