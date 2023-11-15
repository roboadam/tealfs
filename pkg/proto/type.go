package proto

// import "encoding/binary"

// func Hello() NetType {
// 	return NetType{Value: 1}
// }

// func NodeSync() NetType {
// 	return NetType{Value: 2}
// }

// type NetType struct {
// 	Value uint8
// }

// type Header struct {
// 	Typ NetType
// 	Len uint32
// }

// type Payload struct {
// 	Type NetType
// 	Data []byte
// }

// const HeaderLen = 5

// func (header *Header) ToBytes() []byte {
// 	buf := make([]byte, 5)
// 	buf[0] = header.Typ.Value
// 	binary.BigEndian.PutUint32(buf[1:], header.Len)
// 	return buf
// }

// func HeaderFromBytes(data []byte) (Header, []byte) {
// 	typ := NetType{Value: data[0]}
// 	len, remainder := IntFromBytes(data[1:])
// 	return Header{Typ: typ, Len: len}, remainder
// }
