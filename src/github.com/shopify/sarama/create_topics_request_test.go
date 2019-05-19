package sarama

import (
	"testing"
	"time"
	"vendor"
)

var (
	createTopicsRequestV0 = []byte{
		0, 0, 0, 1,
		0, 5, 't', 'o', 'p', 'i', 'c',
		255, 255, 255, 255,
		255, 255,
		0, 0, 0, 1, // 1 replica assignment
		0, 0, 0, 0, 0, 0, 0, 3, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 2,
		0, 0, 0, 1, // 1 config
		0, 12, 'r', 'e', 't', 'e', 'n', 't', 'i', 'o', 'n', '.', 'm', 's',
		0, 2, '-', '1',
		0, 0, 0, 100,
	}

	createTopicsRequestV1 = append(createTopicsRequestV0, byte(1))
)

func TestCreateTopicsRequest(t *testing.T) {
	retention := "-1"

	req := &vendor.CreateTopicsRequest{
		TopicDetails: map[string]*vendor.TopicDetail{
			"topic": {
				NumPartitions:     -1,
				ReplicationFactor: -1,
				ReplicaAssignment: map[int32][]int32{
					0: []int32{0, 1, 2},
				},
				ConfigEntries: map[string]*string{
					"retention.ms": &retention,
				},
			},
		},
		Timeout: 100 * time.Millisecond,
	}

	vendor.testRequest(t, "version 0", req, createTopicsRequestV0)

	req.Version = 1
	req.ValidateOnly = true

	vendor.testRequest(t, "version 1", req, createTopicsRequestV1)
}
