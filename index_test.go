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
	if !res.Contains(ob.Info().Key()) {
		t.Errorf("Response doesn't contain original key...?")
	}

	hasCorrectIndex := func(o Object) bool {
		val := o.Info().GetIndex("testIdx")
		if val != "myValue" {
			t.Logf("Found incorrect: %v", o.Info().idxs)
			return false
		}
		return true
	}

	ncorrect, err := res.Which(ob, hasCorrectIndex)
	if err != nil {
		t.Fatal(err)
	}
	if len(ncorrect) != res.Len() {
		t.Errorf("Ncorrect is %d; response length is %d", len(ncorrect), res.Len())
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

	if !res.Contains(ob.Info().Key()) {
		t.Errorf("Response doesn't contain original key...?")
	}
	t.Logf("Found %d keys.", res.Len())
}
