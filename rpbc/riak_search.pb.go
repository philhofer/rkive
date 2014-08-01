// Code generated by protoc-gen-gogo.
// source: riak_search.proto
// DO NOT EDIT!

/*
	Package rpbc is a generated protocol buffer package.

	It is generated from these files:
		riak_search.proto

	It has these top-level messages:
		RpbSearchDoc
		RpbSearchQueryReq
		RpbSearchQueryResp
*/
package rpbc

import proto "code.google.com/p/gogoprotobuf/proto"
import json "encoding/json"
import math "math"

// discarding unused import gogoproto "gogo.pb"

import io1 "io"
import unsafe "unsafe"
import code_google_com_p_gogoprotobuf_proto1 "code.google.com/p/gogoprotobuf/proto"

import unsafe1 "unsafe"

import bytes1 "bytes"

// Reference proto, json, and math imports to suppress error if they are not otherwise used.
var _ = proto.Marshal
var _ = &json.SyntaxError{}
var _ = math.Inf

type RpbSearchDoc struct {
	Fields           []*RpbPair `protobuf:"bytes,1,rep,name=fields" json:"fields,omitempty"`
	XXX_unrecognized []byte     `json:"-"`
}

func (m *RpbSearchDoc) Reset()         { *m = RpbSearchDoc{} }
func (m *RpbSearchDoc) String() string { return proto.CompactTextString(m) }
func (*RpbSearchDoc) ProtoMessage()    {}

func (m *RpbSearchDoc) GetFields() []*RpbPair {
	if m != nil {
		return m.Fields
	}
	return nil
}

type RpbSearchQueryReq struct {
	Q                []byte   `protobuf:"bytes,1,req,name=q" json:"q,omitempty"`
	Index            []byte   `protobuf:"bytes,2,req,name=index" json:"index,omitempty"`
	Rows             *uint32  `protobuf:"varint,3,opt,name=rows" json:"rows,omitempty"`
	Start            *uint32  `protobuf:"varint,4,opt,name=start" json:"start,omitempty"`
	Sort             []byte   `protobuf:"bytes,5,opt,name=sort" json:"sort,omitempty"`
	Filter           []byte   `protobuf:"bytes,6,opt,name=filter" json:"filter,omitempty"`
	Df               []byte   `protobuf:"bytes,7,opt,name=df" json:"df,omitempty"`
	Op               []byte   `protobuf:"bytes,8,opt,name=op" json:"op,omitempty"`
	Fl               [][]byte `protobuf:"bytes,9,rep,name=fl" json:"fl,omitempty"`
	Presort          []byte   `protobuf:"bytes,10,opt,name=presort" json:"presort,omitempty"`
	XXX_unrecognized []byte   `json:"-"`
}

func (m *RpbSearchQueryReq) Reset()         { *m = RpbSearchQueryReq{} }
func (m *RpbSearchQueryReq) String() string { return proto.CompactTextString(m) }
func (*RpbSearchQueryReq) ProtoMessage()    {}

func (m *RpbSearchQueryReq) GetQ() []byte {
	if m != nil {
		return m.Q
	}
	return nil
}

func (m *RpbSearchQueryReq) GetIndex() []byte {
	if m != nil {
		return m.Index
	}
	return nil
}

func (m *RpbSearchQueryReq) GetRows() uint32 {
	if m != nil && m.Rows != nil {
		return *m.Rows
	}
	return 0
}

func (m *RpbSearchQueryReq) GetStart() uint32 {
	if m != nil && m.Start != nil {
		return *m.Start
	}
	return 0
}

func (m *RpbSearchQueryReq) GetSort() []byte {
	if m != nil {
		return m.Sort
	}
	return nil
}

func (m *RpbSearchQueryReq) GetFilter() []byte {
	if m != nil {
		return m.Filter
	}
	return nil
}

func (m *RpbSearchQueryReq) GetDf() []byte {
	if m != nil {
		return m.Df
	}
	return nil
}

