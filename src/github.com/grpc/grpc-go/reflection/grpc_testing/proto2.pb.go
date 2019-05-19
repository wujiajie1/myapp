// Code generated by protoc-gen-go. DO NOT EDIT.
// source: proto2.proto

package grpc_testing

import (
	"vendor"
)
import "fmt"
import "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = vendor.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = vendor.ProtoPackageIsVersion2 // please upgrade the proto package

type ToBeExtended struct {
	Foo                           *int32   `protobuf:"varint,1,req,name=foo" json:"foo,omitempty"`
	XXX_NoUnkeyedLiteral          struct{} `json:"-"`
	vendor.XXX_InternalExtensions `json:"-"`
	XXX_unrecognized              []byte `json:"-"`
	XXX_sizecache                 int32  `json:"-"`
}

func (m *ToBeExtended) Reset()         { *m = ToBeExtended{} }
func (m *ToBeExtended) String() string { return vendor.CompactTextString(m) }
func (*ToBeExtended) ProtoMessage()    {}
func (*ToBeExtended) Descriptor() ([]byte, []int) {
	return fileDescriptor_proto2_b16f7a513d0acdc0, []int{0}
}

var extRange_ToBeExtended = []vendor.ExtensionRange{
	{Start: 10, End: 30},
}

func (*ToBeExtended) ExtensionRangeArray() []vendor.ExtensionRange {
	return extRange_ToBeExtended
}
func (m *ToBeExtended) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ToBeExtended.Unmarshal(m, b)
}
func (m *ToBeExtended) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ToBeExtended.Marshal(b, m, deterministic)
}
func (dst *ToBeExtended) XXX_Merge(src vendor.Message) {
	xxx_messageInfo_ToBeExtended.Merge(dst, src)
}
func (m *ToBeExtended) XXX_Size() int {
	return xxx_messageInfo_ToBeExtended.Size(m)
}
func (m *ToBeExtended) XXX_DiscardUnknown() {
	xxx_messageInfo_ToBeExtended.DiscardUnknown(m)
}

var xxx_messageInfo_ToBeExtended vendor.InternalMessageInfo

func (m *ToBeExtended) GetFoo() int32 {
	if m != nil && m.Foo != nil {
		return *m.Foo
	}
	return 0
}

func init() {
	vendor.RegisterType((*ToBeExtended)(nil), "grpc.testing.ToBeExtended")
}

func init() { vendor.RegisterFile("proto2.proto", fileDescriptor_proto2_b16f7a513d0acdc0) }

var fileDescriptor_proto2_b16f7a513d0acdc0 = []byte{
	// 86 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0xe2, 0x29, 0x28, 0xca, 0x2f,
	0xc9, 0x37, 0xd2, 0x03, 0x53, 0x42, 0x3c, 0xe9, 0x45, 0x05, 0xc9, 0x7a, 0x25, 0xa9, 0xc5, 0x25,
	0x99, 0x79, 0xe9, 0x4a, 0x6a, 0x5c, 0x3c, 0x21, 0xf9, 0x4e, 0xa9, 0xae, 0x15, 0x25, 0xa9, 0x79,
	0x29, 0xa9, 0x29, 0x42, 0x02, 0x5c, 0xcc, 0x69, 0xf9, 0xf9, 0x12, 0x8c, 0x0a, 0x4c, 0x1a, 0xac,
	0x41, 0x20, 0xa6, 0x16, 0x0b, 0x07, 0x97, 0x80, 0x3c, 0x20, 0x00, 0x00, 0xff, 0xff, 0x74, 0x86,
	0x9c, 0x08, 0x44, 0x00, 0x00, 0x00,
}
