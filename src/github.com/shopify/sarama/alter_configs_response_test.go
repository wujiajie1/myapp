package sarama

import (
	"testing"
	"vendor"
)

var (
	alterResponseEmpty = []byte{
		0, 0, 0, 0, //throttle
		0, 0, 0, 0, // no configs
	}

	alterResponsePopulated = []byte{
		0, 0, 0, 0, //throttle
		0, 0, 0, 1, // response
		0, 0, //errorcode
		0, 0, //string
		2, // topic
		0, 3, 'f', 'o', 'o',
	}
)

func TestAlterConfigsResponse(t *testing.T) {
	var response *vendor.AlterConfigsResponse

	response = &vendor.AlterConfigsResponse{
		Resources: []*vendor.AlterConfigsResourceResponse{},
	}
	vendor.testVersionDecodable(t, "empty", response, alterResponseEmpty, 0)
	if len(response.Resources) != 0 {
		t.Error("Expected no groups")
	}

	response = &vendor.AlterConfigsResponse{
		Resources: []*vendor.AlterConfigsResourceResponse{
			&vendor.AlterConfigsResourceResponse{
				ErrorCode: 0,
				ErrorMsg:  "",
				Type:      vendor.TopicResource,
				Name:      "foo",
			},
		},
	}
	vendor.testResponse(t, "response with error", response, alterResponsePopulated)
}
