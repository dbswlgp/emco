// Code generated by protoc-gen-go. DO NOT EDIT.
// source: migrationcontroller.proto

/*
Package migrationcontroller is a generated protocol buffer package.

It is generated from these files:
	migrationcontroller.proto

It has these top-level messages:
	MigrationRequest
	MigrationResponse
*/
package migrationcontroller

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

type MigrationRequest struct {
	AppContext string `protobuf:"bytes,1,opt,name=appContext" json:"appContext,omitempty"`
	App        string `protobuf:"bytes,2,opt,name=app" json:"app,omitempty"`
}

func (m *MigrationRequest) Reset()                    { *m = MigrationRequest{} }
func (m *MigrationRequest) String() string            { return proto.CompactTextString(m) }
func (*MigrationRequest) ProtoMessage()               {}
func (*MigrationRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

func (m *MigrationRequest) GetAppContext() string {
	if m != nil {
		return m.AppContext
	}
	return ""
}

func (m *MigrationRequest) GetApp() string {
	if m != nil {
		return m.App
	}
	return ""
}

type MigrationResponse struct {
	AppContext string `protobuf:"bytes,1,opt,name=appContext" json:"appContext,omitempty"`
	Status     bool   `protobuf:"varint,2,opt,name=status" json:"status,omitempty"`
	Message    string `protobuf:"bytes,3,opt,name=message" json:"message,omitempty"`
}

func (m *MigrationResponse) Reset()                    { *m = MigrationResponse{} }
func (m *MigrationResponse) String() string            { return proto.CompactTextString(m) }
func (*MigrationResponse) ProtoMessage()               {}
func (*MigrationResponse) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

func (m *MigrationResponse) GetAppContext() string {
	if m != nil {
		return m.AppContext
	}
	return ""
}

func (m *MigrationResponse) GetStatus() bool {
	if m != nil {
		return m.Status
	}
	return false
}

func (m *MigrationResponse) GetMessage() string {
	if m != nil {
		return m.Message
	}
	return ""
}

func init() {
	proto.RegisterType((*MigrationRequest)(nil), "MigrationRequest")
	proto.RegisterType((*MigrationResponse)(nil), "MigrationResponse")
}










// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// MigrationControllerClient is the client API for MigrationController service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type MigrationControllerClient interface {
        MigrationApps(ctx context.Context, in *MigrationRequest, opts ...grpc.CallOption) (*MigrationResponse, error)
}

type migrationControllerClient struct {
        cc *grpc.ClientConn
}

func NewMigrationControllerClient(cc *grpc.ClientConn) MigrationControllerClient {
        return &migrationControllerClient{cc}
}

func (c *migrationControllerClient) MigrationApps(ctx context.Context, in *MigrationRequest, opts ...grpc.CallOption) (*MigrationResponse, error) {
        out := new(MigrationResponse)
        err := c.cc.Invoke(ctx, "/MigrationController/MigrationApps", in, out, opts...)
        if err != nil {
                return nil, err
        }
        return out, nil
}

// MigrationControllerServer is the server API for MigrationController service.
type MigrationControllerServer interface {
        MigrationApps(context.Context, *MigrationRequest) (*MigrationResponse, error)
}

func RegisterMigrationControllerServer(s *grpc.Server, srv MigrationControllerServer) {
        s.RegisterService(&_MigrationController_serviceDesc, srv)
}

func _MigrationController_MigrationApps_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
        in := new(MigrationRequest)
        if err := dec(in); err != nil {
                return nil, err
        }
        if interceptor == nil {
                return srv.(MigrationControllerServer).MigrationApps(ctx, in)
        }
        info := &grpc.UnaryServerInfo{
                Server:     srv,
                FullMethod: "/MigrationController/MigrationApps",
        }
        handler := func(ctx context.Context, req interface{}) (interface{}, error) {
                return srv.(MigrationControllerServer).MigrationApps(ctx, req.(*MigrationRequest))
        }
        return interceptor(ctx, in, info, handler)
}

var _MigrationController_serviceDesc = grpc.ServiceDesc{
        ServiceName: "MigrationController",
        HandlerType: (*MigrationControllerServer)(nil),
        Methods: []grpc.MethodDesc{
                {
                        MethodName: "MigrationApps",
                        Handler:    _MigrationController_MigrationApps_Handler,
                },
        },
        Streams:  []grpc.StreamDesc{},
        Metadata: "migrationcontroller.proto",
}












func init() { proto.RegisterFile("migrationcontroller.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 184 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0x92, 0xcc, 0xcd, 0x4c, 0x2f,
	0x4a, 0x2c, 0xc9, 0xcc, 0xcf, 0x4b, 0xce, 0xcf, 0x2b, 0x29, 0xca, 0xcf, 0xc9, 0x49, 0x2d, 0xd2,
	0x2b, 0x28, 0xca, 0x2f, 0xc9, 0x57, 0x72, 0xe1, 0x12, 0xf0, 0x85, 0x49, 0x06, 0xa5, 0x16, 0x96,
	0xa6, 0x16, 0x97, 0x08, 0xc9, 0x71, 0x71, 0x25, 0x16, 0x14, 0x38, 0xe7, 0xe7, 0x95, 0xa4, 0x56,
	0x94, 0x48, 0x30, 0x2a, 0x30, 0x6a, 0x70, 0x06, 0x21, 0x89, 0x08, 0x09, 0x70, 0x31, 0x27, 0x16,
	0x14, 0x48, 0x30, 0x81, 0x25, 0x40, 0x4c, 0xa5, 0x54, 0x2e, 0x41, 0x24, 0x53, 0x8a, 0x0b, 0xf2,
	0xf3, 0x8a, 0x53, 0x09, 0x1a, 0x23, 0xc6, 0xc5, 0x56, 0x5c, 0x92, 0x58, 0x52, 0x5a, 0x0c, 0x36,
	0x89, 0x23, 0x08, 0xca, 0x13, 0x92, 0xe0, 0x62, 0xcf, 0x4d, 0x2d, 0x2e, 0x4e, 0x4c, 0x4f, 0x95,
	0x60, 0x06, 0x6b, 0x82, 0x71, 0x8d, 0xfc, 0xb9, 0x84, 0xe1, 0xd6, 0x38, 0xc3, 0x7d, 0x22, 0x64,
	0xc1, 0xc5, 0x0b, 0x17, 0x76, 0x2c, 0x28, 0x28, 0x16, 0x12, 0xd4, 0x43, 0xf7, 0x93, 0x94, 0x90,
	0x1e, 0x86, 0x03, 0x95, 0x18, 0x92, 0xd8, 0xc0, 0x81, 0x60, 0x0c, 0x08, 0x00, 0x00, 0xff, 0xff,
	0x25, 0x98, 0x90, 0x9f, 0x21, 0x01, 0x00, 0x00,
}
