package rkive

import (
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/philhofer/rkive/rpbc"
	"log"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

var (
	// ErrAck is returned when all of the connections
	// in the connection pool are unavailable for longer
	// than 250ms, or when the pool is closing
	ErrAck = errors.New("connection unavailable")
)

// read timeout (ms)
const readTimeout = 1000

// write timeout (ms)
const writeTimeout = 1000

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

// A Node is a Riak physical node.
type Node struct {
	Addr   string // address (e.g. 127.0.0.1:8078)
	NConns uint   // Number of simultaneous TCP connections
}

// Dial creates a client connected to one
// or many Riak nodes. It will try to reconnect to
// downed nodes in the background. Dial verifies
// that it is able to connect to at least one node
// before it returns; otherwise, it will return an error.
func Dial(nodes []Node, clientID string) (*Client, error) {

	// count total connections
	nconns := 0
	for _, node := range nodes {
		nconns += int(node.NConns)
	}

	naddrs := make([]*net.TCPAddr, len(nodes))

	var err error
	for i, node := range nodes {
		naddrs[i], err = net.ResolveTCPAddr("tcp", node.Addr)
		if err != nil {
			return nil, err
		}
	}

	cl := &Client{
		tag:   0,
		id:    []byte(clientID),
		dones: make(chan *node, nconns),
		lock:  make(chan struct{}, 3),
		wg:    new(sync.WaitGroup),
	}

	// dial up all the nodes
	for i, naddr := range naddrs {
		for j := 0; j < int(nodes[i].NConns); j++ {
			tnode := &node{
				addr:        naddr,
				parent:      cl,
				isConnected: false,
				conn:        nil,
			}
			cl.wg.Add(1)
			go cl.redialLoop(tnode)
		}
	}
	cl.wg.Add(1)
	go cl.pingLoop()
	cl.dunlock()
	
	// make sure we're able to dial
	// at least one of the nodes after
	// 5 seconds
	select {
	case n := <- cl.dones:
	        cl.dones <- n
	case <-time.After(5 * time.Second):
	        cl.Close()
	        return nil, errors.New("unable to dial any nodes")
	}
	
	return cl, nil
}

// DialOne returns a client with one TCP connection
// to a single Riak node.
func DialOne(addr string, clientID string) (*Client, error) {
	return Dial([]Node{{Addr: addr, NConns: 1}}, clientID)
}

// Close() closes the client. This cannot
// be reversed.
func (c *Client) Close() {
	if !atomic.CompareAndSwapInt64(&c.tag, 0, 1) {
		return
	}
	c.dlock()
	close(c.dones)
	for node := range c.dones {
		log.Printf("Closing TCP tunnel to %s", node.addr.String())
		node.conn.Close()
	}
	c.dunlock()
	c.wg.Wait()
	return
}

// lock completely (for closing)
func (c *Client) dlock() {
	<-c.lock
	<-c.lock
	<-c.lock
}

// unlock completely
func (c *Client) dunlock() {
	c.lock <- struct{}{}
	c.lock <- struct{}{}
	c.lock <- struct{}{}
}

func (c *Client) closed() bool {
	return atomic.LoadInt64(&c.tag) == 1
}

type Client struct {
	tag   int64
	id    []byte
	dones chan *node    // holds good nodes
	lock  chan struct{} // used as RWmutex
	wg    *sync.WaitGroup
}

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
	log.Printf("dialing TCP: %s", n.addr.String())
	for err := n.Dial(); err != nil; nd++ {
		if c.closed() {
			n.Drop()
			goto exit
		}
		log.Printf("dial error #%d for %s: %s", nd, n.addr.String(), err)
		time.Sleep(3 * time.Second)
	}
	c.done(n)
exit:
	c.wg.Done()
}

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
}

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

// acquire node
func (c *Client) ack() (*node, bool) {
	node, ok := <-c.dones
	return node, ok
}

