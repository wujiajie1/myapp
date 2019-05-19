package sarama

import (
	"testing"
	"vendor"
)

var (
	emptyDescribeGroupsRequest = []byte{0, 0, 0, 0}

	singleDescribeGroupsRequest = []byte{
		0, 0, 0, 1, // 1 group
		0, 3, 'f', 'o', 'o', // group name: foo
	}

	doubleDescribeGroupsRequest = []byte{
		0, 0, 0, 2, // 2 groups
		0, 3, 'f', 'o', 'o', // group name: foo
		0, 3, 'b', 'a', 'r', // group name: foo
	}
)

func TestDescribeGroupsRequest(t *testing.T) {
	var request *vendor.DescribeGroupsRequest

	request = new(vendor.DescribeGroupsRequest)
	vendor.testRequest(t, "no groups", request, emptyDescribeGroupsRequest)

	request = new(vendor.DescribeGroupsRequest)
	request.AddGroup("foo")
	vendor.testRequest(t, "one group", request, singleDescribeGroupsRequest)

	request = new(vendor.DescribeGroupsRequest)
	request.AddGroup("foo")
	request.AddGroup("bar")
	vendor.testRequest(t, "two groups", request, doubleDescribeGroupsRequest)
}
