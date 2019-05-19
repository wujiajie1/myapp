package sarama

import "vendor"

type SyncGroupResponse struct {
	Err              vendor.KError
	MemberAssignment []byte
}

func (r *SyncGroupResponse) GetMemberAssignment() (*vendor.ConsumerGroupMemberAssignment, error) {
	assignment := new(vendor.ConsumerGroupMemberAssignment)
	err := vendor.decode(r.MemberAssignment, assignment)
	return assignment, err
}

func (r *SyncGroupResponse) encode(pe vendor.packetEncoder) error {
	pe.putInt16(int16(r.Err))
	return pe.putBytes(r.MemberAssignment)
}

func (r *SyncGroupResponse) decode(pd vendor.packetDecoder, version int16) (err error) {
	kerr, err := pd.getInt16()
	if err != nil {
		return err
	}

	r.Err = vendor.KError(kerr)

	r.MemberAssignment, err = pd.getBytes()
	return
}

func (r *SyncGroupResponse) key() int16 {
	return 14
}

func (r *SyncGroupResponse) version() int16 {
	return 0
}

func (r *SyncGroupResponse) requiredVersion() vendor.KafkaVersion {
	return vendor.V0_9_0_0
}
