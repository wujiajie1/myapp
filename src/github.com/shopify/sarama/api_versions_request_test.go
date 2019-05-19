package sarama

import (
	"testing"
	"vendor"
)

var (
	apiVersionRequest = []byte{}
)

func TestApiVersionsRequest(t *testing.T) {
	var request *vendor.ApiVersionsRequest

	request = new(vendor.ApiVersionsRequest)
	vendor.testRequest(t, "basic", request, apiVersionRequest)
}
