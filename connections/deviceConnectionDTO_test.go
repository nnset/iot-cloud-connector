package connections

import (
	"testing"

	"gotest.tools/assert"
)

func TestCreatingADeviceConnectionDTOFromDeviceConnectionShouldReturnAnInstanceOfDeviceConnectionDTO(t *testing.T) {
	device, _ := NewDeviceConnection("id", "device_id", "name", "type", "user-agent", "address")
	dto := NewDeviceConnectionDTOFromDeviceConnection(device)

	assert.Assert(t, dto != nil)
	assert.Assert(t, dto.DeviceID == device.DeviceID())
}
