package rkive

import (
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/philhofer/rkive/rpbc"
	"log"
	"math/rand"
	"net"
	"os"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

var (
	// ErrAck is returned when the
	// an attempt is made to make a request
	// with a closed clinet
	ErrAck = errors.New("client closed")
	logger = log.New(os.Stderr, "[RKIVE] ", log.LstdFlags)
)

// read timeout (ms)
const readTimeout = 1000

// write timeout (ms)
const writeTimeout = 1000

const maxConns = 20

// RiakError is an error
// returned from the Riak server
// iteself.
type RiakError struct {
	res *rpbc.RpbErrorResp
}

// ErrMultipleResponses is the type
// of error returned when multiple
// siblings are retrieved for an object.
type ErrMultipleResponses struct {
	Bucket      string
	Key         string
	NumSiblings int
}

func (m *ErrMultipleResponses) Error() string {
	return fmt.Sprintf("%d siblings found", m.NumSiblings)
}

// Blob is a generic riak key/value container
type Blob struct {
	RiakInfo *Info
	Content  []byte
}

// generate *ErrMultipleResponses from multiple contents
func handleMultiple(n int, key, bucket string) *ErrMultipleResponses {
	return &ErrMultipleResponses{
		Bucket:      bucket,
		Key:         key,
		NumSiblings: n,
	}
}

// Blob satisfies the Object interface.
func (r *Blob) Info() *Info              { return r.RiakInfo }
func (r *Blob) Unmarshal(b []byte) error { r.Content = b; return nil }
func (r *Blob) Marshal() ([]byte, error) { return r.Content, nil }

func (r RiakError) Error() string {
	return fmt.Sprintf("riak error (0): %s", r.res.GetErrmsg())
}

// Dial creates a client connected to one
// or many Riak nodes. It will try to reconnect to
// downed nodes in the background. Dial verifies
// that it is able to connect to at least one node
// before it returns; otherwise, it will return an error.
func Dial(addrs []string, clientID string) (*Client, error) {
	naddrs := make([]*net.TCPAddr, len(addrs))

	var err error
	for i, node := range addrs {
		naddrs[i], err = net.ResolveTCPAddr("tcp", node)
		if err != nil {
			return nil, err
		}
	}

	cl := &Client{
		tag:   0,
		id:    []byte(clientID),
		pool:  new(sync.Pool),
		addrs: naddrs,
	}

	return cl, nil
}

// DialOne returns a client with one TCP connection
// to a single Riak node.
func DialOne(addr string, clientID string) (*Client, error) {
	return Dial([]string{addr}, clientID)
}

// Close() closes the client. This cannot
// be reversed.
func (c *Client) Close() {
	if !atomic.CompareAndSwapInt64(&c.tag, 0, 1) {
		return
	}
	time.Sleep(10 * time.Millisecond)
	// this (hopefully) clears
	// the pool
	runtime.GC()
}

func (c *Client) closed() bool {
	return atomic.LoadInt64(&c.tag) == 1
}

// Client represents a pool of connections
// to a Riak cluster.
type Client struct {
	conns int64
	tag   int64
	id    []byte
	pool  *sync.Pool
	addrs []*net.TCPAddr
}

// can we add another connection?
// if so, increment
func (c *Client) try() bool {
	new := atomic.AddInt64(&c.conns, 1)
	if new > maxConns {
		atomic.AddInt64(&c.conns, -1)
		return false
	}
	return true
}

func (c *Client) dec() {
	atomic.AddInt64(&c.conns, -1)
}

// newconn tries to return a valid
// tcp connection to a node
func (c *Client) newconn() (*net.TCPConn, error) {
	ntry := 0
try:
	addr := c.addrs[rand.Intn(len(c.addrs))]
	logger.Printf("dialing TCP %s", addr.String())
	conn, err := net.DialTCP("tcp", nil, addr)
	if err != nil {
		ntry++
		c.dec()
		logger.Printf("error dialing TCP %s: %s", addr.String(), err)
		if ntry < len(c.addrs) {
			goto try
		}
		return nil, err
	}
	conn.SetKeepAlive(true)
	conn.SetNoDelay(true)
	err = c.writeClientID(conn)
	if err != nil {
		ntry++
		c.dec()
		conn.Close()
		logger.Printf("error writing client ID: %s", err)
		if ntry < len(c.addrs) {
			goto try
		}
		return nil, err
	}
	runtime.SetFinalizer(conn, cclose)
	return conn, nil
}

// pop connection
func (c *Client) popConn() (*net.TCPConn, error) {
	if c.closed() {
		return nil, ErrAck
	}
	// spinlock (sort of)
	// on acquiring a connection
	for {
		conn, ok := c.pool.Get().(*net.TCPConn)
		if ok || conn != nil {
			return conn, nil
		}
		if c.try() {
			return c.newconn()
		}
		runtime.Gosched()
	}
}

/*
// redialLoop is responsible
// for attempting to dial nodes
// NOTE: the client's waitgroup
// must be incremented before starting
// an async redialLoop
func (c *Client) redialLoop(n *node) {
	var nd int
	if c.closed() {
		n.Drop()
		goto exit
	}
	logger.Printf("dialing TCP: %s", n.addr.String())
	for err := n.Dial(); err != nil; nd++ {
		if c.closed() {
			n.Drop()
			goto exit
		}
		logger.Printf("dial error #%d for %s: %s", nd, n.addr.String(), err)
		time.Sleep(3 * time.Second)
	}
	c.done(n)
exit:
	c.wg.Done()
}*/

/*
// ping nodes
//
// NOTE: the client's waitgroup/
// must be incremented before starting
// an async pingLoop
//
// pingLoop spends most of its time
// sleeping if all the nodes are reachable,
// so it shouldn't cause serious issues
// with contention over c.dones
func (c *Client) pingLoop() {
	var node *node
	var err error
	var ok bool
inspect:
	for {
	check:
		if c.closed() {
			goto exit
		}
		select {
		case node, ok = <-c.dones:
			if !ok {
				goto exit
			}
			err = node.Ping()

			// we don't sleep
			// if an error is returned;
			// instead, we start a redial
			// on the node. otherwise, we
			// sleep.
			if err != nil {
				if c.closed() {
					node.Drop()
					goto exit
				}
				c.wg.Add(1)
				go c.redialLoop(node)
			} else {
				c.done(node)
				time.Sleep(2 * time.Second)
			}
			continue inspect

			// don't block forever;
			// we could be closing
		case <-time.After(1 * time.Second):
			goto check
		}
	}
exit:
	c.wg.Done()
}*/

func (c *Client) writeClientID(conn *net.TCPConn) error {
	if c.id == nil {
		return nil
	}
	req := &rpbc.RpbSetClientIdReq{
		ClientId: c.id,
	}
	bts, err := req.Marshal()
	if err != nil {
		return err
	}
	msglen := len(bts) + 1
	msg := make([]byte, msglen+4)
	binary.BigEndian.PutUint32(msg, uint32(msglen))
	msg[4] = 5 // code for RpbSetClientIdReq
	copy(msg[5:], bts)
	conn.SetWriteDeadline(time.Now().Add(writeTimeout * time.Millisecond))
	_, err = conn.Write(msg)
	if err != nil {
		return err
	}
	conn.SetReadDeadline(time.Now().Add(readTimeout * time.Millisecond))
	_, err = conn.Read(msg[:5])
	if err != nil {
		return err
	}
	// expect response code 6
	if msg[4] != 6 {
		return ErrUnexpectedResponse
	}
	return nil
}

// finish node (success)
func (c *Client) done(n *net.TCPConn) {
	c.dec()
	if c.closed() {
		n.Close()
		return
	}
	c.pool.Put(n)
}

// finish node with err
func (c *Client) err(n *net.TCPConn) {
	c.dec()
	if c.closed() {
		n.Close()
		return
	}
	err := ping(n)
	if err != nil {
		n.Close()
	}
	c.done(n)
}

// writes the message to the node with
// the appropriate leading message size
// and the given message code
func writeMsg(n *net.TCPConn, msg []byte, code byte) error {
	// bigendian length + code byte
	var lead [5]byte
	msglen := uint32(len(msg) + 1)
	binary.BigEndian.PutUint32(lead[:4], msglen)
	lead[4] = code

	// keep this on the stack -
	// don't allocate just for the
	// five byte frame
	mbd := make([]byte, len(msg)+5)
	copy(mbd, lead[:])
	copy(mbd[5:], msg)

	// send the message
	n.SetWriteDeadline(time.Now().Add(writeTimeout * time.Millisecond))
	_, err := n.Write(mbd)
	if err != nil {
		return err
	}
	return nil
}

// readLead reads the size of the inbound message
func readLead(n *net.TCPConn) (int, byte, error) {
	var lead [5]byte
	n.SetReadDeadline(time.Now().Add(readTimeout * time.Millisecond))
	_, err := n.Read(lead[:])
	if err != nil {
		return 0, lead[4], err
	}
	msglen := binary.BigEndian.Uint32(lead[:4]) - 1
	rescode := lead[4]
	return int(msglen), rescode, nil
}

// readBody reads from the node into 'body'
// body should be sized by the result from readLead
func readBody(n *net.TCPConn, body []byte) error {
	n.SetReadDeadline(time.Now().Add(readTimeout * time.Millisecond))
	_, err := n.Read(body)
	return err
}

// send the contents of a buffer and receive a response
// back in the same buffer
func (c *Client) doBuf(code byte, msg []byte) ([]byte, byte, error) {
	node, err := c.popConn()
	if err != nil {
		return nil, 0, err
	}

	err = writeMsg(node, msg, code)
	if err != nil {
		c.err(node)
		return nil, 0, err
	}

	// read lead
	var msglen int
	// read length and code
	msglen, code, err = readLead(node)
	if err != nil {
		c.err(node)
		return nil, code, err
	}
	if msglen == 0 {
		msg = msg[0:0] // mark empty (necessary for ErrNotFound)
		goto exit
	}
	// no alloc if response is smaller
	// than request
	if msglen > cap(msg) {
		msg = make([]byte, msglen)
	} else {
		msg = msg[0:msglen]
	}

	// read body
	err = readBody(node, msg)
	if err != nil {
		c.err(node)
		return msg, code, err
	}

exit:
	c.done(node)
	return msg, code, nil
}

func (c *Client) req(msg protom, code byte, res unmarshaler) (byte, error) {
	buf := getBuf() // maybe we've already allocated
	err := buf.Set(msg)
	if err != nil {
		return 0, fmt.Errorf("rkive: client.Req marshal err: %s", err)
	}
	resbts, rescode, err := c.doBuf(code, buf.Body)
	buf.Body = resbts // save the largest-cap byte slice
	if err != nil {
		putBuf(buf)
		return 0, fmt.Errorf("rkive: doBuf err: %s", err)
	}
	if rescode == 0 {
		riakerr := new(rpbc.RpbErrorResp)
		err = riakerr.Unmarshal(resbts)
		putBuf(buf)
		if err != nil {
			return 0, err
		}
		return 0, RiakError{res: riakerr}
	}
	if res != nil {
		// expected response body,
		// but we got none
		if len(resbts) == 0 {
			putBuf(buf)
			return 0, ErrNotFound
		}
		err = res.Unmarshal(resbts)
		if err != nil {
			err = fmt.Errorf("rkive: unmarshal err: %s", err)
		}
	}
	putBuf(buf) // save the bytes we allocated
	return rescode, err
}

type protoStream interface {
	Unmarshal([]byte) error
	GetDone() bool
}

type unmarshaler interface {
	Unmarshal([]byte) error
	ProtoMessage()
}

// streaming response -
// returns a primed connection
type streamRes struct {
	c    *Client
	node *net.TCPConn
}

// unmarshals; returns done / code / error
func (s *streamRes) unmarshal(res protoStream) (bool, byte, error) {
	var msglen int
	var code byte
	var err error

	msglen, code, err = readLead(s.node)
	if err != nil {
		s.c.err(s.node)
		return true, code, err
	}

	buf := getBuf()
	buf.setSz(msglen)

	// read into s.bts
	err = readBody(s.node, buf.Body)
	if err != nil {
		s.c.err(s.node)
		putBuf(buf)
		return true, code, err
	}
	// handle a code 0
	if code == 0 {
		// we're done
		s.close()

		riakerr := new(rpbc.RpbErrorResp)
		err = riakerr.Unmarshal(buf.Body)
		putBuf(buf)
		if err != nil {
			return true, 0, err
		}
		return true, 0, RiakError{res: riakerr}
	}

	err = res.Unmarshal(buf.Body)
	putBuf(buf)
	if err != nil {
		s.close()
		return true, code, err
	}
	done := res.GetDone()
	if done {
		s.close()
	}
	return done, code, nil
}

// return the connection to the client
func (s *streamRes) close() { s.c.done(s.node) }

func (c *Client) streamReq(req protom, code byte) (*streamRes, error) {

	buf := getBuf()
	err := buf.Set(req)
	//bts, err := req.Marshal()
	if err != nil {
		putBuf(buf)
		return nil, err
	}
	node, err := c.popConn()
	if err != nil {
		return nil, err
	}

	err = writeMsg(node, buf.Body, code)
	putBuf(buf)
	if err != nil {
		c.err(node)
		return nil, err
	}
	return &streamRes{c: c, node: node}, nil
}

func (c *Client) Ping() error {
	conn, err := c.popConn()
	if err != nil {
		return err
	}
	err = ping(conn)
	if err != nil {
		c.dec()
		conn.Close()
		return err
	}
	c.done(conn)
	return nil
}

func ping(conn *net.TCPConn) error {
	conn.SetWriteDeadline(time.Now().Add(writeTimeout * time.Millisecond))
	_, err := conn.Write([]byte{0, 0, 0, 1, 1})
	if err != nil {
		return err
	}
	var res [5]byte
	conn.SetReadDeadline(time.Now().Add(readTimeout * time.Millisecond))
	_, err = conn.Read(res[:])
	return err
}

func cclose(conn *net.TCPConn) {
	if conn == nil {
		return
	}
	logger.Printf("closing TCP connection to %s", conn.RemoteAddr().String())
	conn.Close()
}
