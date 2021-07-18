package events

const (
	SystemMetricsAllocatedMemoryTopic string = "system_metrics::allocated_memory"
	SystemMetricsNumGoRoutinesTopic   string = "system_metrics::num_go_routines"
	ConnectionEstablishedTopic        string = "connections::established"
	ConnectionClosedTopic             string = "connections::closed"
	MessageReceivedTopic              string = "connections::message_received"
	MessageSentTopic                  string = "connections::message_sent"
)
