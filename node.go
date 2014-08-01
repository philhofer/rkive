package riakpb

import (
	"log"
	"net"
	"time"
)

type node struct {
	isConnected bool
	parent      *Client
	addr        *net.TCPAddr
	conn        *net.TCPConn
}

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
		log.Printf("Riakpb: Error dialing %s: %s", n.addr.String(), err)
		return err
	}

	err = n.parent.writeClientID(n.conn)
	if err != nil {
		n.conn.Close()
		log.Printf("Riak error: %s", err)
		return err
	}
	n.conn.SetKeepAlive(true)
	n.isConnected = true
	return nil
}

func (n *node) IsConnected() bool {
	return n.isConnected
}

func (n *node) Write(b []byte) (int, error) {
	n.conn.SetWriteDeadline(time.Now().Add(writeTimeout * time.Millisecond))
	return n.conn.Write(b)
}

func (n *node) Read(b []byte) (int, error) {
	n.conn.SetReadDeadline(time.Now().Add(readTimeout * time.Millisecond))
	return n.conn.Read(b)
}

func (n *node) Done() {
	n.parent.done(n)
}

func (n *node) Err() {
	n.parent.err(n)
}

func (n *node) Drop() {
	log.Printf("Closing TCP connection to %s", n.addr.String())
	n.conn.Close()
	n.isConnected = false
}
