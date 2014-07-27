package riakpb

import (
	"errors"
	"github.com/philhofer/riakpb/rpbc"
)

const (
	DefaultReqTimeout = 500
)

var (
	// ErrRiakError is returned when riak returns a 5XX code
	ErrRiakError = errors.New("riak error")

	// ErrUnexpectedResponse is returned when riak returns the wrong
	// message type
	ErrUnexpectedResponse = errors.New("unexpected response")

	// ErrNotFound is returned when
	// no objects are returned for
	// a read operation
	ErrNotFound = errors.New("not found")

	// ErrMultiple is returned when
	// multiple objects are returned
	// for a read operation
	ErrMultiple = errors.New("multiple responses")

	dfltreq uint32 = DefaultReqTimeout
)

// ReadOpts are read options
// that can be specified when
// doing a read operation. All
// of these default to the default
// bucket properties.
type ReadOpts struct {
	R            *uint32 // number of reads
	Pr           *uint32 // number of primary replica reads
	BasicQuorum  *bool   // basic quorum required
	SloppyQuorum *bool   // sloppy quorum required
	NotfoundOk   *bool   // treat not-found as a read for 'R'
	IfModified   []byte  // vclock of object, if doing a conditional read
	NVal         *uint32 // 'n_val'
}

// Fetch puts whatever exists at the provided bucket+key
// into the provided Object. It has undefined behavior
// if the object supplied does not know how to unmarshal
// the bytes returned from riak.
func (c *Client) Fetch(o Object, bucket string, key string, opts *ReadOpts) error {
	// make request object
	req := &rpbc.RpbGetReq{
		Bucket: []byte(bucket),
		Key:    []byte(key),
	}
	// set 500ms reqeust timeout
	req.Timeout = &dfltreq
	// get opts
	if opts != nil {
		if opts.R != nil {
			req.R = opts.R
		}
		if opts.Pr != nil {
			req.Pr = opts.Pr
		}
		if opts.BasicQuorum != nil {
			req.BasicQuorum = opts.BasicQuorum
		}
		if opts.NVal != nil {
			req.NVal = opts.NVal
		}
		if opts.NotfoundOk != nil {
			req.NotfoundOk = opts.NotfoundOk
		}
		if opts.IfModified != nil {
			req.IfModified = opts.IfModified
		}
	}

	res := rpbc.RpbGetResp{}
	rescode, err := c.req(req, 9, &res)
	if err != nil {
		return err
	}
	if rescode != 10 {
		return ErrUnexpectedResponse
	}
	if len(res.GetContent()) == 0 {
		return ErrNotFound
	}
	if len(res.GetContent()) > 1 {
		return ErrMultiple
	}
	err = readContent(o, res.GetContent()[0])
	o.Info().key = req.Key
	o.Info().bucket = req.Bucket
	o.Info().vclock = res.GetVclock()
	return err
}
