package storage

import (
	"errors"
	"sync"

	"github.com/google/uuid"
	"github.com/nnset/iot-cloud-connector/connections"
)

/*
InMemoryDeviceConnectionsStatsStorage Concurrency safe in memory implementation of
DeviceConnectionsStorageInterface
*/
type InMemoryDeviceConnectionsStatsStorage struct {
	id                    string
	activeConnections     map[string]*connections.DeviceConnectionStats
	dataMutex             sync.Mutex
	totalSentMessages     uint
	totalReceivedMessages uint
}

/*
NewInMemoryDeviceConnectionsStatsStorage Returns a new instance
*/
func NewInMemoryDeviceConnectionsStatsStorage() *InMemoryDeviceConnectionsStatsStorage {
	return &InMemoryDeviceConnectionsStatsStorage{
		id:                uuid.New().String(),
		activeConnections: make(map[string]*connections.DeviceConnectionStats),
		dataMutex:         sync.Mutex{},
	}
}

/*
Add Adds a new connection
*/
func (storage *InMemoryDeviceConnectionsStatsStorage) Add(connectionID, deviceID, deviceType, userAgent, remoteAddress string) error {
	storage.dataMutex.Lock()
	defer storage.dataMutex.Unlock()

	_, alreadyConnected := storage.activeConnections[connectionID]

	if alreadyConnected {
		return errors.New("Connection already established")
	}

	storage.activeConnections[connectionID] =
		connections.NewDeviceConnectionStats(connectionID, deviceID, deviceType, userAgent, remoteAddress)

	return nil
}

/*
Delete Deletes a connection
*/
func (storage *InMemoryDeviceConnectionsStatsStorage) Delete(connectionID string) error {
	storage.dataMutex.Lock()
	defer storage.dataMutex.Unlock()

	_, exists := storage.activeConnections[connectionID]

	if !exists {
		return nil
	}

	delete(storage.activeConnections, connectionID)

	return nil
}

/*
Get Gets a copy of an existing connection
*/
func (storage *InMemoryDeviceConnectionsStatsStorage) Get(connectionID string) (connections.DeviceConnectionStats, error) {
	connection, exists := storage.activeConnections[connectionID]

	if !exists {
		return connections.DeviceConnectionStats{}, errors.New("Connection not found")
	}

	connectionCopy := *connection

	return connectionCopy, nil
}

/*
GetByDeviceID Gets a copy of an existing connection
*/
func (storage *InMemoryDeviceConnectionsStatsStorage) GetByDeviceID(connectionID string) (connections.DeviceConnectionStats, error) {
	// TODO
	return connections.DeviceConnectionStats{}, nil
}

/*
IncomingMessageReceived Increments icoming message
*/
func (storage *InMemoryDeviceConnectionsStatsStorage) IncomingMessageReceived(connectionID string) error {
	storage.dataMutex.Lock()
	defer storage.dataMutex.Unlock()

	_, exists := storage.activeConnections[connectionID]

	if !exists {
		return errors.New("Connection nor found")
	}

	storage.activeConnections[connectionID].MessageReceived()
	storage.totalReceivedMessages++

	return nil
}

/*
OutgoingMessageSent Updates a connection when a message is sent
*/
func (storage *InMemoryDeviceConnectionsStatsStorage) OutgoingMessageSent(connectionID string) error {
	storage.dataMutex.Lock()
	defer storage.dataMutex.Unlock()

	_, exists := storage.activeConnections[connectionID]

	if !exists {
		return errors.New("Connection nor found")
	}

	storage.activeConnections[connectionID].MessageSent()
	storage.totalSentMessages++

	return nil
}

/*
IsDeviceConnected Is there a connection with the given Device?
*/
func (storage *InMemoryDeviceConnectionsStatsStorage) IsDeviceConnected(connectionID string) bool {
	_, exists := storage.activeConnections[connectionID]

	return exists
}

/*
OpenConnections How many connections are established
*/
func (storage *InMemoryDeviceConnectionsStatsStorage) OpenConnections() uint {
	return uint(len(storage.activeConnections))
}

/*
IncomingMessages How many messages current server received from a given connection.
If connection does not exists 0 is returned.
*/
func (storage *InMemoryDeviceConnectionsStatsStorage) IncomingMessages(connectionID string) uint {
	_, exists := storage.activeConnections[connectionID]

	if !exists {
		return 0
	}

	return storage.activeConnections[connectionID].ReceivedMessages()
}

/*
OutgoingMessages How many messages current server sent to a given connection.
If connection does not exists 0 is returned.
*/
func (storage *InMemoryDeviceConnectionsStatsStorage) OutgoingMessages(connectionID string) uint {
	_, exists := storage.activeConnections[connectionID]

	if !exists {
		return 0
	}

	return storage.activeConnections[connectionID].SentMessages()
}

/*
TotalIncomingMessages
*/
func (storage *InMemoryDeviceConnectionsStatsStorage) TotalIncomingMessages() uint {
	return storage.totalReceivedMessages
}

/*
TotalSentMessages
*/
func (storage *InMemoryDeviceConnectionsStatsStorage) TotalOutgoingMessages() uint {
	return storage.totalSentMessages
}
