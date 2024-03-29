// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: index/indexproto/protos/index.proto

package indexproto

import (
	fmt "fmt"
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

type CidEntry struct {
	Size_      uint64 `protobuf:"varint,1,opt,name=size,proto3" json:"size,omitempty"`
	CreateTime int64  `protobuf:"varint,2,opt,name=createTime,proto3" json:"createTime,omitempty"`
	UpdateTime int64  `protobuf:"varint,3,opt,name=updateTime,proto3" json:"updateTime,omitempty"`
	Refs       int32  `protobuf:"varint,4,opt,name=refs,proto3" json:"refs,omitempty"`
	Version    uint32 `protobuf:"varint,5,opt,name=version,proto3" json:"version,omitempty"`
}

func (m *CidEntry) Reset()         { *m = CidEntry{} }
func (m *CidEntry) String() string { return proto.CompactTextString(m) }
func (*CidEntry) ProtoMessage()    {}
func (*CidEntry) Descriptor() ([]byte, []int) {
	return fileDescriptor_f1f29953df8d243b, []int{0}
}
func (m *CidEntry) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *CidEntry) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_CidEntry.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *CidEntry) XXX_Merge(src proto.Message) {
	xxx_messageInfo_CidEntry.Merge(m, src)
}
func (m *CidEntry) XXX_Size() int {
	return m.Size()
}
func (m *CidEntry) XXX_DiscardUnknown() {
	xxx_messageInfo_CidEntry.DiscardUnknown(m)
}

var xxx_messageInfo_CidEntry proto.InternalMessageInfo

func (m *CidEntry) GetSize_() uint64 {
	if m != nil {
		return m.Size_
	}
	return 0
}

func (m *CidEntry) GetCreateTime() int64 {
	if m != nil {
		return m.CreateTime
	}
	return 0
}

func (m *CidEntry) GetUpdateTime() int64 {
	if m != nil {
		return m.UpdateTime
	}
	return 0
}

func (m *CidEntry) GetRefs() int32 {
	if m != nil {
		return m.Refs
	}
	return 0
}

func (m *CidEntry) GetVersion() uint32 {
	if m != nil {
		return m.Version
	}
	return 0
}

type CidList struct {
	Cids [][]byte `protobuf:"bytes,1,rep,name=cids,proto3" json:"cids,omitempty"`
}

func (m *CidList) Reset()         { *m = CidList{} }
func (m *CidList) String() string { return proto.CompactTextString(m) }
func (*CidList) ProtoMessage()    {}
func (*CidList) Descriptor() ([]byte, []int) {
	return fileDescriptor_f1f29953df8d243b, []int{1}
}
func (m *CidList) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *CidList) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_CidList.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *CidList) XXX_Merge(src proto.Message) {
	xxx_messageInfo_CidList.Merge(m, src)
}
func (m *CidList) XXX_Size() int {
	return m.Size()
}
func (m *CidList) XXX_DiscardUnknown() {
	xxx_messageInfo_CidList.DiscardUnknown(m)
}

var xxx_messageInfo_CidList proto.InternalMessageInfo

func (m *CidList) GetCids() [][]byte {
	if m != nil {
		return m.Cids
	}
	return nil
}

type GroupEntry struct {
	GroupId      string   `protobuf:"bytes,1,opt,name=groupId,proto3" json:"groupId,omitempty"`
	CreateTime   int64    `protobuf:"varint,2,opt,name=createTime,proto3" json:"createTime,omitempty"`
	UpdateTime   int64    `protobuf:"varint,3,opt,name=updateTime,proto3" json:"updateTime,omitempty"`
	Size_        uint64   `protobuf:"varint,4,opt,name=size,proto3" json:"size,omitempty"`
	CidCount     uint64   `protobuf:"varint,5,opt,name=cidCount,proto3" json:"cidCount,omitempty"`
	SpaceIds     []string `protobuf:"bytes,6,rep,name=spaceIds,proto3" json:"spaceIds,omitempty"`
	Limit        uint64   `protobuf:"varint,7,opt,name=limit,proto3" json:"limit,omitempty"`
	AccountLimit uint64   `protobuf:"varint,8,opt,name=accountLimit,proto3" json:"accountLimit,omitempty"`
}

