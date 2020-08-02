package servers

type SystemMetricsStreamInterface interface {
	Start()
	Stop()
	SubscribeToSystemMetricsStream(channel chan SystemMetricChangedMessage)
	UnSubscribeToSystemMetricsStream(channel chan SystemMetricChangedMessage)
	SystemMetricsStreamSubscriptions() uint
}
