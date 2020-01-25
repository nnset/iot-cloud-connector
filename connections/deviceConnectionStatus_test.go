package connections

import(
    "testing"
    "time"
    "gotest.tools/assert"
)

func TestDeviceConnectionStatusNamedConstructorShouldReturnAPointerToDeviceConnectionStatus(t *testing.T) {
    deviceConnectionStatus := NewDeviceConnectionStatus("abc-123", "device_id", "device_type", "agent", "192.168.1.100")

    assert.Assert(t, deviceConnectionStatus != nil)
    assert.Assert(t, deviceConnectionStatus.RemoteAddress() == "192.168.1.100")
}

func TestDeviceConectionStatusDurationShouldReturnHowManySecondsTheConnectionsHasBeenActive(t *testing.T) {
    deviceConnectionStatus := NewDeviceConnectionStatus("abc-123", "device_id", "device_type", "agent", "192.168.1.100")
    
    duration, _ := deviceConnectionStatus.Uptime()
    assert.Assert(t, duration == 0)

    time.Sleep(1 * time.Second)

    duration, _ = deviceConnectionStatus.Uptime()
    assert.Assert(t, duration == 1)
}

func TestWhenAMessageIsSentStatsAreUpdated(t *testing.T) {
    deviceConnectionStatus := NewDeviceConnectionStatus("abc-123", "device_id", "device_type", "agent", "192.168.1.100")
    assert.Assert(t, deviceConnectionStatus.SentMessages() == 0)
    assert.Assert(t, deviceConnectionStatus.LastSentMessageTimeStamp() == 0)

    deviceConnectionStatus.MessageSent()

    assert.Assert(t, deviceConnectionStatus.SentMessages() == 1)
    assert.Assert(t, deviceConnectionStatus.LastSentMessageTimeStamp() == time.Now().Unix())
}

func TestWhenAMessageIsReceivedStatsAreUpdated(t *testing.T) {
    deviceConnectionStatus := NewDeviceConnectionStatus("abc-123", "device_id", "device_type", "agent", "192.168.1.100")
    assert.Assert(t, deviceConnectionStatus.ReceivedMessages() == 0)
    assert.Assert(t, deviceConnectionStatus.LastReceivedMessageTimeStamp() == 0)

    deviceConnectionStatus.MessageReceived()

    assert.Assert(t, deviceConnectionStatus.ReceivedMessages() == 1)
    assert.Assert(t, deviceConnectionStatus.LastReceivedMessageTimeStamp() == time.Now().Unix())
}
