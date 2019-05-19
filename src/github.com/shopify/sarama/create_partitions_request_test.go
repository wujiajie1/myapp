package sarama

import (
	"testing"
	"time"
	"vendor"
)

var (
	createPartitionRequestNoAssignment = []byte{
		0, 0, 0, 1, // one topic
		0, 5, 't', 'o', 'p', 'i', 'c',
		0, 0, 0, 3, // 3 partitions
		255, 255, 255, 255, // no assignments
		0, 0, 0, 100, // timeout
		0, // validate only = false
	}

	createPartitionRequestAssignment = []byte{
		0, 0, 0, 1,
		0, 5, 't', 'o', 'p', 'i', 'c',
		0, 0, 0, 3, // 3 partitions
		0, 0, 0, 2,
		0, 0, 0, 2,
		0, 0, 0, 2, 0, 0, 0, 3,
		0, 0, 0, 2,
		0, 0, 0, 3, 0, 0, 0, 1,
		0, 0, 0, 100,
		1, // validate only = true
	}
)

func TestCreatePartitionsRequest(t *testing.T) {
	req := &vendor.CreatePartitionsRequest{
		TopicPartitions: map[string]*vendor.TopicPartition{
			"topic": &vendor.TopicPartition{
				Count: 3,
			},
		},
		Timeout: 100 * time.Millisecond,
	}

	buf := vendor.testRequestEncode(t, "no assignment", req, createPartitionRequestNoAssignment)
	vendor.testRequestDecode(t, "no assignment", req, buf)

	req.ValidateOnly = true
	req.TopicPartitions["topic"].Assignment = [][]int32{{2, 3}, {3, 1}}

	buf = vendor.testRequestEncode(t, "assignment", req, createPartitionRequestAssignment)
	vendor.testRequestDecode(t, "assignment", req, buf)
}
