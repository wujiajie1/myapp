package sarama

import (
	"testing"
	"time"
	"vendor"
)

func TestFindCoordinatorResponse(t *testing.T) {
	errMsg := "kaboom"

	for _, tc := range []struct {
		desc     string
		response *vendor.FindCoordinatorResponse
		encoded  []byte
	}{{
		desc: "version 0 - no error",
		response: &vendor.FindCoordinatorResponse{
			Version: 0,
			Err:     vendor.ErrNoError,
			Coordinator: &vendor.Broker{
				id:   7,
				addr: "host:9092",
			},
		},
		encoded: []byte{
			0, 0, // Err
			0, 0, 0, 7, // Coordinator.ID
			0, 4, 'h', 'o', 's', 't', // Coordinator.Host
			0, 0, 35, 132, // Coordinator.Port
		},
	}, {
		desc: "version 1 - no error",
		response: &vendor.FindCoordinatorResponse{
			Version:      1,
			ThrottleTime: 100 * time.Millisecond,
			Err:          vendor.ErrNoError,
			Coordinator: &vendor.Broker{
				id:   7,
				addr: "host:9092",
			},
		},
		encoded: []byte{
			0, 0, 0, 100, // ThrottleTime
			0, 0, // Err
			255, 255, // ErrMsg: empty
			0, 0, 0, 7, // Coordinator.ID
			0, 4, 'h', 'o', 's', 't', // Coordinator.Host
			0, 0, 35, 132, // Coordinator.Port
		},
	}, {
		desc: "version 0 - error",
		response: &vendor.FindCoordinatorResponse{
			Version:     0,
			Err:         vendor.ErrConsumerCoordinatorNotAvailable,
			Coordinator: vendor.NoNode,
		},
		encoded: []byte{
			0, 15, // Err
			255, 255, 255, 255, // Coordinator.ID: -1
			0, 0, // Coordinator.Host: ""
			255, 255, 255, 255, // Coordinator.Port: -1
		},
	}, {
		desc: "version 1 - error",
		response: &vendor.FindCoordinatorResponse{
			Version:      1,
			ThrottleTime: 100 * time.Millisecond,
			Err:          vendor.ErrConsumerCoordinatorNotAvailable,
			ErrMsg:       &errMsg,
			Coordinator:  vendor.NoNode,
		},
		encoded: []byte{
			0, 0, 0, 100, // ThrottleTime
			0, 15, // Err
			0, 6, 'k', 'a', 'b', 'o', 'o', 'm', // ErrMsg
			255, 255, 255, 255, // Coordinator.ID: -1
			0, 0, // Coordinator.Host: ""
			255, 255, 255, 255, // Coordinator.Port: -1
		},
	}} {
		vendor.testResponse(t, tc.desc, tc.response, tc.encoded)
	}
}
