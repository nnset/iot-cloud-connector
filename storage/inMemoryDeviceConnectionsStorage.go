package storage

import (
	"errors"
	"sync"

	"github.com/google/uuid"
	"github.com/nnset/iot-cloud-connector/connections"
)

// InMemoryDeviceConnectionsStorage Thread safe in memory implementation, of DeviceConnectionsStorageInterface
type InMemoryDeviceConnectionsStorage struct {
	id                    string
	activeConnections     map[string]*connections.DeviceConnection
	dataMutex             sync.Mutex
	totalSentMessages     uint
	totalReceivedMessages uint
}

// NewInMemoryDeviceConnectionsStorage Returns a new instance
func NewInMemoryDeviceConnectionsStorage() *InMemoryDeviceConnectionsStorage {
	return &InMemoryDeviceConnectionsStorage{
		id:                uuid.New().String(),
		activeConnections: make(map[string]*connections.DeviceConnection),
		dataMutex:         sync.Mutex{},
	}
}

// Add Adds a new connection
func (storage *InMemoryDeviceConnectionsStorage) Add(connectionID, deviceID, deviceName, deviceType, userAgent, remoteAddress string) error {
	storage.dataMutex.Lock()
	defer storage.dataMutex.Unlock()

	_, alreadyConnected := storage.activeConnections[deviceID]

	if alreadyConnected {
		return errors.New("Connection already established")
	}

	connection, err :=
		connections.NewDeviceConnection(connectionID, deviceID, deviceName, deviceType, userAgent, remoteAddress)

	if err == nil {
		storage.activeConnections[deviceID] = connection
	}

	return err
}

// Delete Deletes a connection
func (storage *InMemoryDeviceConnectionsStorage) Delete(deviceID string) error {
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
func (storage *InMemoryDeviceConnectionsStorage) Get(deviceID string) (connections.DeviceConnection, error) {
	connection, exists := storage.activeConnections[deviceID]

	if !exists {
		return connections.DeviceConnection{}, errors.New("Connection not found")
	}

	connectionCopy := *connection

	return connectionCopy, nil
}

// MessageWasReceived Increments incoming message
func (storage *InMemoryDeviceConnectionsStorage) MessageWasReceived(deviceID string) error {
	storage.dataMutex.Lock()
	defer storage.dataMutex.Unlock()

	_, exists := storage.activeConnections[deviceID]

	if !exists {
		return errors.New("Connection not found")
	}

	storage.activeConnections[deviceID].MessageReceived()
	storage.totalReceivedMessages++

	return nil
}

// MessageWasSent Updates a connection when a message is sent
func (storage *InMemoryDeviceConnectionsStorage) MessageWasSent(deviceID string) error {
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
func (storage *InMemoryDeviceConnectionsStorage) IsDeviceConnected(deviceID string) bool {
	_, exists := storage.activeConnections[deviceID]

	return exists
}

// OpenConnections How many connections are established
func (storage *InMemoryDeviceConnectionsStorage) OpenConnections() uint {
	return uint(len(storage.activeConnections))
}

// ConnectedDevices A list of connected Devices IDs
func (storage *InMemoryDeviceConnectionsStorage) ConnectedDevices() []*connections.DeviceConnectionDTO {
	storage.dataMutex.Lock()
	defer storage.dataMutex.Unlock()

	devices := make([]*connections.DeviceConnectionDTO, len(storage.activeConnections))

	i := 0
	for _, connection := range storage.activeConnections {
		devices[i] = connections.NewDeviceConnectionDTOFromDeviceConnection(connection)
		i++
	}

	return devices
}

// ReceivedMessages How many messages current server received from a given connection.
// If connection does not exists 0 is returned.
func (storage *InMemoryDeviceConnectionsStorage) ReceivedMessages(deviceID string) uint {
	_, exists := storage.activeConnections[deviceID]

	if !exists {
		return 0
	}

	return storage.activeConnections[deviceID].ReceivedMessages()
}

// SentMessages How many messages current server sent to a given connection.
// If connection does not exists 0 is returned.
func (storage *InMemoryDeviceConnectionsStorage) SentMessages(deviceID string) uint {
	_, exists := storage.activeConnections[deviceID]

	if !exists {
		return 0
	}

	return storage.activeConnections[deviceID].SentMessages()
}

// TotalReceivedMessages How many messages were received from all IoT devices
func (storage *InMemoryDeviceConnectionsStorage) TotalReceivedMessages() uint {
	return storage.totalReceivedMessages
}

// TotalSentMessages How many messages were sent to all IoT devices
func (storage *InMemoryDeviceConnectionsStorage) TotalSentMessages() uint {
	return storage.totalSentMessages
}

func (storage *InMemoryDeviceConnectionsStorage) ID() string {
	return storage.id
}
