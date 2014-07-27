// +build riak

package riakpb

import (
	"testing"
)

type TestObject struct {
	Data []byte
	info *Info
}

func (t *TestObject) Unmarshal(b []byte) error {
	t.Data = b
	return nil
}

func (t *TestObject) Marshal(_ []byte) ([]byte, error) {
	return t.Data, nil
}

func (t *TestObject) Info() *Info { return t.info }

func TestFetchNotFound(t *testing.T) {
	nconn := 1
	cl, err := NewClient("localhost:8087", "testClient", &nconn)
	if err != nil {
		t.Fatal(err)
	}
	ob := &TestObject{}

	err = cl.Fetch(ob, "anybucket", "dne", nil)
	if err == nil {
		t.Error("'err' should not be nil")
	}
	if err != ErrNotFound {
		t.Errorf("err is not ErrNotFound: %q", err)
	}
}
