Riak (protocol buffers) Client for Go
================

A Riak client that doesn't use `interface{}`, but does use protocol buffers over
raw TCP. 

## Status

In progress.


## Usage

Satisfy the `Object` interface and you're off to the races. Here's the simplest possible example:

```go
import (
       "github.com/philhofer/riakpb"
)

type SimpleObject struct {
     Data []byte
     info *riakpb.Info
}

// Objects need to know how to marshal
// themselves to bytes. This value is used as the
// value stored in the database.
func (s *SimpleObject) Marshal(_ []byte) ([]byte, error) {
     return s.Data
}

// Objects also need to know how to unmarshal
// themselves from bytes. This is exactly the
// reverse of the function above.
func (s *SimpleObject) Unmarshal(b []byte) error {
     s.Data = b
     return nil
}

// Info() needs to return a pointer
// to a valid riakpb.Info object. It stores
// Riak metadata like secondary indexes, links
// user metadata, vector clocks, etc.
func (s *SimpleObject) Info() *riakpb.Info { return s.info }

```
