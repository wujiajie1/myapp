package sarama

import (
	"fmt"
	"testing"
	"vendor"
)

var (
	emptyOffsetFetchResponse = []byte{
		0x00, 0x00, 0x00, 0x00}

	emptyOffsetFetchResponseV2 = []byte{
		0x00, 0x00, 0x00, 0x00,
		0x00, 0x2A}

	emptyOffsetFetchResponseV3 = []byte{
		0x00, 0x00, 0x00, 0x09,
		0x00, 0x00, 0x00, 0x00,
		0x00, 0x2A}
)

func TestEmptyOffsetFetchResponse(t *testing.T) {
	for version := 0; version <= 1; version++ {
		response := vendor.OffsetFetchResponse{Version: int16(version)}
		vendor.testResponse(t, fmt.Sprintf("empty v%d", version), &response, emptyOffsetFetchResponse)
	}

	responseV2 := vendor.OffsetFetchResponse{Version: 2, Err: vendor.ErrInvalidRequest}
	vendor.testResponse(t, "empty V2", &responseV2, emptyOffsetFetchResponseV2)

	for version := 3; version <= 5; version++ {
		responseV3 := vendor.OffsetFetchResponse{Version: int16(version), Err: vendor.ErrInvalidRequest, ThrottleTimeMs: 9}
		vendor.testResponse(t, fmt.Sprintf("empty v%d", version), &responseV3, emptyOffsetFetchResponseV3)
	}
}

func TestNormalOffsetFetchResponse(t *testing.T) {
	// The response encoded form cannot be checked for it varies due to
	// unpredictable map traversal order.
	// Hence the 'nil' as byte[] parameter in the 'testResponse(..)' calls

	for version := 0; version <= 1; version++ {
		response := vendor.OffsetFetchResponse{Version: int16(version)}
		response.AddBlock("t", 0, &vendor.OffsetFetchResponseBlock{0, 0, "md", vendor.ErrRequestTimedOut})
		response.Blocks["m"] = nil
		vendor.testResponse(t, fmt.Sprintf("Normal v%d", version), &response, nil)
	}

	responseV2 := vendor.OffsetFetchResponse{Version: 2, Err: vendor.ErrInvalidRequest}
	responseV2.AddBlock("t", 0, &vendor.OffsetFetchResponseBlock{0, 0, "md", vendor.ErrRequestTimedOut})
	responseV2.Blocks["m"] = nil
	vendor.testResponse(t, "normal V2", &responseV2, nil)

	for version := 3; version <= 4; version++ {
		responseV3 := vendor.OffsetFetchResponse{Version: int16(version), Err: vendor.ErrInvalidRequest, ThrottleTimeMs: 9}
		responseV3.AddBlock("t", 0, &vendor.OffsetFetchResponseBlock{0, 0, "md", vendor.ErrRequestTimedOut})
		responseV3.Blocks["m"] = nil
		vendor.testResponse(t, fmt.Sprintf("Normal v%d", version), &responseV3, nil)
	}

	responseV5 := vendor.OffsetFetchResponse{Version: 5, Err: vendor.ErrInvalidRequest, ThrottleTimeMs: 9}
	responseV5.AddBlock("t", 0, &vendor.OffsetFetchResponseBlock{Offset: 10, LeaderEpoch: 100, Metadata: "md", Err: vendor.ErrRequestTimedOut})
	responseV5.Blocks["m"] = nil
	vendor.testResponse(t, "normal V5", &responseV5, nil)
}
