// +build bench

package rkive

import (
	"fmt"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

func init() {
	fmt.Println("RKIVE: BENCHMARK MODE")
}

// Client represents a pool of connections
// to a Riak cluster.
type Client struct {
	conns int32 // total live conns
	pad1  [4]byte
	inuse int32 // conns in use
	pad2  [4]byte
	tag   int32 // 0 = open; 1 = closed
	pad3  [4]byte

	// BENCH-SPECIFIC
	nwait uint64 // number of tcp.Read/Write calls
	twait uint64 // time in iowait

	id    []byte
	pool  sync.Pool
	addrs []*net.TCPAddr
}

func (c *Client) doBuf(code byte, msg []byte) ([]byte, byte, error) {
	node, err := c.popConn()
	if err != nil {
		return nil, 0, err
	}

	msg[4] = code
	atomic.AddUint64(&c.nwait, 1)
	startwrite := time.Now()
	_, err = node.Write(msg)
	if err != nil {
		c.err(node)
		return nil, 0, err
	}
	if err != nil {
		c.err(node)
		return nil, 0, err
	}
	msg, code, err = readResponse(node, msg)
	atomic.AddUint64(&c.twait, uint64(time.Since(startwrite).Nanoseconds()))
	if err == nil {
		c.done(node)
	} else {
		c.err(node)
	}
	return msg, code, nil
}

func (c *Client) AvgWait() uint64 { return atomic.LoadUint64(&c.twait) / atomic.LoadUint64(&c.nwait) }
