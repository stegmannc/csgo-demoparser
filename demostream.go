package main

import (
	"bytes"
	"encoding/binary"
	"github.com/golang/protobuf/proto"
	"io"
)

type DemoStream struct {
	reader   io.ReadSeeker
	position int
	length   int32
}

func NewDemoStream(reader io.ReadSeeker, length int32) *DemoStream {
	stream := DemoStream{reader: reader, position: 0, length: length}
	return &stream
}

func (d *DemoStream) GetVarInt() uint64 {
	var x uint64
	var s uint
	buf := make([]byte, 1)
	for i := 0; ; i++ {
		_, err := d.reader.Read(buf)
		d.position++
		if err != nil {
			panic(err)
		}
		if buf[0] < 0x80 {
			if i > 9 || i == 9 && buf[0] > 1 {
				panic("overflow")
			}
			return x | uint64(buf[0])<<s
		}
		x |= uint64(buf[0]&0x7f) << s
		s += 7
	}
}

func (d *DemoStream) IsProcessed() bool {
	return int32(d.position+1) >= d.length
}

func (d *DemoStream) GetCurrentOffset() int {
	return d.position
}
func (d *DemoStream) GetByte() byte {
	buf := make([]byte, 1)
	n, err := d.reader.Read(buf)
	if err != nil {
		panic(err)
	}
	d.position += n
	return buf[0]
}
func (d *DemoStream) GetInt() int32 {
	var x int32
	err := binary.Read(d.reader, binary.LittleEndian, &x)
	if err != nil {
		panic(err)
	}
	d.position += 4
	return x
}

func (d *DemoStream) GetUInt8() uint8 {
	var x uint8
	err := binary.Read(d.reader, binary.LittleEndian, &x)
	if err != nil {
		panic(err)
	}
	d.position += 1
	return x
}

func (d *DemoStream) GetDataTableString() string {
	buffer := make([]byte, 0)
	for b := d.GetByte(); b != 0; b = d.GetByte() {
		buffer = append(buffer, b)
	}

	return string(buffer)
}

func (d *DemoStream) GetInt16() int16 {
	var x int16
	err := binary.Read(d.reader, binary.LittleEndian, &x)
	if err != nil {
		panic(err)
	}
	d.position += 2
	return x
}

func (d *DemoStream) Read(out []byte) (int, error) {
	n, err := d.reader.Read(out)
	d.position += n
	return n, err
}
func (d *DemoStream) Skip(n int64) {
	d.position += int(n)
	d.reader.Seek(n, 1)
}

func (d *DemoStream) ParseToStruct(msg proto.Message, messageLength uint64) error {
	d.position += int(messageLength)
	buf := make([]byte, messageLength)
	_, err := d.Read(buf)
	if err != nil {
		return err
	}
	err = proto.Unmarshal(buf, msg)
	if err != nil {
		return err
	}
	return nil
}

func (d *DemoStream) CreatePacketStream() *DemoStream {
	packetSize := d.GetInt()
	buffer := make([]byte, packetSize)
	if _, err := d.Read(buffer); err != nil {
		panic(err)
	}
	return NewDemoStream(bytes.NewReader(buffer), packetSize)
}

func (d *DemoStream) readCommandHeader() *DemoCmdHeader {
	return &DemoCmdHeader{
		Cmd:        d.GetUInt8(),
		Tick:       d.GetInt(),
		Playerslot: d.GetUInt8(),
	}
}
