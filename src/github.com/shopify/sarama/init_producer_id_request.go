package sarama

import (
	"time"
	"vendor"
)

type InitProducerIDRequest struct {
	TransactionalID    *string
	TransactionTimeout time.Duration
}

func (i *InitProducerIDRequest) encode(pe vendor.packetEncoder) error {
	if err := pe.putNullableString(i.TransactionalID); err != nil {
		return err
	}
	pe.putInt32(int32(i.TransactionTimeout / time.Millisecond))

	return nil
}

func (i *InitProducerIDRequest) decode(pd vendor.packetDecoder, version int16) (err error) {
	if i.TransactionalID, err = pd.getNullableString(); err != nil {
		return err
	}

	timeout, err := pd.getInt32()
	if err != nil {
		return err
	}
	i.TransactionTimeout = time.Duration(timeout) * time.Millisecond

	return nil
}

func (i *InitProducerIDRequest) key() int16 {
	return 22
}

func (i *InitProducerIDRequest) version() int16 {
	return 0
}

func (i *InitProducerIDRequest) requiredVersion() vendor.KafkaVersion {
	return vendor.V0_11_0_0
}
