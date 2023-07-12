// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v4.23.3
// source: rabbitMQ.proto

package proto

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// SideCarClient is the client API for SideCar service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type SideCarClient interface {
	InitRabbitMq(ctx context.Context, in *ServiceRequest, opts ...grpc.CallOption) (*emptypb.Empty, error)
	Consume(ctx context.Context, in *ConsumeRequest, opts ...grpc.CallOption) (SideCar_ConsumeClient, error)
	SendRequestApproval(ctx context.Context, in *RequestApproval, opts ...grpc.CallOption) (*emptypb.Empty, error)
	SendValidationResponse(ctx context.Context, in *ValidationResponse, opts ...grpc.CallOption) (*emptypb.Empty, error)
	SendCompositionRequest(ctx context.Context, in *CompositionRequest, opts ...grpc.CallOption) (*emptypb.Empty, error)
	SendSqlDataRequest(ctx context.Context, in *SqlDataRequest, opts ...grpc.CallOption) (*emptypb.Empty, error)
	// rpc SendSqlDataRequestResponse(SqlDataRequestResponse) returns  (google.protobuf.Empty) {}
	SendTest(ctx context.Context, in *SqlDataRequest, opts ...grpc.CallOption) (*emptypb.Empty, error)
	SendMicroserviceComm(ctx context.Context, in *MicroserviceCommunication, opts ...grpc.CallOption) (*emptypb.Empty, error)
}

type sideCarClient struct {
	cc grpc.ClientConnInterface
}

func NewSideCarClient(cc grpc.ClientConnInterface) SideCarClient {
	return &sideCarClient{cc}
}

func (c *sideCarClient) InitRabbitMq(ctx context.Context, in *ServiceRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, "/proto.SideCar/InitRabbitMq", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *sideCarClient) Consume(ctx context.Context, in *ConsumeRequest, opts ...grpc.CallOption) (SideCar_ConsumeClient, error) {
	stream, err := c.cc.NewStream(ctx, &SideCar_ServiceDesc.Streams[0], "/proto.SideCar/Consume", opts...)
	if err != nil {
		return nil, err
	}
	x := &sideCarConsumeClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type SideCar_ConsumeClient interface {
	Recv() (*RabbitMQMessage, error)
	grpc.ClientStream
}

type sideCarConsumeClient struct {
	grpc.ClientStream
}

func (x *sideCarConsumeClient) Recv() (*RabbitMQMessage, error) {
	m := new(RabbitMQMessage)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *sideCarClient) SendRequestApproval(ctx context.Context, in *RequestApproval, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, "/proto.SideCar/SendRequestApproval", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *sideCarClient) SendValidationResponse(ctx context.Context, in *ValidationResponse, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, "/proto.SideCar/SendValidationResponse", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *sideCarClient) SendCompositionRequest(ctx context.Context, in *CompositionRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, "/proto.SideCar/SendCompositionRequest", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *sideCarClient) SendSqlDataRequest(ctx context.Context, in *SqlDataRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, "/proto.SideCar/SendSqlDataRequest", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *sideCarClient) SendTest(ctx context.Context, in *SqlDataRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, "/proto.SideCar/SendTest", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *sideCarClient) SendMicroserviceComm(ctx context.Context, in *MicroserviceCommunication, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, "/proto.SideCar/SendMicroserviceComm", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// SideCarServer is the server API for SideCar service.
// All implementations must embed UnimplementedSideCarServer
// for forward compatibility
type SideCarServer interface {
	InitRabbitMq(context.Context, *ServiceRequest) (*emptypb.Empty, error)
	Consume(*ConsumeRequest, SideCar_ConsumeServer) error
	SendRequestApproval(context.Context, *RequestApproval) (*emptypb.Empty, error)
	SendValidationResponse(context.Context, *ValidationResponse) (*emptypb.Empty, error)
	SendCompositionRequest(context.Context, *CompositionRequest) (*emptypb.Empty, error)
	SendSqlDataRequest(context.Context, *SqlDataRequest) (*emptypb.Empty, error)
	// rpc SendSqlDataRequestResponse(SqlDataRequestResponse) returns  (google.protobuf.Empty) {}
	SendTest(context.Context, *SqlDataRequest) (*emptypb.Empty, error)
	SendMicroserviceComm(context.Context, *MicroserviceCommunication) (*emptypb.Empty, error)
	mustEmbedUnimplementedSideCarServer()
}

// UnimplementedSideCarServer must be embedded to have forward compatible implementations.
type UnimplementedSideCarServer struct {
}

func (UnimplementedSideCarServer) InitRabbitMq(context.Context, *ServiceRequest) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method InitRabbitMq not implemented")
}
func (UnimplementedSideCarServer) Consume(*ConsumeRequest, SideCar_ConsumeServer) error {
	return status.Errorf(codes.Unimplemented, "method Consume not implemented")
}
func (UnimplementedSideCarServer) SendRequestApproval(context.Context, *RequestApproval) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SendRequestApproval not implemented")
}
func (UnimplementedSideCarServer) SendValidationResponse(context.Context, *ValidationResponse) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SendValidationResponse not implemented")
}
func (UnimplementedSideCarServer) SendCompositionRequest(context.Context, *CompositionRequest) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SendCompositionRequest not implemented")
}
func (UnimplementedSideCarServer) SendSqlDataRequest(context.Context, *SqlDataRequest) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SendSqlDataRequest not implemented")
}
func (UnimplementedSideCarServer) SendTest(context.Context, *SqlDataRequest) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SendTest not implemented")
}
func (UnimplementedSideCarServer) SendMicroserviceComm(context.Context, *MicroserviceCommunication) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SendMicroserviceComm not implemented")
}
func (UnimplementedSideCarServer) mustEmbedUnimplementedSideCarServer() {}

// UnsafeSideCarServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to SideCarServer will
// result in compilation errors.
type UnsafeSideCarServer interface {
	mustEmbedUnimplementedSideCarServer()
}

func RegisterSideCarServer(s grpc.ServiceRegistrar, srv SideCarServer) {
	s.RegisterService(&SideCar_ServiceDesc, srv)
}

func _SideCar_InitRabbitMq_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ServiceRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SideCarServer).InitRabbitMq(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.SideCar/InitRabbitMq",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SideCarServer).InitRabbitMq(ctx, req.(*ServiceRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _SideCar_Consume_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(ConsumeRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(SideCarServer).Consume(m, &sideCarConsumeServer{stream})
}

type SideCar_ConsumeServer interface {
	Send(*RabbitMQMessage) error
	grpc.ServerStream
}

type sideCarConsumeServer struct {
	grpc.ServerStream
}

func (x *sideCarConsumeServer) Send(m *RabbitMQMessage) error {
	return x.ServerStream.SendMsg(m)
}

func _SideCar_SendRequestApproval_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RequestApproval)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SideCarServer).SendRequestApproval(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.SideCar/SendRequestApproval",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SideCarServer).SendRequestApproval(ctx, req.(*RequestApproval))
	}
	return interceptor(ctx, in, info, handler)
}

func _SideCar_SendValidationResponse_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ValidationResponse)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SideCarServer).SendValidationResponse(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.SideCar/SendValidationResponse",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SideCarServer).SendValidationResponse(ctx, req.(*ValidationResponse))
	}
	return interceptor(ctx, in, info, handler)
}

func _SideCar_SendCompositionRequest_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CompositionRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SideCarServer).SendCompositionRequest(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.SideCar/SendCompositionRequest",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SideCarServer).SendCompositionRequest(ctx, req.(*CompositionRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _SideCar_SendSqlDataRequest_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SqlDataRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SideCarServer).SendSqlDataRequest(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.SideCar/SendSqlDataRequest",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SideCarServer).SendSqlDataRequest(ctx, req.(*SqlDataRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _SideCar_SendTest_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SqlDataRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SideCarServer).SendTest(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.SideCar/SendTest",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SideCarServer).SendTest(ctx, req.(*SqlDataRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _SideCar_SendMicroserviceComm_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MicroserviceCommunication)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SideCarServer).SendMicroserviceComm(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.SideCar/SendMicroserviceComm",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SideCarServer).SendMicroserviceComm(ctx, req.(*MicroserviceCommunication))
	}
	return interceptor(ctx, in, info, handler)
}

// SideCar_ServiceDesc is the grpc.ServiceDesc for SideCar service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var SideCar_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "proto.SideCar",
	HandlerType: (*SideCarServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "InitRabbitMq",
			Handler:    _SideCar_InitRabbitMq_Handler,
		},
		{
			MethodName: "SendRequestApproval",
			Handler:    _SideCar_SendRequestApproval_Handler,
		},
		{
			MethodName: "SendValidationResponse",
			Handler:    _SideCar_SendValidationResponse_Handler,
		},
		{
			MethodName: "SendCompositionRequest",
			Handler:    _SideCar_SendCompositionRequest_Handler,
		},
		{
			MethodName: "SendSqlDataRequest",
			Handler:    _SideCar_SendSqlDataRequest_Handler,
		},
		{
			MethodName: "SendTest",
			Handler:    _SideCar_SendTest_Handler,
		},
		{
			MethodName: "SendMicroserviceComm",
			Handler:    _SideCar_SendMicroserviceComm_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "Consume",
			Handler:       _SideCar_Consume_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "rabbitMQ.proto",
}
