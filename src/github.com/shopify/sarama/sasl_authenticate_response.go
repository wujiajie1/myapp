package sarama

import "vendor"

type SaslAuthenticateResponse struct {
	Err           vendor.KError
	ErrorMessage  *string
	SaslAuthBytes []byte
}

func (r *SaslAuthenticateResponse) encode(pe vendor.packetEncoder) error {
	pe.putInt16(int16(r.Err))
	if err := pe.putNullableString(r.ErrorMessage); err != nil {
		return err
	}
	return pe.putBytes(r.SaslAuthBytes)
}

func (r *SaslAuthenticateResponse) decode(pd vendor.packetDecoder, version int16) error {
	kerr, err := pd.getInt16()
	if err != nil {
		return err
	}

	r.Err = vendor.KError(kerr)

	if r.ErrorMessage, err = pd.getNullableString(); err != nil {
		return err
	}

	r.SaslAuthBytes, err = pd.getBytes()

	return err
}

func (r *SaslAuthenticateResponse) key() int16 {
	return vendor.APIKeySASLAuth
}

func (r *SaslAuthenticateResponse) version() int16 {
	return 0
}

func (r *SaslAuthenticateResponse) requiredVersion() vendor.KafkaVersion {
	return vendor.V1_0_0_0
}
