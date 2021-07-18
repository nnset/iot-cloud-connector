package connector

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nnset/iot-cloud-connector/services"

	"github.com/nnset/iot-cloud-connector/bus"

	"gotest.tools/assert"
)

func TestCreatingNewCloudConenctorShouldSetItsStateToCreated(t *testing.T) {
	eventBus, _ := bus.NewInMemoryEventBus()
	var services []services.ServiceInterface
	services = append(services, &DummyConnectionsHandler{})

	connector := NewCloudConnector(eventBus, services, "", LogErrorLevel, 5)

	assert.Assert(t, connector.State == CloudConnectorCreated)
}

func TestStartingCloudConenctorShouldSetItsStateToStarted(t *testing.T) {
	eventBus, _ := bus.NewInMemoryEventBus()
	var services []services.ServiceInterface
	services = append(services, &DummyConnectionsHandler{})

	connector := NewCloudConnector(eventBus, services, "", LogErrorLevel, 5)

	go connector.Start()

	time.Sleep(20 * time.Millisecond)

	assert.Assert(t, connector.State == CloudConnectorStarted)

	connector.Stop()
}

func TestStoppingCloudConenctorShouldSetItsStateToStopped(t *testing.T) {
	eventBus, _ := bus.NewInMemoryEventBus()
	var services []services.ServiceInterface
	services = append(services, &DummyConnectionsHandler{})

	connector := NewCloudConnector(eventBus, services, "", LogErrorLevel, 5)

	go connector.Start()

	time.Sleep(20 * time.Millisecond)

	connector.Stop()

	time.Sleep(20 * time.Millisecond)

	assert.Assert(t, connector.State == CloudConnectorStopped)
}

func TestStartingCloudConnectorShouldStartAllServices(t *testing.T) {
	eventBus, _ := bus.NewInMemoryEventBus()
	var services []services.ServiceInterface
	services = append(services, &DummyConnectionsHandler{})
	services = append(services, &DummyConnectionsHandler{})

	connector := NewCloudConnector(eventBus, services, "", LogErrorLevel, 5)

	assert.Assert(t, (services[0].(*DummyConnectionsHandler)).HasStarted == false)
	assert.Assert(t, (services[1].(*DummyConnectionsHandler)).HasStarted == false)

	go connector.Start()

	time.Sleep(20 * time.Millisecond)

	assert.Assert(t, (services[0].(*DummyConnectionsHandler)).HasStarted == true)
	assert.Assert(t, (services[1].(*DummyConnectionsHandler)).HasStarted == true)

	connector.Stop()
	time.Sleep(20 * time.Millisecond)
}

func TestStoppingCloudConnectorShouldStopAllServices(t *testing.T) {
	eventBus, _ := bus.NewInMemoryEventBus()
	var services []services.ServiceInterface
	services = append(services, &DummyConnectionsHandler{})
	services = append(services, &DummyConnectionsHandler{})

	connector := NewCloudConnector(eventBus, services, "", LogErrorLevel, 5)

	assert.Assert(t, (services[0].(*DummyConnectionsHandler)).IsStopped == false)
	assert.Assert(t, (services[1].(*DummyConnectionsHandler)).IsStopped == false)

	go connector.Start()

	time.Sleep(20 * time.Millisecond)

	connector.Stop()
	time.Sleep(20 * time.Millisecond)

	assert.Assert(t, (services[0].(*DummyConnectionsHandler)).IsStopped == true)
	assert.Assert(t, (services[1].(*DummyConnectionsHandler)).IsStopped == true)
}

// Mocks

type DummyConnectionsHandler struct {
	connectionsHandlerIsShutdown chan bool
	shutdownService              chan bool
	id                           string
	HasStarted                   bool
	IsStopped                    bool
}

func (handler *DummyConnectionsHandler) Id() string {
	return handler.id
}

func (handler *DummyConnectionsHandler) Init(shutdownService chan bool) error {
	handler.shutdownService = shutdownService
	handler.connectionsHandlerIsShutdown = make(chan bool)
	handler.id = uuid.New().String()
	handler.HasStarted = false
	handler.IsStopped = false

	return nil
}

func (handler *DummyConnectionsHandler) Start() {
	handler.HasStarted = true

	go func() {
		<-handler.shutdownService
		handler.IsStopped = true
		handler.connectionsHandlerIsShutdown <- true
	}()
}

func (handler *DummyConnectionsHandler) ShutdownChannel() chan bool {
	return handler.connectionsHandlerIsShutdown
}
