package riakpb

import (
	"bytes"
	"errors"
	"github.com/philhofer/riakpb/rpbc"
)

var (
	ErrNoPath   = errors.New("bucket or key is not defined")
	ErrModified = errors.New("object has been modified since last read")
	ErrExists   = errors.New("object already exists")
)

type WriteOpts struct {
	W  *uint32 // writes
	DW *uint32 // Durable writes
	PW *uint32 // Primary replica writes
}

// put options into request
func parseOpts(opts *WriteOpts, req *rpbc.RpbPutReq) {
	if opts == nil || req == nil {
		return
	}
	if opts.W != nil {
		req.W = opts.W
	}
	if opts.DW != nil {
		req.Dw = opts.DW
	}
	if opts.PW != nil {
		req.Pw = opts.PW
	}
}

func (c *Client) New(o Object, bucket string, key *string, opts *WriteOpts) error {
	req := &rpbc.RpbPutReq{
		Bucket:  []byte(bucket),
		Content: new(rpbc.RpbContent),
	}
	// set the bucket, because the user can't
	o.Info().bucket = req.Bucket

	// return head
	rth := true
	req.ReturnHead = &rth

	// set keys if specified
	if key != nil {
		req.Key = []byte(*key)
		req.IfNoneMatch = &rth
		o.Info().key = req.Key
	}
	// write content to request
	err := writeContent(o, req.Content)
	if err != nil {
		return err
	}
	// parse options
	parseOpts(opts, req)

	res := &rpbc.RpbPutResp{}
	rescode, err := c.req(req, 11, res)
	if err != nil {
		// riak returns "match_found" on failure
		if rke, ok := err.(RiakError); ok {
			if bytes.Contains(rke.res.GetErrmsg(), []byte("match_found")) {
				return ErrExists
			}
		}
		return err
	}
	// not what we expected...
	if rescode != 12 {
		return ErrUnexpectedResponse
	}
	// multiple content items
	if len(res.GetContent()) != 1 {
		return ErrMultiple
	}
	// pull info from content
	readHeader(o, res.GetContent()[0])
	o.Info().vclock = res.Vclock
	// pull the new key if
	// we didn't already set it
	if key == nil {
		o.Info().key = res.GetKey()
	}
	return err
}

// Store makes a basic write to the database
func (c *Client) Store(o Object, opts *WriteOpts) error {
	if o.Info().bucket == nil || o.Info().key == nil {
		return ErrNoPath
	}
	req := &rpbc.RpbPutReq{
		Bucket:  o.Info().bucket,
		Key:     o.Info().key,
		Content: new(rpbc.RpbContent),
	}
	req.Vclock = o.Info().vclock
	rth := true
	req.ReturnHead = &rth
	if o.Info().vclock != nil {
		req.Vclock = o.Info().vclock
	}
	parseOpts(opts, req)

	// write content
	err := writeContent(o, req.Content)
	if err != nil {
		return err
	}
	res := &rpbc.RpbPutResp{}
	rescode, err := c.req(req, 11, res)
	if err != nil {
		return err
	}
	if rescode != 12 {
		return ErrUnexpectedResponse
	}
	if len(res.GetContent()) > 1 {
		return ErrMultiple
	}
	readHeader(o, res.GetContent()[0])
	o.Info().vclock = res.Vclock

	return err
}

// Push makes a conditional (if-not-modified) write
// to the database
func (c *Client) Push(o Object, opts *WriteOpts) error {
	if o.Info().bucket == nil || o.Info().key == nil {
		return ErrNoPath
	}
	req := &rpbc.RpbPutReq{
		Bucket:  o.Info().bucket,
		Key:     o.Info().key,
		Content: new(rpbc.RpbContent),
		Vclock:  o.Info().vclock,
	}
	rth := true
	req.ReturnHead = &rth
	req.IfNotModified = &rth
	parseOpts(opts, req)

	res := &rpbc.RpbPutResp{}
	err := writeContent(o, req.Content)
	if err != nil {
		return err
	}

	rescode, err := c.req(req, 11, res)
	if err != nil {
		return err
	}
	if rescode != 12 {
		return ErrUnexpectedResponse
	}
	if res.GetVclock() == nil || len(res.GetContent()) == 0 {
		return ErrModified
	}
	if len(res.GetContent()) > 1 {
		return ErrMultiple
	}
	return readContent(o, res.GetContent()[0])
}
