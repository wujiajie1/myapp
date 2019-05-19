package sarama

import (
	"testing"
	"time"
	"vendor"
)

var (
	addOffsetsToTxnResponse = []byte{
		0, 0, 0, 100,
		0, 47,
	}
)

func TestAddOffsetsToTxnResponse(t *testing.T) {
	resp := &vendor.AddOffsetsToTxnResponse{
		ThrottleTime: 100 * time.Millisecond,
		Err:          vendor.ErrInvalidProducerEpoch,
	}

	vendor.testResponse(t, "", resp, addOffsetsToTxnResponse)
}
