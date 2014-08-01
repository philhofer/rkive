package riakpb

import (
	"bytes"
	"errors"
	"github.com/philhofer/riakpb/rpbc"
)

var (
	ErrNoPath   = errors.New("bucket and/or key not defined")
	ErrModified = errors.New("object has been modified since last read")
	ErrExists   = errors.New("object already exists")
)

// WriteOpts are options available
// for all write opertations.
type WriteOpts struct {
	W  *uint32 // Required write acknowledgements
	DW *uint32 // 'Durable' (to disk) write
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

// New writes a new object into the database. If 'key'
// is non-nil, New will attempt to use that key, and return
// ErrExists if an object already exists at that key-bucket pair.
// Riak will assign this object a key if 'key' is nil.
func (c *Client) New(o Object, bucket string, key *string, opts *WriteOpts) error {
	req := &rpbc.RpbPutReq{
		Bucket:  ustr(bucket),
		Content: new(rpbc.RpbContent),
	}
	// set the bucket, because the user can't

	// return head
	rth := true
	req.ReturnHead = &rth

	// set keys if specified
	if key != nil {
		req.Key = ustr(*key)
		req.IfNoneMatch = &rth
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
	// multiple content items... unlikely
	if len(res.GetContent()) > 1 {
		return handleMultiple(res.Content)
	}
	// pull info from content
	readHeader(o, res.GetContent()[0])
	// set data
	o.Info().vclock = res.Vclock
	o.Info().bucket = req.Bucket
	o.Info().key = res.GetKey()
	return err
}

// Store makes a basic write to the database. Store
// will return ErrNoPath if the object does not already
// have a key and bucket defined. (Use New() if this object
// isn't already in the database.)
func (c *Client) Store(o Object, opts *WriteOpts) error {
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
		return handleMultiple(res.Content)
	}
	readHeader(o, res.GetContent()[0])
	o.Info().vclock = res.Vclock

	return err
}

// Push makes a conditional (if-not-modified) write
// to the database. This is the recommended way of making
// writes to the database, as it minimizes the chances
// of producing sibling objects.
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

	// Return-Head = true; If-Not-Modified = true
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
		if rke, ok := err.(RiakError); ok {
			if bytes.Contains(rke.res.Errmsg, []byte("modified")) {
				return ErrModified
			}
		}
		return err
	}
	if rescode != 12 {
		return ErrUnexpectedResponse
	}
	if res.Vclock == nil || len(res.Content) == 0 {
		return ErrNotFound
	}
	if len(res.Content) > 1 {
		return handleMultiple(res.Content)
	}
	o.Info().vclock = res.Vclock
	readHeader(o, res.Content[0])
	return nil
}
