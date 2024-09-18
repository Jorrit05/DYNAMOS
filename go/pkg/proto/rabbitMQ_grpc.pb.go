// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.5.1
// - protoc             v3.12.4
// source: rabbitMQ.proto

package proto

import (
	context "context"
	empty "github.com/golang/protobuf/ptypes/empty"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.64.0 or later.
const _ = grpc.SupportPackageIsVersion9

const (
	RabbitMQ_InitRabbitMq_FullMethodName                = "/dynamos.RabbitMQ/InitRabbitMq"
	RabbitMQ_InitRabbitForChain_FullMethodName          = "/dynamos.RabbitMQ/InitRabbitForChain"
	RabbitMQ_StopReceivingRabbit_FullMethodName         = "/dynamos.RabbitMQ/StopReceivingRabbit"
	RabbitMQ_Consume_FullMethodName                     = "/dynamos.RabbitMQ/Consume"
	RabbitMQ_ChainConsume_FullMethodName                = "/dynamos.RabbitMQ/ChainConsume"
	RabbitMQ_SendRequestApproval_FullMethodName         = "/dynamos.RabbitMQ/SendRequestApproval"
	RabbitMQ_SendValidationResponse_FullMethodName      = "/dynamos.RabbitMQ/SendValidationResponse"
	RabbitMQ_SendCompositionRequest_FullMethodName      = "/dynamos.RabbitMQ/SendCompositionRequest"
	RabbitMQ_SendSqlDataRequest_FullMethodName          = "/dynamos.RabbitMQ/SendSqlDataRequest"
	RabbitMQ_SendPolicyUpdate_FullMethodName            = "/dynamos.RabbitMQ/SendPolicyUpdate"
	RabbitMQ_SendTest_FullMethodName                    = "/dynamos.RabbitMQ/SendTest"
	RabbitMQ_SendMicroserviceComm_FullMethodName        = "/dynamos.RabbitMQ/SendMicroserviceComm"
	RabbitMQ_CreateQueue_FullMethodName                 = "/dynamos.RabbitMQ/CreateQueue"
	RabbitMQ_DeleteQueue_FullMethodName                 = "/dynamos.RabbitMQ/DeleteQueue"
	RabbitMQ_SendRequestApprovalResponse_FullMethodName = "/dynamos.RabbitMQ/SendRequestApprovalResponse"
	RabbitMQ_SendRequestApprovalRequest_FullMethodName  = "/dynamos.RabbitMQ/SendRequestApprovalRequest"
)

// RabbitMQClient is the client API for RabbitMQ service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
//
// RPC calls for RabbitMQ.
type RabbitMQClient interface {
	InitRabbitMq(ctx context.Context, in *InitRequest, opts ...grpc.CallOption) (*empty.Empty, error)
	InitRabbitForChain(ctx context.Context, in *ChainRequest, opts ...grpc.CallOption) (*empty.Empty, error)
	StopReceivingRabbit(ctx context.Context, in *StopRequest, opts ...grpc.CallOption) (*empty.Empty, error)
	Consume(ctx context.Context, in *ConsumeRequest, opts ...grpc.CallOption) (grpc.ServerStreamingClient[SideCarMessage], error)
	ChainConsume(ctx context.Context, in *ConsumeRequest, opts ...grpc.CallOption) (grpc.ServerStreamingClient[SideCarMessage], error)
	SendRequestApproval(ctx context.Context, in *RequestApproval, opts ...grpc.CallOption) (*empty.Empty, error)
	SendValidationResponse(ctx context.Context, in *ValidationResponse, opts ...grpc.CallOption) (*empty.Empty, error)
	SendCompositionRequest(ctx context.Context, in *CompositionRequest, opts ...grpc.CallOption) (*empty.Empty, error)
	SendSqlDataRequest(ctx context.Context, in *SqlDataRequest, opts ...grpc.CallOption) (*empty.Empty, error)
	SendPolicyUpdate(ctx context.Context, in *PolicyUpdate, opts ...grpc.CallOption) (*empty.Empty, error)
	SendTest(ctx context.Context, in *SqlDataRequest, opts ...grpc.CallOption) (*empty.Empty, error)
	SendMicroserviceComm(ctx context.Context, in *MicroserviceCommunication, opts ...grpc.CallOption) (*empty.Empty, error)
	CreateQueue(ctx context.Context, in *QueueInfo, opts ...grpc.CallOption) (*empty.Empty, error)
	DeleteQueue(ctx context.Context, in *QueueInfo, opts ...grpc.CallOption) (*empty.Empty, error)
	SendRequestApprovalResponse(ctx context.Context, in *RequestApprovalResponse, opts ...grpc.CallOption) (*empty.Empty, error)
	SendRequestApprovalRequest(ctx context.Context, in *RequestApproval, opts ...grpc.CallOption) (*empty.Empty, error)
}

type rabbitMQClient struct {
	cc grpc.ClientConnInterface
}

func NewRabbitMQClient(cc grpc.ClientConnInterface) RabbitMQClient {
	return &rabbitMQClient{cc}
}

func (c *rabbitMQClient) InitRabbitMq(ctx context.Context, in *InitRequest, opts ...grpc.CallOption) (*empty.Empty, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(empty.Empty)
	err := c.cc.Invoke(ctx, RabbitMQ_InitRabbitMq_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *rabbitMQClient) InitRabbitForChain(ctx context.Context, in *ChainRequest, opts ...grpc.CallOption) (*empty.Empty, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(empty.Empty)
	err := c.cc.Invoke(ctx, RabbitMQ_InitRabbitForChain_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *rabbitMQClient) StopReceivingRabbit(ctx context.Context, in *StopRequest, opts ...grpc.CallOption) (*empty.Empty, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(empty.Empty)
	err := c.cc.Invoke(ctx, RabbitMQ_StopReceivingRabbit_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *rabbitMQClient) Consume(ctx context.Context, in *ConsumeRequest, opts ...grpc.CallOption) (grpc.ServerStreamingClient[SideCarMessage], error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	stream, err := c.cc.NewStream(ctx, &RabbitMQ_ServiceDesc.Streams[0], RabbitMQ_Consume_FullMethodName, cOpts...)
	if err != nil {
		return nil, err
	}
	x := &grpc.GenericClientStream[ConsumeRequest, SideCarMessage]{ClientStream: stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

// This type alias is provided for backwards compatibility with existing code that references the prior non-generic stream type by name.
type RabbitMQ_ConsumeClient = grpc.ServerStreamingClient[SideCarMessage]

func (c *rabbitMQClient) ChainConsume(ctx context.Context, in *ConsumeRequest, opts ...grpc.CallOption) (grpc.ServerStreamingClient[SideCarMessage], error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	stream, err := c.cc.NewStream(ctx, &RabbitMQ_ServiceDesc.Streams[1], RabbitMQ_ChainConsume_FullMethodName, cOpts...)
	if err != nil {
		return nil, err
	}
	x := &grpc.GenericClientStream[ConsumeRequest, SideCarMessage]{ClientStream: stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

// This type alias is provided for backwards compatibility with existing code that references the prior non-generic stream type by name.
type RabbitMQ_ChainConsumeClient = grpc.ServerStreamingClient[SideCarMessage]

func (c *rabbitMQClient) SendRequestApproval(ctx context.Context, in *RequestApproval, opts ...grpc.CallOption) (*empty.Empty, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(empty.Empty)
	err := c.cc.Invoke(ctx, RabbitMQ_SendRequestApproval_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *rabbitMQClient) SendValidationResponse(ctx context.Context, in *ValidationResponse, opts ...grpc.CallOption) (*empty.Empty, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(empty.Empty)
	err := c.cc.Invoke(ctx, RabbitMQ_SendValidationResponse_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *rabbitMQClient) SendCompositionRequest(ctx context.Context, in *CompositionRequest, opts ...grpc.CallOption) (*empty.Empty, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(empty.Empty)
	err := c.cc.Invoke(ctx, RabbitMQ_SendCompositionRequest_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *rabbitMQClient) SendSqlDataRequest(ctx context.Context, in *SqlDataRequest, opts ...grpc.CallOption) (*empty.Empty, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(empty.Empty)
	err := c.cc.Invoke(ctx, RabbitMQ_SendSqlDataRequest_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *rabbitMQClient) SendPolicyUpdate(ctx context.Context, in *PolicyUpdate, opts ...grpc.CallOption) (*empty.Empty, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(empty.Empty)
	err := c.cc.Invoke(ctx, RabbitMQ_SendPolicyUpdate_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *rabbitMQClient) SendTest(ctx context.Context, in *SqlDataRequest, opts ...grpc.CallOption) (*empty.Empty, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(empty.Empty)
	err := c.cc.Invoke(ctx, RabbitMQ_SendTest_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *rabbitMQClient) SendMicroserviceComm(ctx context.Context, in *MicroserviceCommunication, opts ...grpc.CallOption) (*empty.Empty, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(empty.Empty)
	err := c.cc.Invoke(ctx, RabbitMQ_SendMicroserviceComm_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *rabbitMQClient) CreateQueue(ctx context.Context, in *QueueInfo, opts ...grpc.CallOption) (*empty.Empty, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(empty.Empty)
	err := c.cc.Invoke(ctx, RabbitMQ_CreateQueue_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *rabbitMQClient) DeleteQueue(ctx context.Context, in *QueueInfo, opts ...grpc.CallOption) (*empty.Empty, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(empty.Empty)
	err := c.cc.Invoke(ctx, RabbitMQ_DeleteQueue_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *rabbitMQClient) SendRequestApprovalResponse(ctx context.Context, in *RequestApprovalResponse, opts ...grpc.CallOption) (*empty.Empty, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(empty.Empty)
	err := c.cc.Invoke(ctx, RabbitMQ_SendRequestApprovalResponse_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *rabbitMQClient) SendRequestApprovalRequest(ctx context.Context, in *RequestApproval, opts ...grpc.CallOption) (*empty.Empty, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(empty.Empty)
	err := c.cc.Invoke(ctx, RabbitMQ_SendRequestApprovalRequest_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// RabbitMQServer is the server API for RabbitMQ service.
// All implementations must embed UnimplementedRabbitMQServer
// for forward compatibility.
//
// RPC calls for RabbitMQ.
type RabbitMQServer interface {
	InitRabbitMq(context.Context, *InitRequest) (*empty.Empty, error)
	InitRabbitForChain(context.Context, *ChainRequest) (*empty.Empty, error)
	StopReceivingRabbit(context.Context, *StopRequest) (*empty.Empty, error)
	Consume(*ConsumeRequest, grpc.ServerStreamingServer[SideCarMessage]) error
	ChainConsume(*ConsumeRequest, grpc.ServerStreamingServer[SideCarMessage]) error
	SendRequestApproval(context.Context, *RequestApproval) (*empty.Empty, error)
	SendValidationResponse(context.Context, *ValidationResponse) (*empty.Empty, error)
	SendCompositionRequest(context.Context, *CompositionRequest) (*empty.Empty, error)
	SendSqlDataRequest(context.Context, *SqlDataRequest) (*empty.Empty, error)
	SendPolicyUpdate(context.Context, *PolicyUpdate) (*empty.Empty, error)
	SendTest(context.Context, *SqlDataRequest) (*empty.Empty, error)
	SendMicroserviceComm(context.Context, *MicroserviceCommunication) (*empty.Empty, error)
	CreateQueue(context.Context, *QueueInfo) (*empty.Empty, error)
	DeleteQueue(context.Context, *QueueInfo) (*empty.Empty, error)
	SendRequestApprovalResponse(context.Context, *RequestApprovalResponse) (*empty.Empty, error)
	SendRequestApprovalRequest(context.Context, *RequestApproval) (*empty.Empty, error)
	mustEmbedUnimplementedRabbitMQServer()
}

// UnimplementedRabbitMQServer must be embedded to have
// forward compatible implementations.
//
// NOTE: this should be embedded by value instead of pointer to avoid a nil
// pointer dereference when methods are called.
type UnimplementedRabbitMQServer struct{}

func (UnimplementedRabbitMQServer) InitRabbitMq(context.Context, *InitRequest) (*empty.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method InitRabbitMq not implemented")
}
func (UnimplementedRabbitMQServer) InitRabbitForChain(context.Context, *ChainRequest) (*empty.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method InitRabbitForChain not implemented")
}
func (UnimplementedRabbitMQServer) StopReceivingRabbit(context.Context, *StopRequest) (*empty.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method StopReceivingRabbit not implemented")
}
func (UnimplementedRabbitMQServer) Consume(*ConsumeRequest, grpc.ServerStreamingServer[SideCarMessage]) error {
	return status.Errorf(codes.Unimplemented, "method Consume not implemented")
}
func (UnimplementedRabbitMQServer) ChainConsume(*ConsumeRequest, grpc.ServerStreamingServer[SideCarMessage]) error {
	return status.Errorf(codes.Unimplemented, "method ChainConsume not implemented")
}
func (UnimplementedRabbitMQServer) SendRequestApproval(context.Context, *RequestApproval) (*empty.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SendRequestApproval not implemented")
}
func (UnimplementedRabbitMQServer) SendValidationResponse(context.Context, *ValidationResponse) (*empty.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SendValidationResponse not implemented")
}
func (UnimplementedRabbitMQServer) SendCompositionRequest(context.Context, *CompositionRequest) (*empty.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SendCompositionRequest not implemented")
}
func (UnimplementedRabbitMQServer) SendSqlDataRequest(context.Context, *SqlDataRequest) (*empty.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SendSqlDataRequest not implemented")
}
func (UnimplementedRabbitMQServer) SendPolicyUpdate(context.Context, *PolicyUpdate) (*empty.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SendPolicyUpdate not implemented")
}
func (UnimplementedRabbitMQServer) SendTest(context.Context, *SqlDataRequest) (*empty.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SendTest not implemented")
}
func (UnimplementedRabbitMQServer) SendMicroserviceComm(context.Context, *MicroserviceCommunication) (*empty.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SendMicroserviceComm not implemented")
}
func (UnimplementedRabbitMQServer) CreateQueue(context.Context, *QueueInfo) (*empty.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateQueue not implemented")
}
func (UnimplementedRabbitMQServer) DeleteQueue(context.Context, *QueueInfo) (*empty.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteQueue not implemented")
}
func (UnimplementedRabbitMQServer) SendRequestApprovalResponse(context.Context, *RequestApprovalResponse) (*empty.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SendRequestApprovalResponse not implemented")
}
func (UnimplementedRabbitMQServer) SendRequestApprovalRequest(context.Context, *RequestApproval) (*empty.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SendRequestApprovalRequest not implemented")
}
func (UnimplementedRabbitMQServer) mustEmbedUnimplementedRabbitMQServer() {}
func (UnimplementedRabbitMQServer) testEmbeddedByValue()                  {}

// UnsafeRabbitMQServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to RabbitMQServer will
// result in compilation errors.
type UnsafeRabbitMQServer interface {
	mustEmbedUnimplementedRabbitMQServer()
}

func RegisterRabbitMQServer(s grpc.ServiceRegistrar, srv RabbitMQServer) {
	// If the following call pancis, it indicates UnimplementedRabbitMQServer was
	// embedded by pointer and is nil.  This will cause panics if an
	// unimplemented method is ever invoked, so we test this at initialization
	// time to prevent it from happening at runtime later due to I/O.
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&RabbitMQ_ServiceDesc, srv)
}

func _RabbitMQ_InitRabbitMq_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(InitRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RabbitMQServer).InitRabbitMq(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: RabbitMQ_InitRabbitMq_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RabbitMQServer).InitRabbitMq(ctx, req.(*InitRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _RabbitMQ_InitRabbitForChain_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ChainRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RabbitMQServer).InitRabbitForChain(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: RabbitMQ_InitRabbitForChain_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RabbitMQServer).InitRabbitForChain(ctx, req.(*ChainRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _RabbitMQ_StopReceivingRabbit_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(StopRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RabbitMQServer).StopReceivingRabbit(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: RabbitMQ_StopReceivingRabbit_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RabbitMQServer).StopReceivingRabbit(ctx, req.(*StopRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _RabbitMQ_Consume_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(ConsumeRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(RabbitMQServer).Consume(m, &grpc.GenericServerStream[ConsumeRequest, SideCarMessage]{ServerStream: stream})
}

// This type alias is provided for backwards compatibility with existing code that references the prior non-generic stream type by name.
type RabbitMQ_ConsumeServer = grpc.ServerStreamingServer[SideCarMessage]

func _RabbitMQ_ChainConsume_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(ConsumeRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(RabbitMQServer).ChainConsume(m, &grpc.GenericServerStream[ConsumeRequest, SideCarMessage]{ServerStream: stream})
}

// This type alias is provided for backwards compatibility with existing code that references the prior non-generic stream type by name.
type RabbitMQ_ChainConsumeServer = grpc.ServerStreamingServer[SideCarMessage]

func _RabbitMQ_SendRequestApproval_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RequestApproval)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RabbitMQServer).SendRequestApproval(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: RabbitMQ_SendRequestApproval_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RabbitMQServer).SendRequestApproval(ctx, req.(*RequestApproval))
	}
	return interceptor(ctx, in, info, handler)
}

func _RabbitMQ_SendValidationResponse_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ValidationResponse)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RabbitMQServer).SendValidationResponse(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: RabbitMQ_SendValidationResponse_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RabbitMQServer).SendValidationResponse(ctx, req.(*ValidationResponse))
	}
	return interceptor(ctx, in, info, handler)
}

func _RabbitMQ_SendCompositionRequest_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CompositionRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RabbitMQServer).SendCompositionRequest(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: RabbitMQ_SendCompositionRequest_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RabbitMQServer).SendCompositionRequest(ctx, req.(*CompositionRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _RabbitMQ_SendSqlDataRequest_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SqlDataRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RabbitMQServer).SendSqlDataRequest(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: RabbitMQ_SendSqlDataRequest_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RabbitMQServer).SendSqlDataRequest(ctx, req.(*SqlDataRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _RabbitMQ_SendPolicyUpdate_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PolicyUpdate)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RabbitMQServer).SendPolicyUpdate(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: RabbitMQ_SendPolicyUpdate_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RabbitMQServer).SendPolicyUpdate(ctx, req.(*PolicyUpdate))
	}
	return interceptor(ctx, in, info, handler)
}

func _RabbitMQ_SendTest_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SqlDataRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RabbitMQServer).SendTest(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: RabbitMQ_SendTest_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RabbitMQServer).SendTest(ctx, req.(*SqlDataRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _RabbitMQ_SendMicroserviceComm_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MicroserviceCommunication)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RabbitMQServer).SendMicroserviceComm(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: RabbitMQ_SendMicroserviceComm_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RabbitMQServer).SendMicroserviceComm(ctx, req.(*MicroserviceCommunication))
	}
	return interceptor(ctx, in, info, handler)
}

func _RabbitMQ_CreateQueue_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QueueInfo)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RabbitMQServer).CreateQueue(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: RabbitMQ_CreateQueue_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RabbitMQServer).CreateQueue(ctx, req.(*QueueInfo))
	}
	return interceptor(ctx, in, info, handler)
}

func _RabbitMQ_DeleteQueue_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QueueInfo)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RabbitMQServer).DeleteQueue(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: RabbitMQ_DeleteQueue_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RabbitMQServer).DeleteQueue(ctx, req.(*QueueInfo))
	}
	return interceptor(ctx, in, info, handler)
}

func _RabbitMQ_SendRequestApprovalResponse_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RequestApprovalResponse)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RabbitMQServer).SendRequestApprovalResponse(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: RabbitMQ_SendRequestApprovalResponse_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RabbitMQServer).SendRequestApprovalResponse(ctx, req.(*RequestApprovalResponse))
	}
	return interceptor(ctx, in, info, handler)
}

