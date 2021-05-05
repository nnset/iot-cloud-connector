package connector

import (
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/nnset/iot-cloud-connector/bus"
	"github.com/nnset/iot-cloud-connector/services"

	"github.com/google/uuid"
)

// CloudConnectorState Server's state
type CloudConnectorState string

// CloudConnectors go across some status:
//   - CloudConnectorCreated
//   - CloudConnectorStarted
//   - CloudConnectorStopped
//   - CloudConnectorGracefullyStopped
const (
	CloudConnectorCreated           CloudConnectorState = "created"
	CloudConnectorStarted           CloudConnectorState = "started"
	CloudConnectorStopped           CloudConnectorState = "stopped"
	CloudConnectorGracefullyStopped CloudConnectorState = "gracefully_stopped"
)

type CloudConnector struct {
	Id                                 string
	StartTime                          int64
	State                              CloudConnectorState
	eventBus                           bus.MessageBus
	services                           []services.ServiceInterface
	servicesGracefullShutdowns         map[string]chan bool
	shutdownServices                   chan bool
	servicesGracefullShutdownWaitGroup sync.WaitGroup
	serverFullShutdownWaitGroup        sync.WaitGroup
	operatingSystemSignal              chan os.Signal
}

// NewCloudConnector Creates a new instance of CloudConnector
func NewCloudConnector(
	eventBus bus.MessageBus,
	services []services.ServiceInterface,
) *CloudConnector {
	return &CloudConnector{
		Id:                                 uuid.New().String(),
		StartTime:                          time.Now().Unix(),
		eventBus:                           eventBus,
		services:                           services,
		servicesGracefullShutdowns:         make(map[string]chan bool),
		servicesGracefullShutdownWaitGroup: sync.WaitGroup{},
		serverFullShutdownWaitGroup:        sync.WaitGroup{},
		shutdownServices:                   make(chan bool),
		operatingSystemSignal:              make(chan os.Signal),
		State:                              CloudConnectorCreated,
	}
}

// Start TODO
func (cc *CloudConnector) Start() {

	cc.startServices()

	cc.serverFullShutdownWaitGroup.Add(1)

	cc.State = CloudConnectorStarted

	signal.Notify(cc.operatingSystemSignal, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)

	<-cc.operatingSystemSignal // Blocked until operating system shutdows CloudConnector (or Stop() is called)

	go cc.notifyAllServicesToShutDown()

	go cc.waitServicesToShutdown()

	cc.serverFullShutdownWaitGroup.Wait()

	cc.State = CloudConnectorStopped
}

func (cc *CloudConnector) startServices() {
	for _, service := range cc.services {

		err := service.Init(cc.shutdownServices)

		if err == nil {
			go service.Start()

			cc.servicesGracefullShutdowns[service.Id()] = service.ShutdownChannel()
			cc.servicesGracefullShutdownWaitGroup.Add(1)

			go cc.waitForServiceShutdown(service.Id())
		}
	}
}

func (cc *CloudConnector) waitForServiceShutdown(serviceID string) {
	<-cc.servicesGracefullShutdowns[serviceID] // Blocked until service reports its shutdown

	cc.servicesGracefullShutdownWaitGroup.Done()

	close(cc.servicesGracefullShutdowns[serviceID])
	delete(cc.servicesGracefullShutdowns, serviceID)
}

// waitServicesToShutdown Waits for all services to gracefully shutdown or return
// if services shutdown, took more time than the allowed timeout.
func (cc *CloudConnector) waitServicesToShutdown() {
	c := make(chan struct{})

	go func() {
		defer close(c)

		cc.servicesGracefullShutdownWaitGroup.Wait()
	}()

	select {
	case <-c:
		// TODO log
		cc.serverFullShutdownWaitGroup.Done()
	case <-time.After(10 * time.Second):
		// TODO log
		cc.serverFullShutdownWaitGroup.Done()
	}
}

func (cc *CloudConnector) notifyAllServicesToShutDown() {
	for range cc.services {
		cc.shutdownServices <- true
	}
}

// Stop If you want to stop CloudConnector programatically instead of waiting for an
// operating system signal, use this method.
func (cc *CloudConnector) Stop() {
	cc.operatingSystemSignal <- syscall.SIGINT
}
