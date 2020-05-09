package servers

/*
CloudConnectorAPIInterface
*/
type CloudConnectorAPIInterface interface {
	Start() error
	Stop()
}
