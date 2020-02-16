package storage

import(
    "sync"
    "errors"

    "github.com/nnset/iot-cloud-connector/connections"
    "github.com/google/uuid"
)

/*
InMemoryDeviceConnectionsStorage Concurrency safe in memory implementation of 
DeviceConnectionsStorageInterface
*/
type InMemoryDeviceConnectionsStorage struct {
    id                string    
    activeConnections map[string]*connections.DeviceConnectionStatus
    dataMutex         sync.Mutex
    totalSentMessages uint
    totalReceivedMessages uint
}

/*
NewInMemoryDeviceConnectionsStorage Returns a new instance
*/
func NewInMemoryDeviceConnectionsStorage() *InMemoryDeviceConnectionsStorage {
    return &InMemoryDeviceConnectionsStorage {
        id: uuid.New().String(),
        activeConnections: make(map[string]*connections.DeviceConnectionStatus),
        dataMutex: sync.Mutex{},
    }
}

/*
Add Adds a new connection
*/
func (storage *InMemoryDeviceConnectionsStorage) Add(connectionID, deviceID, deviceType, userAgent, remoteAddress string) error {
    storage.dataMutex.Lock()
    defer storage.dataMutex.Unlock()

    _, alreadyConnected := storage.activeConnections[connectionID]

    if alreadyConnected {
        return errors.New("Connection already established")
    }

    storage.activeConnections[connectionID] = 
        connections.NewDeviceConnectionStatus(connectionID, deviceID, deviceType, userAgent, remoteAddress)

    return nil    
}

/*
Delete Deletes a connection
*/
func (storage *InMemoryDeviceConnectionsStorage) Delete(connectionID string) error {
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
func (storage *InMemoryDeviceConnectionsStorage) Get(connectionID string) (connections.DeviceConnectionStatus, error) {
    connection, exists := storage.activeConnections[connectionID]

    if !exists {
        return connections.DeviceConnectionStatus{}, errors.New("Connection not found")
    }

    connectionCopy := *connection

    return connectionCopy, nil
}

/*
GetByDeviceID Gets a copy of an existing connection
*/
func (storage *InMemoryDeviceConnectionsStorage) GetByDeviceID(connectionID string) (connections.DeviceConnectionStatus, error) {
    // TODO
    return connections.DeviceConnectionStatus{}, nil
}

/*
MessageReceived Updates a connection when a message is received
*/
func (storage *InMemoryDeviceConnectionsStorage) IncomingMessageReceived(connectionID string) error {
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
MessageSent Updates a connection when a message is sent
*/
func (storage *InMemoryDeviceConnectionsStorage) OutgoingMessageSent(connectionID string) error {
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
func (storage *InMemoryDeviceConnectionsStorage) IsDeviceConnected(connectionID string) bool {
    _, exists := storage.activeConnections[connectionID]

    return exists
}

/*
TotalConnections How many connections are established
*/
func (storage *InMemoryDeviceConnectionsStorage) OpenConnections() uint {
    return uint(len(storage.activeConnections))
}

/*
ReceivedMessages How many messages current server received from a given connection.
If connection does not exists 0 is returned.
*/
func (storage *InMemoryDeviceConnectionsStorage) IncomingMessages(connectionID string) uint {
    _, exists := storage.activeConnections[connectionID]

    if !exists {
        return 0
    }
    
    return storage.activeConnections[connectionID].ReceivedMessages()
}

/*
SentMessages How many messages current server sent to a given connection.
If connection does not exists 0 is returned.
*/
func (storage *InMemoryDeviceConnectionsStorage) OutgoingMessages(connectionID string) uint {
    _, exists := storage.activeConnections[connectionID]

    if !exists {
        return 0
    }
    
    return storage.activeConnections[connectionID].SentMessages()
}

/*
TotalReceivedMessages
*/
func (storage *InMemoryDeviceConnectionsStorage) TotalIncomingMessages() uint {
    return storage.totalReceivedMessages
}

/*
TotalSentMessages
*/
func (storage *InMemoryDeviceConnectionsStorage) TotalOutgoingMessages() uint {
    return storage.totalSentMessages
}
