Riak (protocol buffers) Client for Go [![docs examples](https://sourcegraph.com/api/repos/github.com/philhofer/riakpb/.badges/docs-examples.png)](https://sourcegraph.com/github.com/philhofer/riakpb) [![dependencies](https://sourcegraph.com/api/repos/github.com/philhofer/riakpb/.badges/dependencies.png)](https://sourcegraph.com/github.com/philhofer/riakpb)
================

A Riak client that doesn't use `interface{}`, but does use protocol buffers over
raw TCP. 

## Status

Work in progress. Passes existing test cases.

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

## Design

This package is focused on using Riak the way it was intended: with `allow_mult` set to `true`. This library will *always* use vclocks when getting and setting values. Additionally, this library adheres strictly to Riak's read-before-write policy.

Internally, Return-Head is always set to `true`, so every Push() or Store() operation updates the local object's metadata. Consequently, you can carry out a series of transactions on an object in parallel, and avoid conflicts by using Push() (which enforces an If-Not-Modified precondition). You can retreive the latest version of an object by calling Update().

The core "verbs" of this library (New, Fetch, Store, Push, Update, Delete) are meant to have intuitive and sane default behavior. For instance, New always asks Riak to abort the transaction if object already exists at the given key, and Update doesn't return the whole
body of the object back from the database if it hasn't been modified.

The Client object is designed to maintain many long-lived TCP connections to many different Riak nodes. It re-dials closed connections in the background. More TCP connections per client mean less contention for connections, at the cost of more system resources on both ends of the connection. Care should be taken to tune this parameter appropriately for your use case.

