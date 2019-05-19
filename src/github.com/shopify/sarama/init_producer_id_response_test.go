package sarama

import (
	"testing"
	"time"
	"vendor"
)

var (
	initProducerIDResponse = []byte{
		0, 0, 0, 100,
		0, 0,
		0, 0, 0, 0, 0, 0, 31, 64, // producerID = 8000
		0, 0, // epoch
	}

	initProducerIDRequestError = []byte{
		0, 0, 0, 100,
		0, 51,
		255, 255, 255, 255, 255, 255, 255, 255,
		0, 0,
	}
)

func TestInitProducerIDResponse(t *testing.T) {
	resp := &vendor.InitProducerIDResponse{
		ThrottleTime:  100 * time.Millisecond,
		ProducerID:    8000,
		ProducerEpoch: 0,
	}

	vendor.testResponse(t, "", resp, initProducerIDResponse)

	resp.Err = vendor.ErrConcurrentTransactions
	resp.ProducerID = -1

	vendor.testResponse(t, "with error", resp, initProducerIDRequestError)
}
