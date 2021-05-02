package services

type ServiceInterface interface {
	// Init
	// shutdownService is the channel where this service will receive a message from Cloud Connector
	// to shut down.
	Init(shutdownService chan bool) error
	// Start Starts the service.
	Start()

	Id() string
	// ShutdownChannel
	// Returned chan bool, is the channel where this service will notify Cloud Connector
	// that this service was gracefully shutdown.
	ShutdownChannel() chan bool
}