func (m *GroupEntry) Reset()         { *m = GroupEntry{} }
func (m *GroupEntry) String() string { return proto.CompactTextString(m) }
func (*GroupEntry) ProtoMessage()    {}
func (*GroupEntry) Descriptor() ([]byte, []int) {
	return fileDescriptor_f1f29953df8d243b, []int{2}
}
func (m *GroupEntry) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *GroupEntry) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_GroupEntry.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *GroupEntry) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GroupEntry.Merge(m, src)
}
func (m *GroupEntry) XXX_Size() int {
	return m.Size()
}
func (m *GroupEntry) XXX_DiscardUnknown() {
	xxx_messageInfo_GroupEntry.DiscardUnknown(m)
}

var xxx_messageInfo_GroupEntry proto.InternalMessageInfo

func (m *GroupEntry) GetGroupId() string {
	if m != nil {
		return m.GroupId
	}
	return ""
}

func (m *GroupEntry) GetCreateTime() int64 {
	if m != nil {
		return m.CreateTime
	}
	return 0
}

func (m *GroupEntry) GetUpdateTime() int64 {
	if m != nil {
		return m.UpdateTime
	}
	return 0
}

func (m *GroupEntry) GetSize_() uint64 {
	if m != nil {
		return m.Size_
	}
	return 0
}

func (m *GroupEntry) GetCidCount() uint64 {
	if m != nil {
		return m.CidCount
	}
	return 0
}

func (m *GroupEntry) GetSpaceIds() []string {
	if m != nil {
		return m.SpaceIds
	}
	return nil
}

func (m *GroupEntry) GetLimit() uint64 {
	if m != nil {
		return m.Limit
	}
	return 0
}

func (m *GroupEntry) GetAccountLimit() uint64 {
	if m != nil {
		return m.AccountLimit
	}
	return 0
}

type SpaceEntry struct {
	GroupId    string `protobuf:"bytes,1,opt,name=groupId,proto3" json:"groupId,omitempty"`
	CreateTime int64  `protobuf:"varint,2,opt,name=createTime,proto3" json:"createTime,omitempty"`
	UpdateTime int64  `protobuf:"varint,3,opt,name=updateTime,proto3" json:"updateTime,omitempty"`
	Size_      uint64 `protobuf:"varint,4,opt,name=size,proto3" json:"size,omitempty"`
	FileCount  uint32 `protobuf:"varint,5,opt,name=fileCount,proto3" json:"fileCount,omitempty"`
	CidCount   uint64 `protobuf:"varint,6,opt,name=cidCount,proto3" json:"cidCount,omitempty"`
	Limit      uint64 `protobuf:"varint,7,opt,name=limit,proto3" json:"limit,omitempty"`
}

func (m *SpaceEntry) Reset()         { *m = SpaceEntry{} }
func (m *SpaceEntry) String() string { return proto.CompactTextString(m) }
func (*SpaceEntry) ProtoMessage()    {}
func (*SpaceEntry) Descriptor() ([]byte, []int) {
	return fileDescriptor_f1f29953df8d243b, []int{3}
}
func (m *SpaceEntry) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *SpaceEntry) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_SpaceEntry.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *SpaceEntry) XXX_Merge(src proto.Message) {
	xxx_messageInfo_SpaceEntry.Merge(m, src)
}
func (m *SpaceEntry) XXX_Size() int {
	return m.Size()
}
func (m *SpaceEntry) XXX_DiscardUnknown() {
	xxx_messageInfo_SpaceEntry.DiscardUnknown(m)
}

var xxx_messageInfo_SpaceEntry proto.InternalMessageInfo

func (m *SpaceEntry) GetGroupId() string {
	if m != nil {
		return m.GroupId
	}
	return ""
}

func (m *SpaceEntry) GetCreateTime() int64 {
	if m != nil {
		return m.CreateTime
	}
	return 0
}

func (m *SpaceEntry) GetUpdateTime() int64 {
	if m != nil {
		return m.UpdateTime
	}
	return 0
}

func (m *SpaceEntry) GetSize_() uint64 {
	if m != nil {
		return m.Size_
	}
	return 0
}

func (m *SpaceEntry) GetFileCount() uint32 {
	if m != nil {
		return m.FileCount
	}
	return 0
}

