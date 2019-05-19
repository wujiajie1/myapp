package sarama

import (
	"testing"
	"vendor"
)

var (
	txnOffsetCommitRequest = []byte{
		0, 3, 't', 'x', 'n',
		0, 7, 'g', 'r', 'o', 'u', 'p', 'i', 'd',
		0, 0, 0, 0, 0, 0, 31, 64, // producer ID
		0, 1, // producer epoch
		0, 0, 0, 1, // 1 topic
		0, 5, 't', 'o', 'p', 'i', 'c',
		0, 0, 0, 1, // 1 partition
		0, 0, 0, 2, // partition no 2
		0, 0, 0, 0, 0, 0, 0, 123,
		255, 255, // no meta data
	}
)

func TestTxnOffsetCommitRequest(t *testing.T) {
	req := &vendor.TxnOffsetCommitRequest{
		TransactionalID: "txn",
		GroupID:         "groupid",
		ProducerID:      8000,
		ProducerEpoch:   1,
		Topics: map[string][]*vendor.PartitionOffsetMetadata{
			"topic": []*vendor.PartitionOffsetMetadata{{
				Offset:    123,
				Partition: 2,
			}},
		},
	}

	vendor.testRequest(t, "", req, txnOffsetCommitRequest)
}
