package connections

import (
	"testing"
	"time"

	"gotest.tools/assert"
)

func TestDeviceConnectionNamedConstructorShouldReturnAPointerToDeviceConnection(t *testing.T) {
	deviceConnection := NewDeviceConnection("abc-123", "device_id", "device_type", "agent", "192.168.1.100")

	assert.Assert(t, deviceConnection != nil)
	assert.Assert(t, deviceConnection.RemoteAddress() == "192.168.1.100")
}

func TestDeviceConectionStatusDurationShouldReturnHowManySecondsTheConnectionsHasBeenActive(t *testing.T) {
	deviceConnection := NewDeviceConnection("abc-123", "device_id", "device_type", "agent", "192.168.1.100")

	duration, _ := deviceConnection.Uptime()
	assert.Assert(t, duration == 0)

	time.Sleep(1 * time.Second)

	duration, _ = deviceConnection.Uptime()
	assert.Assert(t, duration == 1)
}

func TestWhenAMessageIsSentStatsAreUpdated(t *testing.T) {
	deviceConnection := NewDeviceConnection("abc-123", "device_id", "device_type", "agent", "192.168.1.100")
	assert.Assert(t, deviceConnection.SentMessages() == 0)
	assert.Assert(t, deviceConnection.LastSentMessageTimeStamp() == 0)

	deviceConnection.MessageSent()

	assert.Assert(t, deviceConnection.SentMessages() == 1)
	assert.Assert(t, deviceConnection.LastSentMessageTimeStamp() == time.Now().Unix())
}

func TestWhenAMessageIsReceivedStatsAreUpdated(t *testing.T) {
	deviceConnection := NewDeviceConnection("abc-123", "device_id", "device_type", "agent", "192.168.1.100")
	assert.Assert(t, deviceConnection.ReceivedMessages() == 0)
	assert.Assert(t, deviceConnection.LastReceivedMessageTimeStamp() == 0)

	deviceConnection.MessageReceived()

	assert.Assert(t, deviceConnection.ReceivedMessages() == 1)
	assert.Assert(t, deviceConnection.LastReceivedMessageTimeStamp() == time.Now().Unix())
}
