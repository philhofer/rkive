package riakpb

import (
	"code.google.com/p/gogoprotobuf/proto"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/philhofer/riakpb/rpbc"
	"io"
	"log"
	"net"
	"syscall"
	"time"
)

// default open connections
const dfltConns = 3

// read timeout (ms)
const readTimeout = 1000

// write timeout (ms)
const writeTimeout = 1000

var (
	ErrWriteTimeout = errors.New("TCP write timeout")
	ErrReadTimeout  = errors.New("TCP read timeout")
)

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
	info    *Info
	Content []byte
}

// generate *ErrMultipleResponses from multiple contents
func handleMultiple(vs []*rpbc.RpbContent) *ErrMultipleResponses {
	nc := len(vs)
	em := &ErrMultipleResponses{
		Responses: make([]*Blob, nc),
	}
	for i, ctnt := range vs {
		em.Responses[i] = &Blob{info: &Info{}, Content: nil}
		_ = readContent(em.Responses[i], ctnt)
	}
	return em
}

// Blob satisfies the Object interface.
func (r *Blob) Info() *Info                      { return r.info }
func (r *Blob) Unmarshal(b []byte) error         { r.Content = b; return nil }
func (r *Blob) Marshal(_ []byte) ([]byte, error) { return r.Content, nil }

func (r RiakError) Error() string {
	return fmt.Sprintf("riak error (0): %s", r.res.GetErrmsg())
}

func NewClient(addr string, clientID string, nconns *int) (*Client, error) {
	if nconns == nil {
		nconns = new(int)
		*nconns = dfltConns
	}

	netaddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		return nil, err
	}

	cl := &Client{
		addr:  netaddr,
		conns: make(chan *net.TCPConn, *nconns),
		id:    []byte(clientID),
	}

	// store conns until initialization
	// is completed successfully
	temp := make([]*net.TCPConn, *nconns)
	for i := 0; i < *nconns; i++ {
		conn, err := cl.dial()
		if err != nil {
			// cleanup other connections
			for j := i - 1; j >= 0; j-- {
				temp[j].Close()
				return nil, err
			}
		}
		temp[i] = conn
	}
	log.Printf("Successfully opened %d connections to %s", *nconns, addr)
	// send
	for _, conn := range temp {
		cl.conns <- conn
	}
	return cl, nil
}

type Client struct {
	addr  *net.TCPAddr
	conns chan *net.TCPConn
	id    []byte
}

