package entities

import (
	"testing"
	"time"

	"gotest.tools/assert"
)

func TestConnectionNamedConstructorShouldReturnAPointerToAConnection(t *testing.T) {
	connection, _ := NewConnection("device_id", "device_name", "device_type", "agent", "192.168.1.100")

	assert.Assert(t, connection != nil)
	assert.Assert(t, connection.RemoteAddress == "192.168.1.100")
}

func TestConnectionNamedConstructorShouldReturnErrorIfDeviceIdIsEmpty(t *testing.T) {
	_, err := NewConnection("", "device_name", "device_type", "agent", "192.168.1.100")

	assert.Error(t, err, "can not create a new connection: empty deviceID")
}

func TestConnectionNamedConstructorShouldReturnErrorIfRemoteAddressIsEmpty(t *testing.T) {
	_, err := NewConnection("device_id", "device_name", "device_type", "agent", "")

	assert.Error(t, err, "can not create a new connection: empty remoteAddress")
}

func TestCreatingAConnectionFromDefaultPayloadShouldReturnAnIsntanceOfConnection(t *testing.T) {

	connection, err := NewConnectionFromDefaultPayload("{\"device_id\": \"abc-123\"}", "192.168.1.100")

	assert.NilError(t, err)
	assert.Assert(t, connection.DeviceID == "abc-123")
	assert.Assert(t, connection.DeviceName == "")
	assert.Assert(t, connection.DeviceType == "")
	assert.Assert(t, connection.RemoteAddress == "192.168.1.100")
}

func TestCreatingAConnectionFromDefaultPayloadWithNoDeviceIdShouldReturnAnError(t *testing.T) {

	_, err := NewConnectionFromDefaultPayload("{\"dummy\": \"hello\"}", "192.168.1.100")

	assert.Error(t, err, "can not create a new connection: empty deviceID")
}

func TestConectionUptimeShouldReturnHowManySecondsTheConnectionsHasBeenActive(t *testing.T) {
	connection, _ := NewConnection("device_id", "device_name", "device_type", "agent", "192.168.1.100")

	duration, _ := connection.Uptime()
	assert.Assert(t, duration == 0)

	time.Sleep(1 * time.Second)

	duration, _ = connection.Uptime()
	assert.Assert(t, duration == 1)
}

func TestWhenAMessageIsSentStatsAreUpdated(t *testing.T) {
	connection, _ := NewConnection("device_id", "device_name", "device_type", "agent", "192.168.1.100")
	assert.Assert(t, connection.SentMessages == 0)
	assert.Assert(t, connection.LastSentMessageTimeStamp == 0)

	connection.MessageSent()

	assert.Assert(t, connection.SentMessages == 1)
	assert.Assert(t, connection.LastSentMessageTimeStamp == time.Now().Unix())
}

func TestWhenAMessageIsReceivedStatsAreUpdated(t *testing.T) {
	deviceConnection, _ := NewConnection("device_id", "device_name", "device_type", "agent", "192.168.1.100")
	assert.Assert(t, deviceConnection.ReceivedMessages == 0)
	assert.Assert(t, deviceConnection.LastReceivedMessageTimeStamp == 0)

	deviceConnection.MessageReceived()

	assert.Assert(t, deviceConnection.ReceivedMessages == 1)
	assert.Assert(t, deviceConnection.LastReceivedMessageTimeStamp == time.Now().Unix())
}
