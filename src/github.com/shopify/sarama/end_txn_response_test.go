package sarama

import (
	"testing"
	"time"
	"vendor"
)

var (
	endTxnResponse = []byte{
		0, 0, 0, 100,
		0, 49,
	}
)

func TestEndTxnResponse(t *testing.T) {
	resp := &vendor.EndTxnResponse{
		ThrottleTime: 100 * time.Millisecond,
		Err:          vendor.ErrInvalidProducerIDMapping,
	}

	vendor.testResponse(t, "", resp, endTxnResponse)
}