func (m *SpaceEntry) GetCidCount() uint64 {
	if m != nil {
		return m.CidCount
	}
	return 0
}

func (m *SpaceEntry) GetLimit() uint64 {
	if m != nil {
		return m.Limit
	}
	return 0
}

type FileEntry struct {
	Cids       []string `protobuf:"bytes,1,rep,name=cids,proto3" json:"cids,omitempty"`
	Size_      uint64   `protobuf:"varint,2,opt,name=size,proto3" json:"size,omitempty"`
	CreateTime int64    `protobuf:"varint,3,opt,name=createTime,proto3" json:"createTime,omitempty"`
	UpdateTime int64    `protobuf:"varint,4,opt,name=updateTime,proto3" json:"updateTime,omitempty"`
}

func (m *FileEntry) Reset()         { *m = FileEntry{} }
func (m *FileEntry) String() string { return proto.CompactTextString(m) }
func (*FileEntry) ProtoMessage()    {}
func (*FileEntry) Descriptor() ([]byte, []int) {
	return fileDescriptor_f1f29953df8d243b, []int{4}
}
func (m *FileEntry) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *FileEntry) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_FileEntry.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *FileEntry) XXX_Merge(src proto.Message) {
	xxx_messageInfo_FileEntry.Merge(m, src)
}
func (m *FileEntry) XXX_Size() int {
	return m.Size()
}
func (m *FileEntry) XXX_DiscardUnknown() {
	xxx_messageInfo_FileEntry.DiscardUnknown(m)
}

var xxx_messageInfo_FileEntry proto.InternalMessageInfo

func (m *FileEntry) GetCids() []string {
	if m != nil {
		return m.Cids
	}
	return nil
}

func (m *FileEntry) GetSize_() uint64 {
	if m != nil {
		return m.Size_
	}
	return 0
}

func (m *FileEntry) GetCreateTime() int64 {
	if m != nil {
		return m.CreateTime
	}
	return 0
}

func (m *FileEntry) GetUpdateTime() int64 {
	if m != nil {
		return m.UpdateTime
	}
	return 0
}

func init() {
	proto.RegisterType((*CidEntry)(nil), "fileIndexProto.CidEntry")
	proto.RegisterType((*CidList)(nil), "fileIndexProto.CidList")
	proto.RegisterType((*GroupEntry)(nil), "fileIndexProto.GroupEntry")
	proto.RegisterType((*SpaceEntry)(nil), "fileIndexProto.SpaceEntry")
	proto.RegisterType((*FileEntry)(nil), "fileIndexProto.FileEntry")
}

func init() {
	proto.RegisterFile("index/indexproto/protos/index.proto", fileDescriptor_f1f29953df8d243b)
}

