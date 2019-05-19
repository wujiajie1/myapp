package sarama

import (
	"testing"
	"vendor"
)

var (
	baseSaslRequest = []byte{
		0, 3, 'f', 'o', 'o', // Mechanism
	}
)

func TestSaslHandshakeRequest(t *testing.T) {
	var request *vendor.SaslHandshakeRequest

	request = new(vendor.SaslHandshakeRequest)
	request.Mechanism = "foo"
	vendor.testRequest(t, "basic", request, baseSaslRequest)
}
