package riakpb

import (
	"code.google.com/p/gogoprotobuf/proto"
	"encoding/binary"
	"fmt"
	"github.com/philhofer/riakpb/rpbc"
	"log"
	"net"
	"time"
)

// read timeout (ms)
const readTimeout = 1000

// write timeout (ms)
const writeTimeout = 1000

type RiakError struct {
	res *rpbc.RpbErrorResp
}

// ErrMultipleResponses is the type
// of error returned when multiple
// siblings are retrieved for an object.
type ErrMultipleResponses struct {
	Responses []*Blob
}

func (m *ErrMultipleResponses) Error() string {
	return fmt.Sprintf("%d siblings found", len(m.Responses))
}

// Blob is a generic riak container
type Blob struct {
	RiakInfo *Info
	Content  []byte
}

// generate *ErrMultipleResponses from multiple contents
func handleMultiple(vs []*rpbc.RpbContent) *ErrMultipleResponses {
	nc := len(vs)
	em := &ErrMultipleResponses{
		Responses: make([]*Blob, nc),
	}
	for i, ctnt := range vs {
		em.Responses[i] = &Blob{RiakInfo: &Info{}, Content: nil}
		_ = readContent(em.Responses[i], ctnt)
	}
	return em
}

// Blob satisfies the Object interface.
func (r *Blob) Info() *Info              { return r.RiakInfo }
func (r *Blob) Unmarshal(b []byte) error { r.Content = b; return nil }
func (r *Blob) Marshal() ([]byte, error) { return r.Content, nil }

func (r RiakError) Error() string {
	return fmt.Sprintf("riak error (0): %s", r.res.GetErrmsg())
}

func Dial(addrs []string, clientID string) (*Client, error) {
	cl := &Client{
		id:    []byte(clientID),
		dones: make(chan *node, len(addrs)),
	}

	naddrs := make([]*net.TCPAddr, len(addrs))

	var err error
	for i, addr := range addrs {
		naddrs[i], err = net.ResolveTCPAddr("tcp", addr)
		if err != nil {
			return nil, err
		}
	}

	for _, naddr := range naddrs {
		tnode := &node{
			addr:        naddr,
			parent:      cl,
			isConnected: false,
			conn:        nil,
		}
		go cl.redialLoop(tnode)
	}

	return cl, nil
}

func (c *Client) Close() {
	close(c.dones)
	for node := range c.dones {
		log.Printf("Closing TCP tunnel to %s", node.addr.String())
		node.conn.Close()
	}
	return
}

type Client struct {
	id    []byte
	dones chan *node // holds good nodes
}

func (c *Client) redialLoop(n *node) {
	log.Printf("Redialing TCP %s...", n.addr.String())
	var nd int
	for err := n.Dial(); err != nil; nd++ {
		log.Printf("Dial error #%d: %s", n, err)
		time.Sleep(3 * time.Second)
	}
	c.dones <- n
}

