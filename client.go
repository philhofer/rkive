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
	log.Printf("Successfully opened %d connections to %s", dfltConns, addr)
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

// send the contents of a buffer and receive a response
// back in the same buffer
func (c *Client) doBuf(code byte, msg []byte) ([]byte, byte, error) {
	ct, err := c.ack()
	if err != nil {
		// something went pretty wrong
		return nil, 0, err
	}

	// bigendian length + code byte
	var lead [5]byte

	msglen := uint32(len(msg) + 1)
	binary.BigEndian.PutUint32(lead[:4], msglen)
	lead[4] = code
	msg = append(lead[:], msg...)

	// send the message
	ct.SetWriteDeadline(time.Now().Add(writeTimeout * time.Millisecond))
	_, err = ct.Write(msg)
	if err != nil {
		if err == io.EOF {
			return nil, 0, err
		}
		if operr, ok := err.(*net.OpError); ok {
			if operr.Temporary() {
				c.done(ct)
				return nil, 0, err
			}
			if errno, ok := operr.Err.(syscall.Errno); ok {
				if errno == syscall.EPIPE {
					ct.Close()
					return nil, 0, err
				}
			}
		}
	}

	/////////////////////
	// 	 RESPONSE 	  ///
	/////////////////////

	// read lead
	ct.SetReadDeadline(time.Now().Add(readTimeout * time.Millisecond))
	_, err = ct.Read(lead[:])
	if err != nil {
		if err == io.EOF {
			return nil, 0, err
		}
		if operr, ok := err.(*net.OpError); ok {
			if operr.Temporary() {
				c.done(ct)
				return nil, 0, err
			}
			if errno, ok := operr.Err.(syscall.Errno); ok {
				if errno == syscall.EPIPE {
					ct.Close()
					return nil, 0, err
				}
			}
		}
	}

	// response message size
	msglen = binary.BigEndian.Uint32(lead[:4]) - 1
	rescode := lead[4]
	log.Printf("Returning %d-byte message; code %d\n", msglen, rescode)
	if int(msglen) > cap(msg) {
		msg = make([]byte, int(msglen))
	} else {
		msg = msg[0:int(msglen)]
	}
	if msglen == 0 {
		msg = nil
		goto exit
	}

	ct.SetReadDeadline(time.Now().Add(readTimeout * time.Millisecond))
	_, err = ct.Read(msg)
	if err != nil {
		if err == io.EOF {
			return msg, rescode, err
		}
		if operr, ok := err.(*net.OpError); ok {
			if operr.Temporary() {
				c.done(ct)
				return msg, rescode, err
			}
			if errno, ok := operr.Err.(syscall.Errno); ok {
				if errno == syscall.EPIPE {
					ct.Close()
					return msg, rescode, err
				}
			}
		}
		c.done(ct)
		return msg, rescode, err
	}

exit:
	log.Printf("Message length %d", len(msg))
	log.Printf("Returned message: %x", msg)
	c.done(ct)
	return msg, rescode, nil
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
