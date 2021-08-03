// Code generated by protoc-gen-go. DO NOT EDIT.
// source: doc.proto

package protos

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

type Doc struct {
	Doc                  string   `protobuf:"bytes,1,opt,name=doc" json:"doc,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Doc) Reset()         { *m = Doc{} }
func (m *Doc) String() string { return proto.CompactTextString(m) }
func (*Doc) ProtoMessage()    {}
func (*Doc) Descriptor() ([]byte, []int) {
	return fileDescriptor_doc_ddb39afaa46ee6d6, []int{0}
}
func (m *Doc) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Doc.Unmarshal(m, b)
}
func (m *Doc) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Doc.Marshal(b, m, deterministic)
}
func (dst *Doc) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Doc.Merge(dst, src)
}
func (m *Doc) XXX_Size() int {
	return xxx_messageInfo_Doc.Size(m)
}
func (m *Doc) XXX_DiscardUnknown() {
	xxx_messageInfo_Doc.DiscardUnknown(m)
}

var xxx_messageInfo_Doc proto.InternalMessageInfo

func (m *Doc) GetDoc() string {
	if m != nil {
		return m.Doc
	}
	return ""
}

func init() {
	proto.RegisterType((*Doc)(nil), "protos.Doc")
}

func init() { proto.RegisterFile("doc.proto", fileDescriptor_doc_ddb39afaa46ee6d6) }

var fileDescriptor_doc_ddb39afaa46ee6d6 = []byte{
	// 66 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0xe2, 0x4c, 0xc9, 0x4f, 0xd6,
	0x2b, 0x28, 0xca, 0x2f, 0xc9, 0x17, 0x62, 0x03, 0x53, 0xc5, 0x4a, 0xe2, 0x5c, 0xcc, 0x2e, 0xf9,
	0xc9, 0x42, 0x02, 0x5c, 0xcc, 0x29, 0xf9, 0xc9, 0x12, 0x8c, 0x0a, 0x8c, 0x1a, 0x9c, 0x41, 0x20,
	0x66, 0x12, 0x44, 0x81, 0x31, 0x20, 0x00, 0x00, 0xff, 0xff, 0x98, 0xdb, 0x15, 0x9b, 0x34, 0x00,
	0x00, 0x00,
}