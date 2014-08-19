package rkive

import (
	"bytes"
	"errors"
	"github.com/philhofer/rkive/rpbc"
	"sync"
)

const (
	// maximum number of times we attempt
	// to handle contended stores, either
	// via conflicting vclocks or modifications
	maxMerges = 10
)

var (
	ErrNoPath   = errors.New("bucket and/or key not defined")
	ErrModified = errors.New("object has been modified since last read")
	ErrExists   = errors.New("object already exists")
	ctntPool    *sync.Pool // pool for RpbContent
	hdrPool     *sync.Pool // pool for RpbPutResp
)

func init() {
	ctntPool = new(sync.Pool)
	ctntPool.New = func() interface{} { return &rpbc.RpbContent{} }
	hdrPool = new(sync.Pool)
	hdrPool.New = func() interface{} { return &rpbc.RpbPutResp{} }
}

// push content
func ctput(c *rpbc.RpbContent) {
	ctntPool.Put(c)
}

// create RpbContent from object
func ctpop(o Object) (*rpbc.RpbContent, error) {
	ctnt := ctntPool.Get().(*rpbc.RpbContent)
	err := writeContent(o, ctnt)
	return ctnt, err
}

// pop putresp
func hdrpop() *rpbc.RpbPutResp {
	return hdrPool.Get().(*rpbc.RpbPutResp)
}

// put putresp; zeros fields
func hdrput(r *rpbc.RpbPutResp) {
	r.Content = r.Content[0:0]
	r.Key = r.Key[0:0]
	r.Vclock = r.Vclock[0:0]
	hdrPool.Put(r)
}

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
	req := rpbc.RpbPutReq{
		Bucket: []byte(bucket),
	}

	// return head
	rth := true
	req.ReturnHead = &rth

	// set keys if specified
	if key != nil {
		req.Key = ustr(*key)
		req.IfNoneMatch = &rth
		o.Info().key = append(o.Info().key[0:0], req.Key...)
	}
	var err error
	req.Content, err = ctpop(o)
	if err != nil {
		return err
	}
	// parse options
	parseOpts(opts, &req)
	res := hdrpop()
	rescode, err := c.req(&req, 11, res)
	ctput(req.Content)
	if err != nil {
		hdrput(res)
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
	if len(res.GetContent()) > 1 {
		return handleMultiple(len(res.GetContent()), string(req.Key), string(req.Bucket))
	}
	// pull info from content
	readHeader(o, res.GetContent()[0])
	// set data
	o.Info().vclock = append(o.Info().vclock[0:0], res.Vclock...)
	o.Info().bucket = append(o.Info().bucket[0:0], req.Bucket...)
	if len(res.Key) > 0 {
		o.Info().key = append(o.Info().key[0:0], res.Key...)
	}
	hdrput(res)
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
	ntry := 0 // merge attempts

dostore:
	req := rpbc.RpbPutReq{
		Bucket: o.Info().bucket,
		Key:    o.Info().key,
		Vclock: o.Info().vclock,
	}

	rth := true
	req.ReturnHead = &rth
	if o.Info().vclock != nil {
		req.Vclock = append(req.Vclock, o.Info().vclock...)
	}
	parseOpts(opts, &req)

	// write content
	var err error
	req.Content, err = ctpop(o)
	if err != nil {
		return err
	}
	res := hdrpop()
	rescode, err := c.req(&req, 11, res)
	ctput(req.Content)
	if err != nil {
		return err
	}
	if rescode != 12 {
		return ErrUnexpectedResponse
	}
	if len(res.GetContent()) > 1 {
		if ntry > maxMerges {
			return handleMultiple(len(res.GetContent()), o.Info().Key(), o.Info().Bucket())
		}
		// repair if possible
		if om, ok := o.(ObjectM); ok {
			hdrput(res)
			// load the old value(s) into nom
			nom := om.NewEmpty()
			err = c.Fetch(nom, om.Info().Bucket(), om.Info().Key(), nil)
			if err != nil {
				return err
			}
			// merge old values
			om.Merge(nom)
			om.Info().vclock = nom.Info().vclock
			ntry++
			// retry the store
			goto dostore
		} else {
			return handleMultiple(len(res.GetContent()), o.Info().Key(), o.Info().Bucket())
		}
	}
	readHeader(o, res.GetContent()[0])
	o.Info().vclock = append(o.Info().vclock[0:0], res.Vclock...)
	hdrput(res)
	return nil
}

// Push makes a conditional (if-not-modified) write
// to the database. This is the recommended way of making
// writes to the database, as it minimizes the chances
// of producing sibling objects.
func (c *Client) Push(o Object, opts *WriteOpts) error {
	if o.Info().bucket == nil || o.Info().key == nil || o.Info().vclock == nil {
		return ErrNoPath
	}

	req := rpbc.RpbPutReq{
		Bucket: o.Info().bucket,
		Key:    o.Info().key,
		Vclock: o.Info().vclock,
	}

	// Return-Head = true; If-Not-Modified = true
	rth := true
	req.ReturnHead = &rth
	req.IfNotModified = &rth
	parseOpts(opts, &req)
	ntry := 0

dopush:
	var err error
	req.Content, err = ctpop(o)
	if err != nil {
		return err
	}
	res := hdrpop()
	rescode, err := c.req(&req, 11, res)
	ctput(req.Content)
	if err != nil {
		hdrput(res)
		if rke, ok := err.(RiakError); ok {
			if bytes.Contains(rke.res.Errmsg, []byte("modified")) {
				return ErrModified
			}
		}
		return err
	}
	if rescode != 12 {
		hdrput(res)
		return ErrUnexpectedResponse
	}
	if res.Vclock == nil || len(res.Content) == 0 {
		hdrput(res)
		return ErrNotFound
	}
	if len(res.Content) > 1 {
		// repair if possible
		if om, ok := o.(ObjectM); ok {
			if ntry > maxMerges {
				return handleMultiple(len(res.Content), o.Info().Key(), o.Info().Bucket())
			}
			nom := om.NewEmpty()
			// fetch carries out the local merge on read
			err = c.Fetch(nom, om.Info().Bucket(), om.Info().Key(), nil)
			if err != nil {
				return err
			}
			om.Merge(nom)
			om.Info().vclock = append(om.Info().vclock[0:0], nom.Info().vclock...)
			ntry++
			goto dopush
		} else {
			return handleMultiple(len(res.Content), o.Info().Key(), o.Info().Bucket())
		}
	}
	o.Info().vclock = append(o.Info().vclock[0:0], res.Vclock...)
	readHeader(o, res.GetContent()[0])
	hdrput(res)
	return nil
}
