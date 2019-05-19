package sarama

import "vendor"

type DescribeConfigsRequest struct {
	Version         int16
	Resources       []*ConfigResource
	IncludeSynonyms bool
}

type ConfigResource struct {
	Type        vendor.ConfigResourceType
	Name        string
	ConfigNames []string
}

func (r *DescribeConfigsRequest) encode(pe vendor.packetEncoder) error {
	if err := pe.putArrayLength(len(r.Resources)); err != nil {
		return err
	}

	for _, c := range r.Resources {
		pe.putInt8(int8(c.Type))
		if err := pe.putString(c.Name); err != nil {
			return err
		}

		if len(c.ConfigNames) == 0 {
			pe.putInt32(-1)
			continue
		}
		if err := pe.putStringArray(c.ConfigNames); err != nil {
			return err
		}
	}

	if r.Version >= 1 {
		pe.putBool(r.IncludeSynonyms)
	}

	return nil
}

func (r *DescribeConfigsRequest) decode(pd vendor.packetDecoder, version int16) (err error) {
	n, err := pd.getArrayLength()
	if err != nil {
		return err
	}

	r.Resources = make([]*ConfigResource, n)

	for i := 0; i < n; i++ {
		r.Resources[i] = &ConfigResource{}
		t, err := pd.getInt8()
		if err != nil {
			return err
		}
		r.Resources[i].Type = vendor.ConfigResourceType(t)
		name, err := pd.getString()
		if err != nil {
			return err
		}
		r.Resources[i].Name = name

		confLength, err := pd.getArrayLength()

		if err != nil {
			return err
		}

		if confLength == -1 {
			continue
		}

		cfnames := make([]string, confLength)
		for i := 0; i < confLength; i++ {
			s, err := pd.getString()
			if err != nil {
				return err
			}
			cfnames[i] = s
		}
		r.Resources[i].ConfigNames = cfnames
	}
	r.Version = version
	if r.Version >= 1 {
		b, err := pd.getBool()
		if err != nil {
			return err
		}
		r.IncludeSynonyms = b
	}

	return nil
}

func (r *DescribeConfigsRequest) key() int16 {
	return 32
}

func (r *DescribeConfigsRequest) version() int16 {
	return r.Version
}

func (r *DescribeConfigsRequest) requiredVersion() vendor.KafkaVersion {
	switch r.Version {
	case 1:
		return vendor.V1_1_0_0
	case 2:
		return vendor.V2_0_0_0
	default:
		return vendor.V0_11_0_0
	}
}
