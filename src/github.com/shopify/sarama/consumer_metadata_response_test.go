package sarama

import (
	"testing"
	"vendor"
)

var (
	consumerMetadataResponseError = []byte{
		0x00, 0x0E,
		0x00, 0x00, 0x00, 0x00,
		0x00, 0x00,
		0x00, 0x00, 0x00, 0x00}

	consumerMetadataResponseSuccess = []byte{
		0x00, 0x00,
		0x00, 0x00, 0x00, 0xAB,
		0x00, 0x03, 'f', 'o', 'o',
		0x00, 0x00, 0xCC, 0xDD}
)

func TestConsumerMetadataResponseError(t *testing.T) {
	response := &vendor.ConsumerMetadataResponse{Err: vendor.ErrOffsetsLoadInProgress}
	vendor.testEncodable(t, "", response, consumerMetadataResponseError)

	decodedResp := &vendor.ConsumerMetadataResponse{}
	if err := vendor.versionedDecode(consumerMetadataResponseError, decodedResp, 0); err != nil {
		t.Error("could not decode: ", err)
	}

	if decodedResp.Err != vendor.ErrOffsetsLoadInProgress {
		t.Errorf("got %s, want %s", decodedResp.Err, vendor.ErrOffsetsLoadInProgress)
	}
}

func TestConsumerMetadataResponseSuccess(t *testing.T) {
	broker := vendor.NewBroker("foo:52445")
	broker.id = 0xAB
	response := vendor.ConsumerMetadataResponse{
		Coordinator:     broker,
		CoordinatorID:   0xAB,
		CoordinatorHost: "foo",
		CoordinatorPort: 0xCCDD,
		Err:             vendor.ErrNoError,
	}
	vendor.testResponse(t, "success", &response, consumerMetadataResponseSuccess)
}
