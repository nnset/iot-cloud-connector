package servers

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/nnset/iot-cloud-connector/storage"

	"github.com/sirupsen/logrus"
	"gotest.tools/assert"

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
	connectionsStats storage.DeviceConnectionsStatsStorageInterface
}

func (d *dummyConnectionsHandler) Listen(shutdownChannel, shutdownIsCompleteChannel *chan bool, log *logrus.Logger) error {
	return nil
}

func (d *dummyConnectionsHandler) Stats() storage.DeviceConnectionsStatsStorageInterface {
	return d.connectionsStats
}

func TestCloudServerNamedConstructorShouldReturnAPointerToANewInstance(t *testing.T) {
	log := createLogger()

	connectionsHandler := dummyConnectionsHandler{
		connectionsStats: storage.NewInMemoryDeviceConnectionsStatsStorage(),
	}

	s := NewCloudConnector("localhost", "9090", "tcp", log, &connectionsHandler, nil)

	assert.Assert(t, s != nil)

	_, err := uuid.Parse(s.ID())
	assert.Assert(t, err == nil, err)
}

func TestCreatingACloudServerShouldSetItsStateToCreated(t *testing.T) {
	log := createLogger()
	connectionsHandler := dummyConnectionsHandler{}

	s := NewCloudConnector("localhost", "9090", "tcp", log, &connectionsHandler, nil)

	assert.Assert(t, s.State() == CloudConnectorCreated)
}

func TestStartingACloudServerShouldSetItsStateToStarted(t *testing.T) {
	log := createLogger()
	connectionsHandler := dummyConnectionsHandler{}

	s := NewCloudConnector("localhost", "9090", "tcp", log, &connectionsHandler, nil)

	assert.Assert(t, s.State() == CloudConnectorCreated)

	defer s.Kill()
	go s.Start()
	time.Sleep(20 * time.Millisecond)

	assert.Assert(t, s.State() == CloudConnectorStarted)
}

func TestIncomingMessagesCanBeFilteredByConnectedDeviceID(t *testing.T) {
	log := createLogger()

	inMemoryConnectionsStats := storage.NewInMemoryDeviceConnectionsStatsStorage()
	inMemoryConnectionsStats.Add("abc-123", "device_abc", "sensor", "userAgent", "192.168.1.100")
	inMemoryConnectionsStats.Add("abc-456", "device_xyz", "sensor", "userAgent", "192.168.1.101")
	inMemoryConnectionsStats.IncomingMessageReceived("device_abc")
	inMemoryConnectionsStats.IncomingMessageReceived("device_xyz")

	connectionsHandler := dummyConnectionsHandler{
		connectionsStats: inMemoryConnectionsStats,
	}

	s := NewCloudConnector("localhost", "9091", "tcp", log, &connectionsHandler, nil)

	assert.Assert(t, s.State() == CloudConnectorCreated)

	defer s.Kill()
	go s.Start()
	time.Sleep(20 * time.Millisecond)

	assert.Assert(t, s.State() == CloudConnectorStarted)
	assert.Assert(t, s.IncomingMessages("device_abc") == 1)
	assert.Assert(t, s.IncomingMessages("device_xyz") == 1)
	assert.Assert(t, s.IncomingMessages("") == 2)
}

func TestConnectedDevicesIDsShouldBeListable(t *testing.T) {
	log := createLogger()

	inMemoryConnectionsStats := storage.NewInMemoryDeviceConnectionsStatsStorage()
	inMemoryConnectionsStats.Add("abc-123", "device_abc", "sensor", "userAgent", "192.168.1.100")
	inMemoryConnectionsStats.Add("abc-456", "device_xyz", "sensor", "userAgent", "192.168.1.101")

	connectionsHandler := dummyConnectionsHandler{
		connectionsStats: inMemoryConnectionsStats,
	}

	s := NewCloudConnector("localhost", "9091", "tcp", log, &connectionsHandler, nil)

	expectedIDs := []string{"device_abc", "device_xyz"}

	assert.DeepEqual(t, expectedIDs, s.ConnectedDevices())
}
