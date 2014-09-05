package rkive

import (
	"bytes"
	"github.com/philhofer/rkive/rpbc"
	"io"
	"strconv"
	"sync"
)

// IndexQueryRes is the response to an index
// query.
type IndexQueryRes struct {
	c      *Client
	ftchd  int
	bucket []byte
	keys   [][]byte
}

// Contains returns whether or not the query
// response contains this particular key
func (i *IndexQueryRes) Contains(key string) bool {
	kb := ustr(key)
	for _, kv := range i.keys {
		if bytes.Equal(kv, kb) {
			return true
		}
	}
	return false
}

// Len returns the number of keys in the response
func (i *IndexQueryRes) Len() int { return len(i.keys) }

// Keys returns the complete list of response keys
func (i *IndexQueryRes) Keys() []string {
	out := make([]string, i.Len())
	for i, kv := range i.keys {
		out[i] = string(kv)
	}
	return out
}

// Fetch fetches the next object in the query. Fetch
// returns whether or not there are objects remaining
// in the query result, and any error encountered in
// fetching that object.
func (i *IndexQueryRes) FetchNext(o Object) (done bool, err error) {
	if i.ftchd >= len(i.keys) {
		return true, io.EOF
	}

	err = i.c.Fetch(o, string(i.bucket), string(i.keys[i.ftchd]), nil)
	i.ftchd++
	if i.ftchd == len(i.keys) {
		done = true
	}
	return
}

// Which searches within the query result for objects that satisfy
// the given condition functions.
func (i *IndexQueryRes) Which(o Object, conds ...func(Object) bool) ([]string, error) {
	var out []string
	bckt := string(i.bucket)
search:
	for j := 0; j < i.Len(); j++ {
		key := string(i.keys[j])
		err := i.c.Fetch(o, bckt, key, nil)
		if err != nil {
			return out, err
		}
		for _, cond := range conds {
			if !cond(o) {
				continue search
			}
		}
		out = append(out, key)
	}
	return out, nil
}

// AsyncFetch represents the output of an
// asynchronous fetch operation. 'Value' is
// never nil, but 'Error' may or may not be nil.
// If 'Error' is non-nil, then 'Value' is usually
// the zero value of the underlying object.
type AsyncFetch struct {
	Value Object
	Error error
}

// FetchAsync returns a channel on which all of the objects
// in the query are returned. 'procs' determines the
// number of goroutines actively fetching. The channel will be closed once
// all the objects have been returned. Objects are fetched
// asynchronously. The (underlying) type of every object returned in each
// AsyncFetch is the same as returned from o.NewEmpty().
func (i *IndexQueryRes) FetchAsync(o Duplicator, procs int) <-chan *AsyncFetch {
	nw := procs
	if i.Len() < nw || nw <= 0 {
		nw = i.Len()
	}
	// keys to fetch
	keys := make(chan string, nw)

	// responses from fetch
	outs := make(chan *AsyncFetch, i.Len())

	// start 'nw' workers
	wg := new(sync.WaitGroup)
	for j := 0; j < nw; j++ {
		wg.Add(1)
		go func(ks chan string, outs chan *AsyncFetch, o Duplicator, wg *sync.WaitGroup) {
			ob := o.NewEmpty()
			for key := range ks {
				err := i.c.Fetch(ob, string(i.bucket), key, nil)
				outs <- &AsyncFetch{Value: ob, Error: err}
			}
			wg.Done()
		}(keys, outs, o, wg)

	}

	// close outs when all
	// workers have exited.
	go func(wg *sync.WaitGroup, os chan *AsyncFetch) {
		wg.Wait()
		close(os)
	}(wg, outs)

	for _, key := range i.Keys() {
		keys <- key
	}
	close(keys)

	return outs
}

// IndexLookup returns the keys that match the index-value pair specified. You
// can specify the maximum number of returned keys ('max'). Index queries are
// performed in "streaming" mode.
func (c *Client) IndexLookup(bucket string, index string, value string, max *int) (*IndexQueryRes, error) {
	bckt := []byte(bucket)
	idx := make([]byte, len(index)+4)
	copy(idx[0:], index)
	copy(idx[len(index):], []byte("_bin"))
	kv := []byte(value)
	rth := true
	var qtype rpbc.RpbIndexReq_IndexQueryType = 0
	req := &rpbc.RpbIndexReq{
		Bucket: bckt,
		Index:  idx,
		Key:    kv,
		Qtype:  &qtype,
		Stream: &rth,
	}

	if max != nil {
		mxr := uint32(*max)
		req.MaxResults = &mxr
	}

	queryres := &IndexQueryRes{
		c:      c,
		bucket: bckt,
	}

	res := &rpbc.RpbIndexResp{}

	// make a stream request
	stream, err := c.streamReq(req, 25)
	if err != nil {
		return nil, err
	}

	// Retrieve streaming responses
	done := false
	for !done {
		var code byte
		done, code, err = stream.unmarshal(res)
		if err != nil {
			return queryres, err
		}
		if code != 26 {
			return queryres, ErrUnexpectedResponse
		}

		queryres.keys = append(queryres.keys, res.Keys...)
		res.Reset()
	}
	return queryres, nil
}

// IndexRange returns the keys that match the index range query. You can specify
// the maximum number of returned results ('max'). Index queries are performed in
// "streaming" mode.
func (c *Client) IndexRange(bucket string, index string, min int64, max int64, maxret *int) (*IndexQueryRes, error) {
	bckt := []byte(bucket)
	idx := make([]byte, len(index)+4)
	copy(idx[0:], index)
	copy(idx[len(index):], []byte("_int"))
	rth := true
	var qtype rpbc.RpbIndexReq_IndexQueryType = 1
	req := &rpbc.RpbIndexReq{
		Bucket:   bckt,
		Index:    idx,
		Qtype:    &qtype,
		Stream:   &rth,
		RangeMin: strconv.AppendInt([]byte{}, min, 10),
		RangeMax: strconv.AppendInt([]byte{}, max, 10),
	}
	if maxret != nil {
		msr := uint32(*maxret)
		req.MaxResults = &msr
	}

	queryres := &IndexQueryRes{
		bucket: bckt,
	}

	res := &rpbc.RpbIndexResp{}
	stream, err := c.streamReq(req, 25)
	if err != nil {
		return nil, err
	}

	done := false
	for !done {
		var code byte
		done, code, err = stream.unmarshal(res)
		if err != nil {
			return queryres, err
		}
		if code != 26 {
			return queryres, ErrUnexpectedResponse
		}
		queryres.keys = append(queryres.keys, res.Keys...)
		res.Reset()
	}
	return queryres, nil
}
