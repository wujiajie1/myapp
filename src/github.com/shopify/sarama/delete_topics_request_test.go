package sarama

import (
	"testing"
	"time"
	"vendor"
)

var deleteTopicsRequest = []byte{
	0, 0, 0, 2,
	0, 5, 't', 'o', 'p', 'i', 'c',
	0, 5, 'o', 't', 'h', 'e', 'r',
	0, 0, 0, 100,
}

func TestDeleteTopicsRequestV0(t *testing.T) {
	req := &vendor.DeleteTopicsRequest{
		Version: 0,
		Topics:  []string{"topic", "other"},
		Timeout: 100 * time.Millisecond,
	}

	vendor.testRequest(t, "", req, deleteTopicsRequest)
}

func TestDeleteTopicsRequestV1(t *testing.T) {
	req := &vendor.DeleteTopicsRequest{
		Version: 1,
		Topics:  []string{"topic", "other"},
		Timeout: 100 * time.Millisecond,
	}

	vendor.testRequest(t, "", req, deleteTopicsRequest)
}
