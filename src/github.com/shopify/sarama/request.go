package sarama

import (
	"encoding/binary"
	"fmt"
	"io"
	"vendor"
)

type protocolBody interface {
	vendor.encoder
	vendor.versionedDecoder
	key() int16
	version() int16
	requiredVersion() vendor.KafkaVersion
}

type request struct {
	correlationID int32
	clientID      string
	body          protocolBody
}

func (r *request) encode(pe vendor.packetEncoder) error {
	pe.push(&vendor.lengthField{})
	pe.putInt16(r.body.key())
	pe.putInt16(r.body.version())
	pe.putInt32(r.correlationID)

	err := pe.putString(r.clientID)
	if err != nil {
		return err
	}

	err = r.body.encode(pe)
	if err != nil {
		return err
	}

	return pe.pop()
}

func (r *request) decode(pd vendor.packetDecoder) (err error) {
	key, err := pd.getInt16()
	if err != nil {
		return err
	}

	version, err := pd.getInt16()
	if err != nil {
		return err
	}

	r.correlationID, err = pd.getInt32()
	if err != nil {
		return err
	}

	r.clientID, err = pd.getString()
	if err != nil {
		return err
	}

	r.body = allocateBody(key, version)
	if r.body == nil {
		return vendor.PacketDecodingError{fmt.Sprintf("unknown request key (%d)", key)}
	}

	return r.body.decode(pd, version)
}

func decodeRequest(r io.Reader) (*request, int, error) {
	var (
		bytesRead   int
		lengthBytes = make([]byte, 4)
	)

	if _, err := io.ReadFull(r, lengthBytes); err != nil {
		return nil, bytesRead, err
	}

	bytesRead += len(lengthBytes)
	length := int32(binary.BigEndian.Uint32(lengthBytes))

	if length <= 4 || length > vendor.MaxRequestSize {
		return nil, bytesRead, vendor.PacketDecodingError{fmt.Sprintf("message of length %d too large or too small", length)}
	}

	encodedReq := make([]byte, length)
	if _, err := io.ReadFull(r, encodedReq); err != nil {
		return nil, bytesRead, err
	}

	bytesRead += len(encodedReq)

	req := &request{}
	if err := vendor.decode(encodedReq, req); err != nil {
		return nil, bytesRead, err
	}

	return req, bytesRead, nil
}

func allocateBody(key, version int16) protocolBody {
	switch key {
	case 0:
		return &vendor.ProduceRequest{}
	case 1:
		return &vendor.FetchRequest{}
	case 2:
		return &vendor.OffsetRequest{Version: version}
	case 3:
		return &vendor.MetadataRequest{}
	case 8:
		return &vendor.OffsetCommitRequest{Version: version}
	case 9:
		return &vendor.OffsetFetchRequest{}
	case 10:
		return &vendor.FindCoordinatorRequest{}
	case 11:
		return &vendor.JoinGroupRequest{}
	case 12:
		return &vendor.HeartbeatRequest{}
	case 13:
		return &vendor.LeaveGroupRequest{}
	case 14:
		return &vendor.SyncGroupRequest{}
	case 15:
		return &vendor.DescribeGroupsRequest{}
	case 16:
		return &vendor.ListGroupsRequest{}
	case 17:
		return &vendor.SaslHandshakeRequest{}
	case 18:
		return &vendor.ApiVersionsRequest{}
	case 19:
		return &vendor.CreateTopicsRequest{}
	case 20:
		return &vendor.DeleteTopicsRequest{}
	case 21:
		return &vendor.DeleteRecordsRequest{}
	case 22:
		return &vendor.InitProducerIDRequest{}
	case 24:
		return &vendor.AddPartitionsToTxnRequest{}
	case 25:
		return &vendor.AddOffsetsToTxnRequest{}
	case 26:
		return &vendor.EndTxnRequest{}
	case 28:
		return &vendor.TxnOffsetCommitRequest{}
	case 29:
		return &vendor.DescribeAclsRequest{}
	case 30:
		return &vendor.CreateAclsRequest{}
	case 31:
		return &vendor.DeleteAclsRequest{}
	case 32:
		return &vendor.DescribeConfigsRequest{}
	case 33:
		return &vendor.AlterConfigsRequest{}
	case 36:
		return &vendor.SaslAuthenticateRequest{}
	case 37:
		return &vendor.CreatePartitionsRequest{}
	case 42:
		return &vendor.DeleteGroupsRequest{}
	}
	return nil
}
