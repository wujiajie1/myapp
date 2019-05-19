package sarama

import "vendor"

type OffsetCommitResponse struct {
	Version        int16
	ThrottleTimeMs int32
	Errors         map[string]map[int32]vendor.KError
}

func (r *OffsetCommitResponse) AddError(topic string, partition int32, kerror vendor.KError) {
	if r.Errors == nil {
		r.Errors = make(map[string]map[int32]vendor.KError)
	}
	partitions := r.Errors[topic]
	if partitions == nil {
		partitions = make(map[int32]vendor.KError)
		r.Errors[topic] = partitions
	}
	partitions[partition] = kerror
}

func (r *OffsetCommitResponse) encode(pe vendor.packetEncoder) error {
	if r.Version >= 3 {
		pe.putInt32(r.ThrottleTimeMs)
	}
	if err := pe.putArrayLength(len(r.Errors)); err != nil {
		return err
	}
	for topic, partitions := range r.Errors {
		if err := pe.putString(topic); err != nil {
			return err
		}
		if err := pe.putArrayLength(len(partitions)); err != nil {
			return err
		}
		for partition, kerror := range partitions {
			pe.putInt32(partition)
			pe.putInt16(int16(kerror))
		}
	}
	return nil
}

func (r *OffsetCommitResponse) decode(pd vendor.packetDecoder, version int16) (err error) {
	r.Version = version

	if version >= 3 {
		r.ThrottleTimeMs, err = pd.getInt32()
		if err != nil {
			return err
		}
	}

	numTopics, err := pd.getArrayLength()
	if err != nil || numTopics == 0 {
		return err
	}

	r.Errors = make(map[string]map[int32]vendor.KError, numTopics)
	for i := 0; i < numTopics; i++ {
		name, err := pd.getString()
		if err != nil {
			return err
		}

		numErrors, err := pd.getArrayLength()
		if err != nil {
			return err
		}

		r.Errors[name] = make(map[int32]vendor.KError, numErrors)

		for j := 0; j < numErrors; j++ {
			id, err := pd.getInt32()
			if err != nil {
				return err
			}

			tmp, err := pd.getInt16()
			if err != nil {
				return err
			}
			r.Errors[name][id] = vendor.KError(tmp)
		}
	}

	return nil
}

func (r *OffsetCommitResponse) key() int16 {
	return 8
}

func (r *OffsetCommitResponse) version() int16 {
	return r.Version
}

func (r *OffsetCommitResponse) requiredVersion() vendor.KafkaVersion {
	switch r.Version {
	case 1:
		return vendor.V0_8_2_0
	case 2:
		return vendor.V0_9_0_0
	case 3:
		return vendor.V0_11_0_0
	case 4:
		return vendor.V2_0_0_0
	default:
		return vendor.MinVersion
	}
}
