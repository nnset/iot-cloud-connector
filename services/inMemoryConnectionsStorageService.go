package services

import (
	"fmt"
	"sync"

	"github.com/google/uuid"
	"github.com/nnset/iot-cloud-connector/bus"
	"github.com/nnset/iot-cloud-connector/entities"
	"github.com/nnset/iot-cloud-connector/events"
)

// InMemoryConnectionsStorageService Thread safe in memory connections storage.
// Only ONE connection per device is allowed.
// This Storage will listen to eventBus ConnectionEstablished and
// ConnectionClosed messages in order to keep track of all active connections,
// without keeping any historical data, regardless a couple of global counters:
// totalSentMessages and totalReceivedMessages
type InMemoryConnectionsStorageService struct {
	id                            string
	eventBus                      bus.MessageBus
	activeConnections             map[string]*entities.Connection
	dataMutex                     sync.Mutex
	serviceIsShutdown             chan bool
	shutdownService               chan bool
	totalSentMessages             uint
	totalReceivedMessages         uint
	activeConnectionsCount        uint
	connectionsEstablishedChannel chan events.Message
	connectionsClosedChannel      chan events.Message
	gracefullShutdownWaitGroup    sync.WaitGroup
}

// NewInMemoryConnectionsStorageService Creates a new instance of InMemoryConnectionsStorageService
func NewInMemoryConnectionsStorageService(
	eventBus bus.MessageBus,
) (*InMemoryConnectionsStorageService, error) {
	return &InMemoryConnectionsStorageService{
		id:                            uuid.New().String(),
		eventBus:                      eventBus,
		activeConnections:             make(map[string]*entities.Connection),
		dataMutex:                     sync.Mutex{},
		connectionsEstablishedChannel: make(chan events.Message),
		connectionsClosedChannel:      make(chan events.Message),
		gracefullShutdownWaitGroup:    sync.WaitGroup{},
	}, nil
}

func (service *InMemoryConnectionsStorageService) Id() string {
	return service.id
}

func (service *InMemoryConnectionsStorageService) Init(shutdownService chan bool) error {
	service.shutdownService = shutdownService
	service.serviceIsShutdown = make(chan bool)

	service.eventBus.Subscribe(events.ConnectionEstablished, &service.connectionsEstablishedChannel)
	service.eventBus.Subscribe(events.ConnectionClosed, &service.connectionsClosedChannel)

	return nil
}

func (service *InMemoryConnectionsStorageService) Start() {
	service.gracefullShutdownWaitGroup.Add(1)
	shutdownEstablishedConnections := make(chan bool)
	go service.handleEstablishedConnections(shutdownEstablishedConnections)

	service.gracefullShutdownWaitGroup.Add(1)
	shutdownClosedConnections := make(chan bool)
	go service.handleClosedConnections(shutdownClosedConnections)

	<-service.shutdownService
	// TODO add Timeout here
	shutdownEstablishedConnections <- true
	// TODO add Timeout here
	shutdownClosedConnections <- true

	service.serviceIsShutdown <- true
}

func (service *InMemoryConnectionsStorageService) handleEstablishedConnections(shutdownChannel chan bool) {
	for {
		select {
		case m := <-service.connectionsEstablishedChannel:
			service.addConnection(m)

		case <-shutdownChannel:
			service.gracefullShutdownWaitGroup.Done()
			return
		}
	}
}

func (service *InMemoryConnectionsStorageService) handleClosedConnections(shutdownChannel chan bool) {
	for {
		select {
		case m := <-service.connectionsClosedChannel:
			service.removeConnection(m)

		case <-shutdownChannel:
			service.gracefullShutdownWaitGroup.Done()
			return
		}
	}
}

func (service *InMemoryConnectionsStorageService) addConnection(message events.Message) error {
	service.dataMutex.Lock()
	defer service.dataMutex.Unlock()

	connection, err := entities.NewConnectionFromDefaultPayload(message.Payload, message.OriginRemoteAddress)

	if err != nil {
		return err
	}

	_, alreadyConnected := service.activeConnections[connection.DeviceID]

	if alreadyConnected {
		return fmt.Errorf("device %s already connected", connection.DeviceID)
	}

	service.activeConnections[connection.DeviceID] = connection

	service.activeConnectionsCount++

	return nil
}

func (service *InMemoryConnectionsStorageService) removeConnection(message events.Message) error {
	service.dataMutex.Lock()
	defer service.dataMutex.Unlock()

	connection, err := entities.NewConnectionFromDefaultPayload(message.Payload, message.OriginRemoteAddress)

	if err != nil {
		return err
	}

	_, exists := service.activeConnections[connection.DeviceID]

	if !exists {
		return nil
	}

	delete(service.activeConnections, connection.DeviceID)

	service.activeConnectionsCount--

	return nil
}

func (service *InMemoryConnectionsStorageService) ShutdownChannel() chan bool {
	return service.serviceIsShutdown
}

func (service *InMemoryConnectionsStorageService) TotalSentMessages() uint {
	return service.totalSentMessages
}

func (service *InMemoryConnectionsStorageService) TotalReceivedMessages() uint {
	return service.totalReceivedMessages
}

func (service *InMemoryConnectionsStorageService) ActiveConnectionsCount() uint {
	return service.activeConnectionsCount
}

// ActiveConnections Returns a copy of the currently active connections map
func (service *InMemoryConnectionsStorageService) ActiveConnections() map[string]*entities.Connection {
	service.dataMutex.Lock()
	defer service.dataMutex.Unlock()

	cloned := make(map[string]*entities.Connection)

	for k, v := range service.activeConnections {
		cloned[k] = v
	}

	return cloned
}
