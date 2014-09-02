package rkive

import (
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/philhofer/rkive/rpbc"
	"io"
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
	// ErrClosed is returned when the
	// an attempt is made to make a request
	// with a closed clinet
	ErrClosed = errors.New("client closed")

	// ErrUnavail is returned when the client
	// is unable to successfully dial any
	// Riak node.
	ErrUnavail = errors.New("no connection to could be established")

	logger = log.New(os.Stderr, "[RKIVE] ", log.LstdFlags)
)

// read timeout (ms)
const readTimeout = 1000

// write timeout (ms)
const writeTimeout = 1000

// max connection limit
const maxConns = 30

// RiakError is an error
// returned from the Riak server
// iteself.
type RiakError struct {
	res *rpbc.RpbErrorResp
}

func (r RiakError) Error() string {
	return fmt.Sprintf("riak error (0): %s", r.res.GetErrmsg())
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

// Blob is a generic riak key/value container that
// implements the Object interface.
type Blob struct {
	RiakInfo Info
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

// Info implements part of the Object interface.
func (r *Blob) Info() *Info { return &r.RiakInfo }

// Unmarshal implements part of the Object interface
func (r *Blob) Unmarshal(b []byte) error { r.Content = b; return nil }

// Marshal implements part of the Object interface
func (r *Blob) Marshal() ([]byte, error) { return r.Content, nil }

// conn is a connection
type conn struct {
	*net.TCPConn
	parent   *Client
	isClosed bool
}

// write wraps the TCP write
func (c *conn) Write(b []byte) (int, error) {
	c.SetWriteDeadline(time.Now().Add(writeTimeout * time.Millisecond))
	return c.TCPConn.Write(b)
}

// read wraps the TCP read
func (c *conn) Read(b []byte) (int, error) {
	c.SetReadDeadline(time.Now().Add(readTimeout * time.Millisecond))
	return c.TCPConn.Read(b)
}

// Close idempotently closes
// the connection and decrements
// the parent conn counter
func (c *conn) Close() {
	if c.isClosed {
		return
	}
	c.isClosed = true
	logger.Printf("closing TCP connection to %s", c.RemoteAddr().String())
	c.Close()
	c.parent.dec()
}

// Dial creates a client connected to one
// or many Riak nodes. The client will attempt
// to avoid using downed nodes. Dial returns an error
// if it is unable to reach a good node.
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

	// fail on no dial-able nodes
	err = cl.Ping()
	if err != nil {
		cl.Close()
		return nil, err
	}

	return cl, nil
}

// DialOne returns a client that
// always dials the same node. (See: Dial)
func DialOne(addr string, clientID string) (*Client, error) {
	return Dial([]string{addr}, clientID)
}

// Close() idempotently closes the client.
func (c *Client) Close() {
	if !atomic.CompareAndSwapInt32(&c.tag, 0, 1) {
		return
	}

	// wait for all connetions to end
	// up in the pool
	for atomic.LoadInt32(&c.inuse) > 0 {
		time.Sleep(2 * time.Millisecond)
	}

	// we'll hang if we don't make
	// the connection pool immediately
	// GC-able
	c.pool = nil

	runtime.GC()
	nspin := 0
	maxspin := 50
	// give up after 100ms just in case GC
	// fails to work as desired.
	for ; atomic.LoadInt32(&c.conns) > 0 && nspin < maxspin; nspin++ {
		time.Sleep(2 * time.Millisecond)
	}
	nc := atomic.LoadInt32(&c.conns)
	if nc > 0 {
		logger.Printf("unable to close %d conns after 100ms", nc)
	}
}

func (c *Client) closed() bool {
	return atomic.LoadInt32(&c.tag) == 1
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
	id    []byte
	pool  *sync.Pool
	addrs []*net.TCPAddr
}

// can we add another connection?
// if so, increment
func (c *Client) try() bool {
	new := atomic.AddInt32(&c.conns, 1)
	if new > maxConns {
		atomic.AddInt32(&c.conns, -1)
		return false
	}
	return true
}

// decrement conn counter
// MUST BE CALLED WHENEVER A CONNECTION
// IS CLOSED, OR WE WILL HAVE PROBLEMS.
func (c *Client) dec() {
	atomic.AddInt32(&c.conns, -1)
}

// newconn tries to return a valid
// tcp connection to a node, dropping
// failed connections. it should only
// be called by popConn().
func (c *Client) newconn() (*conn, error) {

	// randomly shuffle the list
	// of addresses and then dial
	// them in (shuffled) order until
	// success
	perm := rand.Perm(len(c.addrs))

	for _, v := range perm {
		addr := c.addrs[v]
		logger.Printf("dialing TCP %s", addr)
		tcpconn, err := net.DialTCP("tcp", nil, addr)
		if err != nil {
			logger.Printf("error dialing %s: %s", addr, err)
			continue
		}
		tcpconn.SetKeepAlive(true)
		tcpconn.SetNoDelay(true)
		out := &conn{
			TCPConn:  tcpconn,
			parent:   c,
			isClosed: false,
		}
		err = c.writeClientID(out)
		if err != nil {
			// call the tcp connection's
			// close method, because otherwise
			// the client conn counter will
			// be decremented
			out.TCPConn.Close()
			logger.Printf("error writing client ID: %s", err)
			continue
		}
		runtime.SetFinalizer(out, (*conn).Close)
		return out, nil
	}
	c.dec()
	return nil, ErrUnavail
}

// pop connection
func (c *Client) popConn() (*conn, error) {
	// spinlock (sort of)
	// on acquiring a connection
	for {
		if c.closed() {
			return nil, ErrClosed
		}
		cn, ok := c.pool.Get().(*conn)
		if ok && cn != nil {
			atomic.AddInt32(&c.inuse, 1)
			return cn, nil
		}
		if c.try() {
			cn, err := c.newconn()
			if err != nil {
				return nil, err
			}
			atomic.AddInt32(&c.inuse, 1)
			return cn, nil
		}
		runtime.Gosched()
	}
}

func (c *Client) writeClientID(cn *conn) error {
	if c.id == nil {
		// writeClientID is used
		// to test if a node is actually
		// live, so we need to do *something*
		return ping(cn)
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
	_, err = cn.Write(msg)
	if err != nil {
		return err
	}
	_, err = io.ReadFull(cn, msg[:5])
	if err != nil {
		return err
	}
	// expect response code 6
	if msg[4] != 6 {
		return ErrUnexpectedResponse
	}
	return nil
}

// readLead reads the size of the inbound message
func readLead(n *conn) (int, byte, error) {
	var lead [5]byte
	_, err := io.ReadFull(n, lead[:])
	if err != nil {
		return 0, lead[4], err
	}
	msglen := binary.BigEndian.Uint32(lead[:4]) - 1
	rescode := lead[4]
	return int(msglen), rescode, nil
}

// read response into 'b'; truncate or append if necessary
func readResponse(c *conn, b []byte) ([]byte, byte, error) {
	var n int
	var nn int
	nn, err := c.Read(b)
	n += nn
	if err != nil {
		return nil, 0, err
	}
	mlen := int(binary.BigEndian.Uint32(b[:4]) - 1)
	var scratch [512]byte
	for n < (mlen + 5) {
		nn, err = c.Read(scratch[:])
		n += nn
		if err != nil {
			return nil, b[4], err
		}
		b = append(b, scratch[:nn]...)
	}
	return b[5:n], b[4], err
}

// send the contents of a buffer and receive a response
// back in the same buffer
func (c *Client) doBuf(code byte, msg []byte) ([]byte, byte, error) {
	node, err := c.popConn()
	if err != nil {
		return nil, 0, err
	}

	msg[4] = code
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
	if err != nil {
		return nil, code, err
	}
	/*
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
		_, err = io.ReadFull(node, msg)
		if err != nil {
			c.err(node)
			return msg, code, err
		}
	*/
	//exit:
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
	node *conn
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
	_, err = io.ReadFull(s.node, buf.Body)
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
	if err != nil {
		putBuf(buf)
		return nil, err
	}
	node, err := c.popConn()
	if err != nil {
		return nil, err
	}

	buf.Body[4] = code
	_, err = node.Write(buf.Body)
	putBuf(buf)
	if err != nil {
		c.err(node)
		return nil, err
	}
	return &streamRes{c: c, node: node}, nil
}

// Ping pings a random node. It will
// return an error if no nodes can be dialed.
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

func ping(cn *conn) error {
	_, err := cn.Write([]byte{0, 0, 0, 1, 1})
	if err != nil {
		return err
	}
	var res [5]byte
	_, err = io.ReadFull(cn, res[:])
	return err
}
