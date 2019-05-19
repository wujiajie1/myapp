package sarama

import (
	"testing"
	"vendor"
)

var (
	describeConfigsResponseEmpty = []byte{
		0, 0, 0, 0, //throttle
		0, 0, 0, 0, // no configs
	}

	describeConfigsResponsePopulatedv0 = []byte{
		0, 0, 0, 0, //throttle
		0, 0, 0, 1, // response
		0, 0, //errorcode
		0, 0, //string
		2, // topic
		0, 3, 'f', 'o', 'o',
		0, 0, 0, 1, //configs
		0, 10, 's', 'e', 'g', 'm', 'e', 'n', 't', '.', 'm', 's',
		0, 4, '1', '0', '0', '0',
		0, // ReadOnly
		0, // Default
		0, // Sensitive
	}

	describeConfigsResponsePopulatedv1 = []byte{
		0, 0, 0, 0, //throttle
		0, 0, 0, 1, // response
		0, 0, //errorcode
		0, 0, //string
		2, // topic
		0, 3, 'f', 'o', 'o',
		0, 0, 0, 1, //configs
		0, 10, 's', 'e', 'g', 'm', 'e', 'n', 't', '.', 'm', 's',
		0, 4, '1', '0', '0', '0',
		0,          // ReadOnly
		4,          // Source
		0,          // Sensitive
		0, 0, 0, 0, // No Synonym
	}

	describeConfigsResponseWithSynonymv1 = []byte{
		0, 0, 0, 0, //throttle
		0, 0, 0, 1, // response
		0, 0, //errorcode
		0, 0, //string
		2, // topic
		0, 3, 'f', 'o', 'o',
		0, 0, 0, 1, //configs
		0, 10, 's', 'e', 'g', 'm', 'e', 'n', 't', '.', 'm', 's',
		0, 4, '1', '0', '0', '0',
		0,          // ReadOnly
		4,          // Source
		0,          // Sensitive
		0, 0, 0, 1, // 1 Synonym
		0, 14, 'l', 'o', 'g', '.', 's', 'e', 'g', 'm', 'e', 'n', 't', '.', 'm', 's',
		0, 4, '1', '0', '0', '0',
		4, // Source
	}
)

func TestDescribeConfigsResponsev0(t *testing.T) {
	var response *vendor.DescribeConfigsResponse

	response = &vendor.DescribeConfigsResponse{
		Resources: []*vendor.ResourceResponse{},
	}
	vendor.testVersionDecodable(t, "empty", response, describeConfigsResponseEmpty, 0)
	if len(response.Resources) != 0 {
		t.Error("Expected no groups")
	}

	response = &vendor.DescribeConfigsResponse{
		Version: 0, Resources: []*vendor.ResourceResponse{
			&vendor.ResourceResponse{
				ErrorCode: 0,
				ErrorMsg:  "",
				Type:      vendor.TopicResource,
				Name:      "foo",
				Configs: []*vendor.ConfigEntry{
					&vendor.ConfigEntry{
						Name:      "segment.ms",
						Value:     "1000",
						ReadOnly:  false,
						Default:   false,
						Sensitive: false,
					},
				},
			},
		},
	}
	vendor.testResponse(t, "response with error", response, describeConfigsResponsePopulatedv0)
}

func TestDescribeConfigsResponsev1(t *testing.T) {
	var response *vendor.DescribeConfigsResponse

	response = &vendor.DescribeConfigsResponse{
		Resources: []*vendor.ResourceResponse{},
	}
	vendor.testVersionDecodable(t, "empty", response, describeConfigsResponseEmpty, 0)
	if len(response.Resources) != 0 {
		t.Error("Expected no groups")
	}

	response = &vendor.DescribeConfigsResponse{
		Version: 1,
		Resources: []*vendor.ResourceResponse{
			&vendor.ResourceResponse{
				ErrorCode: 0,
				ErrorMsg:  "",
				Type:      vendor.TopicResource,
				Name:      "foo",
				Configs: []*vendor.ConfigEntry{
					&vendor.ConfigEntry{
						Name:      "segment.ms",
						Value:     "1000",
						ReadOnly:  false,
						Source:    vendor.SourceStaticBroker,
						Sensitive: false,
						Synonyms:  []*vendor.ConfigSynonym{},
					},
				},
			},
		},
	}
	vendor.testResponse(t, "response with error", response, describeConfigsResponsePopulatedv1)
}

func TestDescribeConfigsResponseWithSynonym(t *testing.T) {
	var response *vendor.DescribeConfigsResponse

	response = &vendor.DescribeConfigsResponse{
		Resources: []*vendor.ResourceResponse{},
	}
	vendor.testVersionDecodable(t, "empty", response, describeConfigsResponseEmpty, 0)
	if len(response.Resources) != 0 {
		t.Error("Expected no groups")
	}

	response = &vendor.DescribeConfigsResponse{
		Version: 1,
		Resources: []*vendor.ResourceResponse{
			&vendor.ResourceResponse{
				ErrorCode: 0,
				ErrorMsg:  "",
				Type:      vendor.TopicResource,
				Name:      "foo",
				Configs: []*vendor.ConfigEntry{
					&vendor.ConfigEntry{
						Name:      "segment.ms",
						Value:     "1000",
						ReadOnly:  false,
						Source:    vendor.SourceStaticBroker,
						Sensitive: false,
						Synonyms: []*vendor.ConfigSynonym{
							{
								ConfigName:  "log.segment.ms",
								ConfigValue: "1000",
								Source:      vendor.SourceStaticBroker,
							},
						},
					},
				},
			},
		},
	}
	vendor.testResponse(t, "response with error", response, describeConfigsResponseWithSynonymv1)
}
