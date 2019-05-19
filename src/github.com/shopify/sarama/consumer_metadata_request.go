package sarama

import "vendor"

//ConsumerMetadataRequest is used for metadata requests
type ConsumerMetadataRequest struct {
	ConsumerGroup string
}

func (r *ConsumerMetadataRequest) encode(pe vendor.packetEncoder) error {
	tmp := new(vendor.FindCoordinatorRequest)
	tmp.CoordinatorKey = r.ConsumerGroup
	tmp.CoordinatorType = vendor.CoordinatorGroup
	return tmp.encode(pe)
}

func (r *ConsumerMetadataRequest) decode(pd vendor.packetDecoder, version int16) (err error) {
	tmp := new(vendor.FindCoordinatorRequest)
	if err := tmp.decode(pd, version); err != nil {
		return err
	}
	r.ConsumerGroup = tmp.CoordinatorKey
	return nil
}

func (r *ConsumerMetadataRequest) key() int16 {
	return 10
}

func (r *ConsumerMetadataRequest) version() int16 {
	return 0
}

func (r *ConsumerMetadataRequest) requiredVersion() vendor.KafkaVersion {
	return vendor.V0_8_2_0
}
