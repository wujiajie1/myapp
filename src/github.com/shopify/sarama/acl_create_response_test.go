package sarama

import (
	"testing"
	"time"
	"vendor"
)

var (
	createResponseWithError = []byte{
		0, 0, 0, 100,
		0, 0, 0, 1,
		0, 42,
		0, 5, 'e', 'r', 'r', 'o', 'r',
	}

	createResponseArray = []byte{
		0, 0, 0, 100,
		0, 0, 0, 2,
		0, 42,
		0, 5, 'e', 'r', 'r', 'o', 'r',
		0, 0,
		255, 255,
	}
)

func TestCreateAclsResponse(t *testing.T) {
	errmsg := "error"
	resp := &vendor.CreateAclsResponse{
		ThrottleTime: 100 * time.Millisecond,
		AclCreationResponses: []*vendor.AclCreationResponse{{
			Err:    vendor.ErrInvalidRequest,
			ErrMsg: &errmsg,
		}},
	}

	vendor.testResponse(t, "response with error", resp, createResponseWithError)

	resp.AclCreationResponses = append(resp.AclCreationResponses, new(vendor.AclCreationResponse))

	vendor.testResponse(t, "response array", resp, createResponseArray)
}
