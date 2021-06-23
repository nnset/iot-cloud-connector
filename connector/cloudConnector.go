package connector

import (
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/nnset/iot-cloud-connector/bus"
	"github.com/nnset/iot-cloud-connector/services"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// Cloud Connector have its own logging system, it will not interact in any way, with service's logging
// configuration.
const (
	// LogPanicLevel level, highest level of severity. Logs and then calls panic with the
	// message passed to Debug, Info, ...
	LogPanicLevel uint32 = 0
	// LogFatalLevel level. Logs and then calls `logger.Exit(1)`. It will exit even if the
	// logging level is set to Panic.
	LogFatalLevel uint32 = 1
	// LogErrorLevel level. Logs. Used for errors that should definitely be noted.
	// Commonly used for hooks to send errors to an error tracking service.
	LogErrorLevel uint32 = 2
	// LogWarnLevel level. Non-critical entries that deserve eyes.
	LogWarnLevel uint32 = 3
	// LogInfoLevel level. General operational entries about what's going on inside the
	// application.
	LogInfoLevel uint32 = 4
	// LogDebugLevel level. Usually only enabled when debugging. Very verbose logging.
	LogDebugLevel uint32 = 5
	// LogTraceLevel level. Designates finer-grained informational events than the Debug.
	LogTraceLevel uint32 = 6
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

// CloudConnector is the central component of your IoT server, it has the responsability
// to start all your services (your businness logic and any other dependency that
// you may need) and try a gracefull shutdown when the time comes.
// We encourage you to follow an asynchronous/event driven
// approach with your IoT devices.
type CloudConnector struct {
	Id                                 string
	StartTime                          int64
	State                              CloudConnectorState
	LogFilePath                        string
	LogDebugLevel                      uint32
	ShutdownTimeout                    uint // In seconds
	eventBus                           bus.MessageBus
	services                           []services.ServiceInterface
	servicesGracefullShutdowns         map[string]chan bool
	shutdownServices                   chan bool
	servicesGracefullShutdownWaitGroup sync.WaitGroup
	serverFullShutdownWaitGroup        sync.WaitGroup
	operatingSystemSignal              chan os.Signal
	log                                *logrus.Logger
}

// NewCloudConnector Creates a new instance of CloudConnector
func NewCloudConnector(
	eventBus bus.MessageBus,
	services []services.ServiceInterface,
	logFilePath string,
	logDebugLevel uint32,
	shutdownTimeout uint,
) *CloudConnector {
	return &CloudConnector{
		Id:                                 uuid.New().String(),
		StartTime:                          time.Now().Unix(),
		State:                              CloudConnectorCreated,
		LogFilePath:                        logFilePath,
		LogDebugLevel:                      logDebugLevel,
		ShutdownTimeout:                    shutdownTimeout,
		eventBus:                           eventBus,
		services:                           services,
		servicesGracefullShutdowns:         make(map[string]chan bool),
		servicesGracefullShutdownWaitGroup: sync.WaitGroup{},
		serverFullShutdownWaitGroup:        sync.WaitGroup{},
		shutdownServices:                   make(chan bool),
		operatingSystemSignal:              make(chan os.Signal),
	}
}

// Start TODO
func (cc *CloudConnector) Start() {
	cc.setupLogging()

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

func (cc *CloudConnector) setupLogging() {

	cc.log = logrus.New()

	cc.log.SetLevel(logrus.Level(cc.LogDebugLevel))
	cc.log.Out = os.Stdout

	file, err := os.OpenFile(cc.LogFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)

	if err == nil {
		cc.log.Out = file
	} else {
		fmt.Println("Using Stdout for log")
	}
}

func (cc *CloudConnector) startServices() {
	for _, service := range cc.services {

		err := service.Init(cc.shutdownServices)

		if err == nil {
			go service.Start()

			cc.servicesGracefullShutdowns[service.Id()] = service.ShutdownChannel()
			cc.servicesGracefullShutdownWaitGroup.Add(1)

			go cc.waitForServiceShutdown(service.Id())
		} else {
			cc.log.Error(err)
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
		cc.serverFullShutdownWaitGroup.Done()
	case <-time.After(time.Duration(cc.ShutdownTimeout) * time.Second):
		cc.log.Warning("Unable to gracefully shutdown services. Cloud Conenctor will shutdown.")
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
