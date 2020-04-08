package connections

import (
	"testing"
	"time"

	"gotest.tools/assert"
)

func TestDeviceConnectionStatsNamedConstructorShouldReturnAPointerToDeviceConnectionStats(t *testing.T) {
	deviceConnectionStats := NewDeviceConnectionStats("abc-123", "device_id", "device_type", "agent", "192.168.1.100")

	assert.Assert(t, deviceConnectionStats != nil)
	assert.Assert(t, deviceConnectionStats.RemoteAddress() == "192.168.1.100")
}

func TestDeviceConectionStatusDurationShouldReturnHowManySecondsTheConnectionsHasBeenActive(t *testing.T) {
	deviceConnectionStats := NewDeviceConnectionStats("abc-123", "device_id", "device_type", "agent", "192.168.1.100")

	duration, _ := deviceConnectionStats.Uptime()
	assert.Assert(t, duration == 0)

	time.Sleep(1 * time.Second)

	duration, _ = deviceConnectionStats.Uptime()
	assert.Assert(t, duration == 1)
}

func TestWhenAMessageIsSentStatsAreUpdated(t *testing.T) {
	deviceConnectionStats := NewDeviceConnectionStats("abc-123", "device_id", "device_type", "agent", "192.168.1.100")
	assert.Assert(t, deviceConnectionStats.SentMessages() == 0)
	assert.Assert(t, deviceConnectionStats.LastSentMessageTimeStamp() == 0)

	deviceConnectionStats.MessageSent()

	assert.Assert(t, deviceConnectionStats.SentMessages() == 1)
	assert.Assert(t, deviceConnectionStats.LastSentMessageTimeStamp() == time.Now().Unix())
}

func TestWhenAMessageIsReceivedStatsAreUpdated(t *testing.T) {
	deviceConnectionStats := NewDeviceConnectionStats("abc-123", "device_id", "device_type", "agent", "192.168.1.100")
	assert.Assert(t, deviceConnectionStats.ReceivedMessages() == 0)
	assert.Assert(t, deviceConnectionStats.LastReceivedMessageTimeStamp() == 0)

	deviceConnectionStats.MessageReceived()

	assert.Assert(t, deviceConnectionStats.ReceivedMessages() == 1)
	assert.Assert(t, deviceConnectionStats.LastReceivedMessageTimeStamp() == time.Now().Unix())
}
