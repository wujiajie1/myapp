package sarama

import (
	"testing"
	"time"
	"vendor"
)

var (
	txnOffsetCommitResponse = []byte{
		0, 0, 0, 100,
		0, 0, 0, 1, // 1 topic
		0, 5, 't', 'o', 'p', 'i', 'c',
		0, 0, 0, 1, // 1 partition response
		0, 0, 0, 2, // partition number 2
		0, 47, // err
	}
)

func TestTxnOffsetCommitResponse(t *testing.T) {
	resp := &vendor.TxnOffsetCommitResponse{
		ThrottleTime: 100 * time.Millisecond,
		Topics: map[string][]*vendor.PartitionError{
			"topic": []*vendor.PartitionError{{
				Partition: 2,
				Err:       vendor.ErrInvalidProducerEpoch,
			}},
		},
	}

	vendor.testResponse(t, "", resp, txnOffsetCommitResponse)
}
