package sarama

import (
	"fmt"
	"vendor"
)

const responseLengthSize = 4
const correlationIDSize = 4

type responseHeader struct {
	length        int32
	correlationID int32
}

func (r *responseHeader) decode(pd vendor.packetDecoder) (err error) {
	r.length, err = pd.getInt32()
	if err != nil {
		return err
	}
	if r.length <= 4 || r.length > vendor.MaxResponseSize {
		return vendor.PacketDecodingError{fmt.Sprintf("message of length %d too large or too small", r.length)}
	}

	r.correlationID, err = pd.getInt32()
	return err
}
