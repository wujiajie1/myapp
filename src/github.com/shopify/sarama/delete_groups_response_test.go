package sarama

import (
	"testing"
	"vendor"
)

var (
	emptyDeleteGroupsResponse = []byte{
		0, 0, 0, 0, // does not violate any quota
		0, 0, 0, 0, // no groups
	}

	errorDeleteGroupsResponse = []byte{
		0, 0, 0, 0, // does not violate any quota
		0, 0, 0, 1, // 1 group
		0, 3, 'f', 'o', 'o', // group name
		0, 31, // error ErrClusterAuthorizationFailed
	}

	noErrorDeleteGroupsResponse = []byte{
		0, 0, 0, 0, // does not violate any quota
		0, 0, 0, 1, // 1 group
		0, 3, 'f', 'o', 'o', // group name
		0, 0, // no error
	}
)

func TestDeleteGroupsResponse(t *testing.T) {
	var response *vendor.DeleteGroupsResponse

	response = new(vendor.DeleteGroupsResponse)
	vendor.testVersionDecodable(t, "empty", response, emptyDeleteGroupsResponse, 0)
	if response.ThrottleTime != 0 {
		t.Error("Expected no violation")
	}
	if len(response.GroupErrorCodes) != 0 {
		t.Error("Expected no groups")
	}

	response = new(vendor.DeleteGroupsResponse)
	vendor.testVersionDecodable(t, "error", response, errorDeleteGroupsResponse, 0)
	if response.ThrottleTime != 0 {
		t.Error("Expected no violation")
	}
	if response.GroupErrorCodes["foo"] != vendor.ErrClusterAuthorizationFailed {
		t.Error("Expected error ErrClusterAuthorizationFailed, found:", response.GroupErrorCodes["foo"])
	}

	response = new(vendor.DeleteGroupsResponse)
	vendor.testVersionDecodable(t, "no error", response, noErrorDeleteGroupsResponse, 0)
	if response.ThrottleTime != 0 {
		t.Error("Expected no violation")
	}
	if response.GroupErrorCodes["foo"] != vendor.ErrNoError {
		t.Error("Expected error ErrClusterAuthorizationFailed, found:", response.GroupErrorCodes["foo"])
	}
}