var fileDescriptor_f1f29953df8d243b = []byte{
	// 372 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xc4, 0x53, 0xcd, 0x4a, 0xf3, 0x40,
	0x14, 0xed, 0x34, 0x69, 0xd3, 0x5c, 0xda, 0x8f, 0x8f, 0xc1, 0xc5, 0x20, 0x3a, 0x84, 0xb8, 0x09,
	0x2e, 0xec, 0xc2, 0x37, 0xb0, 0xa8, 0x14, 0xba, 0x90, 0xd1, 0x95, 0xbb, 0x9a, 0x99, 0xca, 0x40,
	0x9b, 0x84, 0x4c, 0x2a, 0xea, 0x33, 0xb8, 0xf0, 0xb1, 0xc4, 0x55, 0x97, 0x2e, 0xa5, 0xdd, 0xfa,
	0x10, 0x32, 0x37, 0x34, 0xfd, 0x41, 0x71, 0x23, 0xb8, 0x99, 0xdc, 0x73, 0x0e, 0x77, 0x38, 0xe7,
	0xce, 0x0d, 0x1c, 0xe8, 0x44, 0xaa, 0xfb, 0x2e, 0x9e, 0x59, 0x9e, 0x16, 0x69, 0x17, 0x4f, 0x53,
	0x32, 0x47, 0x08, 0xe8, 0xbf, 0x91, 0x1e, 0xab, 0xbe, 0x25, 0x2e, 0x2c, 0x0e, 0x9f, 0x08, 0xb4,
	0x7a, 0x5a, 0x9e, 0x26, 0x45, 0xfe, 0x40, 0x29, 0xb8, 0x46, 0x3f, 0x2a, 0x46, 0x02, 0x12, 0xb9,
	0x02, 0x6b, 0xca, 0x01, 0xe2, 0x5c, 0x0d, 0x0b, 0x75, 0xa5, 0x27, 0x8a, 0xd5, 0x03, 0x12, 0x39,
	0x62, 0x8d, 0xb1, 0xfa, 0x34, 0x93, 0x4b, 0xdd, 0x29, 0xf5, 0x15, 0x63, 0xef, 0xcc, 0xd5, 0xc8,
	0x30, 0x37, 0x20, 0x51, 0x43, 0x60, 0x4d, 0x19, 0x78, 0x77, 0x2a, 0x37, 0x3a, 0x4d, 0x58, 0x23,
	0x20, 0x51, 0x47, 0x2c, 0x61, 0xb8, 0x0f, 0x5e, 0x4f, 0xcb, 0x81, 0x36, 0x85, 0x6d, 0x8c, 0xb5,
	0x34, 0x8c, 0x04, 0x4e, 0xd4, 0x16, 0x58, 0x87, 0x1f, 0x04, 0xe0, 0x3c, 0x4f, 0xa7, 0x59, 0xe9,
	0x97, 0x81, 0x77, 0x6b, 0x51, 0x5f, 0xa2, 0x65, 0x5f, 0x2c, 0xe1, 0x6f, 0xb8, 0xc6, 0x49, 0xb8,
	0x6b, 0x93, 0xd8, 0x85, 0x56, 0xac, 0x65, 0x2f, 0x9d, 0x26, 0x05, 0xda, 0x76, 0x45, 0x85, 0xad,
	0x66, 0xb2, 0x61, 0xac, 0xfa, 0xd2, 0xb0, 0x66, 0xe0, 0x44, 0xbe, 0xa8, 0x30, 0xdd, 0x81, 0xc6,
	0x58, 0x4f, 0x74, 0xc1, 0x3c, 0x6c, 0x2a, 0x01, 0x0d, 0xa1, 0x3d, 0x8c, 0x63, 0xdb, 0x3c, 0x40,
	0xb1, 0x85, 0xe2, 0x06, 0x17, 0xbe, 0x12, 0x80, 0x4b, 0x7b, 0xcd, 0x5f, 0xc4, 0xdd, 0x03, 0xdf,
	0xee, 0xca, 0x2a, 0x6f, 0x47, 0xac, 0x88, 0x8d, 0x61, 0x34, 0xb7, 0x86, 0xf1, 0x65, 0xe0, 0xd0,
	0x80, 0x7f, 0xa6, 0xc7, 0xaa, 0xda, 0xb4, 0xea, 0x71, 0xfd, 0xf2, 0x71, 0x2b, 0x13, 0xf5, 0x6f,
	0xb7, 0xcf, 0xf9, 0x21, 0x98, 0xbb, 0x1d, 0xec, 0xe4, 0xf0, 0x65, 0xce, 0xc9, 0x6c, 0xce, 0xc9,
	0xfb, 0x9c, 0x93, 0xe7, 0x05, 0xaf, 0xcd, 0x16, 0xbc, 0xf6, 0xb6, 0xe0, 0xb5, 0xeb, 0xff, 0xdb,
	0x7f, 0xcb, 0x4d, 0x13, 0x3f, 0xc7, 0x9f, 0x01, 0x00, 0x00, 0xff, 0xff, 0x64, 0x5b, 0xf6, 0x08,
	0x48, 0x03, 0x00, 0x00,
}

func (m *CidEntry) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *CidEntry) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *CidEntry) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.Version != 0 {
		i = encodeVarintIndex(dAtA, i, uint64(m.Version))
		i--
		dAtA[i] = 0x28
	}
	if m.Refs != 0 {
		i = encodeVarintIndex(dAtA, i, uint64(m.Refs))
		i--
		dAtA[i] = 0x20
	}
	if m.UpdateTime != 0 {
		i = encodeVarintIndex(dAtA, i, uint64(m.UpdateTime))
		i--
		dAtA[i] = 0x18
	}
	if m.CreateTime != 0 {
		i = encodeVarintIndex(dAtA, i, uint64(m.CreateTime))
		i--
		dAtA[i] = 0x10
	}
	if m.Size_ != 0 {
		i = encodeVarintIndex(dAtA, i, uint64(m.Size_))
		i--
		dAtA[i] = 0x8
	}
	return len(dAtA) - i, nil
}

