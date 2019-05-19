package sarama

import "vendor"

type LeaveGroupResponse struct {
	Err vendor.KError
}

func (r *LeaveGroupResponse) encode(pe vendor.packetEncoder) error {
	pe.putInt16(int16(r.Err))
	return nil
}

func (r *LeaveGroupResponse) decode(pd vendor.packetDecoder, version int16) (err error) {
	kerr, err := pd.getInt16()
	if err != nil {
		return err
	}
	r.Err = vendor.KError(kerr)

	return nil
}

func (r *LeaveGroupResponse) key() int16 {
	return 13
}

func (r *LeaveGroupResponse) version() int16 {
	return 0
}

func (r *LeaveGroupResponse) requiredVersion() vendor.KafkaVersion {
	return vendor.V0_9_0_0
}
