package servers

import (
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/nnset/iot-cloud-connector/storage"

	"github.com/sirupsen/logrus"
	"gotest.tools/assert"
	is "gotest.tools/assert/cmp"

	"github.com/google/uuid"
)

func createLogger() *logrus.Logger {
	var log = logrus.New()

	log.SetLevel(logrus.DebugLevel)
	log.Out = os.Stderr

	file, err := os.OpenFile("../var/log/cloud-connector-test.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)

	if err == nil {
		log.Out = file
	} else {
		fmt.Println("Using stdErr for log")
	}

	return log
}

type dummyConnectionsHandler struct {
	connections storage.DeviceConnectionsStorageInterface
}

func (d *dummyConnectionsHandler) Start(shutdownChannel, shutdownIsCompleteChannel *chan bool, connections storage.DeviceConnectionsStorageInterface, log *logrus.Logger) error {
	return nil
}

func (d *dummyConnectionsHandler) Stats() storage.DeviceConnectionsStorageInterface {
	return d.connections
}

func (d *dummyConnectionsHandler) SendCommand(payload, deviceID string) (string, int, error) {
	if deviceID == "dummy_id" {
		return "Command OK", 200, nil
	}

	return "", 404, errors.New("Device not connected")
}

func (d *dummyConnectionsHandler) SendQuery(payload, deviceID string) (string, int, error) {
	if deviceID == "dummy_id" {
		return "Query OK", 200, nil
	}

	return "", 404, errors.New("Device not connected")
}

func (d *dummyConnectionsHandler) QueriesWaiting() uint {
	return 1
}

func (d *dummyConnectionsHandler) CommandsWaiting() uint {
	return 2
}

func (d *dummyConnectionsHandler) AuthenticateNewConnection(authData string) error {
	return nil
}

func TestCloudServerNamedConstructorShouldReturnAPointerToANewInstance(t *testing.T) {
	log := createLogger()

	connectionsHandler := dummyConnectionsHandler{}

	s := NewCloudConnector(log, &connectionsHandler, storage.NewInMemoryDeviceConnectionsStorage(), nil, nil)

	assert.Assert(t, s != nil)

	_, err := uuid.Parse(s.ID())
	assert.Assert(t, err == nil, err)
}

func TestCreatingACloudServerShouldSetItsStateToCreated(t *testing.T) {
	log := createLogger()
	connectionsHandler := dummyConnectionsHandler{}

	s := NewCloudConnector(log, &connectionsHandler, storage.NewInMemoryDeviceConnectionsStorage(), nil, nil)

	assert.Assert(t, s.State() == CloudConnectorCreated)
}

func TestIncomingMessagesCanBeFilteredByConnectedDeviceID(t *testing.T) {
	log := createLogger()

	inMemoryConnections := storage.NewInMemoryDeviceConnectionsStorage()
	inMemoryConnections.Add("abc-123", "device_abc", "device_ab_name", "sensor", "userAgent", "192.168.1.100")
	inMemoryConnections.Add("abc-456", "device_xyz", "device_xyz_name", "sensor", "userAgent", "192.168.1.101")
	inMemoryConnections.MessageWasReceived("device_abc")
	inMemoryConnections.MessageWasReceived("device_xyz")

	connectionsHandler := dummyConnectionsHandler{}

	s := NewCloudConnector(log, &connectionsHandler, inMemoryConnections, nil, nil)

	defer func() {
		s.Kill()
		time.Sleep(20 * time.Millisecond)
	}()
	go s.Start(5)
	time.Sleep(20 * time.Millisecond)

	assert.Assert(t, s.ReceivedMessages("device_abc") == 1)
	assert.Assert(t, s.ReceivedMessages("device_xyz") == 1)
	assert.Assert(t, s.ReceivedMessages("") == 2)
}

func TestConnectedDevicesIDsShouldBeListable(t *testing.T) {
	log := createLogger()

	inMemoryConnections := storage.NewInMemoryDeviceConnectionsStorage()
	inMemoryConnections.Add("abc-123", "device_abc", "device_abc_name", "sensor", "userAgent", "192.168.1.100")
	inMemoryConnections.Add("abc-456", "device_xyz", "device_xyz_name", "sensor", "userAgent", "192.168.1.101")

	connectionsHandler := dummyConnectionsHandler{}

	s := NewCloudConnector(log, &connectionsHandler, inMemoryConnections, nil, nil)

	assert.Equal(t, 2, len(s.ConnectedDevices()))
}

func TestQueriesWaitingShouldReturnWhatConnectionsHandlerReturns(t *testing.T) {
	log := createLogger()

	connectionsHandler := dummyConnectionsHandler{}

	s := NewCloudConnector(log, &connectionsHandler, storage.NewInMemoryDeviceConnectionsStorage(), nil, nil)

	assert.Equal(t, uint(1), s.QueriesWaiting())
}

func TestCommandsWaitingShouldReturnWhatConnectionsHandlerReturns(t *testing.T) {
	log := createLogger()

	connectionsHandler := dummyConnectionsHandler{}

	s := NewCloudConnector(log, &connectionsHandler, storage.NewInMemoryDeviceConnectionsStorage(), nil, nil)

	assert.Equal(t, uint(2), s.CommandsWaiting())
}

func TestSendingACommandShouldReturnWhatConnectionsHandlerReturns(t *testing.T) {
	log := createLogger()

	connectionsHandler := dummyConnectionsHandler{}

	s := NewCloudConnector(log, &connectionsHandler, storage.NewInMemoryDeviceConnectionsStorage(), nil, nil)

	response, responseCode, err := s.SendCommand("payload", "dummy_id")

	assert.Equal(t, "Command OK", response)
	assert.Equal(t, 200, responseCode)
	assert.Equal(t, nil, err)
}

func TestSendingAQueryShouldReturnWhatConnectionsHandlerReturns(t *testing.T) {
	log := createLogger()

	connectionsHandler := dummyConnectionsHandler{}

	s := NewCloudConnector(log, &connectionsHandler, storage.NewInMemoryDeviceConnectionsStorage(), nil, nil)

	response, responseCode, err := s.SendQuery("payload", "dummy_id")

	assert.Equal(t, "Query OK", response)
	assert.Equal(t, 200, responseCode)
	assert.Equal(t, nil, err)
}

func TestSendingAQueryToADeviceThatIsNotConnectedShouldReturnAnError(t *testing.T) {
	log := createLogger()

	connectionsHandler := dummyConnectionsHandler{}

	s := NewCloudConnector(log, &connectionsHandler, storage.NewInMemoryDeviceConnectionsStorage(), nil, nil)

	response, responseCode, err := s.SendQuery("payload", "abc-123")

	assert.Equal(t, "", response)
	assert.Equal(t, 404, responseCode)
	assert.Error(t, err, "Device not connected")
}

func TestSubscribingToSystemMetricsStreamShouldIncreaseTheCountOfSubscribers(t *testing.T) {
	log := createLogger()

	connectionsHandler := dummyConnectionsHandler{}

	s := NewCloudConnector(log, &connectionsHandler, storage.NewInMemoryDeviceConnectionsStorage(), nil, nil)

	defer func() {
		s.Kill()
		time.Sleep(20 * time.Millisecond)
	}()

	go s.Start(5)
	time.Sleep(20 * time.Millisecond)

	messageChannel := make(chan SystemMetricChangedMessage)

	s.SubscribeToSystemMetricsStream(messageChannel)

	assert.Equal(t, uint(1), s.SystemMetricsStreamSubscriptions())
}

func TestUnSubscribingToSystemMetricsStreamShouldDecreaseTheCountOfSubscribers(t *testing.T) {
	log := createLogger()

	connectionsHandler := dummyConnectionsHandler{}

	s := NewCloudConnector(log, &connectionsHandler, storage.NewInMemoryDeviceConnectionsStorage(), nil, nil)

	defer func() {
		s.Kill()
		time.Sleep(20 * time.Millisecond)
	}()

	go s.Start(5)
	time.Sleep(20 * time.Millisecond)

	messageChannel := make(chan SystemMetricChangedMessage)

	s.SubscribeToSystemMetricsStream(messageChannel)
	assert.Equal(t, uint(1), s.SystemMetricsStreamSubscriptions())

	s.UnSubscribeToSystemMetricsStream(messageChannel)
	assert.Equal(t, uint(0), s.SystemMetricsStreamSubscriptions())
}

func TestSubscriptionsToSystemMetricStreamShouldReceiveMessages(t *testing.T) {
	log := createLogger()

	connectionsHandler := dummyConnectionsHandler{}

	s := NewCloudConnector(log, &connectionsHandler, storage.NewInMemoryDeviceConnectionsStorage(), nil, nil)

	defer func() {
		s.Kill()
		time.Sleep(20 * time.Millisecond)
	}()

	go s.Start(1)
	time.Sleep(20 * time.Millisecond)

	messageChannel := make(chan SystemMetricChangedMessage)

	s.SubscribeToSystemMetricsStream(messageChannel)

	availableMetrics := make([]string, 10)

	availableMetrics[0] = string(OpenConnections)
	availableMetrics[1] = string(ReceivedMessages)
	availableMetrics[2] = string(SentMessages)
	availableMetrics[3] = string(SystemMemory)
	availableMetrics[4] = string(AllocatedMemory)
	availableMetrics[5] = string(HeapAllocatedMemory)
	availableMetrics[6] = string(GoRoutines)
	availableMetrics[7] = string(CommandsWaiting)
	availableMetrics[8] = string(QueriesWaiting)
	availableMetrics[9] = string(StartTime)

	select {
	case m := <-messageChannel:
		assert.Assert(t, is.Contains(availableMetrics, m.Metric))
	case <-time.After(3 * time.Second):
		// Timeout, no message received
		assert.Equal(t, 0, 1)
	}
}
