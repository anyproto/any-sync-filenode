// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: service/sync/syncpb/protos/sync.proto

package syncpb

import (
	fmt "fmt"
	aclpb "github.com/anytypeio/go-anytype-infrastructure-experiments/pkg/acl/aclchanges/aclpb"
	proto "github.com/gogo/protobuf/proto"
	io "io"
	math "math"
	math_bits "math/bits"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.GoGoProtoPackageIsVersion3 // please upgrade the proto package

type Sync struct {
}

func (m *Sync) Reset()         { *m = Sync{} }
func (m *Sync) String() string { return proto.CompactTextString(m) }
func (*Sync) ProtoMessage()    {}
func (*Sync) Descriptor() ([]byte, []int) {
	return fileDescriptor_5f66cdd599c6466f, []int{0}
}
func (m *Sync) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *Sync) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_Sync.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *Sync) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Sync.Merge(m, src)
}
func (m *Sync) XXX_Size() int {
	return m.Size()
}
func (m *Sync) XXX_DiscardUnknown() {
	xxx_messageInfo_Sync.DiscardUnknown(m)
}

var xxx_messageInfo_Sync proto.InternalMessageInfo

type SyncHeadUpdate struct {
	Heads        []string           `protobuf:"bytes,1,rep,name=heads,proto3" json:"heads,omitempty"`
	Changes      []*aclpb.RawChange `protobuf:"bytes,2,rep,name=changes,proto3" json:"changes,omitempty"`
	TreeId       string             `protobuf:"bytes,3,opt,name=treeId,proto3" json:"treeId,omitempty"`
	SnapshotPath []string           `protobuf:"bytes,4,rep,name=snapshotPath,proto3" json:"snapshotPath,omitempty"`
}

func (m *SyncHeadUpdate) Reset()         { *m = SyncHeadUpdate{} }
func (m *SyncHeadUpdate) String() string { return proto.CompactTextString(m) }
func (*SyncHeadUpdate) ProtoMessage()    {}
func (*SyncHeadUpdate) Descriptor() ([]byte, []int) {
	return fileDescriptor_5f66cdd599c6466f, []int{0, 0}
}
func (m *SyncHeadUpdate) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *SyncHeadUpdate) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_SyncHeadUpdate.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *SyncHeadUpdate) XXX_Merge(src proto.Message) {
	xxx_messageInfo_SyncHeadUpdate.Merge(m, src)
}
func (m *SyncHeadUpdate) XXX_Size() int {
	return m.Size()
}
func (m *SyncHeadUpdate) XXX_DiscardUnknown() {
	xxx_messageInfo_SyncHeadUpdate.DiscardUnknown(m)
}

var xxx_messageInfo_SyncHeadUpdate proto.InternalMessageInfo

func (m *SyncHeadUpdate) GetHeads() []string {
	if m != nil {
		return m.Heads
	}
	return nil
}

func (m *SyncHeadUpdate) GetChanges() []*aclpb.RawChange {
	if m != nil {
		return m.Changes
	}
	return nil
}

func (m *SyncHeadUpdate) GetTreeId() string {
	if m != nil {
		return m.TreeId
	}
	return ""
}

func (m *SyncHeadUpdate) GetSnapshotPath() []string {
	if m != nil {
		return m.SnapshotPath
	}
	return nil
}

type SyncFull struct {
}

func (m *SyncFull) Reset()         { *m = SyncFull{} }
func (m *SyncFull) String() string { return proto.CompactTextString(m) }
func (*SyncFull) ProtoMessage()    {}
func (*SyncFull) Descriptor() ([]byte, []int) {
	return fileDescriptor_5f66cdd599c6466f, []int{0, 1}
}
func (m *SyncFull) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *SyncFull) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_SyncFull.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *SyncFull) XXX_Merge(src proto.Message) {
	xxx_messageInfo_SyncFull.Merge(m, src)
}
func (m *SyncFull) XXX_Size() int {
	return m.Size()
}
func (m *SyncFull) XXX_DiscardUnknown() {
	xxx_messageInfo_SyncFull.DiscardUnknown(m)
}

var xxx_messageInfo_SyncFull proto.InternalMessageInfo

// here with send the request with all changes we have (we already know sender's snapshot path)
type SyncFullRequest struct {
	Heads        []string           `protobuf:"bytes,1,rep,name=heads,proto3" json:"heads,omitempty"`
	Changes      []*aclpb.RawChange `protobuf:"bytes,2,rep,name=changes,proto3" json:"changes,omitempty"`
	TreeId       string             `protobuf:"bytes,3,opt,name=treeId,proto3" json:"treeId,omitempty"`
	SnapshotPath []string           `protobuf:"bytes,4,rep,name=snapshotPath,proto3" json:"snapshotPath,omitempty"`
}

