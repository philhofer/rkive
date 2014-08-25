// +build riak

package rkive

import (
	"bytes"
	check "gopkg.in/check.v1"
	"sync"
	"testing"
)

func (s *riakSuite) TestNewObject(c *check.C) {
	ob := &TestObject{
		Data: []byte("Hello World"),
		info: &Info{},
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

	nob := &TestObject{Data: []byte("Blah."), info: &Info{}}
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
}

func (s *riakSuite) TestPushObject(c *check.C) {
	ob := &TestObject{
		Data: []byte("Hello World"),
		info: &Info{},
	}
	// make new
	err := s.cl.New(ob, "testbucket", nil, nil)
	if err != nil {
		c.Fatal(err)
	}

	// fetch 'n store
	newob := &TestObject{
		Data: nil,
		info: &Info{},
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
}

func (s *riakSuite) TestStoreObject(c *check.C) {
	ob := &TestObject{
		Data: []byte("Hello World"),
		info: &Info{},
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
	nob := &TestObject{info: &Info{}}
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
}

func (s *riakSuite) TestPushChangeset(c *check.C) {
	ob := &TestObject{
		Data: []byte("Here's a body."),
		info: &Info{},
	}
	nob := &TestObject{info: &Info{}}

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

	lock := new(sync.Mutex)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err = cl.Store(ob, nil)
		if err != nil {
			lock.Lock()
			b.Fatal(err)
			lock.Unlock()
		}
	}
	b.StopTimer()
	cl.Close()
}

func BenchmarkMultiStore(b *testing.B) {
	NCONN := 5           // nubmer of connections
	nSEND := b.N / NCONN // number of stores/goroutine
	cl, err := DialOne("localhost:8087", "testClient")
	if err != nil {
		b.Fatal(err)
	}

	obs := make([]*TestObject, NCONN)
	for i := range obs {
		obs[i] = &TestObject{
			info: &Info{},
			Data: []byte("Hello World"),
		}
		err = cl.New(obs[i], "testbucket", nil, nil)
		if err != nil {
			b.Fatal(err)
		}
	}

	lock := new(sync.Mutex)
	wg := new(sync.WaitGroup)
	b.ReportAllocs()
	b.ResetTimer()
	for _, ob := range obs {
		wg.Add(1)
		go func(ob *TestObject, wg *sync.WaitGroup) {
			for i := 0; i < nSEND; i++ {
				err := cl.Store(ob, nil)
				if err != nil {
					lock.Lock()
					b.Fatal(err)
					lock.Unlock()
				}
			}
			wg.Done()
		}(ob, wg)
	}
	wg.Wait()
	b.StopTimer()
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

	lock := new(sync.Mutex)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err = cl.Fetch(ob, "testbucket", ob.Info().Key(), nil)
		if err != nil {
			lock.Lock()
			b.Fatal(err)
			lock.Unlock()
		}
	}
	b.StopTimer()
	cl.Close()
}

func BenchmarkMultiFetch(b *testing.B) {
	NCONNS := 5
	nSEND := b.N / NCONNS

	cl, err := DialOne("localhost:8087", "testClient")
	if err != nil {
		b.Fatal(err)
	}

	obs := make([]*TestObject, NCONNS)
	for i := range obs {
		obs[i] = &TestObject{
			info: &Info{},
			Data: []byte("Hello World"),
		}
		err = cl.New(obs[i], "testbucket", nil, nil)
		if err != nil {
			b.Fatal(err)
		}
	}

	lock := new(sync.Mutex)
	wg := new(sync.WaitGroup)
	b.ReportAllocs()
	b.ResetTimer()
	for _, ob := range obs {
		wg.Add(1)
		go func(o *TestObject, wg *sync.WaitGroup) {
			for i := 0; i < nSEND; i++ {
				err := cl.Fetch(o, "testbucket", o.Info().Key(), nil)
				if err != nil {
					lock.Lock()
					b.Fatal(err)
					lock.Unlock()
				}
			}
			wg.Done()
		}(ob, wg)
	}
	wg.Wait()
	b.StopTimer()
	cl.Close()
}
