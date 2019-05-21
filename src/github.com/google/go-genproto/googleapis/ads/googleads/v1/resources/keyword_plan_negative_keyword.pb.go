// Code generated by protoc-gen-go. DO NOT EDIT.
// source: google/ads/googleads/v1/resources/keyword_plan_negative_keyword.proto

package resources

import (
	fmt "fmt"
	math "math"

	proto "github.com/golang/protobuf/proto"
	wrappers "github.com/golang/protobuf/ptypes/wrappers"
	enums "google.golang.org/genproto/googleapis/ads/googleads/v1/enums"
	_ "google.golang.org/genproto/googleapis/api/annotations"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion3 // please upgrade the proto package

// A Keyword Plan negative keyword.
// Max number of keyword plan negative keywords per plan: 1000.
type KeywordPlanNegativeKeyword struct {
	// The resource name of the Keyword Plan negative keyword.
	// KeywordPlanNegativeKeyword resource names have the form:
	//
	//
	// `customers/{customer_id}/keywordPlanNegativeKeywords/{kp_negative_keyword_id}`
	ResourceName string `protobuf:"bytes,1,opt,name=resource_name,json=resourceName,proto3" json:"resource_name,omitempty"`
	// The Keyword Plan campaign to which this negative keyword belongs.
	KeywordPlanCampaign *wrappers.StringValue `protobuf:"bytes,2,opt,name=keyword_plan_campaign,json=keywordPlanCampaign,proto3" json:"keyword_plan_campaign,omitempty"`
	// The ID of the Keyword Plan negative keyword.
	Id *wrappers.Int64Value `protobuf:"bytes,3,opt,name=id,proto3" json:"id,omitempty"`
	// The keyword text.
	Text *wrappers.StringValue `protobuf:"bytes,4,opt,name=text,proto3" json:"text,omitempty"`
	// The keyword match type.
	MatchType            enums.KeywordMatchTypeEnum_KeywordMatchType `protobuf:"varint,5,opt,name=match_type,json=matchType,proto3,enum=google.ads.googleads.v1.enums.KeywordMatchTypeEnum_KeywordMatchType" json:"match_type,omitempty"`
	XXX_NoUnkeyedLiteral struct{}                                    `json:"-"`
	XXX_unrecognized     []byte                                      `json:"-"`
	XXX_sizecache        int32                                       `json:"-"`
}

func (m *KeywordPlanNegativeKeyword) Reset()         { *m = KeywordPlanNegativeKeyword{} }
func (m *KeywordPlanNegativeKeyword) String() string { return proto.CompactTextString(m) }
func (*KeywordPlanNegativeKeyword) ProtoMessage()    {}
func (*KeywordPlanNegativeKeyword) Descriptor() ([]byte, []int) {
	return fileDescriptor_bc44a5c087ed7942, []int{0}
}

func (m *KeywordPlanNegativeKeyword) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_KeywordPlanNegativeKeyword.Unmarshal(m, b)
}
func (m *KeywordPlanNegativeKeyword) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_KeywordPlanNegativeKeyword.Marshal(b, m, deterministic)
}
func (m *KeywordPlanNegativeKeyword) XXX_Merge(src proto.Message) {
	xxx_messageInfo_KeywordPlanNegativeKeyword.Merge(m, src)
}
func (m *KeywordPlanNegativeKeyword) XXX_Size() int {
	return xxx_messageInfo_KeywordPlanNegativeKeyword.Size(m)
}
func (m *KeywordPlanNegativeKeyword) XXX_DiscardUnknown() {
	xxx_messageInfo_KeywordPlanNegativeKeyword.DiscardUnknown(m)
}

var xxx_messageInfo_KeywordPlanNegativeKeyword proto.InternalMessageInfo

func (m *KeywordPlanNegativeKeyword) GetResourceName() string {
	if m != nil {
		return m.ResourceName
	}
	return ""
}

func (m *KeywordPlanNegativeKeyword) GetKeywordPlanCampaign() *wrappers.StringValue {
	if m != nil {
		return m.KeywordPlanCampaign
	}
	return nil
}

func (m *KeywordPlanNegativeKeyword) GetId() *wrappers.Int64Value {
	if m != nil {
		return m.Id
	}
	return nil
}

func (m *KeywordPlanNegativeKeyword) GetText() *wrappers.StringValue {
	if m != nil {
		return m.Text
	}
	return nil
}

func (m *KeywordPlanNegativeKeyword) GetMatchType() enums.KeywordMatchTypeEnum_KeywordMatchType {
	if m != nil {
		return m.MatchType
	}
	return enums.KeywordMatchTypeEnum_UNSPECIFIED
}

func init() {
	proto.RegisterType((*KeywordPlanNegativeKeyword)(nil), "google.ads.googleads.v1.resources.KeywordPlanNegativeKeyword")
}

func init() {
	proto.RegisterFile("google/ads/googleads/v1/resources/keyword_plan_negative_keyword.proto", fileDescriptor_bc44a5c087ed7942)
}

