// +build riak

package rkive

import (
	"bytes"
	"fmt"
	check "gopkg.in/check.v1"
	"os"
	"sync"
	"testing"
	"time"
)

// timed suite - CAN ONLY USE 1 CONNECTION, otherwise timing is useless
type riakSuite struct {
	cl      *Client
	runtime time.Duration // needs to be incremented on tests
}

// untimed suite - can use many connections
type riakAsync struct {
	cl *Client
}

func TestAll(t *testing.T) {
	check.Suite(&riakAsync{})
	check.Suite(&riakSuite{})
	check.TestingT(t)
}

func (s *riakAsync) SetUpSuite(c *check.C) {
	addr := os.Getenv("RIAK_PB_URL")
	if addr == "" {
		addr = "localhost:8087"
	}
	var err error
	s.cl, err = DialOne(addr, "testClient")
	if err != nil {
		fmt.Printf("Couldn't connect to Riak: %s\n", err)
		os.Exit(1)
	}
	err = s.cl.Ping()
	if err != nil {
		c.Fatalf("Error on ping: %s", err)
	}
}

func (s *riakSuite) SetUpSuite(c *check.C) {
	addr := os.Getenv("RIAK_PB_URL")
	if addr == "" {
		addr = "localhost:8087"
	}
	var err error
	s.cl, err = DialOne(addr, "testClient")
	if err != nil {
		fmt.Printf("Couldn't connect to Riak: %s\n", err)
		os.Exit(1)
	}
	err = s.cl.Ping()
	if err != nil {
		c.Fatalf("Error on ping: %s", err)
	}
}

func (s *riakAsync) TearDownSuite(c *check.C) {
	s.cl.Close()
}

func (s *riakSuite) TearDownSuite(c *check.C) {
	s.cl.Close()
	c.Log("------------ STATS -----------")
	c.Logf("Total elapsed time: %s", s.runtime)
	c.Logf("total iowait time:  %s", time.Duration(s.cl.twait))
	c.Logf("total client time:  %s", s.runtime-time.Duration(s.cl.twait))
	c.Logf("non-iowait %%:       %.2f%%", 100*float64((s.runtime-time.Duration(s.cl.twait)))/float64(s.runtime))
	c.Logf("request count:      %d", s.cl.nwait)
	c.Logf("avg request time:   %s", time.Duration(s.cl.AvgWait()))
	c.Logf("nowait rate:        %d req/s", (uint64(time.Second))*(s.cl.nwait)/uint64(s.runtime-time.Duration(s.cl.twait)))
	c.Log("------------------------------")
}

type TestObject struct {
	info Info
	Data []byte
}

func (t *TestObject) Unmarshal(b []byte) error {
	t.Data = b
	return nil
}

func (t *TestObject) Marshal() ([]byte, error) {
	return t.Data, nil
}

func (t *TestObject) Info() *Info { return &t.info }

func (t *TestObject) NewEmpty() Object { return &TestObject{} }

// naive merge
func (t *TestObject) Merge(o Object) {
	tn := o.(*TestObject)
	if len(tn.Data) > len(t.Data) {
		t.Data = tn.Data
	}
}

//func TestMultipleVclocks(t *testing.T) {
func (s *riakSuite) TestMultipleVclocks(c *check.C) {
	startt := time.Now()
	travis := os.Getenv("TRAVIS")
	wercker := os.Getenv("WERCKER")
	if travis != "" || wercker != "" {
		c.Skip("The service doesn't have allow_mult set to true")
	}
	oba := &TestObject{
		Data: []byte("Body 1"),
	}

	obb := &TestObject{
		Data: []byte("Body 2..."),
	}

	// manually create conflict - a user can't ordinarily do this
	oba.info.bucket, oba.info.key = []byte("testbucket"), []byte("conflict")
	obb.info.bucket, obb.info.key = []byte("testbucket"), []byte("conflict")

	// The store operations should not error,
	// because we are doing a fetch and merge
	// when we detect multiple responses on
	// Store()
	err := s.cl.Store(obb, nil)
	if err != nil {
		c.Fatal(err)
	}
	err = s.cl.Store(oba, nil)
	if err != nil {
		c.Fatal(err)
	}

	// Since our Merge() function takes the longer of the
	// two Data fields, the body should always be "Body 2..."
	err = s.cl.Fetch(oba, "testbucket", "conflict", nil)
	if err != nil {
		c.Fatal(err)
	}

	if !bytes.Equal(oba.Data, []byte("Body 2...")) {
		c.Errorf("Data should be %q; got %q", "Body 2...", oba.Data)
	}
	s.runtime += time.Since(startt)
}

