package connectionshandlers

import (
	"github.com/nnset/iot-cloud-connector/storage"
	"github.com/sirupsen/logrus"
)

/*
ConnectionsHandlerInterface This will be your domain layer, you will use any communication protocol,
just implement this interface so CloudConnector will be able to support your domain.
*/
type ConnectionsHandlerInterface interface {
	Listen(shutdownChannel, shutdownIsCompleteChannel *chan bool, log *logrus.Logger) error

	Stats() storage.DeviceConnectionsStatsStorageInterface
}
