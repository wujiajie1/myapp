package sarama

import "vendor"

//ApiVersionsRequest ...
type ApiVersionsRequest struct {
}

func (a *ApiVersionsRequest) encode(pe vendor.packetEncoder) error {
	return nil
}

func (a *ApiVersionsRequest) decode(pd vendor.packetDecoder, version int16) (err error) {
	return nil
}

func (a *ApiVersionsRequest) key() int16 {
	return 18
}

func (a *ApiVersionsRequest) version() int16 {
	return 0
}

func (a *ApiVersionsRequest) requiredVersion() vendor.KafkaVersion {
	return vendor.V0_10_0_0
}