func (m *SyncFullRequest) Reset()         { *m = SyncFullRequest{} }
func (m *SyncFullRequest) String() string { return proto.CompactTextString(m) }
func (*SyncFullRequest) ProtoMessage()    {}
func (*SyncFullRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_5f66cdd599c6466f, []int{0, 1, 0}
}
func (m *SyncFullRequest) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *SyncFullRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_SyncFullRequest.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *SyncFullRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_SyncFullRequest.Merge(m, src)
}
func (m *SyncFullRequest) XXX_Size() int {
	return m.Size()
}
func (m *SyncFullRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_SyncFullRequest.DiscardUnknown(m)
}

var xxx_messageInfo_SyncFullRequest proto.InternalMessageInfo

func (m *SyncFullRequest) GetHeads() []string {
	if m != nil {
		return m.Heads
	}
	return nil
}

func (m *SyncFullRequest) GetChanges() []*aclpb.RawChange {
	if m != nil {
		return m.Changes
	}
	return nil
}

func (m *SyncFullRequest) GetTreeId() string {
	if m != nil {
		return m.TreeId
	}
	return ""
}

func (m *SyncFullRequest) GetSnapshotPath() []string {
	if m != nil {
		return m.SnapshotPath
	}
	return nil
}

type SyncFullResponse struct {
	Heads        []string           `protobuf:"bytes,1,rep,name=heads,proto3" json:"heads,omitempty"`
	Changes      []*aclpb.RawChange `protobuf:"bytes,2,rep,name=changes,proto3" json:"changes,omitempty"`
	TreeId       string             `protobuf:"bytes,3,opt,name=treeId,proto3" json:"treeId,omitempty"`
	SnapshotPath []string           `protobuf:"bytes,4,rep,name=snapshotPath,proto3" json:"snapshotPath,omitempty"`
}

func (m *SyncFullResponse) Reset()         { *m = SyncFullResponse{} }
func (m *SyncFullResponse) String() string { return proto.CompactTextString(m) }
func (*SyncFullResponse) ProtoMessage()    {}
func (*SyncFullResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_5f66cdd599c6466f, []int{0, 1, 1}
}
func (m *SyncFullResponse) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *SyncFullResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_SyncFullResponse.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *SyncFullResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_SyncFullResponse.Merge(m, src)
}
func (m *SyncFullResponse) XXX_Size() int {
	return m.Size()
}
func (m *SyncFullResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_SyncFullResponse.DiscardUnknown(m)
}

var xxx_messageInfo_SyncFullResponse proto.InternalMessageInfo

func (m *SyncFullResponse) GetHeads() []string {
	if m != nil {
		return m.Heads
	}
	return nil
}

func (m *SyncFullResponse) GetChanges() []*aclpb.RawChange {
	if m != nil {
		return m.Changes
	}
	return nil
}

func (m *SyncFullResponse) GetTreeId() string {
	if m != nil {
		return m.TreeId
	}
	return ""
}

func (m *SyncFullResponse) GetSnapshotPath() []string {
	if m != nil {
		return m.SnapshotPath
	}
	return nil
}

func init() {
	proto.RegisterType((*Sync)(nil), "anytype.Sync")
	proto.RegisterType((*SyncHeadUpdate)(nil), "anytype.Sync.HeadUpdate")
	proto.RegisterType((*SyncFull)(nil), "anytype.Sync.Full")
	proto.RegisterType((*SyncFullRequest)(nil), "anytype.Sync.Full.Request")
	proto.RegisterType((*SyncFullResponse)(nil), "anytype.Sync.Full.Response")
}

func init() {
	proto.RegisterFile("service/sync/syncpb/protos/sync.proto", fileDescriptor_5f66cdd599c6466f)
}

