// +build riak

package rkive

import (
	check "gopkg.in/check.v1"
)

func (s *riakSuite) TestGetBucketTypeProperties(c *check.C) {
	c.Skip("not implemented")
}

func (s *riakSuite) TestMakeCache(c *check.C) {
	err := s.cl.Bucket("test-cache").MakeCache()
	if err != nil {
		c.Fatal(err)
	}
}