// finish node (success)
func (c *Client) done(n *node) {
	// we need to lock
	// in order ensure that the
	// channel cannot be closed
	// while sending
	<-c.lock
	if c.closed() {
		c.lock <- struct{}{}
		n.Drop()
		return
	}
	c.dones <- n
	c.lock <- struct{}{}
}

// finish node with err
func (c *Client) err(n *node) {
	if c.closed() {
		n.Drop()
		return
	}
	// do a quick test
	err := n.Ping()
	if err != nil {
		c.wg.Add(1)
		go c.redialLoop(n)
		return
	}
	<-c.lock
	c.dones <- n
	c.lock <- struct{}{}
}

// writes the message to the node with
// the appropriate leading message size
// and the given message code
func writeMsg(n *node, msg []byte, code byte) error {
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
	_, err := n.Write(mbd)
	if err != nil {
		n.Err()
		return err
	}
	return nil
}

// readLead reads the size of the inbound message
func readLead(n *node) (int, byte, error) {
	var lead [5]byte
	_, err := n.Read(lead[:])
	if err != nil {
		n.Err()
		return 0, lead[4], err
	}
	msglen := binary.BigEndian.Uint32(lead[:4]) - 1
	rescode := lead[4]
	return int(msglen), rescode, nil
}

// readBody reads from the node into 'body'
// body should be sized by the result from readLead
func readBody(n *node, body []byte) error {
	_, err := n.Read(body)
	if err != nil {
		n.Err()
	}
	return err
}

// send the contents of a buffer and receive a response
// back in the same buffer
func (c *Client) doBuf(code byte, msg []byte) ([]byte, byte, error) {
	node, ok := c.ack()
	if !ok {
		return nil, 0, ErrAck
	}
	var err error

	err = writeMsg(node, msg, code)
	if err != nil {
		return nil, 0, err
	}

	// read lead
	var msglen int
	// read length and code
	msglen, code, err = readLead(node)
	if err != nil {
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
		return msg, code, err
	}

exit:
	node.Done()
	return msg, code, nil
}

func (c *Client) req(msg protom, code byte, res unmarshaler) (byte, error) {
	buf := getBuf() // maybe we've already allocated
	err := buf.Set(msg)
	if err != nil {
		return 0, fmt.Errorf("riakpb: client.Req marshal err: %s", err)
	}
	resbts, rescode, err := c.doBuf(code, buf.Body)
	buf.Body = resbts // save the largest-cap byte slice
	if err != nil {
		putBuf(buf)
		return 0, fmt.Errorf("riakpb: doBuf err: %s", err)
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
			err = fmt.Errorf("riakpb: unmarshal err: %s", err)
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
	node *node
}

// unmarshals; returns done / code / error
func (s *streamRes) unmarshal(res protoStream) (bool, byte, error) {
	var msglen int
	var code byte
	var err error

	msglen, code, err = readLead(s.node)
	if err != nil {
		return true, code, err
	}

	buf := getBuf()
	buf.setSz(msglen)

	// read into s.bts
	err = readBody(s.node, buf.Body)
	if err != nil {
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
func (s *streamRes) close() { s.node.Done() }

func (c *Client) streamReq(req protom, code byte) (*streamRes, error) {

	buf := getBuf()
	err := buf.Set(req)
	//bts, err := req.Marshal()
	if err != nil {
		putBuf(buf)
		return nil, err
	}
	node, ok := c.ack()
	if !ok {
		putBuf(buf)
		return nil, ErrAck
	}

	err = writeMsg(node, buf.Body, code)
	putBuf(buf)
	if err != nil {
		return nil, err
	}
	return &streamRes{c: c, node: node}, nil
}

func (c *Client) Ping() error {
	node, ok := c.ack()
	if !ok {
		return ErrAck
	}
	err := node.Ping()
	if err != nil {
		node.Err()
		return err
	}
	node.Done()
	return nil
}

func (n *node) Ping() error {
	_, err := n.Write([]byte{0, 0, 0, 1, 1})
	if err != nil {
		return err
	}
	var res [5]byte
	_, err = n.Read(res[:])
	if err != nil {
		return err
	}
	return nil
}
