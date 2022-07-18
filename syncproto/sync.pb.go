// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: syncproto/proto/sync.proto

package syncproto

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

type SyncMessage struct {
	Seq int64 `protobuf:"varint,1,opt,name=seq,proto3" json:"seq,omitempty"`
}

func (m *SyncMessage) Reset()         { *m = SyncMessage{} }
func (m *SyncMessage) String() string { return proto.CompactTextString(m) }
func (*SyncMessage) ProtoMessage()    {}
func (*SyncMessage) Descriptor() ([]byte, []int) {
	return fileDescriptor_4b28dfdd48a89166, []int{0}
}
func (m *SyncMessage) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *SyncMessage) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_SyncMessage.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *SyncMessage) XXX_Merge(src proto.Message) {
	xxx_messageInfo_SyncMessage.Merge(m, src)
}
func (m *SyncMessage) XXX_Size() int {
	return m.Size()
}
func (m *SyncMessage) XXX_DiscardUnknown() {
	xxx_messageInfo_SyncMessage.DiscardUnknown(m)
}

var xxx_messageInfo_SyncMessage proto.InternalMessageInfo

func (m *SyncMessage) GetSeq() int64 {
	if m != nil {
		return m.Seq
	}
	return 0
}

func init() {
	proto.RegisterType((*SyncMessage)(nil), "anytype.SyncMessage")
}

func init() { proto.RegisterFile("syncproto/proto/sync.proto", fileDescriptor_4b28dfdd48a89166) }

var fileDescriptor_4b28dfdd48a89166 = []byte{
	// 119 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0x92, 0x2a, 0xae, 0xcc, 0x4b,
	0x2e, 0x28, 0xca, 0x2f, 0xc9, 0xd7, 0x87, 0x90, 0x20, 0xbe, 0x1e, 0x98, 0x29, 0xc4, 0x9e, 0x98,
	0x57, 0x59, 0x52, 0x59, 0x90, 0xaa, 0x24, 0xcf, 0xc5, 0x1d, 0x5c, 0x99, 0x97, 0xec, 0x9b, 0x5a,
	0x5c, 0x9c, 0x98, 0x9e, 0x2a, 0x24, 0xc0, 0xc5, 0x5c, 0x9c, 0x5a, 0x28, 0xc1, 0xa8, 0xc0, 0xa8,
	0xc1, 0x1c, 0x04, 0x62, 0x3a, 0xa9, 0x9c, 0x78, 0x24, 0xc7, 0x78, 0xe1, 0x91, 0x1c, 0xe3, 0x83,
	0x47, 0x72, 0x8c, 0x13, 0x1e, 0xcb, 0x31, 0x5c, 0x78, 0x2c, 0xc7, 0x70, 0xe3, 0xb1, 0x1c, 0x43,
	0x14, 0x97, 0x3e, 0xdc, 0x82, 0x24, 0x36, 0x30, 0x65, 0x0c, 0x08, 0x00, 0x00, 0xff, 0xff, 0x7e,
	0xbc, 0x46, 0xd2, 0x74, 0x00, 0x00, 0x00,
}

func (m *SyncMessage) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *SyncMessage) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *SyncMessage) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.Seq != 0 {
		i = encodeVarintSync(dAtA, i, uint64(m.Seq))
		i--
		dAtA[i] = 0x8
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
func (m *SyncMessage) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.Seq != 0 {
		n += 1 + sovSync(uint64(m.Seq))
	}
	return n
}

func sovSync(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozSync(x uint64) (n int) {
	return sovSync(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *SyncMessage) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: SyncMessage: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: SyncMessage: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Seq", wireType)
			}
			m.Seq = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowSync
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Seq |= int64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
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