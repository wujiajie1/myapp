package sarama

import (
	"testing"
	"vendor"
)

var (
	emptyDescribeConfigsRequest = []byte{
		0, 0, 0, 0, // 0 configs
	}

	singleDescribeConfigsRequest = []byte{
		0, 0, 0, 1, // 1 config
		2,                   // a topic
		0, 3, 'f', 'o', 'o', // topic name: foo
		0, 0, 0, 1, //1 config name
		0, 10, // 10 chars
		's', 'e', 'g', 'm', 'e', 'n', 't', '.', 'm', 's',
	}

	doubleDescribeConfigsRequest = []byte{
		0, 0, 0, 2, // 2 configs
		2,                   // a topic
		0, 3, 'f', 'o', 'o', // topic name: foo
		0, 0, 0, 2, //2 config name
		0, 10, // 10 chars
		's', 'e', 'g', 'm', 'e', 'n', 't', '.', 'm', 's',
		0, 12, // 12 chars
		'r', 'e', 't', 'e', 'n', 't', 'i', 'o', 'n', '.', 'm', 's',
		2,                   // a topic
		0, 3, 'b', 'a', 'r', // topic name: foo
		0, 0, 0, 1, // 1 config
		0, 10, // 10 chars
		's', 'e', 'g', 'm', 'e', 'n', 't', '.', 'm', 's',
	}

	singleDescribeConfigsRequestAllConfigs = []byte{
		0, 0, 0, 1, // 1 config
		2,                   // a topic
		0, 3, 'f', 'o', 'o', // topic name: foo
		255, 255, 255, 255, // all configs
	}

	singleDescribeConfigsRequestAllConfigsv1 = []byte{
		0, 0, 0, 1, // 1 config
		2,                   // a topic
		0, 3, 'f', 'o', 'o', // topic name: foo
		255, 255, 255, 255, // no configs
		1, //synoms
	}
)

func TestDescribeConfigsRequestv0(t *testing.T) {
	var request *vendor.DescribeConfigsRequest

	request = &vendor.DescribeConfigsRequest{
		Version:   0,
		Resources: []*vendor.ConfigResource{},
	}
	vendor.testRequest(t, "no requests", request, emptyDescribeConfigsRequest)

	configs := []string{"segment.ms"}
	request = &vendor.DescribeConfigsRequest{
		Version: 0,
		Resources: []*vendor.ConfigResource{
			&vendor.ConfigResource{
				Type:        vendor.TopicResource,
				Name:        "foo",
				ConfigNames: configs,
			},
		},
	}

	vendor.testRequest(t, "one config", request, singleDescribeConfigsRequest)

	request = &vendor.DescribeConfigsRequest{
		Version: 0,
		Resources: []*vendor.ConfigResource{
			&vendor.ConfigResource{
				Type:        vendor.TopicResource,
				Name:        "foo",
				ConfigNames: []string{"segment.ms", "retention.ms"},
			},
			&vendor.ConfigResource{
				Type:        vendor.TopicResource,
				Name:        "bar",
				ConfigNames: []string{"segment.ms"},
			},
		},
	}
	vendor.testRequest(t, "two configs", request, doubleDescribeConfigsRequest)

	request = &vendor.DescribeConfigsRequest{
		Version: 0,
		Resources: []*vendor.ConfigResource{
			&vendor.ConfigResource{
				Type: vendor.TopicResource,
				Name: "foo",
			},
		},
	}

	vendor.testRequest(t, "one topic, all configs", request, singleDescribeConfigsRequestAllConfigs)
}

func TestDescribeConfigsRequestv1(t *testing.T) {
	var request *vendor.DescribeConfigsRequest

	request = &vendor.DescribeConfigsRequest{
		Version: 1,
		Resources: []*vendor.ConfigResource{
			{
				Type: vendor.TopicResource,
				Name: "foo",
			},
		},
		IncludeSynonyms: true,
	}

	vendor.testRequest(t, "one topic, all configs", request, singleDescribeConfigsRequestAllConfigsv1)
}
