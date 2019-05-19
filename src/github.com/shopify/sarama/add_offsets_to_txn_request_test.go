package sarama

import (
	"testing"
	"vendor"
)

var (
	addOffsetsToTxnRequest = []byte{
		0, 3, 't', 'x', 'n',
		0, 0, 0, 0, 0, 0, 31, 64,
		0, 0,
		0, 7, 'g', 'r', 'o', 'u', 'p', 'i', 'd',
	}
)

func TestAddOffsetsToTxnRequest(t *testing.T) {
	req := &vendor.AddOffsetsToTxnRequest{
		TransactionalID: "txn",
		ProducerID:      8000,
		ProducerEpoch:   0,
		GroupID:         "groupid",
	}

	vendor.testRequest(t, "", req, addOffsetsToTxnRequest)
}
