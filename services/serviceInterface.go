package services

type ServiceInterface interface {
	// Init
	// shutdownService is the channel where this service will receive a message from Cloud Connector
	// to shut down.
	Init(shutdownService chan bool) error
	// Start Starts the service. This is a blocking operation, waiting for
	// shutdown signal, so run it in a go routine.
	Start()

	Id() string
	// ShutdownChannel
	// Returned chan bool, is the channel where this service will notify Cloud Connector
	// that this service was gracefully shutdown.
	ShutdownChannel() chan bool
}
