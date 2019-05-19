package sarama

import (
	"time"
	"vendor"
)

//AddOffsetsToTxnResponse is a response type for adding offsets to txns
type AddOffsetsToTxnResponse struct {
	ThrottleTime time.Duration
	Err          vendor.KError
}

func (a *AddOffsetsToTxnResponse) encode(pe vendor.packetEncoder) error {
	pe.putInt32(int32(a.ThrottleTime / time.Millisecond))
	pe.putInt16(int16(a.Err))
	return nil
}

func (a *AddOffsetsToTxnResponse) decode(pd vendor.packetDecoder, version int16) (err error) {
	throttleTime, err := pd.getInt32()
	if err != nil {
		return err
	}
	a.ThrottleTime = time.Duration(throttleTime) * time.Millisecond

	kerr, err := pd.getInt16()
	if err != nil {
		return err
	}
	a.Err = vendor.KError(kerr)

	return nil
}

func (a *AddOffsetsToTxnResponse) key() int16 {
	return 25
}

func (a *AddOffsetsToTxnResponse) version() int16 {
	return 0
}

func (a *AddOffsetsToTxnResponse) requiredVersion() vendor.KafkaVersion {
	return vendor.V0_11_0_0
}
