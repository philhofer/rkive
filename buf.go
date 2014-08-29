package rkive

import (
	"encoding/binary"
	"sync"
)

var bufPool *sync.Pool

func init() {
	bufPool = new(sync.Pool)
	bufPool.New = func() interface{} {
		return new(buf)
	}
}

type buf struct {
	Body []byte
}

// proto message w/ Size and MarshalTo
type protom interface {
	MarshalTo([]byte) (n int, err error)
	Size() int
}

// opportunistic MarshalTo; leaves Body[4] open for code
func (b *buf) Set(p protom) error {
	sz := p.Size()
	bsz := sz + 5
	if cap(b.Body) >= bsz {
		b.Body = b.Body[0:bsz]
	} else {
		b.Body = make([]byte, bsz)
	}
	binary.BigEndian.PutUint32(b.Body, uint32(sz+1))
	_, err := p.MarshalTo(b.Body[5:])
	return err
}

func (b *buf) setSz(n int) {
	if cap(b.Body) >= n {
		b.Body = b.Body[0:n]
	} else {
		b.Body = make([]byte, n)
	}
}

func getBuf() *buf {
	return bufPool.Get().(*buf)
}

func putBuf(b *buf) { bufPool.Put(b) }
