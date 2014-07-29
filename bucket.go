package riakpb

import (
	"github.com/philhofer/riakpb/rpbc"
)

// Bucket represents a Riak bucket
type Bucket struct {
	c  *Client
	nm string
}

// Bucket returns a Riak bucket
// with the provided name
func (c *Client) Bucket(name string) *Bucket {
	return &Bucket{c: c, nm: name}
}

// Fetch performs a fetch with the bucket's default properties
func (b *Bucket) Fetch(o Object, key string) error { return b.c.Fetch(o, b.nm, key, nil) }

// New performs a new store with the bucket's default properties
func (b *Bucket) New(o Object, key *string) error { return b.c.New(o, b.nm, key, nil) }

// Push pushes an object with a bucket's default properties
func (b *Bucket) Push(o Object) error { return b.c.Push(o, nil) }

// Store stores an object with a bucket's default properties
func (b *Bucket) Store(o Object) error { return b.c.Store(o, nil) }

func (b *Bucket) Update(o Object) (bool, error) { return b.c.Update(o, nil) }

// GetProperties retreives the properties of the bucket
func (b *Bucket) GetProperties() (*rpbc.RpbBucketProps, error) {
	req := &rpbc.RpbGetBucketReq{
		Bucket: []byte(b.nm),
	}
	res := &rpbc.RpbGetBucketResp{}
	_, err := b.c.req(req, 19, res)
	return res.GetProps(), err
}

// SetProperties sets the properties of the bucket
func (b *Bucket) SetProperties(props *rpbc.RpbBucketProps) error {
	req := &rpbc.RpbSetBucketReq{
		Bucket: []byte(b.nm),
		Props:  props,
	}
	_, err := b.c.req(req, 21, nil)
	return err
}
