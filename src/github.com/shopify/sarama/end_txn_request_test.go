package sarama

import (
	"testing"
	"vendor"
)

var (
	endTxnRequest = []byte{
		0, 3, 't', 'x', 'n',
		0, 0, 0, 0, 0, 0, 31, 64,
		0, 1,
		1,
	}
)

func TestEndTxnRequest(t *testing.T) {
	req := &vendor.EndTxnRequest{
		TransactionalID:   "txn",
		ProducerID:        8000,
		ProducerEpoch:     1,
		TransactionResult: true,
	}

	vendor.testRequest(t, "", req, endTxnRequest)
}