func (m *CidList) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *CidList) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *CidList) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.Cids) > 0 {
		for iNdEx := len(m.Cids) - 1; iNdEx >= 0; iNdEx-- {
			i -= len(m.Cids[iNdEx])
			copy(dAtA[i:], m.Cids[iNdEx])
			i = encodeVarintIndex(dAtA, i, uint64(len(m.Cids[iNdEx])))
			i--
			dAtA[i] = 0xa
		}
	}
	return len(dAtA) - i, nil
}

func (m *GroupEntry) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *GroupEntry) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *GroupEntry) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.AccountLimit != 0 {
		i = encodeVarintIndex(dAtA, i, uint64(m.AccountLimit))
		i--
		dAtA[i] = 0x40
	}
	if m.Limit != 0 {
		i = encodeVarintIndex(dAtA, i, uint64(m.Limit))
		i--
		dAtA[i] = 0x38
	}
	if len(m.SpaceIds) > 0 {
		for iNdEx := len(m.SpaceIds) - 1; iNdEx >= 0; iNdEx-- {
			i -= len(m.SpaceIds[iNdEx])
			copy(dAtA[i:], m.SpaceIds[iNdEx])
			i = encodeVarintIndex(dAtA, i, uint64(len(m.SpaceIds[iNdEx])))
			i--
			dAtA[i] = 0x32
		}
	}
	if m.CidCount != 0 {
		i = encodeVarintIndex(dAtA, i, uint64(m.CidCount))
		i--
		dAtA[i] = 0x28
	}
	if m.Size_ != 0 {
		i = encodeVarintIndex(dAtA, i, uint64(m.Size_))
		i--
		dAtA[i] = 0x20
	}
	if m.UpdateTime != 0 {
		i = encodeVarintIndex(dAtA, i, uint64(m.UpdateTime))
		i--
		dAtA[i] = 0x18
	}
	if m.CreateTime != 0 {
		i = encodeVarintIndex(dAtA, i, uint64(m.CreateTime))
		i--
		dAtA[i] = 0x10
	}
	if len(m.GroupId) > 0 {
		i -= len(m.GroupId)
		copy(dAtA[i:], m.GroupId)
		i = encodeVarintIndex(dAtA, i, uint64(len(m.GroupId)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *SpaceEntry) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *SpaceEntry) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *SpaceEntry) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.Limit != 0 {
		i = encodeVarintIndex(dAtA, i, uint64(m.Limit))
		i--
		dAtA[i] = 0x38
	}
	if m.CidCount != 0 {
		i = encodeVarintIndex(dAtA, i, uint64(m.CidCount))
		i--
		dAtA[i] = 0x30
	}
	if m.FileCount != 0 {
		i = encodeVarintIndex(dAtA, i, uint64(m.FileCount))
		i--
		dAtA[i] = 0x28
	}
	if m.Size_ != 0 {
		i = encodeVarintIndex(dAtA, i, uint64(m.Size_))
		i--
		dAtA[i] = 0x20
	}
	if m.UpdateTime != 0 {
		i = encodeVarintIndex(dAtA, i, uint64(m.UpdateTime))
		i--
		dAtA[i] = 0x18
	}
	if m.CreateTime != 0 {
		i = encodeVarintIndex(dAtA, i, uint64(m.CreateTime))
		i--
		dAtA[i] = 0x10
	}
	if len(m.GroupId) > 0 {
		i -= len(m.GroupId)
		copy(dAtA[i:], m.GroupId)
		i = encodeVarintIndex(dAtA, i, uint64(len(m.GroupId)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *FileEntry) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *FileEntry) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *FileEntry) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.UpdateTime != 0 {
		i = encodeVarintIndex(dAtA, i, uint64(m.UpdateTime))
		i--
		dAtA[i] = 0x20
	}
	if m.CreateTime != 0 {
		i = encodeVarintIndex(dAtA, i, uint64(m.CreateTime))
		i--
		dAtA[i] = 0x18
	}
	if m.Size_ != 0 {
		i = encodeVarintIndex(dAtA, i, uint64(m.Size_))
		i--
		dAtA[i] = 0x10
	}
	if len(m.Cids) > 0 {
		for iNdEx := len(m.Cids) - 1; iNdEx >= 0; iNdEx-- {
			i -= len(m.Cids[iNdEx])
			copy(dAtA[i:], m.Cids[iNdEx])
			i = encodeVarintIndex(dAtA, i, uint64(len(m.Cids[iNdEx])))
			i--
			dAtA[i] = 0xa
		}
	}
	return len(dAtA) - i, nil
}

