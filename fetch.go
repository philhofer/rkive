package rkive

import (
	"errors"
	"github.com/philhofer/rkive/rpbc"
	"sync"
)

const (
	DefaultReqTimeout = 500
)

var (
	// ErrUnexpectedResponse is returned when riak returns the wrong
	// message type
	ErrUnexpectedResponse = errors.New("unexpected response")

	// ErrNotFound is returned when
	// no objects are returned for
	// a read operation
	ErrNotFound = errors.New("not found")

	// ErrDeleted is returned
	// when the object has been marked
	// as deleted, but has not yet been reaped
	ErrDeleted = errors.New("object deleted")

	// default timeout on a request is 500ms
	dfltreq uint32 = DefaultReqTimeout

	// RpbGetResponse pool
	gresPool *sync.Pool
)

func init() {
	gresPool = new(sync.Pool)
	gresPool.New = func() interface{} { return &rpbc.RpbGetResp{} }
}

// pop response from cache
func gresPop() *rpbc.RpbGetResp {
	return gresPool.Get().(*rpbc.RpbGetResp)
}

// push response to cache
func gresPush(r *rpbc.RpbGetResp) {
	r.Content = r.Content[0:0]
	r.Vclock = r.Vclock[0:0]
	r.Unchanged = nil
	gresPool.Put(r)
}

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
	NVal         *uint32 // 'n_val'
}

// parse read options
func parseROpts(req *rpbc.RpbGetReq, opts *ReadOpts) {
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
	}
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
	parseROpts(req, opts)

	res := gresPop()
	rescode, err := c.req(req, 9, res)
	if err != nil {
		return err
	}
	if rescode != 10 {
		return ErrUnexpectedResponse
	}
	// this *should* be handled by req(),
	// but just in case:
	if len(res.GetContent()) == 0 {
		return ErrNotFound
	}
	if len(res.GetContent()) > 1 {
		// merge objects; repair happens
		// on write to prevent sibling
		// explosion
		if om, ok := o.(ObjectM); ok {
			om.Info().key = append(om.Info().key[0:0], req.Key...)
			om.Info().bucket = append(om.Info().bucket[0:0], req.Bucket...)
			om.Info().vclock = append(om.Info().vclock[0:0], res.Vclock...)
			return handleMerge(om, res.Content)
		} else {
			return handleMultiple(len(res.Content), key, bucket)
		}
	}
	err = readContent(o, res.Content[0])
	o.Info().key = append(o.Info().key[0:0], req.Key...)
	o.Info().bucket = append(o.Info().bucket[0:0], req.Bucket...)
	o.Info().vclock = append(o.Info().vclock[0:0], res.Vclock...)
	gresPush(res)
	return err
}

// Update conditionally fetches the object in question
// based on whether or not it has been modified in the database.
// If the object has been changed, the object will be modified
// and Update() will return true. (The object must have a well-defined)
// key, bucket, and vclock.)
func (c *Client) Update(o Object, opts *ReadOpts) (bool, error) {
	if len(o.Info().key) == 0 {
		return false, ErrNoPath
	}
	req := &rpbc.RpbGetReq{
		Bucket:     o.Info().bucket,
		Key:        o.Info().key,
		Timeout:    &dfltreq,
		IfModified: o.Info().vclock,
	}

	parseROpts(req, opts)

	res := gresPop()
	rescode, err := c.req(req, 9, res)
	if err != nil {
		return false, err
	}
	if rescode != 10 {
		return false, ErrUnexpectedResponse
	}
	if res.Unchanged != nil && *res.Unchanged {
		return false, nil
	}
	if len(res.GetContent()) == 0 {
		return false, ErrNotFound
	}
	if len(res.GetContent()) > 1 {
		if om, ok := o.(ObjectM); ok {
			// like Fetch, we merge the results
			// here and hope for reconciliation
			// on write
			om.Info().vclock = append(o.Info().vclock[0:0], res.GetVclock()...)
			err = handleMerge(om, res.Content)
			return true, err
		}
		return false, handleMultiple(len(res.Content), o.Info().Key(), o.Info().Bucket())
	}
	err = readContent(o, res.Content[0])
	o.Info().vclock = append(o.Info().vclock[0:0], res.Vclock...)
	gresPush(res)
	return true, err
}

// FetchHead returns the head (*Info) of an object
// stored in Riak. This is the least expensive way
// to check for the existence of an object.
func (c *Client) FetchHead(bucket string, key string) (*Info, error) {
	rth := true
	req := &rpbc.RpbGetReq{
		Key:     []byte(key),
		Bucket:  []byte(bucket),
		Timeout: &dfltreq,
		Head:    &rth,
	}
	res := gresPop()
	rescode, err := c.req(req, 9, res)
	if err != nil {
		gresPush(res)
		return nil, err
	}
	if rescode != 10 {
		gresPush(res)
		return nil, ErrUnexpectedResponse
	}
	// NotFound is supposed to be handled by
	// c.req, but just in case:
	if len(res.Content) == 0 {
		gresPush(res)
		return nil, ErrNotFound
	}
	if len(res.Content) > 1 {
		gresPush(res)
		return nil, handleMultiple(len(res.Content), key, bucket)
	}
	bl := &Blob{RiakInfo: &Info{}}
	readHeader(bl, res.Content[0])
	bl.Info().vclock = append(bl.Info().vclock[0:0], res.Vclock...)
	bl.Info().key = append(bl.Info().key[0:0], req.Key...)
	bl.Info().bucket = append(bl.Info().bucket[0:0], req.Bucket...)
	gresPush(res)
	return bl.Info(), err
}

// PullHead pulls the latest object metadata into the object.
// The Info() pointed to by the object will be changed if the
// object has been changed in Riak since the last read. If you
// want to read the entire object, use Update() instead.
func (c *Client) PullHead(o Object) error {
	if len(o.Info().key) == 0 {
		return ErrNoPath
	}
	rth := true
	req := &rpbc.RpbGetReq{
		Key:        o.Info().key,
		Bucket:     o.Info().bucket,
		Timeout:    &dfltreq,
		Head:       &rth,
		IfModified: o.Info().vclock,
	}
	res := gresPop()
	code, err := c.req(req, 9, res)
	if err != nil {
		gresPush(res)
		return err
	}
	if code != 10 {
		return ErrUnexpectedResponse
	}
	if res.GetUnchanged() {
		gresPush(res)
		return nil
	}
	if len(res.Content) == 0 {
		return ErrNotFound
	}
	if len(res.Content) > 1 {
		gresPush(res)
		return handleMultiple(len(res.Content), o.Info().Key(), o.Info().Bucket())
	}
	readHeader(o, res.Content[0])
	o.Info().vclock = append(o.Info().vclock[0:0], res.Vclock...)
	gresPush(res)
	return nil
}
