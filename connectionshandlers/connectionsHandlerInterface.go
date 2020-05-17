package connectionshandlers

import (
	"github.com/nnset/iot-cloud-connector/storage"
	"github.com/sirupsen/logrus"
)

// ConnectionsHandlerInterface This will be your domain layer, you may use any communication protocol.
// Just implement this interface so CloudConnector will be able to support your domain logic and your
// IoT devices connections.
type ConnectionsHandlerInterface interface {
	Listen(shutdownChannel, shutdownIsCompleteChannel *chan bool, connectionsStats storage.DeviceConnectionsStatsStorageInterface, log *logrus.Logger) error
	SendCommand(payload, deviceID string) (string, int, error)
	SendQuery(payload, deviceID string) (string, int, error)
	QueriesWaiting() uint
	CommandsWaiting() uint
}
