package events

type ServerShutdownChannels struct {
	ShutdownConnectionsHandler chan bool
	ShutdownIsComplete         chan bool
}

func NewServerShutdownChannels() ServerShutdownChannels {
	return ServerShutdownChannels{make(chan bool), make(chan bool)}
}
