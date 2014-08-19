package rkive

import (
	"github.com/philhofer/rkive/rpbc"
)

// DelOpts are options available on delete
// operations. All values are optional.
type DelOpts struct {
	R  *uint32 // required reads
	W  *uint32 // required writes
	PR *uint32 // required primary node reads
	PW *uint32 // required primary node writes
	RW *uint32 // required replica deletions
	DW *uint32 // required durable (to disk) writes
}

func parseDelOpts(opts *DelOpts, req *rpbc.RpbDelReq) {
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
	if opts.R != nil {
		req.R = opts.R
	}
	if opts.PR != nil {
		req.Pr = opts.R
	}
	if opts.RW != nil {
		req.Rw = opts.RW
	}

}

func (c *Client) Delete(o Object, opts *DelOpts) error {
	if o.Info().bucket == nil || o.Info().key == nil {
		return ErrNoPath
	}
	req := &rpbc.RpbDelReq{
		Bucket: o.Info().bucket,
		Key:    o.Info().key,
		Vclock: o.Info().vclock,
	}

	parseDelOpts(opts, req)

	_, err := c.req(req, 13, nil)
	return err
}
