package services

import (
	"runtime"
	"strconv"
	"time"

	"github.com/nnset/iot-cloud-connector/events"

	"github.com/google/uuid"
	"github.com/nnset/iot-cloud-connector/bus"
)

type DefaultSystemMetricsService struct {
	PublishInterval           int // Seconds
	eventBus                  bus.MessageBus
	serviceIsShutdown         chan bool
	shutdownService           chan bool
	id                        string
	publishMetricsTicker      *time.Ticker
	metricsLastPublishedValue map[string]string
}

// NewDefaultSystemMetricsService Creates a new instance of NewDefaultSystemMetricsService
func NewDefaultSystemMetricsService(
	eventBus bus.MessageBus,
	publishInterval int,
) *DefaultSystemMetricsService {
	return &DefaultSystemMetricsService{
		id:                        uuid.New().String(),
		eventBus:                  eventBus,
		PublishInterval:           publishInterval,
		metricsLastPublishedValue: make(map[string]string),
	}
}

func (service *DefaultSystemMetricsService) Id() string {
	return service.id
}

func (service *DefaultSystemMetricsService) Init(shutdownService chan bool) error {
	service.shutdownService = shutdownService
	service.serviceIsShutdown = make(chan bool)

	if service.PublishInterval == 0 {
		service.PublishInterval = 15
	}

	return nil
}

func (service *DefaultSystemMetricsService) Start() {
	service.publishMetricsTicker = time.NewTicker(time.Duration(service.PublishInterval) * time.Second)

	for {
		select {
		case <-service.shutdownService:
			service.serviceIsShutdown <- true
			return
		case <-service.publishMetricsTicker.C:
			service.publishMetrics()
		}
	}
}

func (service *DefaultSystemMetricsService) ShutdownChannel() chan bool {
	return service.serviceIsShutdown
}

func (service *DefaultSystemMetricsService) publishMetrics() {

	service.publishMetric(
		events.SystemMetricsNumGoRoutinesTopic, strconv.Itoa(service.goRoutinesSpawned()),
	)

	service.publishMetric(
		events.SystemMetricsAllocatedMemoryTopic, strconv.Itoa(service.allocatedMemory()),
	)
}

func (service *DefaultSystemMetricsService) publishMetric(topic string, currentValue string) {
	previousValue, exists := service.metricsLastPublishedValue[topic]

	if !exists || previousValue != currentValue {
		service.eventBus.Publish(topic, events.NewMessage(currentValue, events.Default))
		service.metricsLastPublishedValue[topic] = currentValue
	}
}

func (service *DefaultSystemMetricsService) goRoutinesSpawned() int {
	return runtime.NumGoroutine()
}

// AllocatedMemory mega bytes allocated for heap objects
func (service *DefaultSystemMetricsService) allocatedMemory() int {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return int(m.Alloc / 1024 / 1024)
}
