// +build riak

package rkive

import (
	"bytes"
	check "gopkg.in/check.v1"
	"time"
)

func (s *riakSuite) TestGetBucketTypeProperties(c *check.C) {
	c.Skip("not implemented")
}

func (s *riakSuite) TestCache(c *check.C) {
	startt := time.Now()

	cache := s.cl.Bucket("test-cache")
	err := cache.MakeCache()
	if err != nil {
		c.Fatal(err)
	}
	props, err := cache.GetProperties()
	if err != nil {
		c.Error(err)
	}
	if !bytes.Equal(props.GetBackend(), []byte("cache")) {
		c.Errorf("Expected backend %q; got %q", "cache", props.GetBackend())
	}

	// cache buckets are the only place
	// in which Overwrite is safe...

	ob := &TestObject{
		Data: []byte("Save this."),
	}

	err = cache.New(ob, nil)
	if err != nil {
		c.Error(err)
	}

	ob2 := &TestObject{
		Data: []byte("overwrite!"),
	}

	err = cache.Overwrite(ob2, ob.Info().Key())
	if err != nil {
		c.Error(err)
	}

	var upd bool
	upd, err = cache.Update(ob)
	if err != nil {
		c.Error(err)
	}
	if !upd {
		c.Error("Expected update.")
	}

	if !bytes.Equal(ob.Data, []byte("overwrite!")) {
		c.Errorf("Expected body %q; got %q", []byte("overwrite!"), ob.Data)
	}

	s.runtime += time.Since(startt)
}
