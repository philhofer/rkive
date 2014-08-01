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

func (t *TestObject) Marshal() ([]byte, error) {
	return t.Data, nil
}

func (t *TestObject) Info() *Info { return t.info }

func TestFetchNotFound(t *testing.T) {
	cl, err := DialOne("localhost:8087", "testClient")
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
	cl.Close()
}

func TestUpdate(t *testing.T) {
	cl, err := DialOne("localhost:8087", "testClient")
	if err != nil {
		t.Fatal(err)
	}

	test := cl.Bucket("testbucket")

	lb := &TestObject{
		Data: []byte("flibbertyibbitygibbit"),
		info: &Info{},
	}

	err = test.New(lb, nil)
	if err != nil {
		t.Fatal(err)
	}

	newlb := &TestObject{
		info: &Info{},
	}

	err = test.Fetch(newlb, lb.Info().Key())
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
	err = test.Push(newlb)
	if err != nil {
		t.Fatal(err)
	}

	// this should return true
	upd, err := test.Update(lb)
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
	upd, err = test.Update(newlb)
	if err != nil {
		t.Fatal(err)
	}

	if upd {
		t.Error("Object was spuriously updated...?")
	}
	cl.Close()
}
