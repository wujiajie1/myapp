package sarama

import (
	"time"
	"vendor"
)

var NoNode = &vendor.Broker{id: -1, addr: ":-1"}

type FindCoordinatorResponse struct {
	Version      int16
	ThrottleTime time.Duration
	Err          vendor.KError
	ErrMsg       *string
	Coordinator  *vendor.Broker
}

func (f *FindCoordinatorResponse) decode(pd vendor.packetDecoder, version int16) (err error) {
	if version >= 1 {
		f.Version = version

		throttleTime, err := pd.getInt32()
		if err != nil {
			return err
		}
		f.ThrottleTime = time.Duration(throttleTime) * time.Millisecond
	}

	tmp, err := pd.getInt16()
	if err != nil {
		return err
	}
	f.Err = vendor.KError(tmp)

	if version >= 1 {
		if f.ErrMsg, err = pd.getNullableString(); err != nil {
			return err
		}
	}

	coordinator := new(vendor.Broker)
	// The version is hardcoded to 0, as version 1 of the Broker-decode
	// contains the rack-field which is not present in the FindCoordinatorResponse.
	if err := coordinator.decode(pd, 0); err != nil {
		return err
	}
	if coordinator.addr == ":0" {
		return nil
	}
	f.Coordinator = coordinator

	return nil
}

func (f *FindCoordinatorResponse) encode(pe vendor.packetEncoder) error {
	if f.Version >= 1 {
		pe.putInt32(int32(f.ThrottleTime / time.Millisecond))
	}

	pe.putInt16(int16(f.Err))

	if f.Version >= 1 {
		if err := pe.putNullableString(f.ErrMsg); err != nil {
			return err
		}
	}

	coordinator := f.Coordinator
	if coordinator == nil {
		coordinator = NoNode
	}
	if err := coordinator.encode(pe, 0); err != nil {
		return err
	}
	return nil
}

func (f *FindCoordinatorResponse) key() int16 {
	return 10
}

func (f *FindCoordinatorResponse) version() int16 {
	return f.Version
}

func (f *FindCoordinatorResponse) requiredVersion() vendor.KafkaVersion {
	switch f.Version {
	case 1:
		return vendor.V0_11_0_0
	default:
		return vendor.V0_8_2_0
	}
}