func encodeVarintIndex(dAtA []byte, offset int, v uint64) int {
	offset -= sovIndex(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (m *CidEntry) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.Size_ != 0 {
		n += 1 + sovIndex(uint64(m.Size_))
	}
	if m.CreateTime != 0 {
		n += 1 + sovIndex(uint64(m.CreateTime))
	}
	if m.UpdateTime != 0 {
		n += 1 + sovIndex(uint64(m.UpdateTime))
	}
	if m.Refs != 0 {
		n += 1 + sovIndex(uint64(m.Refs))
	}
	if m.Version != 0 {
		n += 1 + sovIndex(uint64(m.Version))
	}
	return n
}

func (m *CidList) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if len(m.Cids) > 0 {
		for _, b := range m.Cids {
			l = len(b)
			n += 1 + l + sovIndex(uint64(l))
		}
	}
	return n
}

func (m *GroupEntry) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.GroupId)
	if l > 0 {
		n += 1 + l + sovIndex(uint64(l))
	}
	if m.CreateTime != 0 {
		n += 1 + sovIndex(uint64(m.CreateTime))
	}
	if m.UpdateTime != 0 {
		n += 1 + sovIndex(uint64(m.UpdateTime))
	}
	if m.Size_ != 0 {
		n += 1 + sovIndex(uint64(m.Size_))
	}
	if m.CidCount != 0 {
		n += 1 + sovIndex(uint64(m.CidCount))
	}
	if len(m.SpaceIds) > 0 {
		for _, s := range m.SpaceIds {
			l = len(s)
			n += 1 + l + sovIndex(uint64(l))
		}
	}
	if m.Limit != 0 {
		n += 1 + sovIndex(uint64(m.Limit))
	}
	if m.AccountLimit != 0 {
		n += 1 + sovIndex(uint64(m.AccountLimit))
	}
	return n
}

func (m *SpaceEntry) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.GroupId)
	if l > 0 {
		n += 1 + l + sovIndex(uint64(l))
	}
	if m.CreateTime != 0 {
		n += 1 + sovIndex(uint64(m.CreateTime))
	}
	if m.UpdateTime != 0 {
		n += 1 + sovIndex(uint64(m.UpdateTime))
	}
	if m.Size_ != 0 {
		n += 1 + sovIndex(uint64(m.Size_))
	}
	if m.FileCount != 0 {
		n += 1 + sovIndex(uint64(m.FileCount))
	}
	if m.CidCount != 0 {
		n += 1 + sovIndex(uint64(m.CidCount))
	}
	if m.Limit != 0 {
		n += 1 + sovIndex(uint64(m.Limit))
	}
	return n
}

func (m *FileEntry) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if len(m.Cids) > 0 {
		for _, s := range m.Cids {
			l = len(s)
			n += 1 + l + sovIndex(uint64(l))
		}
	}
	if m.Size_ != 0 {
		n += 1 + sovIndex(uint64(m.Size_))
	}
	if m.CreateTime != 0 {
		n += 1 + sovIndex(uint64(m.CreateTime))
	}
	if m.UpdateTime != 0 {
		n += 1 + sovIndex(uint64(m.UpdateTime))
	}
	return n
}

