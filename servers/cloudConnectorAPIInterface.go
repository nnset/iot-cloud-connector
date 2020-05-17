package servers

// CloudConnectorAPIInterface In order to enable a REST API linked CloudConnector
// just implement these methods and let CloudConnector to start the API.
type CloudConnectorAPIInterface interface {
	Start(cloudConnector *CloudConnector) error
	Stop()
}
