package sarama

import (
	"reflect"
	"testing"
	"time"
	"vendor"
)

var (
	createPartitionResponseSuccess = []byte{
		0, 0, 0, 100, // throttleTimeMs
		0, 0, 0, 1,
		0, 5, 't', 'o', 'p', 'i', 'c',
		0, 0, // no error
		255, 255, // no error message
	}

	createPartitionResponseFail = []byte{
		0, 0, 0, 100, // throttleTimeMs
		0, 0, 0, 1,
		0, 5, 't', 'o', 'p', 'i', 'c',
		0, 37, // partition error
		0, 5, 'e', 'r', 'r', 'o', 'r',
	}
)

func TestCreatePartitionsResponse(t *testing.T) {
	resp := &vendor.CreatePartitionsResponse{
		ThrottleTime: 100 * time.Millisecond,
		TopicPartitionErrors: map[string]*vendor.TopicPartitionError{
			"topic": &vendor.TopicPartitionError{},
		},
	}

	vendor.testResponse(t, "success", resp, createPartitionResponseSuccess)
	decodedresp := new(vendor.CreatePartitionsResponse)
	vendor.testVersionDecodable(t, "success", decodedresp, createPartitionResponseSuccess, 0)
	if !reflect.DeepEqual(decodedresp, resp) {
		t.Errorf("Decoding error: expected %v but got %v", decodedresp, resp)
	}

	errMsg := "error"
	resp.TopicPartitionErrors["topic"].Err = vendor.ErrInvalidPartitions
	resp.TopicPartitionErrors["topic"].ErrMsg = &errMsg

	vendor.testResponse(t, "with errors", resp, createPartitionResponseFail)
	decodedresp = new(vendor.CreatePartitionsResponse)
	vendor.testVersionDecodable(t, "with errors", decodedresp, createPartitionResponseFail, 0)
	if !reflect.DeepEqual(decodedresp, resp) {
		t.Errorf("Decoding error: expected %v but got %v", decodedresp, resp)
	}
}

func TestTopicPartitionError(t *testing.T) {
	// Assert that TopicPartitionError satisfies error interface
	var err error = &vendor.TopicPartitionError{
		Err: vendor.ErrTopicAuthorizationFailed,
	}

	got := err.Error()
	want := vendor.ErrTopicAuthorizationFailed.Error()
	if got != want {
		t.Errorf("TopicPartitionError.Error() = %v; want %v", got, want)
	}

	msg := "reason why topic authorization failed"
	err = &vendor.TopicPartitionError{
		Err:    vendor.ErrTopicAuthorizationFailed,
		ErrMsg: &msg,
	}
	got = err.Error()
	want = vendor.ErrTopicAuthorizationFailed.Error() + " - " + msg
	if got != want {
		t.Errorf("TopicPartitionError.Error() = %v; want %v", got, want)
	}
}
