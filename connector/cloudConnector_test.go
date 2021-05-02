package connector

import (
	"testing"
	"time"

	"github.com/nnset/iot-cloud-connector/bus"

	"gotest.tools/assert"
)

func TestCreatingNewCloudConenctorShouldSetItsStateToCreated(t *testing.T) {
	eventBus, _ := bus.NewInMemoryEventBus()

	connector := NewCloudConnector(eventBus, &DummyConnectionsHandler{})

	assert.Assert(t, connector.State == CloudConnectorCreated)
}

func TestStartingCloudConenctorShouldSetItsStateToStarted(t *testing.T) {
	eventBus, _ := bus.NewInMemoryEventBus()

	connector := NewCloudConnector(eventBus, &DummyConnectionsHandler{})

	go connector.Start()

	time.Sleep(20 * time.Millisecond)

	assert.Assert(t, connector.State == CloudConnectorStarted)

	connector.Stop()
}

func TestStoppingCloudConenctorShouldSetItsStateToStopped(t *testing.T) {
	eventBus, _ := bus.NewInMemoryEventBus()

	connector := NewCloudConnector(eventBus, &DummyConnectionsHandler{})

	go connector.Start()

	time.Sleep(20 * time.Millisecond)

	connector.Stop()

	time.Sleep(20 * time.Millisecond)

	assert.Assert(t, connector.State == CloudConnectorStopped)
}

// Mocks

type DummyConnectionsHandler struct {
	connectionsHandlerIsShutdown chan bool
	shutdownService              chan bool
	id                           string
}

func (handler *DummyConnectionsHandler) Init(shutdownService chan bool) error {
	handler.shutdownService = shutdownService
	handler.connectionsHandlerIsShutdown = make(chan bool)
	handler.id = "abc-123"

	return nil
}

func (handler *DummyConnectionsHandler) Start() {
	go func() {
		<-handler.shutdownService
		handler.connectionsHandlerIsShutdown <- true
	}()
}

func (handler *DummyConnectionsHandler) ShutdownChannel() chan bool {
	return handler.connectionsHandlerIsShutdown
}

func (handler *DummyConnectionsHandler) ID() string {
	return handler.id
}

func (handler *DummyConnectionsHandler) Port() string {
	return "123"
}

func (handler *DummyConnectionsHandler) Address() string {
	return "address"
}

func (handler *DummyConnectionsHandler) Network() string {
	return "network"
}

func (handler *DummyConnectionsHandler) OpenConnectionsCount() uint {
	return 0
}
