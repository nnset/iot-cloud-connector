package handlers

import (
	"github.com/sirupsen/logrus"
)

/*
ConnectionsHandlerInterface This will be your domain layer, you will use any communication protocol,
just implement this interface so CloudConnector will be able to support your server.
*/
type ConnectionsHandlerInterface interface {
	Listen(shutdownChannel, shutdownIsCompleteChannel *chan bool, log *logrus.Logger) error

	IncomingMessagesProcessed() uint
	OpenConnections() uint
}
