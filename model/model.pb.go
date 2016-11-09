// Code generated by protoc-gen-go.
// source: model.proto
// DO NOT EDIT!

/*
Package model is a generated protocol buffer package.

It is generated from these files:
	model.proto

It has these top-level messages:
	Report
	Browse
*/
package model

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

type Report struct {
	WorkerID string  `protobuf:"bytes,1,opt,name=WorkerID,json=workerID" json:"WorkerID,omitempty"`
	Browse   *Browse `protobuf:"bytes,2,opt,name=Browse,json=browse" json:"Browse,omitempty"`
	Pcap     []byte  `protobuf:"bytes,3,opt,name=Pcap,json=pcap,proto3" json:"Pcap,omitempty"`
	Log      []byte  `protobuf:"bytes,4,opt,name=Log,json=log,proto3" json:"Log,omitempty"`
}

func (m *Report) Reset()                    { *m = Report{} }
func (m *Report) String() string            { return proto.CompactTextString(m) }
func (*Report) ProtoMessage()               {}
func (*Report) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

func (m *Report) GetBrowse() *Browse {
	if m != nil {
		return m.Browse
	}
	return nil
}

// Browse is a work item.
type Browse struct {
	ID      string `protobuf:"bytes,1,opt,name=ID,json=iD" json:"ID,omitempty"`
	BatchID string `protobuf:"bytes,2,opt,name=BatchID,json=batchID" json:"BatchID,omitempty"`
	URL     string `protobuf:"bytes,3,opt,name=URL,json=uRL" json:"URL,omitempty"`
	Torrc   string `protobuf:"bytes,4,opt,name=Torrc,json=torrc" json:"Torrc,omitempty"`
	Log     bool   `protobuf:"varint,5,opt,name=Log,json=log" json:"Log,omitempty"`
	Timeout int64  `protobuf:"varint,6,opt,name=Timeout,json=timeout" json:"Timeout,omitempty"`
}

func (m *Browse) Reset()                    { *m = Browse{} }
func (m *Browse) String() string            { return proto.CompactTextString(m) }
func (*Browse) ProtoMessage()               {}
func (*Browse) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

func init() {
	proto.RegisterType((*Report)(nil), "model.Report")
	proto.RegisterType((*Browse)(nil), "model.Browse")
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion3

// Client API for Gather service

type GatherClient interface {
	Work(ctx context.Context, in *Report, opts ...grpc.CallOption) (*Browse, error)
}

type gatherClient struct {
	cc *grpc.ClientConn
}

func NewGatherClient(cc *grpc.ClientConn) GatherClient {
	return &gatherClient{cc}
}

func (c *gatherClient) Work(ctx context.Context, in *Report, opts ...grpc.CallOption) (*Browse, error) {
	out := new(Browse)
	err := grpc.Invoke(ctx, "/model.Gather/Work", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for Gather service

type GatherServer interface {
	Work(context.Context, *Report) (*Browse, error)
}

func RegisterGatherServer(s *grpc.Server, srv GatherServer) {
	s.RegisterService(&_Gather_serviceDesc, srv)
}

func _Gather_Work_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Report)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GatherServer).Work(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/model.Gather/Work",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GatherServer).Work(ctx, req.(*Report))
	}
	return interceptor(ctx, in, info, handler)
}

var _Gather_serviceDesc = grpc.ServiceDesc{
	ServiceName: "model.Gather",
	HandlerType: (*GatherServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Work",
			Handler:    _Gather_Work_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: fileDescriptor0,
}

func init() { proto.RegisterFile("model.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 250 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x09, 0x6e, 0x88, 0x02, 0xff, 0x54, 0x90, 0xcf, 0x4a, 0xc3, 0x40,
	0x10, 0xc6, 0xdd, 0xfc, 0xd9, 0x24, 0x53, 0x15, 0x19, 0x3c, 0x2c, 0x3d, 0x85, 0x80, 0x92, 0x53,
	0x91, 0xfa, 0x06, 0x25, 0x20, 0x85, 0x1c, 0x64, 0xa8, 0x78, 0x4e, 0xe2, 0xd2, 0x16, 0x5b, 0x66,
	0x5d, 0xb7, 0xf4, 0x0d, 0x7c, 0x6e, 0xc9, 0xae, 0x11, 0x7a, 0x1a, 0x7e, 0x73, 0xf8, 0x7d, 0x7c,
	0x1f, 0xcc, 0x8e, 0xfc, 0xa1, 0x0f, 0x0b, 0x63, 0xd9, 0x31, 0xa6, 0x1e, 0xaa, 0x2f, 0x90, 0xa4,
	0x0d, 0x5b, 0x87, 0x73, 0xc8, 0xdf, 0xd9, 0x7e, 0x6a, 0xbb, 0x6e, 0x94, 0x28, 0x45, 0x5d, 0x50,
	0x7e, 0xfe, 0x63, 0x7c, 0x00, 0xb9, 0xb2, 0x7c, 0xfe, 0xd6, 0x2a, 0x2a, 0x45, 0x3d, 0x5b, 0xde,
	0x2c, 0x82, 0x2a, 0x3c, 0x49, 0xf6, 0xfe, 0x22, 0x42, 0xf2, 0x3a, 0x74, 0x46, 0xc5, 0xa5, 0xa8,
	0xaf, 0x29, 0x31, 0x43, 0x67, 0xf0, 0x0e, 0xe2, 0x96, 0xb7, 0x2a, 0xf1, 0xaf, 0xf8, 0xc0, 0xdb,
	0xea, 0x47, 0x4c, 0x36, 0xbc, 0x85, 0xe8, 0x3f, 0x2d, 0xda, 0x37, 0xa8, 0x20, 0x5b, 0x75, 0x6e,
	0xd8, 0xad, 0x1b, 0x1f, 0x54, 0x50, 0xd6, 0x07, 0x1c, 0x35, 0x6f, 0xd4, 0x7a, 0x73, 0x41, 0xf1,
	0x89, 0x5a, 0xbc, 0x87, 0x74, 0xc3, 0xd6, 0x0e, 0x5e, 0x5d, 0x50, 0xea, 0x46, 0x98, 0xe2, 0xd2,
	0x52, 0xd4, 0xb9, 0x8f, 0x1b, 0x9d, 0x9b, 0xfd, 0x51, 0xf3, 0xc9, 0x29, 0x59, 0x8a, 0x3a, 0xa6,
	0xcc, 0x05, 0x5c, 0x3e, 0x81, 0x7c, 0xe9, 0xdc, 0x4e, 0x5b, 0x7c, 0x84, 0x64, 0xec, 0x8e, 0x53,
	0xaf, 0x30, 0xc9, 0xfc, 0xb2, 0x66, 0x75, 0xd5, 0x4b, 0xbf, 0xdd, 0xf3, 0x6f, 0x00, 0x00, 0x00,
	0xff, 0xff, 0x61, 0x98, 0xa9, 0xa4, 0x4a, 0x01, 0x00, 0x00,
}