func (m *RpbSearchQueryReq) GetOp() []byte {
	if m != nil {
		return m.Op
	}
	return nil
}

func (m *RpbSearchQueryReq) GetFl() [][]byte {
	if m != nil {
		return m.Fl
	}
	return nil
}

func (m *RpbSearchQueryReq) GetPresort() []byte {
	if m != nil {
		return m.Presort
	}
	return nil
}

type RpbSearchQueryResp struct {
	Docs             []*RpbSearchDoc `protobuf:"bytes,1,rep,name=docs" json:"docs,omitempty"`
	MaxScore         *float32        `protobuf:"fixed32,2,opt,name=max_score" json:"max_score,omitempty"`
	NumFound         *uint32         `protobuf:"varint,3,opt,name=num_found" json:"num_found,omitempty"`
	XXX_unrecognized []byte          `json:"-"`
}

func (m *RpbSearchQueryResp) Reset()         { *m = RpbSearchQueryResp{} }
func (m *RpbSearchQueryResp) String() string { return proto.CompactTextString(m) }
func (*RpbSearchQueryResp) ProtoMessage()    {}

func (m *RpbSearchQueryResp) GetDocs() []*RpbSearchDoc {
	if m != nil {
		return m.Docs
	}
	return nil
}

func (m *RpbSearchQueryResp) GetMaxScore() float32 {
	if m != nil && m.MaxScore != nil {
		return *m.MaxScore
	}
	return 0
}

func (m *RpbSearchQueryResp) GetNumFound() uint32 {
	if m != nil && m.NumFound != nil {
		return *m.NumFound
	}
	return 0
}

