package servers

import (
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/nnset/iot-cloud-connector/connectionshandlers"
	"github.com/nnset/iot-cloud-connector/storage"
	"github.com/sirupsen/logrus"
)

// CloudConnector is the main process, once you Start() it, these processes are spawned :
//  - An instance of connectionshandlers.ConnectionsHandlerInterface
//  - An instance of CloudConnectorAPIInterface
//
// CloudConnector attempts a graceful shutdown (closing all connections) when any of
// these signals are received:
//  - syscall.SIGINT
//  - syscall.SIGTERM
//  - syscall.SIGKILL
type CloudConnector struct {
	id                      string
	address                 string
	port                    string
	network                 string
	startTime               int64
	log                     *logrus.Logger
	serverShutdownWaitGroup sync.WaitGroup
	connectionsHandler      connectionshandlers.ConnectionsHandlerInterface
	statusAPI               CloudConnectorAPIInterface
	state                   CloudConnectorState
	connectionsStats        storage.DeviceConnectionsStatsStorageInterface
	auth                    APIAuthMiddleWare
}

// CloudConnectorState Server's state
type CloudConnectorState string

// CloudConnectors go across some status:
//   - CloudConnectorCreated
//   - CloudConnectorStarted
//   - CloudConnectorStopped
const (
	CloudConnectorCreated CloudConnectorState = "created"
	CloudConnectorStarted CloudConnectorState = "started"
	CloudConnectorStopped CloudConnectorState = "stopped"
)

// NewCloudConnector Creates a new instance of CloudConnector
func NewCloudConnector(
	log *logrus.Logger,
	connectionsHandler connectionshandlers.ConnectionsHandlerInterface,
	connectionsStats storage.DeviceConnectionsStatsStorageInterface,
	statusAPI CloudConnectorAPIInterface) *CloudConnector {
	return &CloudConnector{
		id:                      uuid.New().String(),
		log:                     log,
		serverShutdownWaitGroup: sync.WaitGroup{},
		connectionsHandler:      connectionsHandler,
		statusAPI:               statusAPI,
		startTime:               time.Now().Unix(),
		state:                   CloudConnectorCreated,
		connectionsStats:        connectionsStats,
	}
}

// Start Starts ClodConnector and its child processes, currently:
//  - An instance of connectionshandlers.ConnectionsHandlerInterface via its Listen() method.
//  - An instance of CloudConnectorAPIInterface via its Start() method.
func (cc *CloudConnector) Start() {
	cc.log.Debugf("Starting CloudConnector #%s", cc.id)

	connectionsHandlerShutdownIsComplete, shutdownConnectionsHandler := make(chan bool), make(chan bool)

	cc.serverShutdownWaitGroup.Add(1)
	go cc.waitForShutdownSignal()

	go cc.connectionsHandler.Start(&shutdownConnectionsHandler, &connectionsHandlerShutdownIsComplete, cc.connectionsStats, cc.log)

	if cc.statusAPI != nil {
		go cc.statusAPI.Start(cc)
	}

	cc.state = CloudConnectorStarted

	cc.serverShutdownWaitGroup.Wait()

	cc.log.Info("CloudConnector received shutdown signal")

	cc.shutdown(shutdownConnectionsHandler, connectionsHandlerShutdownIsComplete)
}

func (cc *CloudConnector) waitForShutdownSignal() {
	operatingSystemSignal := make(chan os.Signal)

	signal.Notify(operatingSystemSignal, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)

	go func() {
		sig := <-operatingSystemSignal
		cc.log.Debugf("Signal received : %s", sig)
		cc.log.Debug("Shutting down main")

		cc.serverShutdownWaitGroup.Done()
	}()
}

