// +build riak

package rkive

import (
	"bytes"
	check "gopkg.in/check.v1"
)

func (s *riakSuite) TestGetBucketTypeProperties(c *check.C) {
	c.Skip("not implemented")
}

func (s *riakSuite) TestCache(c *check.C) {
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

}
