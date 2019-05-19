package sarama

import (
	"testing"
	"vendor"
)

var (
	emptySyncGroupRequest = []byte{
		0, 3, 'f', 'o', 'o', // Group ID
		0x00, 0x01, 0x02, 0x03, // Generation ID
		0, 3, 'b', 'a', 'z', // Member ID
		0, 0, 0, 0, // no assignments
	}

	populatedSyncGroupRequest = []byte{
		0, 3, 'f', 'o', 'o', // Group ID
		0x00, 0x01, 0x02, 0x03, // Generation ID
		0, 3, 'b', 'a', 'z', // Member ID
		0, 0, 0, 1, // one assignment
		0, 3, 'b', 'a', 'z', // Member ID
		0, 0, 0, 3, 'f', 'o', 'o', // Member assignment
	}
)

func TestSyncGroupRequest(t *testing.T) {
	var request *vendor.SyncGroupRequest

	request = new(vendor.SyncGroupRequest)
	request.GroupId = "foo"
	request.GenerationId = 66051
	request.MemberId = "baz"
	vendor.testRequest(t, "empty", request, emptySyncGroupRequest)

	request = new(vendor.SyncGroupRequest)
	request.GroupId = "foo"
	request.GenerationId = 66051
	request.MemberId = "baz"
	request.AddGroupAssignment("baz", []byte("foo"))
	vendor.testRequest(t, "populated", request, populatedSyncGroupRequest)
}
