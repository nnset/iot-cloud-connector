package servers

/*
CloudConnectorAPIInterface
*/
type CloudConnectorAPIInterface interface {
	Start(cloudConnector *CloudConnector) error
	Stop()
}
