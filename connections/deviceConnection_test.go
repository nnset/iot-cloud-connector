package connections

import(
    "testing"
    "time"
    "gotest.tools/assert"
    "github.com/google/uuid"
)

type dummyNetworkConnectionConnection struct {
    CloseWasCalled bool
}

func (c *dummyNetworkConnectionConnection) Close(statusCode ConnectionStatusCode, reason string) error {
    c.CloseWasCalled = true
    return nil
}

func TestDeviceConnectionNamedConstructorShouldReturnAPointerToDeviceConnection(t *testing.T) {

    deviceConnection := NewDeviceConnection(&dummyNetworkConnectionConnection{}, "device_id", "192.168.1.100", "agent")

    assert.Assert(t, deviceConnection != nil)
    assert.Assert(t, deviceConnection.RemoteAddress() == "192.168.1.100")
}

func TestEachInstanceOfDeviceConnectionShouldHaveAUniqueIdentifier(t *testing.T) {

    deviceConnection := NewDeviceConnection(&dummyNetworkConnectionConnection{}, "device_id", "192.168.1.100", "agent")

    _, err := uuid.Parse(deviceConnection.ID())

    assert.Assert(t, err == nil, err)
}

func TestDurationShouldReturnHowManySecondsTheConnectionsHasBeenActive(t *testing.T) {
    deviceConnection := NewDeviceConnection(&dummyNetworkConnectionConnection{}, "device_id", "192.168.1.100", "agent")
    
    duration, _ := deviceConnection.Duration()
    assert.Assert(t, duration == 0)

    time.Sleep(1 * time.Second)

    duration, _ = deviceConnection.Duration()
    assert.Assert(t, duration == 1)
}

func TestCloseShouldCloseTheConnectionViaTheNetworkConnectionInterface(t *testing.T) {
    netConnection := &dummyNetworkConnectionConnection{CloseWasCalled: false}
    deviceConnection := NewDeviceConnection(netConnection, "device_id", "192.168.1.100", "agent")
    
    assert.Assert(t, netConnection.CloseWasCalled == false)
    
    deviceConnection.Close(StatusNormalClosure, "testing")

    assert.Assert(t, netConnection.CloseWasCalled == true)
}
