package sarama

import (
	"testing"
	"vendor"
)

var (
	saslAuthenticatResponseErr = []byte{
		0, 58,
		0, 3, 'e', 'r', 'r',
		0, 0, 0, 3, 'm', 's', 'g',
	}
)

func TestSaslAuthenticateResponse(t *testing.T) {

	response := new(vendor.SaslAuthenticateResponse)
	response.Err = vendor.ErrSASLAuthenticationFailed
	msg := "err"
	response.ErrorMessage = &msg
	response.SaslAuthBytes = []byte(`msg`)

	vendor.testResponse(t, "authenticate reponse", response, saslAuthenticatResponseErr)
}
