package gtype

import (
	"crypto/rand"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"sync/atomic"
	"time"
)

var guider = &guidGenerator{
	process: processUniqueBytes(),
	counter: readRandomUint32(),
}

func NewGuid() string {
	//uuid := make([]byte, 16)
	//n, err := io.ReadFull(rand.Reader, uuid)
	//if n != len(uuid) || err != nil {
	//	return ""
	//}

	//uuid[8] = uuid[8]&^0xc0 | 0x80
	//uuid[6] = uuid[6]&^0xf0 | 0x40

	//return fmt.Sprintf("%x%x%x%x%x", uuid[0:4], uuid[4:6], uuid[6:8], uuid[8:10], uuid[10:])

	id := guider.NewIDFromTimestamp(time.Now())
	return hex.EncodeToString(id[:])
}

type guidGenerator struct {
	process [5]byte
	counter uint32
}

func (s *guidGenerator) NewIDFromTimestamp(timestamp time.Time) [12]byte {
	var b [12]byte

	binary.BigEndian.PutUint32(b[0:4], uint32(timestamp.Unix()))
	copy(b[4:9], s.process[:])
	s.putUint24(b[9:12], atomic.AddUint32(&s.counter, 1))

	return b
}

func (s *guidGenerator) putUint24(b []byte, v uint32) {
	b[0] = byte(v >> 16)
	b[1] = byte(v >> 8)
	b[2] = byte(v)
}

func processUniqueBytes() [5]byte {
	var b [5]byte
	_, err := io.ReadFull(rand.Reader, b[:])
	if err != nil {
		panic(fmt.Errorf("cannot initialize guid package with crypto.rand.Reader: %v", err))
	}

	return b
}

func readRandomUint32() uint32 {
	var b [4]byte
	_, err := io.ReadFull(rand.Reader, b[:])
	if err != nil {
		panic(fmt.Errorf("cannot initialize guid package with crypto.rand.Reader: %v", err))
	}

	return (uint32(b[0]) << 0) | (uint32(b[1]) << 8) | (uint32(b[2]) << 16) | (uint32(b[3]) << 24)
}