func (c *Client) writeClientID(conn *net.TCPConn) error {
	if c.id == nil {
		return nil
	}
	req := &rpbc.RpbSetClientIdReq{
		ClientId: c.id,
	}
	bts, err := proto.Marshal(req)
	if err != nil {
		return err
	}
	msglen := len(bts) + 1
	msg := make([]byte, msglen+4)
	binary.BigEndian.PutUint32(msg, uint32(msglen))
	msg[4] = 5 // code for RpbSetClientIdReq
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
func (c *Client) ack() *node {
	return <-c.dones
}

// finish node (success)
func (c *Client) done(n *node) {
	c.dones <- n
}

// finish node (error)
func (c *Client) err(n *node) {
	go c.redialLoop(n)
}

func writeMsg(c *Client, n *node, msg []byte, code byte) ([]byte, error) {
	// bigendian length + code byte
	var lead [5]byte

	msglen := uint32(len(msg) + 1)
	binary.BigEndian.PutUint32(lead[:4], msglen)
	lead[4] = code
	msg = append(lead[:], msg...)

	// send the message
	_, err := n.Write(msg)
	if err != nil {
		n.Err()
		return msg, err
	}
	return msg, nil
}

func readLead(c *Client, n *node) (int, byte, error) {
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

func readBody(c *Client, n *node, body []byte) error {
	_, err := n.Read(body)
	if err != nil {
		n.Err()
		return err
	}
	return nil
}

// send the contents of a buffer and receive a response
// back in the same buffer
func (c *Client) doBuf(code byte, msg []byte) ([]byte, byte, error) {
	node := c.ack()
	var err error

	msg, err = writeMsg(c, node, msg, code)
	if err != nil {
		return nil, 0, err
	}

	// read lead
	var msglen int
	// read length and code
	msglen, code, err = readLead(c, node)
	if err != nil {
		return nil, code, err
	}
	if msglen > cap(msg) {
		msg = make([]byte, int(msglen))
	} else {
		msg = msg[0:msglen]
	}
	if msglen == 0 {
		goto exit
	}

	// read body
	err = readBody(c, node, msg)
	if err != nil {
		return msg, code, err
	}

exit:
	node.Done()
	return msg, code, nil
}

func (c *Client) req(msg proto.Message, code byte, res proto.Message) (byte, error) {
	bts, err := proto.Marshal(msg)
	if err != nil {
		return 0, fmt.Errorf("riakpb: client.Req marshal err: %s", err)
	}
	resbts, rescode, err := c.doBuf(code, bts)
	if err != nil {
		return 0, fmt.Errorf("riakpb: doBuf err: %s", err)
	}
	if rescode == 0 {
		riakerr := new(rpbc.RpbErrorResp)
		err = proto.Unmarshal(resbts, riakerr)
		if err != nil {
			return 0, err
		}
		return 0, RiakError{res: riakerr}
	}
	if res != nil {
		err = proto.Unmarshal(resbts, res)
		if err != nil {
			err = fmt.Errorf("riakpb: unmarshal err: %s", err)
		}
	}
	return rescode, err
}

type protoStream interface {
	proto.Message
	GetDone() bool
}

// streaming response -
// returns a primed connection
type streamRes struct {
	c    *Client
	node *node
	bts  []byte
}

// unmarshals; returns done / code / error
func (s *streamRes) unmarshal(res protoStream) (bool, byte, error) {
	var msglen int
	var code byte
	var err error

	msglen, code, err = readLead(s.c, s.node)
	if err != nil {
		return true, code, err
	}

	if msglen > cap(s.bts) {
		s.bts = make([]byte, msglen)
	} else {
		s.bts = s.bts[0:msglen]
	}

	err = readBody(s.c, s.node, s.bts)
	if err != nil {
		return true, code, err
	}
	// handle a code 0
	if code == 0 {
		// we're done
		s.close()

		riakerr := new(rpbc.RpbErrorResp)
		err = proto.Unmarshal(s.bts, riakerr)
		if err != nil {
			return true, 0, err
		}
		return true, 0, RiakError{res: riakerr}
	}

	err = proto.Unmarshal(s.bts, res)
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

func (c *Client) streamReq(req proto.Message, code byte) (*streamRes, error) {
	node := c.ack()
	var err error

	msg, err := proto.Marshal(req)
	if err != nil {
		return nil, err
	}
	msg, err = writeMsg(c, node, msg, code)
	if err != nil {
		return nil, err
	}
	return &streamRes{c: c, node: node, bts: msg}, nil
}

func (c *Client) Ping() error {
	node := c.ack()
	var err error

	_, err = node.Write([]byte{0, 0, 0, 1, 1})
	if err != nil {
		return err
	}
	var res [5]byte
	_, err = node.Read(res[:])
	if err != nil {
		return err
	}
	node.Done()
	return nil
}
