// +build riak

package riakpb

import (
	"bytes"
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

func TestUpdate(t *testing.T) {
	nconn := 1
	cl, err := NewClient("localhost:8087", "testClient", &nconn)
	if err != nil {
		t.Fatal(err)
	}

	lb := &TestObject{
		Data: []byte("flibbertyibbitygibbit"),
		info: &Info{},
	}

	err = cl.New(lb, "testbucket", nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	newlb := &TestObject{
		info: &Info{},
	}

	err = cl.Fetch(newlb, "testbucket", lb.Info().Key(), nil)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(newlb.Data, lb.Data) {
		t.Logf("Object 1 data: %q", lb.Data)
		t.Logf("Object 2 data: %q", newlb.Data)
		t.Errorf("Objects don't have the same body")
	}

	// make a modification
	newlb.Data = []byte("new data.")
	err = cl.Push(newlb, nil)
	if err != nil {
		t.Fatal(err)
	}

	// this should return true
	upd, err := cl.Update(lb, nil)
	if err != nil {
		t.Fatal(err)
	}

	if !upd {
		t.Error("Object was not updated.")
	}

	if !bytes.Equal(lb.Data, newlb.Data) {
		t.Error("Objects are not equal after update.")
	}

	// this should return false
	upd, err = cl.Update(lb, nil)
	if err != nil {
		t.Fatal(err)
	}

	if upd {
		t.Error("Object was spuriously updated...?")
	}
}
