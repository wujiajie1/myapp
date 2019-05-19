package sarama

import (
	"testing"
	"vendor"
)

var (
	basicLeaveGroupRequest = []byte{
		0, 3, 'f', 'o', 'o',
		0, 3, 'b', 'a', 'r',
	}
)

func TestLeaveGroupRequest(t *testing.T) {
	var request *vendor.LeaveGroupRequest

	request = new(vendor.LeaveGroupRequest)
	request.GroupId = "foo"
	request.MemberId = "bar"
	vendor.testRequest(t, "basic", request, basicLeaveGroupRequest)
}
