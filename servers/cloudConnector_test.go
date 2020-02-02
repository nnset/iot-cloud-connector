package servers

import (
    "os"
    "fmt"
    "time"
    "testing"
    "github.com/sirupsen/logrus"
	"github.com/nnset/iot-cloud-connector/storage"
	"github.com/nnset/iot-cloud-connector/connections"
    "gotest.tools/assert"
    //is "gotest.tools/assert/cmp"
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

type dummyServer struct { }

func (s *dummyServer) Name() string {
    return "Dummy server"
}

func (s *dummyServer) Start(cloudConnector *CloudConnector) error {
    return nil
}

func (s *dummyServer) Shutdown(*chan bool) error {
    return nil
}


type dummyNetworkConnection struct { }

func (d *dummyNetworkConnection) Close(statusCode connections.ConnectionStatusCode, reason string) error {
    return nil
}

func TestCloudConnectorNamedConstructorShouldReturnAPointerToANewInstance(t *testing.T) {
    shutdownChannel := make(chan bool, 1)
    cloudConnector := NewCloudConnector(
        &shutdownChannel, createLogger(), "", "", 
        &dummyServer{}, storage.NewInMemoryDeviceConnectionsStorage(),
    )

    assert.Assert(t, cloudConnector != nil)
    
    _, err := uuid.Parse(cloudConnector.ID())
    assert.Assert(t, err == nil, err)
}

func TestStartingACloudConnectorShouldUpdateItsServerStatusToOnline(t *testing.T) {
    shutdownChannel := make(chan bool, 1)
    log := createLogger()

    cloudConnector := NewCloudConnector(
        &shutdownChannel, log, "", "", 
        &dummyServer{}, storage.NewInMemoryDeviceConnectionsStorage(),
    )
    
    assert.Assert(t, cloudConnector.IsServerOnline() == false)

    go cloudConnector.Start()

    time.Sleep(100 * time.Millisecond)
    assert.Assert(t, cloudConnector.IsServerOnline() == true)
}

func TestShutdowningACloudConnectorShouldUpdateItsServerStatusToOffline(t *testing.T) {
    shutdownChannel := make(chan bool, 1)
    log := createLogger()

    cloudConnector := NewCloudConnector(
        &shutdownChannel, log, "", "", 
        &dummyServer{}, storage.NewInMemoryDeviceConnectionsStorage(),
    )
    
    assert.Assert(t, cloudConnector.IsServerOnline() == false)

    go cloudConnector.Start()

    time.Sleep(500 * time.Millisecond)
    assert.Assert(t, cloudConnector.IsServerOnline() == true)

    shutdownChannel <- true

    time.Sleep(500 * time.Millisecond)

    assert.Assert(t, cloudConnector.IsServerOnline() == false)
}

func TestServerShouldBeRebootable(t *testing.T) {
    shutdownChannel := make(chan bool, 1)
    log := createLogger()

    cloudConnector := NewCloudConnector(
        &shutdownChannel, log, "", "", 
        &dummyServer{}, storage.NewInMemoryDeviceConnectionsStorage(),
    )
    
    assert.Assert(t, cloudConnector.IsServerOnline() == false)
    go cloudConnector.Start()
    time.Sleep(500 * time.Millisecond)
    assert.Assert(t, cloudConnector.IsServerOnline() == true)
    
    shutdownChannel <- true
    time.Sleep(500 * time.Millisecond)
    assert.Assert(t, cloudConnector.IsServerOnline() == false)

    
    go cloudConnector.Start()
    time.Sleep(500 * time.Millisecond)
    assert.Assert(t, cloudConnector.IsServerOnline() == true)

    shutdownChannel <- true
    time.Sleep(500 * time.Millisecond)
    assert.Assert(t, cloudConnector.IsServerOnline() == false)
}

func TestIncomingConnectionsShouldUpdateCloudConnectorStats(t *testing.T) {
    shutdownChannel := make(chan bool, 1)
    log := createLogger()

    cloudConnector := NewCloudConnector(
        &shutdownChannel, log, "", "", 
        &dummyServer{}, storage.NewInMemoryDeviceConnectionsStorage(),
    )

	deviceConnection := connections.NewDeviceConnection(
		&dummyNetworkConnection{}, "device_id", "10.10.10.1", "user_agent",
	)
	
	err := cloudConnector.ConnectionEstablished(deviceConnection)

	assert.Assert(t, err == nil, err)
	assert.Assert(t, cloudConnector.TotalConnections() == 1)
}

func TestClosingConnectionsShouldUpdateCloudConnectorStats(t *testing.T) {
    shutdownChannel := make(chan bool, 1)
    log := createLogger()

    cloudConnector := NewCloudConnector(
        &shutdownChannel, log, "", "", 
        &dummyServer{}, storage.NewInMemoryDeviceConnectionsStorage(),
    )

	deviceConnection := connections.NewDeviceConnection(
		&dummyNetworkConnection{}, "device_id", "10.10.10.1", "user_agent",
	)
	
	err := cloudConnector.ConnectionEstablished(deviceConnection)

	assert.Assert(t, err == nil, err)
	assert.Assert(t, cloudConnector.TotalConnections() == 1)

	err = cloudConnector.ConnectionClosed(deviceConnection.ID(), connections.StatusNormalClosure, "connection closed")
	
	assert.Assert(t, err == nil, err)
	assert.Assert(t, cloudConnector.TotalConnections() == 0)
}

func TestEstablishedConnectionsMustHaveDifferentIds(t *testing.T) {
    shutdownChannel := make(chan bool, 1)
    log := createLogger()

    cloudConnector := NewCloudConnector(
        &shutdownChannel, log, "", "", 
        &dummyServer{}, storage.NewInMemoryDeviceConnectionsStorage(),
    )

	deviceConnection := connections.NewDeviceConnection(
		&dummyNetworkConnection{}, "device_id", "10.10.10.1", "user_agent",
	)
	
	err := cloudConnector.ConnectionEstablished(deviceConnection)

	assert.Assert(t, err == nil, err)
	assert.Assert(t, cloudConnector.TotalConnections() == 1)

	err = cloudConnector.ConnectionEstablished(deviceConnection)
	
	expectedErrorMessage := fmt.Sprintf("Connection rejected. Connection #%s with device #device_id was already established.", deviceConnection.ID())

	assert.Error(t, err, expectedErrorMessage)
	assert.Assert(t, cloudConnector.TotalConnections() == 1)
}

func TestClosingAnInvalidConnectionShouldReturnAnError(t *testing.T) {
    shutdownChannel := make(chan bool, 1)
    log := createLogger()

    cloudConnector := NewCloudConnector(
        &shutdownChannel, log, "", "", 
        &dummyServer{}, storage.NewInMemoryDeviceConnectionsStorage(),
    )

	err := cloudConnector.ConnectionClosed("dummy_id", connections.StatusNormalClosure, "connection closed")
	
	assert.Error(t, err, "Connection not found")
	assert.Assert(t, cloudConnector.TotalConnections() == 0)
}

func TestReceivingMessagesShouldUpdateCloudConnectorStats(t *testing.T) {
    shutdownChannel := make(chan bool, 1)
    log := createLogger()

    cloudConnector := NewCloudConnector(
        &shutdownChannel, log, "", "", 
        &dummyServer{}, storage.NewInMemoryDeviceConnectionsStorage(),
    )

	deviceConnection := connections.NewDeviceConnection(
		&dummyNetworkConnection{}, "device_id", "10.10.10.1", "user_agent",
	)
	
	err := cloudConnector.ConnectionEstablished(deviceConnection)
	assert.Assert(t, err == nil, err)
	
	cloudConnector.MessageReceived(deviceConnection.ID())

	assert.Assert(t, cloudConnector.ReceivedMessages(deviceConnection.ID()) == 1)
	assert.Assert(t, cloudConnector.SentMessages(deviceConnection.ID()) == 0)
}

func TestSendingMessagesToAConnectedDeviceShouldUpdateCloudConnectorStats(t *testing.T) {
    shutdownChannel := make(chan bool, 1)
    log := createLogger()

    cloudConnector := NewCloudConnector(
        &shutdownChannel, log, "", "", 
        &dummyServer{}, storage.NewInMemoryDeviceConnectionsStorage(),
    )

	deviceConnection := connections.NewDeviceConnection(
		&dummyNetworkConnection{}, "device_id", "10.10.10.1", "user_agent",
	)
	
	err := cloudConnector.ConnectionEstablished(deviceConnection)
	assert.Assert(t, err == nil, err)
	
	cloudConnector.MessageSent(deviceConnection.ID())

	assert.Assert(t, cloudConnector.SentMessages(deviceConnection.ID()) == 1)
	assert.Assert(t, cloudConnector.ReceivedMessages(deviceConnection.ID()) == 0)
}