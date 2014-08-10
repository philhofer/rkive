package rkive

import (
	"errors"
	"github.com/philhofer/rkive/rpbc"
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

	// default timeout on a request is 500ms
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

	res := rpbc.RpbGetResp{}
	rescode, err := c.req(req, 9, &res)
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
			om.Info().key = req.Key
			om.Info().bucket = req.Bucket
			om.Info().vclock = res.Vclock
			return handleMerge(om, res.Content)
		} else {
			return handleMultiple(res.Content)
		}
	}
	err = readContent(o, res.GetContent()[0])
	o.Info().key = req.Key
	o.Info().bucket = req.Bucket
	o.Info().vclock = res.GetVclock()
	return err
}

// Update conditionally fetches the object in question
// based on whether or not it has been modified in the database.
// If the object has been changed, the object will be modified
// and Update() will return true. (The object must have a well-defined)
// key, bucket, and vclock.)
func (c *Client) Update(o Object, opts *ReadOpts) (bool, error) {
	if o.Info().key == nil || o.Info().bucket == nil || o.Info().vclock == nil {
		return false, ErrNoPath
	}
	// make request object
	req := &rpbc.RpbGetReq{
		Bucket:     o.Info().bucket,
		Key:        o.Info().key,
		Timeout:    &dfltreq,
		IfModified: o.Info().vclock,
	}

	parseROpts(req, opts)

	res := rpbc.RpbGetResp{}
	rescode, err := c.req(req, 9, &res)
	if err != nil {
		return false, err
	}
	if rescode != 10 {
		return false, ErrUnexpectedResponse
	}
	if res.GetUnchanged() {
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
			om.Info().vclock = res.GetVclock()
			err = handleMerge(om, res.Content)
			return true, err
		}
		return false, handleMultiple(res.Content)
	}
	err = readContent(o, res.Content[0])
	o.Info().vclock = res.Vclock
	return true, err
}
