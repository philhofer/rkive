// +build riak

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
	conns int32   // total live conns
	pad1  [4]byte //
	inuse int32   // conns in use
	pad2  [4]byte //
	tag   int32   // 0 = open; 1 = closed
	pad3  [4]byte //

	// BENCH-SPECIFIC
	nwait uint64 // number of roundtrips
	twait uint64 // time between Write() and Read()

	id    []byte         // client ID
	pool  sync.Pool      // connection pool
	addrs []*net.TCPAddr // node addrs
}

func (c *Client) doBuf(code byte, msg []byte) ([]byte, byte, error) {
	var retried bool
try:
	node, err := c.popConn()
	if err != nil {
		return nil, 0, err
	}

	msg[4] = code

	// this is testing-specific in order to
	// time network i/o
	// BENCHMARKING
	if !retried {
		atomic.AddUint64(&c.nwait, 1)
	}
	startwrite := time.Now()
	// BENCHMARKING

	_, err = node.Write(msg)
	if err != nil {
		go c.err(node)

		// it could be the case that we pulled
		// a bad connection from the pool - we'll
		// attempt one retry
		if !retried {
			retried = true
			goto try
		}

		return nil, 0, err
	}
	msg, code, err = readResponse(node, msg)

	// testing-specific, again
	// BENCHMARKING
	atomic.AddUint64(&c.twait, uint64(time.Since(startwrite).Nanoseconds()))
	// BENCHMARKING

	if err == nil {
		c.done(node)
	} else {
		c.err(node)
	}
	return msg, code, nil
}

func (c *Client) AvgWait() uint64 { return atomic.LoadUint64(&c.twait) / atomic.LoadUint64(&c.nwait) }
func (c *Client) TimerReset()     { atomic.StoreUint64(&c.twait, 0); atomic.StoreUint64(&c.nwait, 0) }
