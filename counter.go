package rkive

import (
	"github.com/philhofer/rkive/rpbc"
)

// Counter is a Riak CRDT that
// acts as a distributed counter.
type Counter struct {
	key    []byte
	bucket []byte
	val    int64
	parent *Client
}

// Val is the value of the counter
func (c *Counter) Val() int64 { return c.val }

// Bucket is the bucket of the counter
func (c *Counter) Bucket() string { return string(c.bucket) }

// Key is the key of the counter
func (c *Counter) Key() string { return string(c.key) }

// Add adds the value 'v' to the counter.
func (c *Counter) Add(v int64) error {
	req := rpbc.RpbCounterUpdateReq{
		Amount:      &v,       // new value
		Returnvalue: &ptrTrue, // return new value
		Key:         c.key,    // key
		Bucket:      c.bucket, // bucket
	}
	res := rpbc.RpbCounterUpdateResp{}
	code, err := c.parent.req(&req, 50, &res)
	if err != nil {
		return err
	}
	if code != 51 {
		return ErrUnexpectedResponse
	}
	c.val = res.GetValue()
	return nil
}

// Refresh gets the latest value of the counter
// from the database
func (c *Counter) Refresh() error {
	req := rpbc.RpbCounterGetReq{
		Key:    c.key,
		Bucket: c.bucket,
	}
	res := rpbc.RpbCounterGetResp{}
	code, err := c.parent.req(&req, 52, &res)
	if err != nil {
		return err
	}
	if code != 53 {
		return ErrUnexpectedResponse
	}
	c.val = res.GetValue()
	return nil
}

// NewCounter creates a new counter with
// an optional starting value.
func (b *Bucket) NewCounter(name string, start int64) (*Counter, error) {
	req := rpbc.RpbCounterUpdateReq{
		Amount:      &start,
		Returnvalue: &ptrTrue,
		Key:         []byte(name),
		Bucket:      []byte(b.nm),
	}
	res := rpbc.RpbCounterUpdateResp{}
	code, err := b.c.req(&req, 50, &res)
	if err != nil {
		return nil, err
	}
	if code != 51 {
		return nil, ErrUnexpectedResponse
	}
	return &Counter{
		key:    req.Key,
		bucket: req.Bucket,
		val:    res.GetValue(),
		parent: b.c,
	}, nil
}

// GetCounter gets a counter
func (b *Bucket) GetCounter(name string) (*Counter, error) {
	req := rpbc.RpbCounterGetReq{
		Key:    []byte(name),
		Bucket: []byte(b.nm),
	}
	res := rpbc.RpbCounterGetResp{}
	code, err := b.c.req(&req, 52, &res)
	if err != nil {
		return nil, err
	}
	if code != 53 {
		return nil, ErrUnexpectedResponse
	}
	return &Counter{
		key:    req.Key,
		bucket: req.Bucket,
		val:    res.GetValue(),
	}, nil
}
