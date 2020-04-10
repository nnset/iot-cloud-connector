package servers

/*
StatusApiInterface
*/
type StatusAPIInterface interface {
	Start() error
	Stop()
}
