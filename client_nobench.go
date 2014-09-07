// +build !riak

package rkive

import (
	"net"
	"sync"
)

// Client represents a pool of connections
// to a Riak cluster.
type Client struct {
	conns int32          // total live conns
	pad1  [4]byte        //
	inuse int32          // conns in use
	pad2  [4]byte        //
	tag   int32          // 0 = open; 1 = closed; others reserved
	pad3  [4]byte        //
	id    []byte         // client ID for writeClientID
	pool  sync.Pool      // connection pool
	addrs []*net.TCPAddr // addresses to dial
}

func (c *Client) doBuf(code byte, msg []byte) ([]byte, byte, error) {
	var retried bool
try:
	node, err := c.popConn()
	if err != nil {
		return nil, 0, err
	}

	msg[4] = code
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
	if err != nil {
		c.err(node)
		return nil, 0, err
	}
	msg, code, err = readResponse(node, msg)
	if err == nil {
		c.done(node)
	} else {
		c.err(node)
	}
	return msg, code, nil
}
