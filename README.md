Riak (protocol buffers) Client for Go
================

A Riak client that doesn't use `interface{}`, but does use protocol buffers over
raw TCP. 

## Status

In progress.


## Usage

Satisfy the `Object` interface and you're off to the races. The included 'Blob' object
is the simplest possible Object implementation.

```go
import (
       "github.com/philhofer/riakpb"
)

riak, err := riakpb.DialOne("127.0.0.1", "test-Client-ID")
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
