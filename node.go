package rkive

import (
	"net"
	"time"
)

// node is a single TCP
// connection to a single physical
// Riak node
type node struct {
	isConnected bool
	parent      *Client
	addr        *net.TCPAddr
	conn        *net.TCPConn
}

// Dial attempts to open the connection
// (closes already-open connections)
func (n *node) Dial() error {
	if n.isConnected {
		if n.conn != nil {
			n.conn.Close()
		}
		n.isConnected = false
	}

	var err error
	n.conn, err = net.DialTCP("tcp", nil, n.addr)
	if err != nil {
		logger.Printf("rkive: error dialing %s: %s", n.addr.String(), err)
		return err
	}

	err = n.parent.writeClientID(n.conn)
	if err != nil {
		n.conn.Close()
		logger.Printf("rkive: error writing client ID: %s", err)
		return err
	}
	n.conn.SetKeepAlive(true)
	n.isConnected = true
	return nil
}

// deadlined write
func (n *node) Write(b []byte) (int, error) {
	n.conn.SetWriteDeadline(time.Now().Add(writeTimeout * time.Millisecond))
	return n.conn.Write(b)
}

// deadlined read
func (n *node) Read(b []byte) (int, error) {
	n.conn.SetReadDeadline(time.Now().Add(readTimeout * time.Millisecond))
	return n.conn.Read(b)
}

// returns the node to the parent pool
func (n *node) Done() {
	n.parent.done(n)
}

// returns the node to the parent
// pool with a suspected error
func (n *node) Err() {
	go n.parent.err(n)
}

// drop closes the node connection
func (n *node) Drop() {
	if !n.isConnected {
		return
	}
	logger.Printf("closing TCP connection to %s", n.addr.String())
	n.conn.Close()
	n.isConnected = false
}
