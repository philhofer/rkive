package riakpb

import (
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

func (b *buf) Set(p protom) error {
	sz := p.Size()
	var err error
	if cap(b.Body) >= sz {
		b.Body = b.Body[0:sz]
		_, err = p.MarshalTo(b.Body)
	} else {
		b.Body = make([]byte, sz)
		_, err = p.MarshalTo(b.Body)
	}
	return err
}

func getBuf() *buf {
	b, ok := bufPool.Get().(*buf)
	if !ok {
		return new(buf)
	}
	return b
}

func putBuf(b *buf) { bufPool.Put(b) }
