package sarama

import (
	"testing"
	"time"
	"vendor"
)

var (
	addPartitionsToTxnResponse = []byte{
		0, 0, 0, 100,
		0, 0, 0, 1,
		0, 5, 't', 'o', 'p', 'i', 'c',
		0, 0, 0, 1, // 1 partition error
		0, 0, 0, 2, // partition 2
		0, 48, // error
	}
)

func TestAddPartitionsToTxnResponse(t *testing.T) {
	resp := &vendor.AddPartitionsToTxnResponse{
		ThrottleTime: 100 * time.Millisecond,
		Errors: map[string][]*vendor.PartitionError{
			"topic": []*vendor.PartitionError{&vendor.PartitionError{
				Err:       vendor.ErrInvalidTxnState,
				Partition: 2,
			}},
		},
	}

	vendor.testResponse(t, "", resp, addPartitionsToTxnResponse)
}
