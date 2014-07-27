// +build riak

package riakpb

import (
	"testing"
)

func TestNewObject(t *testing.T) {
	nconn := 1
	cl, err := NewClient("localhost:8087", "testClient", &nconn)
	if err != nil {
		t.Fatal(err)
	}

	ob := &TestObject{
		Data: []byte("Hello World"),
		info: &Info{},
	}

	// random key assignment
	err = cl.New(ob, "testbucket", nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if ob.Info().Vclock() == "" {
		t.Error("object didn't get assigned a vclock")
	}

	nob := &TestObject{Data: []byte("Blah."), info: &Info{}}
	key := "testkey"
	err = cl.New(nob, "testbucket", &key, nil)
	if err != nil {
		if err != ErrExists {
			t.Fatal(err)
		}
	}
	if ob.Info().Vclock() == "" {
		t.Error("Object didn't get assigned a vclock")
	}
	if ob.Info().Key() == "" {
		t.Errorf("object didn't get assigned a key")
	}

}