func (s *riakSuite) TestFetchNotFound(c *check.C) {
	startt := time.Now()
	ob := &TestObject{}

	err := s.cl.Fetch(ob, "anybucket", "dne", nil)
	if err == nil {
		c.Error("'err' should not be nil")
	}
	if err != ErrNotFound {
		c.Errorf("err is not ErrNotFound: %q", err)
	}
	s.runtime += time.Since(startt)
}

func (s *riakSuite) TestUpdate(c *check.C) {
	startt := time.Now()
	test := s.cl.Bucket("testbucket")

	lb := &TestObject{
		Data: []byte("flibbertyibbitygibbit"),
	}

	err := test.New(lb, nil)
	if err != nil {
		c.Fatal(err)
	}

	newlb := &TestObject{}

	err = test.Fetch(newlb, lb.Info().Key())
	if err != nil {
		c.Fatal(err)
	}

	if !bytes.Equal(newlb.Data, lb.Data) {
		c.Logf("Object 1 data: %q", lb.Data)
		c.Logf("Object 2 data: %q", newlb.Data)
		c.Errorf("Objects don't have the same body")
	}

	// make a modification
	newlb.Data = []byte("new data.")
	err = test.Push(newlb)
	if err != nil {
		c.Fatal(err)
	}

	// this should return true
	upd, err := test.Update(lb)
	if err != nil {
		c.Fatal(err)
	}

	if !upd {
		c.Error("Object was not updated.")
	}

	if !bytes.Equal(lb.Data, newlb.Data) {
		c.Error("Objects are not equal after update.")
	}

	// this should return false
	upd, err = test.Update(newlb)
	if err != nil {
		c.Fatal(err)
	}

	if upd {
		c.Error("Object was spuriously updated...?")
	}
	s.runtime += time.Since(startt)
}

func (s *riakSuite) TestHead(c *check.C) {
	startt := time.Now()
	tests := s.cl.Bucket("testbucket")

	ob := &TestObject{
		Data: []byte("exists."),
	}

	err := tests.New(ob, nil)
	if err != nil {
		c.Fatal(err)
	}

	// fetch head exists
	var info *Info
	info, err = s.cl.FetchHead("testbucket", ob.Info().Key())
	if err != nil {
		c.Fatal(err)
	}

	if !bytes.Equal(info.vclock, ob.info.vclock) {
		c.Errorf("vclocks not equal: %q and %q", info.vclock, ob.info.vclock)
	}

	// fetch dne
	_, err = s.cl.FetchHead("testbucket", "dne")
	if err != ErrNotFound {
		c.Errorf("expected ErrNotFound, got: %q", err)
	}
	s.runtime += time.Since(startt)
}

func (s *riakAsync) TestGoFlood(c *check.C) {
	c.Skip("This isn't necessary unless the connection handler changes.")

	// flood with goroutines
	// to test the stability
	// of the connection cap

	ob := &TestObject{
		Data: []byte("Here's a body."),
	}
	tests := s.cl.Bucket("testbucket")
	err := tests.New(ob, nil)
	if err != nil {
		c.Fatal(err)
	}

	key := ob.Info().Key()
	NGO := 200
	wg := new(sync.WaitGroup)
	lock := new(sync.Mutex)
	for i := 0; i < NGO; i++ {
		wg.Add(1)
		go func(key string, wg *sync.WaitGroup) {
			nob := &TestObject{}
			err := tests.Fetch(nob, key)
			if err != nil {
				lock.Lock()
				c.Error(err)
				lock.Unlock()
			}
			wg.Done()
		}(key, wg)
	}
	wg.Wait()
}
