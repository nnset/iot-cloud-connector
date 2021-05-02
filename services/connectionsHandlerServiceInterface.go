package services

type ConnectionsHandlerServiceInterface interface {
	ServiceInterface
	Port() string
	Address() string
	Network() string
	OpenConnectionsCount() uint
}
