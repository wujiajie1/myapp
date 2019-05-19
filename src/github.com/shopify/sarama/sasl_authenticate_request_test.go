package sarama

import (
	"testing"
	"vendor"
)

var (
	saslAuthenticateRequest = []byte{
		0, 0, 0, 3, 'f', 'o', 'o',
	}
)

func TestSaslAuthenticateRequest(t *testing.T) {
	var request *vendor.SaslAuthenticateRequest

	request = new(vendor.SaslAuthenticateRequest)
	request.SaslAuthBytes = []byte(`foo`)
	vendor.testRequest(t, "basic", request, saslAuthenticateRequest)
}
