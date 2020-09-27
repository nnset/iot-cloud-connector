package servers

import (
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/nnset/iot-cloud-connector/connections"
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
	activeConnections       storage.DeviceConnectionsStorageInterface // Does it have to be here or just at connections handler ???
	auth                    APIAuthMiddleWare
	systemMetricsStream     SystemMetricsStreamInterface
}

// SystemMetricChangedMessage message send to subscribed channels when a system metric changes
type SystemMetricChangedMessage struct {
	Metric string `json:"metric"`
	Value  string `json:"value"`
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

// SystemMetric Server's system metrics
type SystemMetric string

const (
	OpenConnections     SystemMetric = "connections"
	ReceivedMessages    SystemMetric = "received_messages"
	SentMessages        SystemMetric = "sent_messages"
	SystemMemory        SystemMetric = "system_memory"
	AllocatedMemory     SystemMetric = "allocated_memory"
	HeapAllocatedMemory SystemMetric = "heap_allocated_memory"
	GoRoutines          SystemMetric = "go_routines"
	CommandsWaiting     SystemMetric = "commands_waiting"
	QueriesWaiting      SystemMetric = "queries_waiting"
	StartTime           SystemMetric = "start_time"
	SSESubscribers      SystemMetric = "sse_subscribers"
)

// NewCloudConnector Creates a new instance of CloudConnector
func NewCloudConnector(
	log *logrus.Logger,
	connectionsHandler connectionshandlers.ConnectionsHandlerInterface,
	activeConnections storage.DeviceConnectionsStorageInterface,
	statusAPI CloudConnectorAPIInterface,
	systemMetricsStream SystemMetricsStreamInterface,
) *CloudConnector {
	return &CloudConnector{
		id:                      uuid.New().String(),
		log:                     log,
		serverShutdownWaitGroup: sync.WaitGroup{},
		connectionsHandler:      connectionsHandler,
		statusAPI:               statusAPI,
		startTime:               time.Now().Unix(),
		state:                   CloudConnectorCreated,
		activeConnections:       activeConnections,
		systemMetricsStream:     systemMetricsStream,
	}
}

// Start Starts ClodConnector and its child processes, currently:
// systemMetricsStreamPublishInterval is in seconds
//  - An instance of connectionshandlers.ConnectionsHandlerInterface via its Listen() method.
//  - An instance of CloudConnectorAPIInterface via its Start() method.
func (cc *CloudConnector) Start(systemMetricsStreamPublishInterval uint) {
	cc.log.Debugf("Starting CloudConnector #%s", cc.id)

	connectionsHandlerShutdownIsComplete, shutdownConnectionsHandler := make(chan bool), make(chan bool)

	cc.serverShutdownWaitGroup.Add(1)
	go cc.waitForShutdownSignal()

	go cc.connectionsHandler.Start(
		&shutdownConnectionsHandler,
		&connectionsHandlerShutdownIsComplete,
		cc.activeConnections,
		cc.log,
	)

	if cc.statusAPI != nil {
		go cc.statusAPI.Start(cc)
	}

	// By default we always set a SystemMetrics stream
	if cc.systemMetricsStream == nil {
		cc.systemMetricsStream = NewServerSentEventsSystemMetricsStream(systemMetricsStreamPublishInterval, cc)
	}

	go cc.systemMetricsStream.Start()

	cc.state = CloudConnectorStarted

	cc.serverShutdownWaitGroup.Wait()

	cc.log.Info("CloudConnector received shutdown signal")

	cc.systemMetricsStream.Stop()
	cc.shutdown(shutdownConnectionsHandler, connectionsHandlerShutdownIsComplete)
}

func (cc *CloudConnector) waitForShutdownSignal() {
	operatingSystemSignal := make(chan os.Signal)

	signal.Notify(operatingSystemSignal, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)

	go func() {
		sig := <-operatingSystemSignal
		cc.log.Debugf("Signal received : %s", sig)
		cc.log.Debug("Shutting down CloudConnector")

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
	cc.log.Info("  Total received messages processed: ", cc.activeConnections.TotalReceivedMessages())
	cc.log.Info("  Total sent messages processed: ", cc.activeConnections.TotalSentMessages())
	cc.log.Infof("  Uptime: %d seconds", cc.Uptime(""))
}

// SubscribeToSystemMetricsStream Subscribe a SystemMetricChangedMessage channel to receive messages
// every time a System Metric changes.
func (cc *CloudConnector) SubscribeToSystemMetricsStream(channel chan SystemMetricChangedMessage) {
	cc.systemMetricsStream.SubscribeToSystemMetricsStream(channel)
}

// UnSubscribeToSystemMetricsStream UnSubscribe a SystemMetricChangedMessage channel
func (cc *CloudConnector) UnSubscribeToSystemMetricsStream(channel chan SystemMetricChangedMessage) {
	cc.systemMetricsStream.UnSubscribeToSystemMetricsStream(channel)
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

	connection, err := cc.activeConnections.Get(deviceID)

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

func (cc *CloudConnector) SendCommand(command connectionshandlers.Command) (string, int, error) {
	return cc.connectionsHandler.SendCommand(command)
}

func (cc *CloudConnector) SendQuery(query connectionshandlers.Query) (string, int, error) {
	return cc.connectionsHandler.SendQuery(query)
}

func (cc *CloudConnector) StartTime() int64 {
	return cc.startTime
}

// OpenConnections How many connections are currently open
func (cc *CloudConnector) OpenConnections() uint {
	return cc.activeConnections.OpenConnections()
}

// ConnectedDevices Which IoT devices (IDs are displayed) are currently connected
func (cc *CloudConnector) ConnectedDevices() []*connections.DeviceConnectionDTO {
	return cc.activeConnections.ConnectedDevices()
}

// ReceivedMessages How many messages were received from a given IoT Device, or if
// deviceID is empty, globally.
func (cc *CloudConnector) ReceivedMessages(deviceID string) uint {
	if deviceID == "" {
		return cc.activeConnections.TotalReceivedMessages()
	}

	return cc.activeConnections.ReceivedMessages(deviceID)
}

// SentMessages How many messages were sent to a given IoT Device, or if
// deviceID is empty, globally.
func (cc *CloudConnector) SentMessages(deviceID string) uint {
	if deviceID == "" {
		return cc.activeConnections.TotalSentMessages()
	}

	return cc.activeConnections.SentMessages(deviceID)
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

// SystemMetricsStreamSubscriptions How many channels are subscrives to receice System Metrics updates
func (cc *CloudConnector) SystemMetricsStreamSubscriptions() uint {
	return cc.systemMetricsStream.SystemMetricsStreamSubscriptions()
}

func (cc *CloudConnector) SystemMetrics() map[string]string {
	metrics := make(map[string]string)

	metrics[string(OpenConnections)] = strconv.Itoa(int(cc.OpenConnections()))
	metrics[string(ReceivedMessages)] = strconv.Itoa(int(cc.ReceivedMessages("")))
	metrics[string(SentMessages)] = strconv.Itoa(int(cc.SentMessages("")))
	metrics[string(SystemMemory)] = strconv.Itoa(int(cc.SystemMemory()))
	metrics[string(AllocatedMemory)] = strconv.Itoa(int(cc.AllocatedMemory()))
	metrics[string(HeapAllocatedMemory)] = strconv.Itoa(int(cc.HeapAllocatedMemory()))
	metrics[string(GoRoutines)] = strconv.Itoa(int(cc.GoRoutinesSpawned()))
	metrics[string(CommandsWaiting)] = strconv.Itoa(int(cc.CommandsWaiting()))
	metrics[string(QueriesWaiting)] = strconv.Itoa(int(cc.QueriesWaiting()))
	metrics[string(StartTime)] = strconv.Itoa(int(cc.startTime))

	return metrics
}
