package sarama

import (
	"testing"
	"vendor"
)

var (
	consumerMetadataRequestEmpty = []byte{
		0x00, 0x00}

	consumerMetadataRequestString = []byte{
		0x00, 0x06, 'f', 'o', 'o', 'b', 'a', 'r'}
)

func TestConsumerMetadataRequest(t *testing.T) {
	request := new(vendor.ConsumerMetadataRequest)
	vendor.testEncodable(t, "empty string", request, consumerMetadataRequestEmpty)
	vendor.testVersionDecodable(t, "empty string", request, consumerMetadataRequestEmpty, 0)

	request.ConsumerGroup = "foobar"
	vendor.testEncodable(t, "with string", request, consumerMetadataRequestString)
	vendor.testVersionDecodable(t, "with string", request, consumerMetadataRequestString, 0)
}