func (c *Client) dial() (*net.TCPConn, error) {
	log.Printf("Dialing %s...", c.addr)
	conn, err := net.DialTCP("tcp", nil, c.addr)
	if err != nil {
		return nil, err
	}
	err = c.writeClientID(conn)
	if err != nil {
		conn.Close()
		return nil, err
	}
	conn.SetKeepAlive(true)
	return conn, nil
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
	conn.SetWriteDeadline(time.Now().Add(DefaultReqTimeout * time.Millisecond))
	_, err = conn.Write(msg)
	if err != nil {
		return err
	}
	conn.SetReadDeadline(time.Now().Add(DefaultReqTimeout * time.Millisecond))
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

func (c *Client) ack() (*net.TCPConn, error) {
	var con *net.TCPConn
	select {
	case con = <-c.conns:
		return con, nil
	case <-time.After(50 * time.Millisecond):
		con, err := c.dial()
		return con, err
	}
}

func (c *Client) done(ct *net.TCPConn) {
	select {
	case c.conns <- ct:
	default:
	}
}

func writeMsg(c *Client, ct *net.TCPConn, msg []byte, code byte) ([]byte, error) {
	// bigendian length + code byte
	var lead [5]byte

	msglen := uint32(len(msg) + 1)
	binary.BigEndian.PutUint32(lead[:4], msglen)
	lead[4] = code
	msg = append(lead[:], msg...)

	// send the message
	ct.SetWriteDeadline(time.Now().Add(writeTimeout * time.Millisecond))
	_, err := ct.Write(msg)
	if err != nil {
		if err == io.EOF {
			return msg, err
		}
		if operr, ok := err.(*net.OpError); ok {
			if operr.Temporary() {
				c.done(ct)
				return msg, err
			}
			if errno, ok := operr.Err.(syscall.Errno); ok {
				if errno == syscall.EPIPE {
					ct.Close()
					return msg, err
				}
			}
		}
	}
	return msg, nil
}

func readLead(c *Client, ct *net.TCPConn) (int, byte, error) {
	var lead [5]byte
	ct.SetReadDeadline(time.Now().Add(readTimeout * time.Millisecond))
	_, err := ct.Read(lead[:])
	if err != nil {
		if err == io.EOF {
			return 0, 0, err
		}
		if operr, ok := err.(*net.OpError); ok {
			if operr.Temporary() {
				c.done(ct)
				return 0, 0, err
			}
			if errno, ok := operr.Err.(syscall.Errno); ok {
				if errno == syscall.EPIPE {
					ct.Close()
					return 0, 0, err
				}
			}
		}
	}
	msglen := binary.BigEndian.Uint32(lead[:4]) - 1
	rescode := lead[4]
	return int(msglen), rescode, nil
}

func readBody(c *Client, ct *net.TCPConn, body []byte) error {
	ct.SetReadDeadline(time.Now().Add(readTimeout * time.Millisecond))
	_, err := ct.Read(body)
	if err != nil {
		if err == io.EOF {
			return err
		}
		if operr, ok := err.(*net.OpError); ok {
			if operr.Temporary() {
				c.done(ct)
				return err
			}
			if errno, ok := operr.Err.(syscall.Errno); ok {
				if errno == syscall.EPIPE {
					ct.Close()
					return err
				}
			}
		}
		c.done(ct)
		return err
	}
	return err
}

// send the contents of a buffer and receive a response
// back in the same buffer
func (c *Client) doBuf(code byte, msg []byte) ([]byte, byte, error) {
	ct, err := c.ack()
	if err != nil {
		// something went pretty wrong
		return nil, 0, err
	}

	msg, err = writeMsg(c, ct, msg, code)
	if err != nil {
		return nil, 0, err
	}

	// read lead
	var msglen int
	// read length and code
	msglen, code, err = readLead(c, ct)
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
	err = readBody(c, ct, msg)
	if err != nil {
		return msg, code, err
	}

exit:
	c.done(ct)
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
	conn *net.TCPConn
	bts  []byte
}

// unmarshals; returns done / code / error
func (s *streamRes) unmarshal(res protoStream) (bool, byte, error) {
	var msglen int
	var code byte
	var err error

	msglen, code, err = readLead(s.c, s.conn)
	if err != nil {
		return true, code, err
	}

	if msglen > cap(s.bts) {
		s.bts = make([]byte, msglen)
	} else {
		s.bts = s.bts[0:msglen]
	}

	err = readBody(s.c, s.conn, s.bts)
	if err != nil {
		return true, code, err
	}
	// handle a code 0
	if code == 0 {
		riakerr := new(rpbc.RpbErrorResp)
		err = proto.Unmarshal(s.bts, riakerr)
		if err != nil {
			return true, 0, err
		}
		s.close()
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
func (s *streamRes) close() {
	s.c.done(s.conn)
	s.conn = nil
}

func (c *Client) streamReq(req proto.Message, code byte) (*streamRes, error) {
	conn, err := c.ack()
	if err != nil {
		return nil, fmt.Errorf("riakpb: client.Req marshal err: %s", err)
	}
	msg, err := proto.Marshal(req)
	if err != nil {
		return nil, err
	}
	msg, err = writeMsg(c, conn, msg, code)
	if err != nil {
		return nil, err
	}
	return &streamRes{c: c, conn: conn, bts: msg}, nil
}

func (c *Client) Ping() error {
	conn, err := c.ack()
	if err != nil {
		return err
	}
	conn.SetWriteDeadline(time.Now().Add(writeTimeout * time.Millisecond))
	_, err = conn.Write([]byte{0, 0, 0, 1, 1})
	if err != nil {
		return err
	}
	var res [5]byte
	conn.SetReadDeadline(time.Now().Add(readTimeout * time.Millisecond))
	_, err = conn.Read(res[:])
	if err != nil {
		return err
	}
	c.done(conn)
	return nil
}
