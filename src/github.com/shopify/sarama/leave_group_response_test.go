package sarama

import (
	"testing"
	"vendor"
)

var (
	leaveGroupResponseNoError   = []byte{0x00, 0x00}
	leaveGroupResponseWithError = []byte{0, 25}
)

func TestLeaveGroupResponse(t *testing.T) {
	var response *vendor.LeaveGroupResponse

	response = new(vendor.LeaveGroupResponse)
	vendor.testVersionDecodable(t, "no error", response, leaveGroupResponseNoError, 0)
	if response.Err != vendor.ErrNoError {
		t.Error("Decoding error failed: no error expected but found", response.Err)
	}

	response = new(vendor.LeaveGroupResponse)
	vendor.testVersionDecodable(t, "with error", response, leaveGroupResponseWithError, 0)
	if response.Err != vendor.ErrUnknownMemberId {
		t.Error("Decoding error failed: ErrUnknownMemberId expected but found", response.Err)
	}
}
