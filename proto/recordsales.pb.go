// Code generated by protoc-gen-go. DO NOT EDIT.
// source: recordsales.proto

package recordsales

import (
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	context "golang.org/x/net/context"
	grpc "google.golang.org/grpc"
	math "math"
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

type Sale struct {
	InstanceId           int32    `protobuf:"varint,1,opt,name=instance_id,json=instanceId,proto3" json:"instance_id,omitempty"`
	LastUpdateTime       int64    `protobuf:"varint,2,opt,name=last_update_time,json=lastUpdateTime,proto3" json:"last_update_time,omitempty"`
	Price                int32    `protobuf:"varint,3,opt,name=price,proto3" json:"price,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Sale) Reset()         { *m = Sale{} }
func (m *Sale) String() string { return proto.CompactTextString(m) }
func (*Sale) ProtoMessage()    {}
func (*Sale) Descriptor() ([]byte, []int) {
	return fileDescriptor_3961a399f18e1e66, []int{0}
}

func (m *Sale) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Sale.Unmarshal(m, b)
}
func (m *Sale) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Sale.Marshal(b, m, deterministic)
}
func (m *Sale) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Sale.Merge(m, src)
}
func (m *Sale) XXX_Size() int {
	return xxx_messageInfo_Sale.Size(m)
}
func (m *Sale) XXX_DiscardUnknown() {
	xxx_messageInfo_Sale.DiscardUnknown(m)
}

var xxx_messageInfo_Sale proto.InternalMessageInfo

func (m *Sale) GetInstanceId() int32 {
	if m != nil {
		return m.InstanceId
	}
	return 0
}

func (m *Sale) GetLastUpdateTime() int64 {
	if m != nil {
		return m.LastUpdateTime
	}
	return 0
}

func (m *Sale) GetPrice() int32 {
	if m != nil {
		return m.Price
	}
	return 0
}

type Config struct {
	Sales                []*Sale  `protobuf:"bytes,1,rep,name=sales,proto3" json:"sales,omitempty"`
	Archives             []*Sale  `protobuf:"bytes,2,rep,name=archives,proto3" json:"archives,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Config) Reset()         { *m = Config{} }
func (m *Config) String() string { return proto.CompactTextString(m) }
func (*Config) ProtoMessage()    {}
func (*Config) Descriptor() ([]byte, []int) {
	return fileDescriptor_3961a399f18e1e66, []int{1}
}

func (m *Config) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Config.Unmarshal(m, b)
}
func (m *Config) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Config.Marshal(b, m, deterministic)
}
func (m *Config) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Config.Merge(m, src)
}
func (m *Config) XXX_Size() int {
	return xxx_messageInfo_Config.Size(m)
}
func (m *Config) XXX_DiscardUnknown() {
	xxx_messageInfo_Config.DiscardUnknown(m)
}

var xxx_messageInfo_Config proto.InternalMessageInfo

func (m *Config) GetSales() []*Sale {
	if m != nil {
		return m.Sales
	}
	return nil
}

func (m *Config) GetArchives() []*Sale {
	if m != nil {
		return m.Archives
	}
	return nil
}

type GetStaleRequest struct {
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *GetStaleRequest) Reset()         { *m = GetStaleRequest{} }
func (m *GetStaleRequest) String() string { return proto.CompactTextString(m) }
func (*GetStaleRequest) ProtoMessage()    {}
func (*GetStaleRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_3961a399f18e1e66, []int{2}
}

func (m *GetStaleRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_GetStaleRequest.Unmarshal(m, b)
}
func (m *GetStaleRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_GetStaleRequest.Marshal(b, m, deterministic)
}
func (m *GetStaleRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GetStaleRequest.Merge(m, src)
}
func (m *GetStaleRequest) XXX_Size() int {
	return xxx_messageInfo_GetStaleRequest.Size(m)
}
func (m *GetStaleRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_GetStaleRequest.DiscardUnknown(m)
}

var xxx_messageInfo_GetStaleRequest proto.InternalMessageInfo

