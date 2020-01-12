package servers

/*
ServerInterface Defines the required methods in order to start a server that
handles connections with IoT devices and uses CloudConnector for connections
stats.
*/
type ServerInterface interface {
	Name() string
	Start(cloudConnector *CloudConnector) error
	Shutdown(*chan bool) error
}