package sarama

import (
	"time"
	"vendor"
)

type EndTxnResponse struct {
	ThrottleTime time.Duration
	Err          vendor.KError
}

func (e *EndTxnResponse) encode(pe vendor.packetEncoder) error {
	pe.putInt32(int32(e.ThrottleTime / time.Millisecond))
	pe.putInt16(int16(e.Err))
	return nil
}

func (e *EndTxnResponse) decode(pd vendor.packetDecoder, version int16) (err error) {
	throttleTime, err := pd.getInt32()
	if err != nil {
		return err
	}
	e.ThrottleTime = time.Duration(throttleTime) * time.Millisecond

	kerr, err := pd.getInt16()
	if err != nil {
		return err
	}
	e.Err = vendor.KError(kerr)

	return nil
}

func (e *EndTxnResponse) key() int16 {
	return 25
}

func (e *EndTxnResponse) version() int16 {
	return 0
}

func (e *EndTxnResponse) requiredVersion() vendor.KafkaVersion {
	return vendor.V0_11_0_0
}
