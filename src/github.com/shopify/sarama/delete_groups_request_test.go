package sarama

import (
	"testing"
	"vendor"
)

var (
	emptyDeleteGroupsRequest = []byte{0, 0, 0, 0}

	singleDeleteGroupsRequest = []byte{
		0, 0, 0, 1, // 1 group
		0, 3, 'f', 'o', 'o', // group name: foo
	}

	doubleDeleteGroupsRequest = []byte{
		0, 0, 0, 2, // 2 groups
		0, 3, 'f', 'o', 'o', // group name: foo
		0, 3, 'b', 'a', 'r', // group name: foo
	}
)

func TestDeleteGroupsRequest(t *testing.T) {
	var request *vendor.DeleteGroupsRequest

	request = new(vendor.DeleteGroupsRequest)
	vendor.testRequest(t, "no groups", request, emptyDeleteGroupsRequest)

	request = new(vendor.DeleteGroupsRequest)
	request.AddGroup("foo")
	vendor.testRequest(t, "one group", request, singleDeleteGroupsRequest)

	request = new(vendor.DeleteGroupsRequest)
	request.AddGroup("foo")
	request.AddGroup("bar")
	vendor.testRequest(t, "two groups", request, doubleDeleteGroupsRequest)
}
