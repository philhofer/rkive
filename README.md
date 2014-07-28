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

riak, err := riakpb.NewClient("127.0.0.1", "test-Client-ID", nil)
// handle err...


// let's make an object
myBlob := &riakpb.Blob{
	Data: []byte("Hello World!"),
	RiakInfo: &riakpb.Info{},
}

// now let's put it in the database
err = riak.New(myBlob, "blob_bucket", nil, nil)
// handle err...


// since we didn't specify a key, riak assigned
// an available key
fmt.Printf("Our blob key is %s\n", myBlob.Info().Key())

// Let's make a change to the object...
myBlob.Data = []byte("MOAR DATA")

// ... and store it!
err = riak.Push(myBlob, nil)
// riak.Push will return an error (riakpb.ErrModified) if
// the object has been modified since the last
// time you called New(), Push(), Store(), Fetch(), etc. 
// You can retreive the latest copy of the object with:
updated, err := riak.Update(myBlob, nil)
// handle err
if updated { /* the object has been changed! */ }

// you can also fetch a new copy
// of the object like so:

// create a new container
newBlob := &riakpb.Blob{ RiakInfo: &riakpb.Info{} }

// fetch using the bucket and key
err = riak.Fetch(newBlob, myBlob.Info().Bucket(), myBlob.Info().Key(), nil)
// handle err...


```
