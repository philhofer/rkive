Riak for Go [![Build Status](https://travis-ci.org/philhofer/rkive.svg?branch=master)](https://travis-ci.org/philhofer/rkive)[![docs examples](https://sourcegraph.com/api/repos/github.com/philhofer/rkive/.badges/docs-examples.png)](https://sourcegraph.com/github.com/philhofer/rkive) [![dependencies](https://sourcegraph.com/api/repos/github.com/philhofer/rkive/.badges/dependencies.png)](https://sourcegraph.com/github.com/philhofer/rkive)
================

A Riak client designed for performance-critical Go programs.

Complete documentation is at [godoc](http://godoc.org/github.com/philhofer/rkive).

## Status

Core functionality (fetch, store, secondary indexes, links) is complete, but many advanced features (MapReduce, Yokozuna search) are still on the way. There is no short-term guarantee that the API will remain stable. (We are shooting for a beta release in Nov. '14, followed by a "stable" 1.0 in December.) That being said, this code is already being actively tested in some production applications.

## Usage

Satisfy the `Object` interface and you're off to the races. The included 'Blob' object is the simplest possible Object implementation.

```go
import (
       "github.com/philhofer/rkive"
)

// Open up one connection
riak, err := rkive.DialOne("127.0.0.1:8087", "test-Client-ID")
// handle err...

blobs := riak.Bucket("blob_bucket")


// let's make an object
myBlob := &rkive.Blob{ Data: []byte("Hello World!") }

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

newBlob := &rkive.Blob{}

err = blobs.Fetch(newBlob, myBlob.Info().Key())
// handle err...

```

For more worked examples, check out the `/_examples` folder.

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
func (b *Blob) NewEmpty() Object {
     return &Blob{}
}

// Merge should make a best-effort attempt to merge
// data from its argument into the method receiver.
// It should be prepared to handle nil/zero values
// for either object.
func (b *Blob) Merge(o Object) {
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

The `Client` object maintains a pool of ephemeral TCP connections that
are re-used under high load. The size of the pool grows and shrinks at runtime. In practice, it is
unusual to see more than five or six connections open simultaneously. (There is a cap set at 30 live connections
at a time for safety's sake.) Empirically, maintaining ephemeral connections in a `sync.Pool` is *at least* as fast
as using long-lived connections in a buffered channel, and in many cases much faster and less resource-hungry.

## Design & TODOs

This package is focused on using Riak the way it was intended: with `allow_mult` set to `true`. This library will *always* use vclocks when getting and setting values. Additionally, this library adheres strictly to Riak's read-before-write policy.

Internally, Return-Head is always set to `true`, so every Push() or Store() operation updates the local object's metadata. Consequently, you can carry out a series of transactions on an object in parallel, and avoid conflicts by using Push() (which enforces an If-Not-Modified precondition). You can retreive the latest version of an object by calling Update().

The core "verbs" of this library (New, Fetch, Store, Push, Update, Delete) are meant to have intuitive and sane default behavior. For instance, New always asks Riak to abort the transaction if an object already exists at the given key, and Update doesn't return the whole
body of the object back from the database if it hasn't been modified.

Since garbage collection time bottlenecks many Go applications, a lot of effort was put into reducing memory allocations on database reads and writes. Much of the remaining improvement will have to come from carefully re-writing the `Marshal()` and `Unmarshal()` methods of the most frequently used protocol buffers messages. At the very least we may be able to reduce the total number of allocations by taking (better) advantage of `sync.Pool`. Right now we generate 636B (8 allocs) of garbage per read and 754B (5 allocs) of garbage per write. (For comparison, the zero value of an `http.Request` object is well over 1KB.)

There is an open issue for cache buckets, which has the potential to dramatically improve performance in query-heavy (2i, map-reduce, Yokozuna) use cases. There are also some open issues related to implementing Riak 2.0 features.


## License

This code is MIT licensed. You may use it however you see fit. However, I would very much appreciate it if you create PRs in this repo for patches and improvements!

