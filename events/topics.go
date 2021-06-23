package events

const (
	SystemMetricsNumGoRoutinesTopic   string = "system_metrics::num_go_routines"
	SystemMetricsAllocatedMemoryTopic string = "system_metrics::allocated_memory"
	ConnectionEstablished             string = "connections::established"
	ConnectionClosed                  string = "connections::closed"
)
