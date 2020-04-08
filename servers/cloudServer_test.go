package servers

import (
	"fmt"
	"os"
	"testing"
	"time"

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
}

func (d *dummyConnectionsHandler) Listen(shutdownChannel, shutdownIsCompleteChannel *chan bool, log *logrus.Logger) error {
	return nil
}

func (d *dummyConnectionsHandler) IncomingMessages() uint {
	return 0
}

func (d *dummyConnectionsHandler) OutgoingMessages() uint {
	return 0
}

func (d *dummyConnectionsHandler) OpenConnections() uint {
	return 0
}

func TestCloudServerNamedConstructorShouldReturnAPointerToANewInstance(t *testing.T) {
	log := createLogger()
	shutdownServer := make(chan bool, 1)
	connectionsHandler := dummyConnectionsHandler{}

	s := NewCloudServer("localhost", "9090", "tcp", log, &shutdownServer, &connectionsHandler)

	assert.Assert(t, s != nil)

	_, err := uuid.Parse(s.ID())
	assert.Assert(t, err == nil, err)
	shutdownServer <- true
}

func TestCreatingACloudServerShouldSetItsStateToCreated(t *testing.T) {
	log := createLogger()
	shutdownServer := make(chan bool, 1)
	connectionsHandler := dummyConnectionsHandler{}

	s := NewCloudServer("localhost", "9090", "tcp", log, &shutdownServer, &connectionsHandler)

	assert.Assert(t, s.State() == CloudServerCreated)
	shutdownServer <- true
}

func TestStartingACloudServerShouldSetItsStateToStarted(t *testing.T) {
	log := createLogger()
	shutdownServer := make(chan bool, 1)
	connectionsHandler := dummyConnectionsHandler{}

	s := NewCloudServer("localhost", "9090", "tcp", log, &shutdownServer, &connectionsHandler)

	assert.Assert(t, s.State() == CloudServerCreated)

	go s.Start()
	time.Sleep(20 * time.Millisecond)

	assert.Assert(t, s.State() == CloudServerStarted)
	shutdownServer <- true
}
