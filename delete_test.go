// +build riak

package riakpb

import (
	"testing"
)

func TestDelete(t *testing.T) {
	nconns := 1
	cl, err := NewClient("localhost:8087", "testClient", &nconns)
	if err != nil {
		t.Fatal(err)
	}

	ob := &TestObject{
		info: &Info{},
		Data: []byte("Blah."),
	}

	err = cl.New(ob, "testbucket", nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	err = cl.Delete(ob, nil)
	if err != nil {
		t.Fatal(err)
	}

	err = cl.Fetch(ob, ob.Info().Bucket(), ob.Info().Key(), nil)
	if err != ErrNotFound {
		t.Fatalf("Expected ErrNotFound; got %s", err)
	}
}