func _RabbitMQ_SendRequestApprovalRequest_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RequestApproval)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RabbitMQServer).SendRequestApprovalRequest(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: RabbitMQ_SendRequestApprovalRequest_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RabbitMQServer).SendRequestApprovalRequest(ctx, req.(*RequestApproval))
	}
	return interceptor(ctx, in, info, handler)
}

// RabbitMQ_ServiceDesc is the grpc.ServiceDesc for RabbitMQ service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var RabbitMQ_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "dynamos.RabbitMQ",
	HandlerType: (*RabbitMQServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "InitRabbitMq",
			Handler:    _RabbitMQ_InitRabbitMq_Handler,
		},
		{
			MethodName: "InitRabbitForChain",
			Handler:    _RabbitMQ_InitRabbitForChain_Handler,
		},
		{
			MethodName: "StopReceivingRabbit",
			Handler:    _RabbitMQ_StopReceivingRabbit_Handler,
		},
		{
			MethodName: "SendRequestApproval",
			Handler:    _RabbitMQ_SendRequestApproval_Handler,
		},
		{
			MethodName: "SendValidationResponse",
			Handler:    _RabbitMQ_SendValidationResponse_Handler,
		},
		{
			MethodName: "SendCompositionRequest",
			Handler:    _RabbitMQ_SendCompositionRequest_Handler,
		},
		{
			MethodName: "SendSqlDataRequest",
			Handler:    _RabbitMQ_SendSqlDataRequest_Handler,
		},
		{
			MethodName: "SendPolicyUpdate",
			Handler:    _RabbitMQ_SendPolicyUpdate_Handler,
		},
		{
			MethodName: "SendTest",
			Handler:    _RabbitMQ_SendTest_Handler,
		},
		{
			MethodName: "SendMicroserviceComm",
			Handler:    _RabbitMQ_SendMicroserviceComm_Handler,
		},
		{
			MethodName: "CreateQueue",
			Handler:    _RabbitMQ_CreateQueue_Handler,
		},
		{
			MethodName: "DeleteQueue",
			Handler:    _RabbitMQ_DeleteQueue_Handler,
		},
		{
			MethodName: "SendRequestApprovalResponse",
			Handler:    _RabbitMQ_SendRequestApprovalResponse_Handler,
		},
		{
			MethodName: "SendRequestApprovalRequest",
			Handler:    _RabbitMQ_SendRequestApprovalRequest_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "Consume",
			Handler:       _RabbitMQ_Consume_Handler,
			ServerStreams: true,
		},
		{
			StreamName:    "ChainConsume",
			Handler:       _RabbitMQ_ChainConsume_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "rabbitMQ.proto",
}