type GetStaleResponse struct {
	StaleSales           []*Sale  `protobuf:"bytes,1,rep,name=stale_sales,json=staleSales,proto3" json:"stale_sales,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *GetStaleResponse) Reset()         { *m = GetStaleResponse{} }
func (m *GetStaleResponse) String() string { return proto.CompactTextString(m) }
func (*GetStaleResponse) ProtoMessage()    {}
func (*GetStaleResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_3961a399f18e1e66, []int{3}
}

func (m *GetStaleResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_GetStaleResponse.Unmarshal(m, b)
}
func (m *GetStaleResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_GetStaleResponse.Marshal(b, m, deterministic)
}
func (m *GetStaleResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GetStaleResponse.Merge(m, src)
}
func (m *GetStaleResponse) XXX_Size() int {
	return xxx_messageInfo_GetStaleResponse.Size(m)
}
func (m *GetStaleResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_GetStaleResponse.DiscardUnknown(m)
}

var xxx_messageInfo_GetStaleResponse proto.InternalMessageInfo

func (m *GetStaleResponse) GetStaleSales() []*Sale {
	if m != nil {
		return m.StaleSales
	}
	return nil
}

func init() {
	proto.RegisterType((*Sale)(nil), "recordsales.Sale")
	proto.RegisterType((*Config)(nil), "recordsales.Config")
	proto.RegisterType((*GetStaleRequest)(nil), "recordsales.GetStaleRequest")
	proto.RegisterType((*GetStaleResponse)(nil), "recordsales.GetStaleResponse")
}

func init() { proto.RegisterFile("recordsales.proto", fileDescriptor_3961a399f18e1e66) }

var fileDescriptor_3961a399f18e1e66 = []byte{
	// 256 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x84, 0x91, 0x41, 0x4b, 0xc3, 0x30,
	0x14, 0xc7, 0xed, 0x6a, 0xc7, 0x78, 0x05, 0x5d, 0x83, 0x87, 0x20, 0x8a, 0x25, 0x17, 0x7b, 0x71,
	0x87, 0xfa, 0x11, 0x04, 0x65, 0xd7, 0x56, 0xc1, 0x5b, 0x8d, 0xe9, 0x73, 0x06, 0xba, 0xa6, 0xe6,
	0x65, 0xfb, 0xfc, 0x92, 0x94, 0xe9, 0x14, 0xc6, 0x6e, 0x2f, 0xbf, 0xf7, 0xcf, 0x3f, 0x3f, 0x08,
	0x64, 0x16, 0x95, 0xb1, 0x2d, 0xc9, 0x0e, 0x69, 0x31, 0x58, 0xe3, 0x0c, 0x4b, 0xf7, 0x90, 0x58,
	0xc1, 0x69, 0x2d, 0x3b, 0x64, 0x37, 0x90, 0xea, 0x9e, 0x9c, 0xec, 0x15, 0x36, 0xba, 0xe5, 0x51,
	0x1e, 0x15, 0x49, 0x05, 0x3b, 0xb4, 0x6c, 0x59, 0x01, 0xf3, 0x4e, 0x92, 0x6b, 0x36, 0x43, 0x2b,
	0x1d, 0x36, 0x4e, 0xaf, 0x91, 0x4f, 0xf2, 0xa8, 0x88, 0xab, 0x33, 0xcf, 0x5f, 0x02, 0x7e, 0xd6,
	0x6b, 0x64, 0x17, 0x90, 0x0c, 0x56, 0x2b, 0xe4, 0x71, 0x28, 0x19, 0x0f, 0xe2, 0x0d, 0xa6, 0x0f,
	0xa6, 0xff, 0xd0, 0x2b, 0x76, 0x0b, 0x49, 0x78, 0x9b, 0x47, 0x79, 0x5c, 0xa4, 0x65, 0xb6, 0xd8,
	0x57, 0xf4, 0x32, 0xd5, 0xb8, 0x67, 0x77, 0x30, 0x93, 0x56, 0x7d, 0xea, 0x2d, 0x12, 0x9f, 0x1c,
	0xca, 0xfe, 0x44, 0x44, 0x06, 0xe7, 0x4f, 0xe8, 0x6a, 0xe7, 0x29, 0x7e, 0x6d, 0x90, 0x9c, 0x78,
	0x84, 0xf9, 0x2f, 0xa2, 0xc1, 0xf4, 0x84, 0xac, 0x84, 0x94, 0x3c, 0x68, 0x8e, 0x48, 0x40, 0x48,
	0xf9, 0x91, 0xca, 0x57, 0x48, 0xfd, 0x50, 0xa3, 0xdd, 0x6a, 0x85, 0x6c, 0x09, 0xb3, 0x5d, 0x2d,
	0xbb, 0xfa, 0x73, 0xf3, 0x9f, 0xc0, 0xe5, 0xf5, 0x81, 0xed, 0xe8, 0x22, 0x4e, 0xde, 0xa7, 0xe1,
	0x4f, 0xee, 0xbf, 0x03, 0x00, 0x00, 0xff, 0xff, 0x98, 0x7c, 0x4e, 0x21, 0xa8, 0x01, 0x00, 0x00,
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// SaleServiceClient is the client API for SaleService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type SaleServiceClient interface {
	GetStale(ctx context.Context, in *GetStaleRequest, opts ...grpc.CallOption) (*GetStaleResponse, error)
}

type saleServiceClient struct {
	cc *grpc.ClientConn
}

func NewSaleServiceClient(cc *grpc.ClientConn) SaleServiceClient {
	return &saleServiceClient{cc}
}

func (c *saleServiceClient) GetStale(ctx context.Context, in *GetStaleRequest, opts ...grpc.CallOption) (*GetStaleResponse, error) {
	out := new(GetStaleResponse)
	err := c.cc.Invoke(ctx, "/recordsales.SaleService/GetStale", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// SaleServiceServer is the server API for SaleService service.
type SaleServiceServer interface {
	GetStale(context.Context, *GetStaleRequest) (*GetStaleResponse, error)
}

func RegisterSaleServiceServer(s *grpc.Server, srv SaleServiceServer) {
	s.RegisterService(&_SaleService_serviceDesc, srv)
}

func _SaleService_GetStale_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetStaleRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SaleServiceServer).GetStale(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/recordsales.SaleService/GetStale",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SaleServiceServer).GetStale(ctx, req.(*GetStaleRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _SaleService_serviceDesc = grpc.ServiceDesc{
	ServiceName: "recordsales.SaleService",
	HandlerType: (*SaleServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetStale",
			Handler:    _SaleService_GetStale_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "recordsales.proto",
}
