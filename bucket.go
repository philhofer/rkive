package rkive

import (
	"github.com/philhofer/rkive/rpbc"
	"sync"
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

// Update updates an object in a bucket
func (b *Bucket) Update(o Object) (bool, error) { return b.c.Update(o, nil) }

// IndexLookup performs a secondary index query on the bucket
func (b *Bucket) IndexLookup(idx string, val string) (*IndexQueryRes, error) {
	return b.c.IndexLookup(b.nm, idx, val, nil)
}

// IndexRange performs a secondary index range query on the bucket
func (b *Bucket) IndexRange(idx string, min int64, max int64) (*IndexQueryRes, error) {
	return b.c.IndexRange(b.nm, idx, min, max, nil)
}

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

var (
	ptrTrue         = true
	ptrFalse        = false
	memNval  uint32 = 1
	memR     uint32 = 1
	memW     uint32 = 1
	// properties for memory-backed cache bucket
	cacheProps = rpbc.RpbBucketProps{
		Backend:       []byte("cache"), // this has to come from the riak.conf
		NotfoundOk:    &ptrTrue,
		AllowMult:     &ptrFalse,
		LastWriteWins: &ptrFalse,
		BasicQuorum:   &ptrFalse,
		NVal:          &memNval,
		R:             &memR,
		W:             &memW,
	}
)

// MakeCache makes a memory-backed bucket. You
// must have the "multi"-backend option enabled
// in your configuration in order for this to
// work.
func (b *Bucket) MakeCache() error {
	return b.SetProperties(&cacheProps)

}

// Reset resets the bucket's properties
func (b *Bucket) Reset() error {
	req := &rpbc.RpbResetBucketReq{
		Bucket: ustr(b.nm),
	}
	code, err := b.c.req(req, 29, nil)
	if err != nil {
		return err
	}
	if code != 30 {
		return ErrUnexpectedResponse
	}
	return nil
}

// MultiFetchAsync returns fetch results as a future. Results
// may return in any order. Every result on the channel will
// have its "Value" field type-assertable to the underlying type of 'o'.
// 'procs' goroutines will be used for fetching.
func (b *Bucket) MultiFetchAsync(o Duplicator, procs int, keys ...string) <-chan *AsyncFetch {
	if procs <= 0 {
		procs = 1
	}
	kc := make(chan string, len(keys))
	out := make(chan *AsyncFetch, len(keys))

	wg := new(sync.WaitGroup)
	for i := 0; i < procs; i++ {
		wg.Add(1)
		go func() {
			for key := range kc {
				v := o.NewEmpty()
				err := b.Fetch(v, key)
				out <- &AsyncFetch{v, err}
			}
			wg.Done()
		}()
	}
	go func() {
		wg.Wait()
		close(out)
	}()
	for _, k := range keys {
		kc <- k
	}
	close(kc)

	return out
}
