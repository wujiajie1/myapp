package sarama

import (
	"testing"
	"vendor"
)

func TestListGroupsRequest(t *testing.T) {
	vendor.testRequest(t, "ListGroupsRequest", &vendor.ListGroupsRequest{}, []byte{})
}
