package connectionshandlers

import (
	"github.com/nnset/iot-cloud-connector/storage"
	"github.com/sirupsen/logrus"
)

// ConnectionsHandlerInterface This will be your domain layer, you may use any communication protocol.
// Just implement this interface so CloudConnector will be able to support your domain logic and your
// IoT devices connections.
type ConnectionsHandlerInterface interface {
	Start(shutdownChannel, shutdownIsCompleteChannel *chan bool, connections storage.DeviceConnectionsStorageInterface, log *logrus.Logger) error

	// Send a Command to a Device and wait for its feedback. This is a Synchronous action.
	SendCommand(command Command) (string, int, error)
	// Send a Query to a Device and wait for the response. This is a Synchronous action.
	SendQuery(query Query) (string, int, error)

	// TODO add a method to allow sending a Message asynchronously

	QueriesWaiting() uint
	CommandsWaiting() uint
}
