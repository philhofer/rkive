// +build riak

package riakpb

import (
	"bytes"
	"testing"
)

func TestNewObject(t *testing.T) {
	cl, err := Dial([]Node{{"localhost:8087", 3}}, "testClient")
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
	if string(ob.Data) != "Hello World" {
		t.Error("Object lost its data")
	}
	if ob.Info().Vclock() == "" {
		t.Error("object didn't get assigned a vclock")
	}

	nob := &TestObject{Data: []byte("Blah."), info: &Info{}}
	key := "testkey"
	err = cl.New(nob, "testbucket", &key, nil)
	if err != nil {
		// we'll allow ErrExists
		// b/c of prior test runs
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
	cl.Close()
}

func TestPushObject(t *testing.T) {
	cl, err := Dial([]Node{{"localhost:8087", 2}}, "testClient")
	if err != nil {
		t.Fatal(err)
	}

	ob := &TestObject{
		Data: []byte("Hello World"),
		info: &Info{},
	}
	// make new
	err = cl.New(ob, "testbucket", nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	// fetch 'n store
	newob := &TestObject{
		Data: nil,
		info: &Info{},
	}
	// fetch the same
	err = cl.Fetch(newob, "testbucket", ob.Info().Key(), nil)
	if err != nil {
		t.Fatal(err)
	}
	// modify the data
	newob.Data = []byte("new conflicting data!")
	// this should work
	err = cl.Push(newob, nil)
	if err != nil {
		t.Fatal(err)
	}

	// modify the old
	ob.Data = []byte("blah blah blah")

	err = cl.Push(ob, nil)
	if err != ErrModified {
		t.Fatalf("Expected ErrModified; got %q", err)
	}

	cl.Close()
}

func TestStoreObject(t *testing.T) {
	cl, err := DialOne("localhost:8087", "testClient")
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
	if string(ob.Data) != "Hello World" {
		t.Fatal("Object lost its data!")
	}

	// fetch the same object
	nob := &TestObject{info: &Info{}}
	err = cl.Fetch(nob, "testbucket", ob.Info().Key(), nil)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(ob.Data, nob.Data) {
		t.Logf("Sent: %q", ob.Data)
		t.Logf("Returned : %q", nob.Data)
		t.Fatal("Objects' 'data' field differs")
	}

	// make a change
	nob.Data = []byte("new information!")
	err = cl.Store(nob, nil)
	if err != nil {
		t.Fatal(err)
	}
	cl.Close()
}

func BenchmarkStore(b *testing.B) {
	b.N /= 100

	cl, err := DialOne("localhost:8087", "testClient")
	if err != nil {
		b.Fatal(err)
	}

	ob := &TestObject{
		Data: []byte("Hello World"),
		info: &Info{},
	}

	err = cl.New(ob, "tesbucket", nil, nil)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		err = cl.Store(ob, nil)
		if err != nil {
			b.Fatal(err)
		}
	}
	cl.Close()
}

func BenchmarkFetch(b *testing.B) {
	b.N /= 100

	cl, err := DialOne("localhost:8087", "testClient")
	if err != nil {
		b.Fatal(err)
	}

	ob := &TestObject{
		info: &Info{},
		Data: []byte("Hello World"),
	}

	err = cl.New(ob, "testbucket", nil, nil)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		err = cl.Fetch(ob, "testbucket", ob.Info().Key(), nil)
		if err != nil {
			b.Fatal(err)
		}
	}

	cl.Close()
}
