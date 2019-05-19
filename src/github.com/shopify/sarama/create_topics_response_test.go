package sarama

import (
	"testing"
	"time"
	"vendor"
)

var (
	createTopicsResponseV0 = []byte{
		0, 0, 0, 1,
		0, 5, 't', 'o', 'p', 'i', 'c',
		0, 42,
	}

	createTopicsResponseV1 = []byte{
		0, 0, 0, 1,
		0, 5, 't', 'o', 'p', 'i', 'c',
		0, 42,
		0, 3, 'm', 's', 'g',
	}

	createTopicsResponseV2 = []byte{
		0, 0, 0, 100,
		0, 0, 0, 1,
		0, 5, 't', 'o', 'p', 'i', 'c',
		0, 42,
		0, 3, 'm', 's', 'g',
	}
)

func TestCreateTopicsResponse(t *testing.T) {
	resp := &vendor.CreateTopicsResponse{
		TopicErrors: map[string]*vendor.TopicError{
			"topic": &vendor.TopicError{
				Err: vendor.ErrInvalidRequest,
			},
		},
	}

	vendor.testResponse(t, "version 0", resp, createTopicsResponseV0)

	resp.Version = 1
	msg := "msg"
	resp.TopicErrors["topic"].ErrMsg = &msg

	vendor.testResponse(t, "version 1", resp, createTopicsResponseV1)

	resp.Version = 2
	resp.ThrottleTime = 100 * time.Millisecond

	vendor.testResponse(t, "version 2", resp, createTopicsResponseV2)
}

func TestTopicError(t *testing.T) {
	// Assert that TopicError satisfies error interface
	var err error = &vendor.TopicError{
		Err: vendor.ErrTopicAuthorizationFailed,
	}

	got := err.Error()
	want := vendor.ErrTopicAuthorizationFailed.Error()
	if got != want {
		t.Errorf("TopicError.Error() = %v; want %v", got, want)
	}

	msg := "reason why topic authorization failed"
	err = &vendor.TopicError{
		Err:    vendor.ErrTopicAuthorizationFailed,
		ErrMsg: &msg,
	}
	got = err.Error()
	want = vendor.ErrTopicAuthorizationFailed.Error() + " - " + msg
	if got != want {
		t.Errorf("TopicError.Error() = %v; want %v", got, want)
	}
}
