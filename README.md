Riak for Go [![docs examples](https://sourcegraph.com/api/repos/github.com/philhofer/riakpb/.badges/docs-examples.png)](https://sourcegraph.com/github.com/philhofer/riakpb) [![dependencies](https://sourcegraph.com/api/repos/github.com/philhofer/riakpb/.badges/dependencies.png)](https://sourcegraph.com/github.com/philhofer/riakpb)
================

A Riak client designed for performance-critical Go programs.

Complete documentation is at [godoc](http://godoc.org/github.com/philhofer/riakpb).

## Status

Core functionality (fetch, store, secondary indexes, links) is complete, but many advanced features (MapReduce, Yokozuna search) are still on the way. There is no short-term guarantee that the API will remain stable. (We are shooting for a beta release in Nov. '14, followed by a "stable" 1.0 in December.) That being said, this code is already being actively tested in some production applications.

## Usage

Satisfy the `Object` interface and you're off to the races. The included 'Blob' object is the simplest possible Object implementation.

```go
import (
       "github.com/philhofer/riakpb"
)

// Open up one connection
riak, err := riakpb.DialOne("127.0.0.1:8087", "test-Client-ID")
// handle err...

blobs := riak.Bucket("blob_bucket")


// let's make an object
myBlob := &riakpb.Blob{
	Data: []byte("Hello World!"),
	RiakInfo: &riakpb.Info{},
}

// now let's put it in the database
err = blobs.New(myBlob, nil)
// handle err...


// since we didn't specify a key, riak assigned
// an available key
fmt.Printf("Our blob key is %s\n", myBlob.Info().Key())

// Let's make a change to the object...
myBlob.Data = []byte("MOAR DATA")

// ... and store it!
err = blobs.Push(myBlob)
// riak.Push will return an error (riakpb.ErrModified) if
// the object has been modified since the last
// time you called New(), Push(), Store(), Fetch(), etc. 
// You can retreive the latest copy of the object with:
updated, err := blobs.Update(myBlob)
// handle err
if updated { /* the object has been changed! */ }

// you can also fetch a new copy
// of the object like so:

// create a new container...
newBlob := &riakpb.Blob{ RiakInfo: &riakpb.Info{} }

// ... and fetch using the old blob's key
err = blobs.Fetch(newBlob, myBlob.Info().Key())
// handle err...

```

If you want to run Riak with `allow_mult=true` (which you should *strongly* consider), take a look
at the `ObjectM` interface, which allows you to specify a `Merge()` operation to be used for
your object when multiple values are encountered on a read or write operation. If you have `allow_mult=true`
and your object does not satisfy the `ObjectM` interface, then read and write operations on a key/bucket
pair with siblings will return a `*ErrMultipleResponses`. (In the degenerate case where 10 consecutive merge 
conflict resolution attempts fail, `*ErrMultipleResponses` will be returned for `ObjectM` operations. This is to 
avoid "sibling explosion.")

As an example, here's what the `Blob` type would have to define (internally) if it were
to satisfy the `ObjectM` interface:

```go
// NewEmpty should always return a properly-initialized
// zero value for the type in question. The client
// will marshal data into this object and pass it to
// Merge().
func (b *Blob) NewEmpty() ObjectM {
     return &Blob{RiakInfo: &Info{}}
}

// Merge should make a best-effort attempt to merge
// data from its argument into the method receiver.
// It should be prepared to handle nil/zero values
// for either object.
func (b *Blob) Merge(o ObjectM) {
     // you can always type-assert the argument
     // to Merge() to be the same type returned
     // by NewEmtpy(), which should also be the
     // same type as the method receiver
     nb := o.(*Blob)

     // we don't really have a good way of handling
     // this conflict, so we'll set the content
     // to be the combination of both
     b.Content = append(b.Content, nb.Content...)
}
```

## Performance

This client library was built with performance in mind.

To run benchmarks, start up Riak locally and run:
`go test -v -tags 'riak' -bench .`

(You must be running the eLevelDB or memory backend with secondary
indexes enabled.)

Remember that TCP is a synchronous protocol, so using multiple
TCP connections per node dramatically improves overall transaction
throughput. In production, you would want to run *about* five live
connections per node. (See: riakpb.Dial()) On my computer, I get 1300 writes and 2340 reads per second with one connection and 4000 writes and 7000 reads per second with 5 connections. Those numbers suggest that my computer would asymptotically approach 9000 writes and 14000 reads per second with an arbitrarily large number of connections. As always, you should benchmark on your production hardware with your specific use case. Remember, too, that Riak is specifically *not* optimized for the single-node case.


## Design & TODOs

This package is focused on using Riak the way it was intended: with `allow_mult` set to `true`. This library will *always* use vclocks when getting and setting values. Additionally, this library adheres strictly to Riak's read-before-write policy.

Internally, Return-Head is always set to `true`, so every Push() or Store() operation updates the local object's metadata. Consequently, you can carry out a series of transactions on an object in parallel, and avoid conflicts by using Push() (which enforces an If-Not-Modified precondition). You can retreive the latest version of an object by calling Update().

The core "verbs" of this library (New, Fetch, Store, Push, Update, Delete) are meant to have intuitive and sane default behavior. For instance, New always asks Riak to abort the transaction if an object already exists at the given key, and Update doesn't return the whole
body of the object back from the database if it hasn't been modified.

The Client object is designed to maintain many long-lived TCP connections to many different Riak nodes. It re-dials closed connections in the background. More TCP connections per client mean less contention for connections, at the cost of more system resources on both ends of the connection. Care should be taken to tune this parameter appropriately for your use case. In the future we may allow for emphemeral connections. (We will need to find a way to do this without slowing down the existing `Client.ack()`.)

Since garbage collection time bottlenecks many Go applications, a lot of effort was put into reducing memory allocations on database reads and writes. Much of the remaining improvement will have to come from carefully re-writing the `Marshal()` and `Unmarshal()` methods of the most frequently used protocol buffers messages. At the very least we may be able to reduce the total number of allocations by taking (better) advantage of `sync.Pool`. Right now we generate 774B (11 allocs) of garbage per read and 1022B (7 allocs) of garbage per write. (For comparison, an `http.Request` object is usually well over 1KB.)

There is an open issue for cache buckets, which has the potential to dramatically improve performance in query-heavy (2i, map-reduce, Yokozuna) use cases. There are also some open issues related to implementing Riak 2.0 features.


## License

This code is MIT licensed. You may use it however you see fit. However, I would very much appreciate it if you create PRs in this repo for patches and improvements!

