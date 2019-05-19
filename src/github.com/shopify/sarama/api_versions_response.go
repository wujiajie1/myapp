package sarama

import "vendor"

//ApiVersionsResponseBlock is an api version reponse block type
type ApiVersionsResponseBlock struct {
	ApiKey     int16
	MinVersion int16
	MaxVersion int16
}

func (b *ApiVersionsResponseBlock) encode(pe vendor.packetEncoder) error {
	pe.putInt16(b.ApiKey)
	pe.putInt16(b.MinVersion)
	pe.putInt16(b.MaxVersion)
	return nil
}

func (b *ApiVersionsResponseBlock) decode(pd vendor.packetDecoder) error {
	var err error

	if b.ApiKey, err = pd.getInt16(); err != nil {
		return err
	}

	if b.MinVersion, err = pd.getInt16(); err != nil {
		return err
	}

	if b.MaxVersion, err = pd.getInt16(); err != nil {
		return err
	}

	return nil
}

//ApiVersionsResponse is an api version response type
type ApiVersionsResponse struct {
	Err         vendor.KError
	ApiVersions []*ApiVersionsResponseBlock
}

func (r *ApiVersionsResponse) encode(pe vendor.packetEncoder) error {
	pe.putInt16(int16(r.Err))
	if err := pe.putArrayLength(len(r.ApiVersions)); err != nil {
		return err
	}
	for _, apiVersion := range r.ApiVersions {
		if err := apiVersion.encode(pe); err != nil {
			return err
		}
	}
	return nil
}

func (r *ApiVersionsResponse) decode(pd vendor.packetDecoder, version int16) error {
	kerr, err := pd.getInt16()
	if err != nil {
		return err
	}

	r.Err = vendor.KError(kerr)

	numBlocks, err := pd.getArrayLength()
	if err != nil {
		return err
	}

	r.ApiVersions = make([]*ApiVersionsResponseBlock, numBlocks)
	for i := 0; i < numBlocks; i++ {
		block := new(ApiVersionsResponseBlock)
		if err := block.decode(pd); err != nil {
			return err
		}
		r.ApiVersions[i] = block
	}

	return nil
}

func (r *ApiVersionsResponse) key() int16 {
	return 18
}

func (r *ApiVersionsResponse) version() int16 {
	return 0
}

func (r *ApiVersionsResponse) requiredVersion() vendor.KafkaVersion {
	return vendor.V0_10_0_0
}
