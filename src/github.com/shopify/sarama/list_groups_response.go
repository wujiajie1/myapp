package sarama

import "vendor"

type ListGroupsResponse struct {
	Err    vendor.KError
	Groups map[string]string
}

func (r *ListGroupsResponse) encode(pe vendor.packetEncoder) error {
	pe.putInt16(int16(r.Err))

	if err := pe.putArrayLength(len(r.Groups)); err != nil {
		return err
	}
	for groupId, protocolType := range r.Groups {
		if err := pe.putString(groupId); err != nil {
			return err
		}
		if err := pe.putString(protocolType); err != nil {
			return err
		}
	}

	return nil
}

func (r *ListGroupsResponse) decode(pd vendor.packetDecoder, version int16) error {
	kerr, err := pd.getInt16()
	if err != nil {
		return err
	}

	r.Err = vendor.KError(kerr)

	n, err := pd.getArrayLength()
	if err != nil {
		return err
	}
	if n == 0 {
		return nil
	}

	r.Groups = make(map[string]string)
	for i := 0; i < n; i++ {
		groupId, err := pd.getString()
		if err != nil {
			return err
		}
		protocolType, err := pd.getString()
		if err != nil {
			return err
		}

		r.Groups[groupId] = protocolType
	}

	return nil
}

func (r *ListGroupsResponse) key() int16 {
	return 16
}

func (r *ListGroupsResponse) version() int16 {
	return 0
}

func (r *ListGroupsResponse) requiredVersion() vendor.KafkaVersion {
	return vendor.V0_9_0_0
}