var fileDescriptor_5f66cdd599c6466f = []byte{
	// 273 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0x52, 0x2d, 0x4e, 0x2d, 0x2a,
	0xcb, 0x4c, 0x4e, 0xd5, 0x2f, 0xae, 0xcc, 0x4b, 0x06, 0x13, 0x05, 0x49, 0xfa, 0x05, 0x45, 0xf9,
	0x25, 0xf9, 0xc5, 0x60, 0x9e, 0x1e, 0x98, 0x2d, 0xc4, 0x9e, 0x98, 0x57, 0x59, 0x52, 0x59, 0x90,
	0x2a, 0x65, 0x50, 0x90, 0x9d, 0xae, 0x9f, 0x98, 0x9c, 0x03, 0xc2, 0xc9, 0x19, 0x89, 0x79, 0xe9,
	0xa9, 0xc5, 0x20, 0x26, 0x42, 0x13, 0x42, 0x1c, 0xa2, 0x55, 0x69, 0x35, 0x33, 0x17, 0x4b, 0x70,
	0x65, 0x5e, 0xb2, 0x54, 0x07, 0x23, 0x17, 0x97, 0x47, 0x6a, 0x62, 0x4a, 0x68, 0x41, 0x4a, 0x62,
	0x49, 0xaa, 0x90, 0x08, 0x17, 0x6b, 0x46, 0x6a, 0x62, 0x4a, 0xb1, 0x04, 0xa3, 0x02, 0xb3, 0x06,
	0x67, 0x10, 0x84, 0x23, 0xa4, 0xc1, 0xc5, 0x0e, 0xd5, 0x2e, 0xc1, 0xa4, 0xc0, 0xac, 0xc1, 0x6d,
	0xc4, 0xa7, 0x97, 0x98, 0x9c, 0xa3, 0x17, 0x94, 0x58, 0xee, 0x0c, 0x16, 0x0e, 0x82, 0x49, 0x0b,
	0x89, 0x71, 0xb1, 0x95, 0x14, 0xa5, 0xa6, 0x7a, 0xa6, 0x48, 0x30, 0x2b, 0x30, 0x6a, 0x70, 0x06,
	0x41, 0x79, 0x42, 0x4a, 0x5c, 0x3c, 0xc5, 0x79, 0x89, 0x05, 0xc5, 0x19, 0xf9, 0x25, 0x01, 0x89,
	0x25, 0x19, 0x12, 0x2c, 0x60, 0xe3, 0x51, 0xc4, 0xa4, 0xa6, 0x33, 0x71, 0xb1, 0xb8, 0x95, 0xe6,
	0xe4, 0x48, 0xb5, 0x32, 0x72, 0xb1, 0x07, 0xa5, 0x16, 0x96, 0xa6, 0x16, 0x97, 0x0c, 0xa8, 0x83,
	0xda, 0x18, 0xb9, 0x38, 0x82, 0x52, 0x8b, 0x0b, 0xf2, 0xf3, 0x8a, 0x07, 0x34, 0x64, 0x9c, 0x14,
	0x4e, 0x3c, 0x92, 0x63, 0xbc, 0xf0, 0x48, 0x8e, 0xf1, 0xc1, 0x23, 0x39, 0xc6, 0x09, 0x8f, 0xe5,
	0x18, 0x2e, 0x3c, 0x96, 0x63, 0xb8, 0xf1, 0x58, 0x8e, 0x21, 0x8a, 0x0d, 0x92, 0x38, 0x92, 0xd8,
	0xc0, 0xd1, 0x6a, 0x0c, 0x08, 0x00, 0x00, 0xff, 0xff, 0x90, 0x29, 0xbc, 0x62, 0x3a, 0x02, 0x00,
	0x00,
}

func (m *Sync) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *Sync) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *Sync) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	return len(dAtA) - i, nil
}

func (m *SyncHeadUpdate) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *SyncHeadUpdate) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *SyncHeadUpdate) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.SnapshotPath) > 0 {
		for iNdEx := len(m.SnapshotPath) - 1; iNdEx >= 0; iNdEx-- {
			i -= len(m.SnapshotPath[iNdEx])
			copy(dAtA[i:], m.SnapshotPath[iNdEx])
			i = encodeVarintSync(dAtA, i, uint64(len(m.SnapshotPath[iNdEx])))
			i--
			dAtA[i] = 0x22
		}
	}
	if len(m.TreeId) > 0 {
		i -= len(m.TreeId)
		copy(dAtA[i:], m.TreeId)
		i = encodeVarintSync(dAtA, i, uint64(len(m.TreeId)))
		i--
		dAtA[i] = 0x1a
	}
	if len(m.Changes) > 0 {
		for iNdEx := len(m.Changes) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.Changes[iNdEx].MarshalToSizedBuffer(dAtA[:i])
				if err != nil {
					return 0, err
				}
				i -= size
				i = encodeVarintSync(dAtA, i, uint64(size))
			}
			i--
			dAtA[i] = 0x12
		}
	}
	if len(m.Heads) > 0 {
		for iNdEx := len(m.Heads) - 1; iNdEx >= 0; iNdEx-- {
			i -= len(m.Heads[iNdEx])
			copy(dAtA[i:], m.Heads[iNdEx])
			i = encodeVarintSync(dAtA, i, uint64(len(m.Heads[iNdEx])))
			i--
			dAtA[i] = 0xa
		}
	}
	return len(dAtA) - i, nil
}

