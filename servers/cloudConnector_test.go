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
	shutdownServer := make(chan bool, 1)
	connectionsHandler := dummyConnectionsHandler{
		connectionsStats: storage.NewInMemoryDeviceConnectionsStatsStorage(),
	}

	s := NewCloudConnector("localhost", "9090", "tcp", log, &shutdownServer, &connectionsHandler, nil)

	assert.Assert(t, s != nil)

	_, err := uuid.Parse(s.ID())
	assert.Assert(t, err == nil, err)
	shutdownServer <- true
}

func TestCreatingACloudServerShouldSetItsStateToCreated(t *testing.T) {
	log := createLogger()
	shutdownServer := make(chan bool, 1)
	connectionsHandler := dummyConnectionsHandler{}

	s := NewCloudConnector("localhost", "9090", "tcp", log, &shutdownServer, &connectionsHandler, nil)

	assert.Assert(t, s.State() == CloudConnectorCreated)
	shutdownServer <- true
}

func TestStartingACloudServerShouldSetItsStateToStarted(t *testing.T) {
	log := createLogger()
	shutdownServer := make(chan bool, 1)
	connectionsHandler := dummyConnectionsHandler{}

	s := NewCloudConnector("localhost", "9090", "tcp", log, &shutdownServer, &connectionsHandler, nil)

	assert.Assert(t, s.State() == CloudConnectorCreated)

	go s.Start()
	time.Sleep(20 * time.Millisecond)

	assert.Assert(t, s.State() == CloudConnectorStarted)
	shutdownServer <- true
}
