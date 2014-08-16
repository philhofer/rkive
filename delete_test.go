// +build riak

package rkive

import (
	"testing"
)

func TestDelete(t *testing.T) {
        t.Parallel()
	cl := testClient

	ob := &TestObject{
		info: &Info{},
		Data: []byte("Blah."),
	}

	err := cl.New(ob, "testbucket", nil, nil)
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