func (cc *CloudConnector) shutdown(shutdownConnectionsHandler, connectionsHandlerShutdownIsComplete chan bool) {
	shutdownConnectionsHandler <- true

	select {
	case <-connectionsHandlerShutdownIsComplete:
		cc.log.Debug("Connections handler successfully shutdown")
	case <-time.After(8 * time.Second):
		cc.log.Error("Connections handler shutdown time out")
	}

	if cc.statusAPI != nil {
		cc.statusAPI.Stop()
	}

	cc.state = CloudConnectorStopped

	cc.log.Info("CloudConnector stopped.")
	cc.log.Info("  Total received messages processed: ", cc.connectionsStats.TotalReceivedMessages())
	cc.log.Info("  Total sent messages processed: ", cc.connectionsStats.TotalSentMessages())
	cc.log.Infof("  Uptime: %d seconds", cc.Uptime(""))
}

// ID CloudConnector's uuid
func (cc *CloudConnector) ID() string {
	return cc.id
}

// Uptime how many seconds CloudConnector has been up
func (cc *CloudConnector) Uptime(deviceID string) int64 {

	if deviceID == "" {
		if cc.startTime == 0 {
			return 0
		}

		return time.Now().Unix() - cc.startTime
	}

	connection, err := cc.connectionsStats.Get(deviceID)

	if err != nil {
		// TODO should we return error or just 0 ?
		return 0
	}

	uptime, err := connection.Uptime()

	if err != nil {
		// TODO should we return error or just 0 ?
		return 0
	}

	return uptime
}

// OpenConnections How many connections are currently open
func (cc *CloudConnector) OpenConnections() uint {
	return cc.connectionsStats.OpenConnections()
}

// ConnectedDevices Which IoT devices (IDs are displayed) are currently connected
func (cc *CloudConnector) ConnectedDevices() []string {
	return cc.connectionsStats.ConnectedDevices()
}

// ReceivedMessages How many messages were received from a given IoT Device, or if
// deviceID is empty, globally.
func (cc *CloudConnector) ReceivedMessages(deviceID string) uint {
	if deviceID == "" {
		return cc.connectionsStats.TotalReceivedMessages()
	}

	return cc.connectionsStats.ReceivedMessages(deviceID)
}

// SentMessages How many messages were sent to a given IoT Device, or if
// deviceID is empty, globally.
func (cc *CloudConnector) SentMessages(deviceID string) uint {
	if deviceID == "" {
		return cc.connectionsStats.TotalSentMessages()
	}

	return cc.connectionsStats.SentMessages(deviceID)
}

// SendCommand Send a command message to a given IoT Device
func (cc *CloudConnector) SendCommand(payload, deviceID string) (string, int, error) {
	return cc.connectionsHandler.SendCommand(payload, deviceID)
}

// SendQuery Send a query message to a given IoT Device
func (cc *CloudConnector) SendQuery(payload, deviceID string) (string, int, error) {
	return cc.connectionsHandler.SendQuery(payload, deviceID)
}

// SystemMemory total mega bytes of memory obtained from the OS by CloudConnector and its
// child processes
func (cc *CloudConnector) SystemMemory() uint {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return uint(m.Sys / 1024 / 1024)
}

// AllocatedMemory mega bytes allocated for heap objects by CloudConnector and its
// child processes
func (cc *CloudConnector) AllocatedMemory() uint {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return uint(m.Alloc / 1024 / 1024)
}

// HeapAllocatedMemory mega bytes of allocated heap objects by CloudConnector and its
// child processes
func (cc *CloudConnector) HeapAllocatedMemory() uint {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return uint(m.HeapAlloc / 1024 / 1024)
}

// GoRoutinesSpawned How many Go routines are currently being executed by CloudConnector and its
// child processes
func (cc *CloudConnector) GoRoutinesSpawned() int {
	return runtime.NumGoroutine()
}

// State CloudConnector current state
func (cc *CloudConnector) State() CloudConnectorState {
	return cc.state
}

// Kill Begins shutdown procedure
func (cc *CloudConnector) Kill() {
	cc.serverShutdownWaitGroup.Done()
}

// QueriesWaiting How many query messages are still waiting for the response of the IoT Device
func (cc *CloudConnector) QueriesWaiting() uint {
	return cc.connectionsHandler.QueriesWaiting()
}

// CommandsWaiting How many commands messages are still waiting for the response of the IoT Device
func (cc *CloudConnector) CommandsWaiting() uint {
	return cc.connectionsHandler.CommandsWaiting()
}
