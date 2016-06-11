package main

import (
	"encoding/binary"
	"io"
	"github.com/golang/protobuf/proto"
)

type Demostream struct {
	reader   io.ReadSeeker
	position int
}

func (d *Demostream) GetVarInt() uint64 {
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
func (d *Demostream) GetCurrentOffset() int {
	return d.position
}
func (d *Demostream) GetByte() byte {
	buf := make([]byte, 1)
	n, err := d.reader.Read(buf)
	if err != nil {
		panic(err)
	}
	d.position += n
	return buf[0]
}
func (d *Demostream) GetInt() int32 {
	var x int32
	err := binary.Read(d.reader, binary.LittleEndian, &x)
	if err != nil {
		panic(err)
	}
	d.position += 4
	return x
}

func (d *Demostream) GetDataTableString() string {
	buffer := make([]byte, 0)
	for b := d.GetByte(); b != 0; b =d.GetByte() {
		buffer = append(buffer, b)
	}

	return string(buffer)
}

func (d *Demostream) GetInt16() int16 {
	var x int16
	err := binary.Read(d.reader, binary.LittleEndian, &x)
	if err != nil {
		panic(err)
	}
	d.position += 2
	return x
}
func NewDemoStream(reader io.ReadSeeker) *Demostream {
	stream := Demostream{reader: reader, position: 0}
	return &stream
}
func (d *Demostream) Read(out []byte) (int, error) {
	n, err := d.reader.Read(out)
	d.position += n
	return n, err
}
func (d *Demostream) Skip(n int64) {
	d.position += int(n)
	d.reader.Seek(n, 1)
}

func (d *Demostream) ParseToStruct(msg proto.Message, messageLength uint64) (error){
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