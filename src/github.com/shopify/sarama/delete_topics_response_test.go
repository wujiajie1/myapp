package sarama

import (
	"testing"
	"time"
	"vendor"
)

var (
	deleteTopicsResponseV0 = []byte{
		0, 0, 0, 1,
		0, 5, 't', 'o', 'p', 'i', 'c',
		0, 0,
	}

	deleteTopicsResponseV1 = []byte{
		0, 0, 0, 100,
		0, 0, 0, 1,
		0, 5, 't', 'o', 'p', 'i', 'c',
		0, 0,
	}
)

func TestDeleteTopicsResponse(t *testing.T) {
	resp := &vendor.DeleteTopicsResponse{
		TopicErrorCodes: map[string]vendor.KError{
			"topic": vendor.ErrNoError,
		},
	}

	vendor.testResponse(t, "version 0", resp, deleteTopicsResponseV0)

	resp.Version = 1
	resp.ThrottleTime = 100 * time.Millisecond

	vendor.testResponse(t, "version 1", resp, deleteTopicsResponseV1)
}
