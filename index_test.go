// +build riak

package riakpb

import (
	"testing"
)

func TestIndexLookup(t *testing.T) {
	nconns := 1
	cl, err := NewClient("localhost:8087", "testClient", &nconns)
	if err != nil {
		t.Fatal(err)
	}

	ob := &TestObject{
		info: &Info{},
		Data: []byte("Hello world!"),
	}

	ob.Info().AddIndex("testIdx", "myValue")

	err = cl.New(ob, "testbucket", nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	res, err := cl.IndexLookup("testbucket", "testIdx", "myValue", nil)
	if err != nil {
		t.Fatal(err)
	}

	if res.Len() < 1 {
		t.Fatalf("Expected multiple keys; got %d.", res.Len())
	}
	t.Logf("Found %d keys.", res.Len())
}

func TestIndexRange(t *testing.T) {
	nconns := 1
	cl, err := NewClient("localhost:8087", "testClient", &nconns)
	if err != nil {
		t.Fatal(err)
	}

	ob := &TestObject{
		info: &Info{},
		Data: []byte("Hello world!"),
	}

	ob.Info().AddIndexInt("testNum", 35)

	err = cl.New(ob, "testbucket", nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	res, err := cl.IndexRange("testbucket", "testNum", 30, 40, nil)
	if err != nil {
		t.Fatal(err)
	}

	if res.Len() < 1 {
		t.Fatalf("Expected multiple keys; got %d", res.Len())
	}
	t.Logf("Found %d keys.", res.Len())
}
