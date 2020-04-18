package servers

/*
StatusAPIInterface
*/
type StatusAPIInterface interface {
	Start() error
	Stop()
}