func (m *SyncFull) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *SyncFull) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *SyncFull) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	return len(dAtA) - i, nil
}

func (m *SyncFullRequest) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *SyncFullRequest) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *SyncFullRequest) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.SnapshotPath) > 0 {
		for iNdEx := len(m.SnapshotPath) - 1; iNdEx >= 0; iNdEx-- {
			i -= len(m.SnapshotPath[iNdEx])
			copy(dAtA[i:], m.SnapshotPath[iNdEx])
			i = encodeVarintSync(dAtA, i, uint64(len(m.SnapshotPath[iNdEx])))
			i--
			dAtA[i] = 0x22
		}
	}
	if len(m.TreeId) > 0 {
		i -= len(m.TreeId)
		copy(dAtA[i:], m.TreeId)
		i = encodeVarintSync(dAtA, i, uint64(len(m.TreeId)))
		i--
		dAtA[i] = 0x1a
	}
	if len(m.Changes) > 0 {
		for iNdEx := len(m.Changes) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.Changes[iNdEx].MarshalToSizedBuffer(dAtA[:i])
				if err != nil {
					return 0, err
				}
				i -= size
				i = encodeVarintSync(dAtA, i, uint64(size))
			}
			i--
			dAtA[i] = 0x12
		}
	}
	if len(m.Heads) > 0 {
		for iNdEx := len(m.Heads) - 1; iNdEx >= 0; iNdEx-- {
			i -= len(m.Heads[iNdEx])
			copy(dAtA[i:], m.Heads[iNdEx])
			i = encodeVarintSync(dAtA, i, uint64(len(m.Heads[iNdEx])))
			i--
			dAtA[i] = 0xa
		}
	}
	return len(dAtA) - i, nil
}

func (m *SyncFullResponse) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *SyncFullResponse) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *SyncFullResponse) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.SnapshotPath) > 0 {
		for iNdEx := len(m.SnapshotPath) - 1; iNdEx >= 0; iNdEx-- {
			i -= len(m.SnapshotPath[iNdEx])
			copy(dAtA[i:], m.SnapshotPath[iNdEx])
			i = encodeVarintSync(dAtA, i, uint64(len(m.SnapshotPath[iNdEx])))
			i--
			dAtA[i] = 0x22
		}
	}
	if len(m.TreeId) > 0 {
		i -= len(m.TreeId)
		copy(dAtA[i:], m.TreeId)
		i = encodeVarintSync(dAtA, i, uint64(len(m.TreeId)))
		i--
		dAtA[i] = 0x1a
	}
	if len(m.Changes) > 0 {
		for iNdEx := len(m.Changes) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.Changes[iNdEx].MarshalToSizedBuffer(dAtA[:i])
				if err != nil {
					return 0, err
				}
				i -= size
				i = encodeVarintSync(dAtA, i, uint64(size))
			}
			i--
			dAtA[i] = 0x12
		}
	}
	if len(m.Heads) > 0 {
		for iNdEx := len(m.Heads) - 1; iNdEx >= 0; iNdEx-- {
			i -= len(m.Heads[iNdEx])
			copy(dAtA[i:], m.Heads[iNdEx])
			i = encodeVarintSync(dAtA, i, uint64(len(m.Heads[iNdEx])))
			i--
			dAtA[i] = 0xa
		}
	}
	return len(dAtA) - i, nil
}

func encodeVarintSync(dAtA []byte, offset int, v uint64) int {
	offset -= sovSync(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (m *Sync) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	return n
}

func (m *SyncHeadUpdate) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if len(m.Heads) > 0 {
		for _, s := range m.Heads {
			l = len(s)
			n += 1 + l + sovSync(uint64(l))
		}
	}
	if len(m.Changes) > 0 {
		for _, e := range m.Changes {
			l = e.Size()
			n += 1 + l + sovSync(uint64(l))
		}
	}
	l = len(m.TreeId)
	if l > 0 {
		n += 1 + l + sovSync(uint64(l))
	}
	if len(m.SnapshotPath) > 0 {
		for _, s := range m.SnapshotPath {
			l = len(s)
			n += 1 + l + sovSync(uint64(l))
		}
	}
	return n
}

