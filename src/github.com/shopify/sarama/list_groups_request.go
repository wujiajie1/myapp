package sarama

import "vendor"

type ListGroupsRequest struct {
}

func (r *ListGroupsRequest) encode(pe vendor.packetEncoder) error {
	return nil
}

func (r *ListGroupsRequest) decode(pd vendor.packetDecoder, version int16) (err error) {
	return nil
}

func (r *ListGroupsRequest) key() int16 {
	return 16
}

func (r *ListGroupsRequest) version() int16 {
	return 0
}

func (r *ListGroupsRequest) requiredVersion() vendor.KafkaVersion {
	return vendor.V0_9_0_0
}