func sovIndex(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozIndex(x uint64) (n int) {
	return sovIndex(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *CidEntry) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowIndex
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
			return fmt.Errorf("proto: CidEntry: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: CidEntry: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Size_", wireType)
			}
			m.Size_ = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowIndex
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Size_ |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 2:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field CreateTime", wireType)
			}
			m.CreateTime = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowIndex
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.CreateTime |= int64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 3:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field UpdateTime", wireType)
			}
			m.UpdateTime = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowIndex
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.UpdateTime |= int64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 4:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Refs", wireType)
			}
			m.Refs = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowIndex
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Refs |= int32(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 5:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Version", wireType)
			}
			m.Version = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowIndex
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Version |= uint32(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		default:
			iNdEx = preIndex
			skippy, err := skipIndex(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthIndex
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
func (m *CidList) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowIndex
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
			return fmt.Errorf("proto: CidList: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: CidList: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Cids", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowIndex
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				byteLen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if byteLen < 0 {
				return ErrInvalidLengthIndex
			}
			postIndex := iNdEx + byteLen
			if postIndex < 0 {
				return ErrInvalidLengthIndex
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Cids = append(m.Cids, make([]byte, postIndex-iNdEx))
			copy(m.Cids[len(m.Cids)-1], dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipIndex(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthIndex
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
func (m *GroupEntry) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowIndex
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
			return fmt.Errorf("proto: GroupEntry: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: GroupEntry: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field GroupId", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowIndex
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
				return ErrInvalidLengthIndex
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthIndex
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.GroupId = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field CreateTime", wireType)
			}
			m.CreateTime = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowIndex
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.CreateTime |= int64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 3:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field UpdateTime", wireType)
			}
			m.UpdateTime = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowIndex
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.UpdateTime |= int64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 4:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Size_", wireType)
			}
			m.Size_ = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowIndex
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Size_ |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 5:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field CidCount", wireType)
			}
			m.CidCount = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowIndex
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.CidCount |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 6:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field SpaceIds", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowIndex
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
				return ErrInvalidLengthIndex
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthIndex
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.SpaceIds = append(m.SpaceIds, string(dAtA[iNdEx:postIndex]))
			iNdEx = postIndex
		case 7:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Limit", wireType)
			}
			m.Limit = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowIndex
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Limit |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 8:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field AccountLimit", wireType)
			}
			m.AccountLimit = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowIndex
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.AccountLimit |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		default:
			iNdEx = preIndex
			skippy, err := skipIndex(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthIndex
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
func (m *SpaceEntry) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowIndex
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
			return fmt.Errorf("proto: SpaceEntry: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: SpaceEntry: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field GroupId", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowIndex
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
				return ErrInvalidLengthIndex
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthIndex
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.GroupId = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field CreateTime", wireType)
			}
			m.CreateTime = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowIndex
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.CreateTime |= int64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 3:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field UpdateTime", wireType)
			}
			m.UpdateTime = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowIndex
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.UpdateTime |= int64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 4:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Size_", wireType)
			}
			m.Size_ = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowIndex
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Size_ |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 5:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field FileCount", wireType)
			}
			m.FileCount = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowIndex
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.FileCount |= uint32(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 6:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field CidCount", wireType)
			}
			m.CidCount = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowIndex
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.CidCount |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 7:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Limit", wireType)
			}
			m.Limit = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowIndex
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Limit |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		default:
			iNdEx = preIndex
			skippy, err := skipIndex(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthIndex
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
func (m *FileEntry) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowIndex
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
			return fmt.Errorf("proto: FileEntry: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: FileEntry: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Cids", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowIndex
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
				return ErrInvalidLengthIndex
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthIndex
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Cids = append(m.Cids, string(dAtA[iNdEx:postIndex]))
			iNdEx = postIndex
		case 2:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Size_", wireType)
			}
			m.Size_ = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowIndex
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Size_ |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 3:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field CreateTime", wireType)
			}
			m.CreateTime = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowIndex
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.CreateTime |= int64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 4:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field UpdateTime", wireType)
			}
			m.UpdateTime = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowIndex
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.UpdateTime |= int64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		default:
			iNdEx = preIndex
			skippy, err := skipIndex(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthIndex
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
func skipIndex(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowIndex
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
					return 0, ErrIntOverflowIndex
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
					return 0, ErrIntOverflowIndex
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
				return 0, ErrInvalidLengthIndex
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupIndex
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthIndex
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthIndex        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowIndex          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupIndex = fmt.Errorf("proto: unexpected end of group")
)