func (m *SyncFull) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	return n
}

func (m *SyncFullRequest) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if len(m.Heads) > 0 {
		for _, s := range m.Heads {
			l = len(s)
			n += 1 + l + sovSync(uint64(l))
		}
	}
	if len(m.Changes) > 0 {
		for _, e := range m.Changes {
			l = e.Size()
			n += 1 + l + sovSync(uint64(l))
		}
	}
	l = len(m.TreeId)
	if l > 0 {
		n += 1 + l + sovSync(uint64(l))
	}
	if len(m.SnapshotPath) > 0 {
		for _, s := range m.SnapshotPath {
			l = len(s)
			n += 1 + l + sovSync(uint64(l))
		}
	}
	return n
}

func (m *SyncFullResponse) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if len(m.Heads) > 0 {
		for _, s := range m.Heads {
			l = len(s)
			n += 1 + l + sovSync(uint64(l))
		}
	}
	if len(m.Changes) > 0 {
		for _, e := range m.Changes {
			l = e.Size()
			n += 1 + l + sovSync(uint64(l))
		}
	}
	l = len(m.TreeId)
	if l > 0 {
		n += 1 + l + sovSync(uint64(l))
	}
	if len(m.SnapshotPath) > 0 {
		for _, s := range m.SnapshotPath {
			l = len(s)
			n += 1 + l + sovSync(uint64(l))
		}
	}
	return n
}

func sovSync(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozSync(x uint64) (n int) {
	return sovSync(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *Sync) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowSync
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: Sync: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: Sync: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		default:
			iNdEx = preIndex
			skippy, err := skipSync(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthSync
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *SyncHeadUpdate) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowSync
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: HeadUpdate: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: HeadUpdate: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Heads", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowSync
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthSync
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthSync
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Heads = append(m.Heads, string(dAtA[iNdEx:postIndex]))
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Changes", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowSync
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthSync
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthSync
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Changes = append(m.Changes, &aclpb.RawChange{})
			if err := m.Changes[len(m.Changes)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field TreeId", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowSync
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthSync
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthSync
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.TreeId = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 4:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field SnapshotPath", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowSync
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthSync
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthSync
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.SnapshotPath = append(m.SnapshotPath, string(dAtA[iNdEx:postIndex]))
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipSync(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthSync
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *SyncFull) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowSync
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: Full: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: Full: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		default:
			iNdEx = preIndex
			skippy, err := skipSync(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthSync
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *SyncFullRequest) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowSync
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: Request: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: Request: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Heads", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowSync
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthSync
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthSync
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Heads = append(m.Heads, string(dAtA[iNdEx:postIndex]))
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Changes", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowSync
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthSync
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthSync
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Changes = append(m.Changes, &aclpb.RawChange{})
			if err := m.Changes[len(m.Changes)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field TreeId", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowSync
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthSync
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthSync
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.TreeId = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 4:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field SnapshotPath", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowSync
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthSync
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthSync
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.SnapshotPath = append(m.SnapshotPath, string(dAtA[iNdEx:postIndex]))
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipSync(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthSync
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *SyncFullResponse) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowSync
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: Response: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: Response: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Heads", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowSync
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthSync
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthSync
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Heads = append(m.Heads, string(dAtA[iNdEx:postIndex]))
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Changes", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowSync
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthSync
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthSync
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Changes = append(m.Changes, &aclpb.RawChange{})
			if err := m.Changes[len(m.Changes)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field TreeId", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowSync
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthSync
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthSync
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.TreeId = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 4:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field SnapshotPath", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowSync
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthSync
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthSync
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.SnapshotPath = append(m.SnapshotPath, string(dAtA[iNdEx:postIndex]))
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipSync(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthSync
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func skipSync(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowSync
			}
			if iNdEx >= l {
				return 0, io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		wireType := int(wire & 0x7)
		switch wireType {
		case 0:
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowSync
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				iNdEx++
				if dAtA[iNdEx-1] < 0x80 {
					break
				}
			}
		case 1:
			iNdEx += 8
		case 2:
			var length int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowSync
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				length |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if length < 0 {
				return 0, ErrInvalidLengthSync
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupSync
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthSync
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthSync        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowSync          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupSync = fmt.Errorf("proto: unexpected end of group")
)
