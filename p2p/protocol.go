package p2p

import (
	"bytes"
	"encoding/binary"
)

type Header struct {
	Type uint8
	Flag uint8
	Seq  uint32
	Len  uint32
}

const (
	HEADER_SIZE = uint32(10)
)

// msg to byte
func (header *Header) Serialize() ([]byte, error) {
	buf := new(bytes.Buffer)
	if err := binary.Write(buf, binary.BigEndian, *header); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// byte to msg
func (header *Header) Deserialize(b []byte) error {
	buf := bytes.NewReader(b)
	if err := binary.Read(buf, binary.BigEndian, header); err != nil {
		return err
	}
	return nil
}
