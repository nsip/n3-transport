// Code generated by protoc-gen-go. DO NOT EDIT.
// source: n3msg.proto

/*
Package pb is a generated protocol buffer package.

It is generated from these files:
	n3msg.proto

It has these top-level messages:
	SPOTuple
	N3Message
	TxSummary
*/
package pb

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

import (
	context "golang.org/x/net/context"
	grpc "google.golang.org/grpc"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

// core message type for lowest-level data tuples
type SPOTuple struct {
	Subject   string `protobuf:"bytes,1,opt,name=Subject" json:"Subject,omitempty"`
	Predicate string `protobuf:"bytes,2,opt,name=Predicate" json:"Predicate,omitempty"`
	Object    string `protobuf:"bytes,3,opt,name=Object" json:"Object,omitempty"`
	Version   int64  `protobuf:"varint,4,opt,name=Version" json:"Version,omitempty"`
}

func (m *SPOTuple) Reset()                    { *m = SPOTuple{} }
func (m *SPOTuple) String() string            { return proto.CompactTextString(m) }
func (*SPOTuple) ProtoMessage()               {}
func (*SPOTuple) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

func (m *SPOTuple) GetSubject() string {
	if m != nil {
		return m.Subject
	}
	return ""
}

func (m *SPOTuple) GetPredicate() string {
	if m != nil {
		return m.Predicate
	}
	return ""
}

func (m *SPOTuple) GetObject() string {
	if m != nil {
		return m.Object
	}
	return ""
}

func (m *SPOTuple) GetVersion() int64 {
	if m != nil {
		return m.Version
	}
	return 0
}

type N3Message struct {
	Payload   []byte `protobuf:"bytes,1,opt,name=Payload,proto3" json:"Payload,omitempty"`
	SndId     string `protobuf:"bytes,2,opt,name=SndId" json:"SndId,omitempty"`
	NameSpace string `protobuf:"bytes,3,opt,name=NameSpace" json:"NameSpace,omitempty"`
	CtxName   string `protobuf:"bytes,4,opt,name=CtxName" json:"CtxName,omitempty"`
	DispId    string `protobuf:"bytes,5,opt,name=DispId" json:"DispId,omitempty"`
}

func (m *N3Message) Reset()                    { *m = N3Message{} }
func (m *N3Message) String() string            { return proto.CompactTextString(m) }
func (*N3Message) ProtoMessage()               {}
func (*N3Message) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

func (m *N3Message) GetPayload() []byte {
	if m != nil {
		return m.Payload
	}
	return nil
}

func (m *N3Message) GetSndId() string {
	if m != nil {
		return m.SndId
	}
	return ""
}

func (m *N3Message) GetNameSpace() string {
	if m != nil {
		return m.NameSpace
	}
	return ""
}

func (m *N3Message) GetCtxName() string {
	if m != nil {
		return m.CtxName
	}
	return ""
}

func (m *N3Message) GetDispId() string {
	if m != nil {
		return m.DispId
	}
	return ""
}

type TxSummary struct {
	MsgCount int64 `protobuf:"varint,1,opt,name=MsgCount" json:"MsgCount,omitempty"`
}

func (m *TxSummary) Reset()                    { *m = TxSummary{} }
func (m *TxSummary) String() string            { return proto.CompactTextString(m) }
func (*TxSummary) ProtoMessage()               {}
func (*TxSummary) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{2} }

func (m *TxSummary) GetMsgCount() int64 {
	if m != nil {
		return m.MsgCount
	}
	return 0
}

func init() {
	proto.RegisterType((*SPOTuple)(nil), "pb.SPOTuple")
	proto.RegisterType((*N3Message)(nil), "pb.N3Message")
	proto.RegisterType((*TxSummary)(nil), "pb.TxSummary")
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// Client API for API service

type APIClient interface {
	Publish(ctx context.Context, opts ...grpc.CallOption) (API_PublishClient, error)
}

type aPIClient struct {
	cc *grpc.ClientConn
}

func NewAPIClient(cc *grpc.ClientConn) APIClient {
	return &aPIClient{cc}
}

func (c *aPIClient) Publish(ctx context.Context, opts ...grpc.CallOption) (API_PublishClient, error) {
	stream, err := grpc.NewClientStream(ctx, &_API_serviceDesc.Streams[0], c.cc, "/pb.API/Publish", opts...)
	if err != nil {
		return nil, err
	}
	x := &aPIPublishClient{stream}
	return x, nil
}

type API_PublishClient interface {
	Send(*N3Message) error
	CloseAndRecv() (*TxSummary, error)
	grpc.ClientStream
}

type aPIPublishClient struct {
	grpc.ClientStream
}

func (x *aPIPublishClient) Send(m *N3Message) error {
	return x.ClientStream.SendMsg(m)
}

func (x *aPIPublishClient) CloseAndRecv() (*TxSummary, error) {
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	m := new(TxSummary)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// Server API for API service

type APIServer interface {
	Publish(API_PublishServer) error
}

func RegisterAPIServer(s *grpc.Server, srv APIServer) {
	s.RegisterService(&_API_serviceDesc, srv)
}

func _API_Publish_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(APIServer).Publish(&aPIPublishServer{stream})
}

type API_PublishServer interface {
	SendAndClose(*TxSummary) error
	Recv() (*N3Message, error)
	grpc.ServerStream
}

type aPIPublishServer struct {
	grpc.ServerStream
}

func (x *aPIPublishServer) SendAndClose(m *TxSummary) error {
	return x.ServerStream.SendMsg(m)
}

func (x *aPIPublishServer) Recv() (*N3Message, error) {
	m := new(N3Message)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

var _API_serviceDesc = grpc.ServiceDesc{
	ServiceName: "pb.API",
	HandlerType: (*APIServer)(nil),
	Methods:     []grpc.MethodDesc{},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "Publish",
			Handler:       _API_Publish_Handler,
			ClientStreams: true,
		},
	},
	Metadata: "n3msg.proto",
}

