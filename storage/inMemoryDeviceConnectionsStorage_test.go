package storage

import (
	"testing"

	"gotest.tools/assert"
)

func TestCreatingANewInMemoryDeviceConnectionsStorageShouldReturnItsInstance(t *testing.T) {
	st := NewInMemoryDeviceConnectionsStorage()

	assert.Assert(t, st != nil)
	assert.Assert(t, len(st.ID()) > 0)
}

func TestAddingANewConnectionShouldStoreIt(t *testing.T) {
	st := NewInMemoryDeviceConnectionsStorage()

	err := st.Add("id", "device", "name", "type", "user-agent", "address")

	assert.Assert(t, err == nil)

	con, err := st.Get("device")

	assert.Assert(t, err == nil)
	assert.Assert(t, con.DeviceID() == "device")
}

func TestOnlyOneConnectionPerDeviceIsAllowed(t *testing.T) {
	st := NewInMemoryDeviceConnectionsStorage()

	err := st.Add("id", "device", "name", "type", "user-agent", "address")
	assert.Assert(t, err == nil)

	err = st.Add("id", "device", "name", "type", "user-agent", "address")

	assert.Error(t, err, "Connection already established")
}

func TestDeletingAConnectionShouldRemoveIt(t *testing.T) {
	st := NewInMemoryDeviceConnectionsStorage()
	deviceID := "device"
	st.Add("id", deviceID, "name", "type", "user-agent", "address")

	_, err := st.Get(deviceID)
	assert.Assert(t, err == nil)

	assert.Assert(t, len(st.ConnectedDevices()) == 1)

	err = st.Delete(deviceID)
	assert.Assert(t, err == nil)

	_, err = st.Get(deviceID)
	assert.Error(t, err, "Connection not found")

	assert.Assert(t, len(st.ConnectedDevices()) == 0)
}

func TestWhenAMessageIsSentShouldBeCounted(t *testing.T) {
	deviceID := "device"
	st := NewInMemoryDeviceConnectionsStorage()

	assert.Assert(t, st.SentMessages(deviceID) == 0)
	assert.Assert(t, st.TotalSentMessages() == 0)

	err := st.Add("id", deviceID, "name", "type", "user-agent", "address")

	err = st.MessageWasSent(deviceID)
	assert.Assert(t, err == nil)

	assert.Assert(t, st.SentMessages(deviceID) == 1)
	assert.Assert(t, st.TotalSentMessages() == 1)
}

func TestReportingASentMessageToAnUnnownDeviceShouldReturnError(t *testing.T) {
	st := NewInMemoryDeviceConnectionsStorage()

	err := st.MessageWasSent("device")
	assert.Error(t, err, "Connection not found")
}

func TestWhenAMessageIsReceivedItShouldBeCounted(t *testing.T) {
	deviceID := "device"
	st := NewInMemoryDeviceConnectionsStorage()

	assert.Assert(t, st.ReceivedMessages(deviceID) == 0)
	assert.Assert(t, st.TotalReceivedMessages() == 0)

	err := st.Add("id", deviceID, "name", "type", "user-agent", "address")

	err = st.MessageWasReceived(deviceID)
	assert.Assert(t, err == nil)

	assert.Assert(t, st.ReceivedMessages(deviceID) == 1)
	assert.Assert(t, st.TotalReceivedMessages() == 1)
}

func TestReportingAReceivedMessageToAnUnnownDeviceShouldReturnError(t *testing.T) {
	st := NewInMemoryDeviceConnectionsStorage()

	err := st.MessageWasReceived("device")
	assert.Error(t, err, "Connection not found")
}

func TestConnectedDevicesReturnASliceOfConenctedDevices(t *testing.T) {
	st := NewInMemoryDeviceConnectionsStorage()
	deviceID := "device"

	assert.Assert(t, len(st.ConnectedDevices()) == 0)

	st.Add("id", deviceID, "name", "type", "user-agent", "address")

	assert.Assert(t, len(st.ConnectedDevices()) == 1)

	connections := st.ConnectedDevices()
	connection := connections[len(connections)-1]

	assert.Assert(t, connection.DeviceID == deviceID)
}
