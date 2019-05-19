package sarama

import "vendor"

type CoordinatorType int8

const (
	CoordinatorGroup CoordinatorType = iota
	CoordinatorTransaction
)

type FindCoordinatorRequest struct {
	Version         int16
	CoordinatorKey  string
	CoordinatorType CoordinatorType
}

func (f *FindCoordinatorRequest) encode(pe vendor.packetEncoder) error {
	if err := pe.putString(f.CoordinatorKey); err != nil {
		return err
	}

	if f.Version >= 1 {
		pe.putInt8(int8(f.CoordinatorType))
	}

	return nil
}

func (f *FindCoordinatorRequest) decode(pd vendor.packetDecoder, version int16) (err error) {
	if f.CoordinatorKey, err = pd.getString(); err != nil {
		return err
	}

	if version >= 1 {
		f.Version = version
		coordinatorType, err := pd.getInt8()
		if err != nil {
			return err
		}

		f.CoordinatorType = CoordinatorType(coordinatorType)
	}

	return nil
}

func (f *FindCoordinatorRequest) key() int16 {
	return 10
}

func (f *FindCoordinatorRequest) version() int16 {
	return f.Version
}

func (f *FindCoordinatorRequest) requiredVersion() vendor.KafkaVersion {
	switch f.Version {
	case 1:
		return vendor.V0_11_0_0
	default:
		return vendor.V0_8_2_0
	}
}