func init() { proto.RegisterFile("n3msg.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 266 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x44, 0x90, 0xbd, 0x6e, 0xb3, 0x30,
	0x14, 0x86, 0x3f, 0xc2, 0x97, 0x1f, 0x4e, 0xdb, 0xc5, 0xaa, 0x2a, 0x14, 0x75, 0x88, 0x58, 0x8a,
	0x54, 0x89, 0x21, 0x5c, 0x41, 0x95, 0x2e, 0x0c, 0x49, 0x10, 0x44, 0xdd, 0x6d, 0x6c, 0x51, 0x2a,
	0xc0, 0x16, 0xb6, 0xa5, 0xe4, 0x1a, 0x7a, 0xd3, 0x95, 0x6d, 0x20, 0xe3, 0x73, 0x8e, 0xfc, 0x3e,
	0xaf, 0x0f, 0x3c, 0xf4, 0x69, 0x27, 0xeb, 0x44, 0x0c, 0x5c, 0x71, 0xb4, 0x10, 0x24, 0x52, 0xb0,
	0x29, 0xf3, 0xf3, 0x45, 0x8b, 0x96, 0xa1, 0x10, 0xd6, 0xa5, 0x26, 0x3f, 0xac, 0x52, 0xa1, 0xb7,
	0xf3, 0xe2, 0xa0, 0x98, 0x10, 0xbd, 0x42, 0x90, 0x0f, 0x8c, 0x36, 0x15, 0x56, 0x2c, 0x5c, 0xd8,
	0xdd, 0x7d, 0x80, 0x5e, 0x60, 0x75, 0x76, 0xcf, 0x7c, 0xbb, 0x1a, 0xc9, 0xe4, 0x7d, 0xb1, 0x41,
	0x36, 0xbc, 0x0f, 0xff, 0xef, 0xbc, 0xd8, 0x2f, 0x26, 0x8c, 0x7e, 0x3d, 0x08, 0x4e, 0xe9, 0x91,
	0x49, 0x89, 0x6b, 0xeb, 0xcd, 0xf1, 0xad, 0xe5, 0x98, 0x5a, 0xef, 0x63, 0x31, 0x21, 0x7a, 0x86,
	0x65, 0xd9, 0xd3, 0x8c, 0x8e, 0x4e, 0x07, 0xa6, 0xcd, 0x09, 0x77, 0xac, 0x14, 0xb8, 0x62, 0xa3,
	0xf2, 0x3e, 0x30, 0x69, 0x07, 0x75, 0x35, 0x6c, 0xad, 0x41, 0x31, 0xa1, 0xe9, 0xf9, 0xd9, 0x48,
	0x91, 0xd1, 0x70, 0xe9, 0x7a, 0x3a, 0x8a, 0xde, 0x20, 0xb8, 0x5c, 0x4b, 0xdd, 0x75, 0x78, 0xb8,
	0xa1, 0x2d, 0x6c, 0x8e, 0xb2, 0x3e, 0x70, 0xdd, 0xbb, 0x2b, 0xf8, 0xc5, 0xcc, 0xfb, 0x3d, 0xf8,
	0x1f, 0x79, 0x86, 0xde, 0x61, 0x9d, 0x6b, 0xd2, 0x36, 0xf2, 0x1b, 0x3d, 0x25, 0x82, 0x24, 0xf3,
	0x4f, 0xb6, 0x16, 0xe7, 0xac, 0xe8, 0x5f, 0xec, 0x91, 0x95, 0xbd, 0x75, 0xfa, 0x17, 0x00, 0x00,
	0xff, 0xff, 0x9b, 0x8e, 0x25, 0x73, 0x7a, 0x01, 0x00, 0x00,
}
