package servers

import (
	"strconv"
	"sync"
	"time"
)

type DefaultSystemMetricsStream struct {
	publishInterval                  uint
	cc                               *CloudConnector
	systemMetricsStreamTicker        *time.Ticker
	systemMetricsStreamTickerDone    chan bool
	systemMetricsStreamSubscriptions map[chan SystemMetricChangedMessage]bool
	metrics                          map[string]string
	dataMutex                        sync.Mutex
	previousPublishedSSESubscribers  uint
}

func NewDefaultSystemMetricsStream(publishInterval uint, cc *CloudConnector) *DefaultSystemMetricsStream {

	interval := publishInterval

	if publishInterval == 0 {
		interval = 5
	}

	return &DefaultSystemMetricsStream{
		publishInterval:                  interval,
		cc:                               cc,
		dataMutex:                        sync.Mutex{},
		metrics:                          make(map[string]string),
		systemMetricsStreamSubscriptions: make(map[chan SystemMetricChangedMessage]bool),
		systemMetricsStreamTickerDone:    make(chan bool),
	}
}

func (stream *DefaultSystemMetricsStream) Start() {
	stream.systemMetricsStreamTicker = time.NewTicker(time.Duration(stream.publishInterval) * time.Second)

	stream.run()
}

func (stream *DefaultSystemMetricsStream) Stop() {
	stream.cc.log.Debug("Stoping SystemMetricsStream")
	stream.systemMetricsStreamTicker.Stop()
}

func (stream *DefaultSystemMetricsStream) run() {
	// This stream in order to avoid too much network traffic will perform as a
	// updates buffer and report changes time to time and if value experienced a
	// relevant change.

	for {
		select {
		case <-stream.systemMetricsStreamTickerDone:
			return
		case <-stream.systemMetricsStreamTicker.C:
			previousMetrics := make(map[string]string)

			for metricName, value := range stream.metrics {
				previousMetrics[metricName] = value
			}

			stream.updateMetrics()
			stream.publishChangedMetrics(previousMetrics)

			previousMetrics = nil
		}
	}
}

func (stream *DefaultSystemMetricsStream) updateMetrics() {
	stream.dataMutex.Lock()
	defer stream.dataMutex.Unlock()

	stream.metrics = stream.cc.SystemMetrics()
}

func (stream *DefaultSystemMetricsStream) publishChangedMetrics(previousMetrics map[string]string) {
	stream.dataMutex.Lock()
	defer stream.dataMutex.Unlock()

	for metricName, currentValue := range stream.metrics {
		if previousMetrics[metricName] != currentValue {
			stream.publishMetric(metricName, previousMetrics[metricName], currentValue)
		}
	}

	if stream.sseSubscribersChanged() {
		stream.publishMetric(
			string(SSESubscribers),
			strconv.Itoa(int(stream.previousPublishedSSESubscribers)),
			strconv.Itoa(int(len(stream.systemMetricsStreamSubscriptions))),
		)

		stream.previousPublishedSSESubscribers = uint(len(stream.systemMetricsStreamSubscriptions))
	}
}

func (stream *DefaultSystemMetricsStream) sseSubscribersChanged() bool {

	return stream.previousPublishedSSESubscribers != uint(len(stream.systemMetricsStreamSubscriptions))
}

func (stream *DefaultSystemMetricsStream) publishMetric(metricName, previousValue, currentValue string) {
	stream.cc.log.Debugf("%s changed from %s to %s", metricName, previousValue, currentValue)

	message := SystemMetricChangedMessage{metricName, currentValue}

	for messageChannel := range stream.systemMetricsStreamSubscriptions {
		messageChannel <- message
	}
}

// SubscribeToSystemMetricsStream Subscribe a SystemMetricChangedMessage channel to receive messages
// every time a System Metric changes.
func (stream *DefaultSystemMetricsStream) SubscribeToSystemMetricsStream(channel chan SystemMetricChangedMessage) {
	stream.dataMutex.Lock()
	defer stream.dataMutex.Unlock()

	stream.systemMetricsStreamSubscriptions[channel] = true
}

// UnSubscribeToSystemMetricsStream UnSubscribe a SystemMetricChangedMessage channel
func (stream *DefaultSystemMetricsStream) UnSubscribeToSystemMetricsStream(channel chan SystemMetricChangedMessage) {
	stream.dataMutex.Lock()
	defer stream.dataMutex.Unlock()

	delete(stream.systemMetricsStreamSubscriptions, channel)
}

// SystemMetricsStreamSubscriptions How many channels are subscrives to receice System Metrics updates
func (stream *DefaultSystemMetricsStream) SystemMetricsStreamSubscriptions() uint {
	return uint(len(stream.systemMetricsStreamSubscriptions))
}