func init() {
}
func (m *RpbSearchDoc) Unmarshal(data []byte) error {
	l := len(data)
	index := 0
	for index < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if index >= l {
				return io1.ErrUnexpectedEOF
			}
			b := data[index]
			index++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return proto.ErrWrongType
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if index >= l {
					return io1.ErrUnexpectedEOF
				}
				b := data[index]
				index++
				msglen |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			postIndex := index + msglen
			if postIndex > l {
				return io1.ErrUnexpectedEOF
			}
			m.Fields = append(m.Fields, &RpbPair{})
			m.Fields[len(m.Fields)-1].Unmarshal(data[index:postIndex])
			index = postIndex
		default:
			var sizeOfWire int
			for {
				sizeOfWire++
				wire >>= 7
				if wire == 0 {
					break
				}
			}
			index -= sizeOfWire
			skippy, err := code_google_com_p_gogoprotobuf_proto1.Skip(data[index:])
			if err != nil {
				return err
			}
			if (index + skippy) > l {
				return io1.ErrUnexpectedEOF
			}
			m.XXX_unrecognized = append(m.XXX_unrecognized, data[index:index+skippy]...)
			index += skippy
		}
	}
	return nil
}
func (m *RpbSearchQueryReq) Unmarshal(data []byte) error {
	l := len(data)
	index := 0
	for index < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if index >= l {
				return io1.ErrUnexpectedEOF
			}
			b := data[index]
			index++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return proto.ErrWrongType
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if index >= l {
					return io1.ErrUnexpectedEOF
				}
				b := data[index]
				index++
				byteLen |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			postIndex := index + byteLen
			if postIndex > l {
				return io1.ErrUnexpectedEOF
			}
			m.Q = append(m.Q, data[index:postIndex]...)
			index = postIndex
		case 2:
			if wireType != 2 {
				return proto.ErrWrongType
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if index >= l {
					return io1.ErrUnexpectedEOF
				}
				b := data[index]
				index++
				byteLen |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			postIndex := index + byteLen
			if postIndex > l {
				return io1.ErrUnexpectedEOF
			}
			m.Index = append(m.Index, data[index:postIndex]...)
			index = postIndex
		case 3:
			if wireType != 0 {
				return proto.ErrWrongType
			}
			var v uint32
			for shift := uint(0); ; shift += 7 {
				if index >= l {
					return io1.ErrUnexpectedEOF
				}
				b := data[index]
				index++
				v |= (uint32(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			m.Rows = &v
		case 4:
			if wireType != 0 {
				return proto.ErrWrongType
			}
			var v uint32
			for shift := uint(0); ; shift += 7 {
				if index >= l {
					return io1.ErrUnexpectedEOF
				}
				b := data[index]
				index++
				v |= (uint32(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			m.Start = &v
		case 5:
			if wireType != 2 {
				return proto.ErrWrongType
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if index >= l {
					return io1.ErrUnexpectedEOF
				}
				b := data[index]
				index++
				byteLen |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			postIndex := index + byteLen
			if postIndex > l {
				return io1.ErrUnexpectedEOF
			}
			m.Sort = append(m.Sort, data[index:postIndex]...)
			index = postIndex
		case 6:
			if wireType != 2 {
				return proto.ErrWrongType
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if index >= l {
					return io1.ErrUnexpectedEOF
				}
				b := data[index]
				index++
				byteLen |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			postIndex := index + byteLen
			if postIndex > l {
				return io1.ErrUnexpectedEOF
			}
			m.Filter = append(m.Filter, data[index:postIndex]...)
			index = postIndex
		case 7:
			if wireType != 2 {
				return proto.ErrWrongType
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if index >= l {
					return io1.ErrUnexpectedEOF
				}
				b := data[index]
				index++
				byteLen |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			postIndex := index + byteLen
			if postIndex > l {
				return io1.ErrUnexpectedEOF
			}
			m.Df = append(m.Df, data[index:postIndex]...)
			index = postIndex
		case 8:
			if wireType != 2 {
				return proto.ErrWrongType
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if index >= l {
					return io1.ErrUnexpectedEOF
				}
				b := data[index]
				index++
				byteLen |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			postIndex := index + byteLen
			if postIndex > l {
				return io1.ErrUnexpectedEOF
			}
			m.Op = append(m.Op, data[index:postIndex]...)
			index = postIndex
		case 9:
			if wireType != 2 {
				return proto.ErrWrongType
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if index >= l {
					return io1.ErrUnexpectedEOF
				}
				b := data[index]
				index++
				byteLen |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			postIndex := index + byteLen
			if postIndex > l {
				return io1.ErrUnexpectedEOF
			}
			m.Fl = append(m.Fl, make([]byte, postIndex-index))
			copy(m.Fl[len(m.Fl)-1], data[index:postIndex])
			index = postIndex
		case 10:
			if wireType != 2 {
				return proto.ErrWrongType
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if index >= l {
					return io1.ErrUnexpectedEOF
				}
				b := data[index]
				index++
				byteLen |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			postIndex := index + byteLen
			if postIndex > l {
				return io1.ErrUnexpectedEOF
			}
			m.Presort = append(m.Presort, data[index:postIndex]...)
			index = postIndex
		default:
			var sizeOfWire int
			for {
				sizeOfWire++
				wire >>= 7
				if wire == 0 {
					break
				}
			}
			index -= sizeOfWire
			skippy, err := code_google_com_p_gogoprotobuf_proto1.Skip(data[index:])
			if err != nil {
				return err
			}
			if (index + skippy) > l {
				return io1.ErrUnexpectedEOF
			}
			m.XXX_unrecognized = append(m.XXX_unrecognized, data[index:index+skippy]...)
			index += skippy
		}
	}
	return nil
}
func (m *RpbSearchQueryResp) Unmarshal(data []byte) error {
	l := len(data)
	index := 0
	for index < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if index >= l {
				return io1.ErrUnexpectedEOF
			}
			b := data[index]
			index++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return proto.ErrWrongType
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if index >= l {
					return io1.ErrUnexpectedEOF
				}
				b := data[index]
				index++
				msglen |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			postIndex := index + msglen
			if postIndex > l {
				return io1.ErrUnexpectedEOF
			}
			m.Docs = append(m.Docs, &RpbSearchDoc{})
			m.Docs[len(m.Docs)-1].Unmarshal(data[index:postIndex])
			index = postIndex
		case 2:
			if wireType != 5 {
				return proto.ErrWrongType
			}
			var v float32
			if index+4 > l {
				return io1.ErrUnexpectedEOF
			}
			v = *(*float32)(unsafe.Pointer(&data[index]))
			index += 4
			m.MaxScore = &v
		case 3:
			if wireType != 0 {
				return proto.ErrWrongType
			}
			var v uint32
			for shift := uint(0); ; shift += 7 {
				if index >= l {
					return io1.ErrUnexpectedEOF
				}
				b := data[index]
				index++
				v |= (uint32(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			m.NumFound = &v
		default:
			var sizeOfWire int
			for {
				sizeOfWire++
				wire >>= 7
				if wire == 0 {
					break
				}
			}
			index -= sizeOfWire
			skippy, err := code_google_com_p_gogoprotobuf_proto1.Skip(data[index:])
			if err != nil {
				return err
			}
			if (index + skippy) > l {
				return io1.ErrUnexpectedEOF
			}
			m.XXX_unrecognized = append(m.XXX_unrecognized, data[index:index+skippy]...)
			index += skippy
		}
	}
	return nil
}
func (m *RpbSearchDoc) Marshal() (data []byte, err error) {
	size := m.Size()
	data = make([]byte, size)
	n, err := m.MarshalTo(data)
	if err != nil {
		return nil, err
	}
	return data[:n], nil
}

func (m *RpbSearchDoc) MarshalTo(data []byte) (n int, err error) {
	var i int
	_ = i
	var l int
	_ = l
	if len(m.Fields) > 0 {
		for _, msg := range m.Fields {
			data[i] = 0xa
			i++
			i = encodeVarintRiakSearch(data, i, uint64(msg.Size()))
			n, err := msg.MarshalTo(data[i:])
			if err != nil {
				return 0, err
			}
			i += n
		}
	}
	if m.XXX_unrecognized != nil {
		i += copy(data[i:], m.XXX_unrecognized)
	}
	return i, nil
}
func (m *RpbSearchQueryReq) Marshal() (data []byte, err error) {
	size := m.Size()
	data = make([]byte, size)
	n, err := m.MarshalTo(data)
	if err != nil {
		return nil, err
	}
	return data[:n], nil
}

func (m *RpbSearchQueryReq) MarshalTo(data []byte) (n int, err error) {
	var i int
	_ = i
	var l int
	_ = l
	if m.Q != nil {
		data[i] = 0xa
		i++
		i = encodeVarintRiakSearch(data, i, uint64(len(m.Q)))
		i += copy(data[i:], m.Q)
	}
	if m.Index != nil {
		data[i] = 0x12
		i++
		i = encodeVarintRiakSearch(data, i, uint64(len(m.Index)))
		i += copy(data[i:], m.Index)
	}
	if m.Rows != nil {
		data[i] = 0x18
		i++
		i = encodeVarintRiakSearch(data, i, uint64(*m.Rows))
	}
	if m.Start != nil {
		data[i] = 0x20
		i++
		i = encodeVarintRiakSearch(data, i, uint64(*m.Start))
	}
	if m.Sort != nil {
		data[i] = 0x2a
		i++
		i = encodeVarintRiakSearch(data, i, uint64(len(m.Sort)))
		i += copy(data[i:], m.Sort)
	}
	if m.Filter != nil {
		data[i] = 0x32
		i++
		i = encodeVarintRiakSearch(data, i, uint64(len(m.Filter)))
		i += copy(data[i:], m.Filter)
	}
	if m.Df != nil {
		data[i] = 0x3a
		i++
		i = encodeVarintRiakSearch(data, i, uint64(len(m.Df)))
		i += copy(data[i:], m.Df)
	}
	if m.Op != nil {
		data[i] = 0x42
		i++
		i = encodeVarintRiakSearch(data, i, uint64(len(m.Op)))
		i += copy(data[i:], m.Op)
	}
	if len(m.Fl) > 0 {
		for _, b := range m.Fl {
			data[i] = 0x4a
			i++
			i = encodeVarintRiakSearch(data, i, uint64(len(b)))
			i += copy(data[i:], b)
		}
	}
	if m.Presort != nil {
		data[i] = 0x52
		i++
		i = encodeVarintRiakSearch(data, i, uint64(len(m.Presort)))
		i += copy(data[i:], m.Presort)
	}
	if m.XXX_unrecognized != nil {
		i += copy(data[i:], m.XXX_unrecognized)
	}
	return i, nil
}
func (m *RpbSearchQueryResp) Marshal() (data []byte, err error) {
	size := m.Size()
	data = make([]byte, size)
	n, err := m.MarshalTo(data)
	if err != nil {
		return nil, err
	}
	return data[:n], nil
}

func (m *RpbSearchQueryResp) MarshalTo(data []byte) (n int, err error) {
	var i int
	_ = i
	var l int
	_ = l
	if len(m.Docs) > 0 {
		for _, msg := range m.Docs {
			data[i] = 0xa
			i++
			i = encodeVarintRiakSearch(data, i, uint64(msg.Size()))
			n, err := msg.MarshalTo(data[i:])
			if err != nil {
				return 0, err
			}
			i += n
		}
	}
	if m.MaxScore != nil {
		data[i] = 0x15
		i++
		*(*float32)(unsafe1.Pointer(&data[i])) = *m.MaxScore
		i += 4
	}
	if m.NumFound != nil {
		data[i] = 0x18
		i++
		i = encodeVarintRiakSearch(data, i, uint64(*m.NumFound))
	}
	if m.XXX_unrecognized != nil {
		i += copy(data[i:], m.XXX_unrecognized)
	}
	return i, nil
}
func encodeFixed64RiakSearch(data []byte, offset int, v uint64) int {
	data[offset] = uint8(v)
	data[offset+1] = uint8(v >> 8)
	data[offset+2] = uint8(v >> 16)
	data[offset+3] = uint8(v >> 24)
	data[offset+4] = uint8(v >> 32)
	data[offset+5] = uint8(v >> 40)
	data[offset+6] = uint8(v >> 48)
	data[offset+7] = uint8(v >> 56)
	return offset + 8
}
func encodeFixed32RiakSearch(data []byte, offset int, v uint32) int {
	data[offset] = uint8(v)
	data[offset+1] = uint8(v >> 8)
	data[offset+2] = uint8(v >> 16)
	data[offset+3] = uint8(v >> 24)
	return offset + 4
}
func encodeVarintRiakSearch(data []byte, offset int, v uint64) int {
	for v >= 1<<7 {
		data[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	data[offset] = uint8(v)
	return offset + 1
}
func (m *RpbSearchDoc) Size() (n int) {
	var l int
	_ = l
	if len(m.Fields) > 0 {
		for _, e := range m.Fields {
			l = e.Size()
			n += 1 + l + sovRiakSearch(uint64(l))
		}
	}
	if m.XXX_unrecognized != nil {
		n += len(m.XXX_unrecognized)
	}
	return n
}
func (m *RpbSearchQueryReq) Size() (n int) {
	var l int
	_ = l
	if m.Q != nil {
		l = len(m.Q)
		n += 1 + l + sovRiakSearch(uint64(l))
	}
	if m.Index != nil {
		l = len(m.Index)
		n += 1 + l + sovRiakSearch(uint64(l))
	}
	if m.Rows != nil {
		n += 1 + sovRiakSearch(uint64(*m.Rows))
	}
	if m.Start != nil {
		n += 1 + sovRiakSearch(uint64(*m.Start))
	}
	if m.Sort != nil {
		l = len(m.Sort)
		n += 1 + l + sovRiakSearch(uint64(l))
	}
	if m.Filter != nil {
		l = len(m.Filter)
		n += 1 + l + sovRiakSearch(uint64(l))
	}
	if m.Df != nil {
		l = len(m.Df)
		n += 1 + l + sovRiakSearch(uint64(l))
	}
	if m.Op != nil {
		l = len(m.Op)
		n += 1 + l + sovRiakSearch(uint64(l))
	}
	if len(m.Fl) > 0 {
		for _, b := range m.Fl {
			l = len(b)
			n += 1 + l + sovRiakSearch(uint64(l))
		}
	}
	if m.Presort != nil {
		l = len(m.Presort)
		n += 1 + l + sovRiakSearch(uint64(l))
	}
	if m.XXX_unrecognized != nil {
		n += len(m.XXX_unrecognized)
	}
	return n
}
func (m *RpbSearchQueryResp) Size() (n int) {
	var l int
	_ = l
	if len(m.Docs) > 0 {
		for _, e := range m.Docs {
			l = e.Size()
			n += 1 + l + sovRiakSearch(uint64(l))
		}
	}
	if m.MaxScore != nil {
		n += 5
	}
	if m.NumFound != nil {
		n += 1 + sovRiakSearch(uint64(*m.NumFound))
	}
	if m.XXX_unrecognized != nil {
		n += len(m.XXX_unrecognized)
	}
	return n
}

func sovRiakSearch(x uint64) (n int) {
	for {
		n++
		x >>= 7
		if x == 0 {
			break
		}
	}
	return n
}
func sozRiakSearch(x uint64) (n int) {
	return sovRiakSearch(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func NewPopulatedRpbSearchDoc(r randyRiakSearch, easy bool) *RpbSearchDoc {
	this := &RpbSearchDoc{}
	if r.Intn(10) != 0 {
		v1 := r.Intn(10)
		this.Fields = make([]*RpbPair, v1)
		for i := 0; i < v1; i++ {
			this.Fields[i] = NewPopulatedRpbPair(r, easy)
		}
	}
	if !easy && r.Intn(10) != 0 {
		this.XXX_unrecognized = randUnrecognizedRiakSearch(r, 2)
	}
	return this
}

func NewPopulatedRpbSearchQueryReq(r randyRiakSearch, easy bool) *RpbSearchQueryReq {
	this := &RpbSearchQueryReq{}
	v2 := r.Intn(100)
	this.Q = make([]byte, v2)
	for i := 0; i < v2; i++ {
		this.Q[i] = byte(r.Intn(256))
	}
	v3 := r.Intn(100)
	this.Index = make([]byte, v3)
	for i := 0; i < v3; i++ {
		this.Index[i] = byte(r.Intn(256))
	}
	if r.Intn(10) != 0 {
		v4 := r.Uint32()
		this.Rows = &v4
	}
	if r.Intn(10) != 0 {
		v5 := r.Uint32()
		this.Start = &v5
	}
	if r.Intn(10) != 0 {
		v6 := r.Intn(100)
		this.Sort = make([]byte, v6)
		for i := 0; i < v6; i++ {
			this.Sort[i] = byte(r.Intn(256))
		}
	}
	if r.Intn(10) != 0 {
		v7 := r.Intn(100)
		this.Filter = make([]byte, v7)
		for i := 0; i < v7; i++ {
			this.Filter[i] = byte(r.Intn(256))
		}
	}
	if r.Intn(10) != 0 {
		v8 := r.Intn(100)
		this.Df = make([]byte, v8)
		for i := 0; i < v8; i++ {
			this.Df[i] = byte(r.Intn(256))
		}
	}
	if r.Intn(10) != 0 {
		v9 := r.Intn(100)
		this.Op = make([]byte, v9)
		for i := 0; i < v9; i++ {
			this.Op[i] = byte(r.Intn(256))
		}
	}
	if r.Intn(10) != 0 {
		v10 := r.Intn(100)
		this.Fl = make([][]byte, v10)
		for i := 0; i < v10; i++ {
			v11 := r.Intn(100)
			this.Fl[i] = make([]byte, v11)
			for j := 0; j < v11; j++ {
				this.Fl[i][j] = byte(r.Intn(256))
			}
		}
	}
	if r.Intn(10) != 0 {
		v12 := r.Intn(100)
		this.Presort = make([]byte, v12)
		for i := 0; i < v12; i++ {
			this.Presort[i] = byte(r.Intn(256))
		}
	}
	if !easy && r.Intn(10) != 0 {
		this.XXX_unrecognized = randUnrecognizedRiakSearch(r, 11)
	}
	return this
}

func NewPopulatedRpbSearchQueryResp(r randyRiakSearch, easy bool) *RpbSearchQueryResp {
	this := &RpbSearchQueryResp{}
	if r.Intn(10) != 0 {
		v13 := r.Intn(10)
		this.Docs = make([]*RpbSearchDoc, v13)
		for i := 0; i < v13; i++ {
			this.Docs[i] = NewPopulatedRpbSearchDoc(r, easy)
		}
	}
	if r.Intn(10) != 0 {
		v14 := r.Float32()
		if r.Intn(2) == 0 {
			v14 *= -1
		}
		this.MaxScore = &v14
	}
	if r.Intn(10) != 0 {
		v15 := r.Uint32()
		this.NumFound = &v15
	}
	if !easy && r.Intn(10) != 0 {
		this.XXX_unrecognized = randUnrecognizedRiakSearch(r, 4)
	}
	return this
}

type randyRiakSearch interface {
	Float32() float32
	Float64() float64
	Int63() int64
	Int31() int32
	Uint32() uint32
	Intn(n int) int
}

func randUTF8RuneRiakSearch(r randyRiakSearch) rune {
	res := rune(r.Uint32() % 1112064)
	if 55296 <= res {
		res += 2047
	}
	return res
}
func randStringRiakSearch(r randyRiakSearch) string {
	v16 := r.Intn(100)
	tmps := make([]rune, v16)
	for i := 0; i < v16; i++ {
		tmps[i] = randUTF8RuneRiakSearch(r)
	}
	return string(tmps)
}
func randUnrecognizedRiakSearch(r randyRiakSearch, maxFieldNumber int) (data []byte) {
	l := r.Intn(5)
	for i := 0; i < l; i++ {
		wire := r.Intn(4)
		if wire == 3 {
			wire = 5
		}
		fieldNumber := maxFieldNumber + r.Intn(100)
		data = randFieldRiakSearch(data, r, fieldNumber, wire)
	}
	return data
}
func randFieldRiakSearch(data []byte, r randyRiakSearch, fieldNumber int, wire int) []byte {
	key := uint32(fieldNumber)<<3 | uint32(wire)
	switch wire {
	case 0:
		data = encodeVarintPopulateRiakSearch(data, uint64(key))
		v17 := r.Int63()
		if r.Intn(2) == 0 {
			v17 *= -1
		}
		data = encodeVarintPopulateRiakSearch(data, uint64(v17))
	case 1:
		data = encodeVarintPopulateRiakSearch(data, uint64(key))
		data = append(data, byte(r.Intn(256)), byte(r.Intn(256)), byte(r.Intn(256)), byte(r.Intn(256)), byte(r.Intn(256)), byte(r.Intn(256)), byte(r.Intn(256)), byte(r.Intn(256)))
	case 2:
		data = encodeVarintPopulateRiakSearch(data, uint64(key))
		ll := r.Intn(100)
		data = encodeVarintPopulateRiakSearch(data, uint64(ll))
		for j := 0; j < ll; j++ {
			data = append(data, byte(r.Intn(256)))
		}
	default:
		data = encodeVarintPopulateRiakSearch(data, uint64(key))
		data = append(data, byte(r.Intn(256)), byte(r.Intn(256)), byte(r.Intn(256)), byte(r.Intn(256)))
	}
	return data
}
func encodeVarintPopulateRiakSearch(data []byte, v uint64) []byte {
	for v >= 1<<7 {
		data = append(data, uint8(uint64(v)&0x7f|0x80))
		v >>= 7
	}
	data = append(data, uint8(v))
	return data
}
func (this *RpbSearchDoc) Equal(that interface{}) bool {
	if that == nil {
		if this == nil {
			return true
		}
		return false
	}

	that1, ok := that.(*RpbSearchDoc)
	if !ok {
		return false
	}
	if that1 == nil {
		if this == nil {
			return true
		}
		return false
	} else if this == nil {
		return false
	}
	if len(this.Fields) != len(that1.Fields) {
		return false
	}
	for i := range this.Fields {
		if !this.Fields[i].Equal(that1.Fields[i]) {
			return false
		}
	}
	if !bytes1.Equal(this.XXX_unrecognized, that1.XXX_unrecognized) {
		return false
	}
	return true
}
func (this *RpbSearchQueryReq) Equal(that interface{}) bool {
	if that == nil {
		if this == nil {
			return true
		}
		return false
	}

	that1, ok := that.(*RpbSearchQueryReq)
	if !ok {
		return false
	}
	if that1 == nil {
		if this == nil {
			return true
		}
		return false
	} else if this == nil {
		return false
	}
	if !bytes1.Equal(this.Q, that1.Q) {
		return false
	}
	if !bytes1.Equal(this.Index, that1.Index) {
		return false
	}
	if this.Rows != nil && that1.Rows != nil {
		if *this.Rows != *that1.Rows {
			return false
		}
	} else if this.Rows != nil {
		return false
	} else if that1.Rows != nil {
		return false
	}
	if this.Start != nil && that1.Start != nil {
		if *this.Start != *that1.Start {
			return false
		}
	} else if this.Start != nil {
		return false
	} else if that1.Start != nil {
		return false
	}
	if !bytes1.Equal(this.Sort, that1.Sort) {
		return false
	}
	if !bytes1.Equal(this.Filter, that1.Filter) {
		return false
	}
	if !bytes1.Equal(this.Df, that1.Df) {
		return false
	}
	if !bytes1.Equal(this.Op, that1.Op) {
		return false
	}
	if len(this.Fl) != len(that1.Fl) {
		return false
	}
	for i := range this.Fl {
		if !bytes1.Equal(this.Fl[i], that1.Fl[i]) {
			return false
		}
	}
	if !bytes1.Equal(this.Presort, that1.Presort) {
		return false
	}
	if !bytes1.Equal(this.XXX_unrecognized, that1.XXX_unrecognized) {
		return false
	}
	return true
}
func (this *RpbSearchQueryResp) Equal(that interface{}) bool {
	if that == nil {
		if this == nil {
			return true
		}
		return false
	}

	that1, ok := that.(*RpbSearchQueryResp)
	if !ok {
		return false
	}
	if that1 == nil {
		if this == nil {
			return true
		}
		return false
	} else if this == nil {
		return false
	}
	if len(this.Docs) != len(that1.Docs) {
		return false
	}
	for i := range this.Docs {
		if !this.Docs[i].Equal(that1.Docs[i]) {
			return false
		}
	}
	if this.MaxScore != nil && that1.MaxScore != nil {
		if *this.MaxScore != *that1.MaxScore {
			return false
		}
	} else if this.MaxScore != nil {
		return false
	} else if that1.MaxScore != nil {
		return false
	}
	if this.NumFound != nil && that1.NumFound != nil {
		if *this.NumFound != *that1.NumFound {
			return false
		}
	} else if this.NumFound != nil {
		return false
	} else if that1.NumFound != nil {
		return false
	}
	if !bytes1.Equal(this.XXX_unrecognized, that1.XXX_unrecognized) {
		return false
	}
	return true
}