var fileDescriptor_bc44a5c087ed7942 = []byte{
	// 436 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x8c, 0x52, 0xdd, 0x6a, 0x14, 0x31,
	0x18, 0x65, 0xa6, 0x55, 0x68, 0xfc, 0xb9, 0x18, 0x11, 0x86, 0xb5, 0xe8, 0x56, 0x29, 0x2c, 0x08,
	0x19, 0xa7, 0x4a, 0x2f, 0xc6, 0xab, 0xa9, 0x96, 0xa2, 0x62, 0x59, 0x56, 0xd9, 0x0b, 0x59, 0x18,
	0xd2, 0x49, 0x8c, 0xa1, 0x93, 0x1f, 0x92, 0xcc, 0xd6, 0x7d, 0x07, 0x9f, 0xc2, 0x4b, 0x1f, 0xc5,
	0x17, 0xf0, 0x1d, 0x7c, 0x0a, 0x99, 0x49, 0x32, 0x8b, 0x2c, 0xab, 0xbd, 0x3b, 0xf9, 0xbe, 0x73,
	0xce, 0xf7, 0x17, 0x70, 0x4a, 0xa5, 0xa4, 0x0d, 0xc9, 0x10, 0x36, 0x99, 0x83, 0x1d, 0x5a, 0xe6,
	0x99, 0x26, 0x46, 0xb6, 0xba, 0x26, 0x26, 0xbb, 0x24, 0xab, 0x2b, 0xa9, 0x71, 0xa5, 0x1a, 0x24,
	0x2a, 0x41, 0x28, 0xb2, 0x6c, 0x49, 0x2a, 0x1f, 0x85, 0x4a, 0x4b, 0x2b, 0x93, 0x03, 0xa7, 0x85,
	0x08, 0x1b, 0x38, 0xd8, 0xc0, 0x65, 0x0e, 0x07, 0x9b, 0xd1, 0xf1, 0xb6, 0x4a, 0x44, 0xb4, 0x7c,
	0x5d, 0x85, 0x23, 0x5b, 0x7f, 0xa9, 0xec, 0x4a, 0x11, 0x67, 0x3d, 0x7a, 0xe8, 0x75, 0xfd, 0xeb,
	0xa2, 0xfd, 0x9c, 0x5d, 0x69, 0xa4, 0x14, 0xd1, 0xc6, 0xe7, 0xf7, 0x83, 0xaf, 0x62, 0x19, 0x12,
	0x42, 0x5a, 0x64, 0x99, 0x14, 0x3e, 0xfb, 0xf8, 0x57, 0x0c, 0x46, 0xef, 0x9c, 0xf5, 0xb4, 0x41,
	0xe2, 0xdc, 0xb7, 0xef, 0x43, 0xc9, 0x13, 0x70, 0x27, 0x74, 0x58, 0x09, 0xc4, 0x49, 0x1a, 0x8d,
	0xa3, 0xc9, 0xde, 0xec, 0x76, 0x08, 0x9e, 0x23, 0x4e, 0x92, 0x29, 0xb8, 0xff, 0xd7, 0x0e, 0x6a,
	0xc4, 0x15, 0x62, 0x54, 0xa4, 0xf1, 0x38, 0x9a, 0xdc, 0x3a, 0xda, 0xf7, 0x13, 0xc3, 0xd0, 0x21,
	0xfc, 0x60, 0x35, 0x13, 0x74, 0x8e, 0x9a, 0x96, 0xcc, 0xee, 0x5d, 0xae, 0xab, 0xbf, 0xf2, 0xc2,
	0xe4, 0x29, 0x88, 0x19, 0x4e, 0x77, 0x7a, 0xf9, 0x83, 0x0d, 0xf9, 0x1b, 0x61, 0x8f, 0x5f, 0x38,
	0x75, 0xcc, 0x70, 0xf2, 0x0c, 0xec, 0x5a, 0xf2, 0xd5, 0xa6, 0xbb, 0xd7, 0xa8, 0xd6, 0x33, 0x93,
	0x1a, 0x80, 0xf5, 0x1a, 0xd3, 0x1b, 0xe3, 0x68, 0x72, 0xf7, 0xe8, 0x35, 0xdc, 0x76, 0xa2, 0x7e,
	0xff, 0xd0, 0x6f, 0xe4, 0x7d, 0xa7, 0xfb, 0xb8, 0x52, 0xe4, 0x54, 0xb4, 0x7c, 0x23, 0x38, 0xdb,
	0xe3, 0x01, 0x9e, 0x7c, 0x8b, 0xc1, 0x61, 0x2d, 0x39, 0xfc, 0xef, 0xe5, 0x4f, 0x1e, 0x6d, 0x3f,
	0xc0, 0xb4, 0x1b, 0x62, 0x1a, 0x7d, 0x7a, 0xeb, 0x5d, 0xa8, 0x6c, 0x90, 0xa0, 0x50, 0x6a, 0x9a,
	0x51, 0x22, 0xfa, 0x11, 0xc3, 0x67, 0x51, 0xcc, 0xfc, 0xe3, 0x97, 0xbe, 0x1c, 0xd0, 0xf7, 0x78,
	0xe7, 0xac, 0x2c, 0x7f, 0xc4, 0x07, 0x67, 0xce, 0xb2, 0xc4, 0x06, 0x3a, 0xd8, 0xa1, 0x79, 0x0e,
	0x67, 0x81, 0xf9, 0x33, 0x70, 0x16, 0x25, 0x36, 0x8b, 0x81, 0xb3, 0x98, 0xe7, 0x8b, 0x81, 0xf3,
	0x3b, 0x3e, 0x74, 0x89, 0xa2, 0x28, 0xb1, 0x29, 0x8a, 0x81, 0x55, 0x14, 0xf3, 0xbc, 0x28, 0x06,
	0xde, 0xc5, 0xcd, 0xbe, 0xd9, 0xe7, 0x7f, 0x02, 0x00, 0x00, 0xff, 0xff, 0x24, 0x2f, 0xaa, 0x5a,
	0x51, 0x03, 0x00, 0x00,
}
