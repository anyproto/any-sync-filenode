// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: index/redisindex/indexproto/protos/index.proto

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
}

func (m *CidEntry) Reset()         { *m = CidEntry{} }
func (m *CidEntry) String() string { return proto.CompactTextString(m) }
func (*CidEntry) ProtoMessage()    {}
func (*CidEntry) Descriptor() ([]byte, []int) {
	return fileDescriptor_01af1a9166444478, []int{0}
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

type CidList struct {
	Cids [][]byte `protobuf:"bytes,1,rep,name=cids,proto3" json:"cids,omitempty"`
}

func (m *CidList) Reset()         { *m = CidList{} }
func (m *CidList) String() string { return proto.CompactTextString(m) }
func (*CidList) ProtoMessage()    {}
func (*CidList) Descriptor() ([]byte, []int) {
	return fileDescriptor_01af1a9166444478, []int{1}
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

func init() {
	proto.RegisterType((*CidEntry)(nil), "fileIndexProto.CidEntry")
	proto.RegisterType((*CidList)(nil), "fileIndexProto.CidList")
}

func init() {
	proto.RegisterFile("index/redisindex/indexproto/protos/index.proto", fileDescriptor_01af1a9166444478)
}

var fileDescriptor_01af1a9166444478 = []byte{
	// 206 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0xd2, 0xcb, 0xcc, 0x4b, 0x49,
	0xad, 0xd0, 0x2f, 0x4a, 0x4d, 0xc9, 0x2c, 0x86, 0x30, 0xc1, 0x64, 0x41, 0x51, 0x7e, 0x49, 0xbe,
	0x3e, 0x98, 0x2c, 0x86, 0x88, 0xe8, 0x81, 0x39, 0x42, 0x7c, 0x69, 0x99, 0x39, 0xa9, 0x9e, 0x20,
	0x81, 0x00, 0x10, 0x5f, 0xa9, 0x88, 0x8b, 0xc3, 0x39, 0x33, 0xc5, 0x35, 0xaf, 0xa4, 0xa8, 0x52,
	0x48, 0x88, 0x8b, 0xa5, 0x38, 0xb3, 0x2a, 0x55, 0x82, 0x51, 0x81, 0x51, 0x83, 0x25, 0x08, 0xcc,
	0x16, 0x92, 0xe3, 0xe2, 0x4a, 0x2e, 0x4a, 0x4d, 0x2c, 0x49, 0x0d, 0xc9, 0xcc, 0x4d, 0x95, 0x60,
	0x52, 0x60, 0xd4, 0x60, 0x0e, 0x42, 0x12, 0x01, 0xc9, 0x97, 0x16, 0xa4, 0xc0, 0xe4, 0x99, 0x21,
	0xf2, 0x08, 0x11, 0x90, 0x99, 0x45, 0xa9, 0x69, 0xc5, 0x12, 0x2c, 0x0a, 0x8c, 0x1a, 0xac, 0x41,
	0x60, 0xb6, 0x92, 0x2c, 0x17, 0xbb, 0x73, 0x66, 0x8a, 0x4f, 0x66, 0x71, 0x09, 0x48, 0x3a, 0x39,
	0x33, 0xa5, 0x58, 0x82, 0x51, 0x81, 0x59, 0x83, 0x27, 0x08, 0xcc, 0x76, 0x32, 0x3d, 0xf1, 0x48,
	0x8e, 0xf1, 0xc2, 0x23, 0x39, 0xc6, 0x07, 0x8f, 0xe4, 0x18, 0x27, 0x3c, 0x96, 0x63, 0xb8, 0xf0,
	0x58, 0x8e, 0xe1, 0xc6, 0x63, 0x39, 0x86, 0x28, 0x69, 0x3c, 0x9e, 0x4d, 0x62, 0x03, 0x53, 0xc6,
	0x80, 0x00, 0x00, 0x00, 0xff, 0xff, 0xd6, 0xa3, 0x20, 0x4c, 0x12, 0x01, 0x00, 0x00,
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
