// +build riak

package rkive

import (
	"bytes"
	check "gopkg.in/check.v1"
	"time"
)

func (s *riakSuite) TestNewObject(c *check.C) {
	startt := time.Now()
	ob := &TestObject{
		Data: []byte("Hello World"),
	}

	// random key assignment
	err := s.cl.New(ob, "testbucket", nil, nil)
	if err != nil {
		c.Fatal(err)
	}
	if string(ob.Data) != "Hello World" {
		c.Error("Object lost its data")
	}
	if ob.Info().Vclock() == "" {
		c.Error("object didn't get assigned a vclock")
	}

	nob := &TestObject{Data: []byte("Blah.")}
	key := "testkey"
	err = s.cl.New(nob, "testbucket", &key, nil)
	if err != nil {
		// we'll allow ErrExists
		// b/c of prior test runs
		if err != ErrExists {
			c.Fatal(err)
		}
	}
	if ob.Info().Vclock() == "" {
		c.Error("Object didn't get assigned a vclock")
	}
	if ob.Info().Key() == "" {
		c.Errorf("object didn't get assigned a key")
	}
	s.runtime += time.Since(startt)
}

func (s *riakSuite) TestPushObject(c *check.C) {
	startt := time.Now()
	ob := &TestObject{
		Data: []byte("Hello World"),
	}
	// make new
	err := s.cl.New(ob, "testbucket", nil, nil)
	if err != nil {
		c.Fatal(err)
	}

	// fetch 'n store
	newob := &TestObject{
		Data: nil,
	}
	// fetch the same
	err = s.cl.Fetch(newob, "testbucket", ob.Info().Key(), nil)
	if err != nil {
		c.Fatal(err)
	}
	// modify the data
	newob.Data = []byte("new conflicting data!")
	// this should work
	err = s.cl.Push(newob, nil)
	if err != nil {
		c.Fatal(err)
	}

	// modify the old
	ob.Data = []byte("blah blah blah")

	err = s.cl.Push(ob, nil)
	if err != ErrModified {
		c.Fatalf("Expected ErrModified; got %q", err)
	}
	s.runtime += time.Since(startt)
}

func (s *riakSuite) TestStoreObject(c *check.C) {
	startt := time.Now()
	ob := &TestObject{
		Data: []byte("Hello World"),
	}

	// random key assignment
	err := s.cl.New(ob, "testbucket", nil, nil)
	if err != nil {
		c.Fatal(err)
	}
	if ob.Info().Vclock() == "" {
		c.Error("object didn't get assigned a vclock")
	}
	if string(ob.Data) != "Hello World" {
		c.Fatal("Object lost its data!")
	}

	// fetch the same object
	nob := &TestObject{}
	err = s.cl.Fetch(nob, "testbucket", ob.Info().Key(), nil)
	if err != nil {
		c.Fatal(err)
	}

	if !bytes.Equal(ob.Data, nob.Data) {
		c.Logf("Sent: %q", ob.Data)
		c.Logf("Returned : %q", nob.Data)
		c.Fatal("Objects' 'data' field differs")
	}

	// make a change
	nob.Data = []byte("new information!")
	err = s.cl.Store(nob, nil)
	if err != nil {
		c.Fatal(err)
	}
	s.runtime += time.Since(startt)
}

func (s *riakSuite) TestPushChangeset(c *check.C) {
	startt := time.Now()
	ob := &TestObject{
		Data: []byte("Here's a body."),
	}
	nob := &TestObject{}

	err := s.cl.New(ob, "testbucket", nil, nil)
	if err != nil {
		c.Fatal(err)
	}

	err = s.cl.Fetch(nob, "testbucket", ob.Info().Key(), nil)
	if err != nil {
		c.Fatal(err)
	}

	nob.Data = []byte("Intermediate")
	err = s.cl.Push(nob, nil)
	if err != nil {
		c.Fatal(err)
	}

	// this should fail
	err = s.cl.Push(ob, nil)
	if err != ErrModified {
		c.Fatalf("Expected \"modified\", got %q", err)
	}

	chng := func(o Object) error {
		v := o.(*TestObject)
		if bytes.Equal(v.Data, []byte("New Body")) {
			return ErrDone
		}
		v.Data = []byte("New Body")
		return nil
	}

	// ... and then this should pass
	err = s.cl.PushChangeset(ob, chng, nil)
	if err != nil {
		c.Fatal(err)
	}

	// the other object should reflect the changes
	chngd, err := s.cl.Update(nob, nil)
	if err != nil {
		c.Fatal(err)
	}
	if !chngd {
		c.Error("Expected change; didn't get it")
	}
	if !bytes.Equal(nob.Data, []byte("New Body")) {
		c.Errorf("Wanted data \"New Body\"; got %q", nob.Data)
	}

	if !bytes.Equal(ob.Data, []byte("New Body")) {
		c.Error("ob.Data didn't retain the appropriate value")
	}

	err = s.cl.Fetch(ob, "testbucket", ob.Info().Key(), nil)
	if err != nil {
		c.Fatal(err)
	}
	if !bytes.Equal(ob.Data, []byte("New Body")) {
		c.Errorf(`Expected "New Body"; got %q`, ob.Data)
	}
	s.runtime += time.Since(startt)
}
