package sarama

import (
	"testing"
	"vendor"
)

var (
	heartbeatResponseNoError = []byte{
		0x00, 0x00}
)

func TestHeartbeatResponse(t *testing.T) {
	var response *vendor.HeartbeatResponse

	response = new(vendor.HeartbeatResponse)
	vendor.testVersionDecodable(t, "no error", response, heartbeatResponseNoError, 0)
	if response.Err != vendor.ErrNoError {
		t.Error("Decoding error failed: no error expected but found", response.Err)
	}
}
