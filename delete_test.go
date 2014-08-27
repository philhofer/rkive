// +build riak

package rkive

import (
	check "gopkg.in/check.v1"
)

func (s *riakSuite) TestDelete(c *check.C) {
	ob := &TestObject{
		Data: []byte("Blah."),
	}

	err := s.cl.New(ob, "testbucket", nil, nil)
	if err != nil {
		c.Fatal(err)
	}

	err = s.cl.Delete(ob, nil)
	if err != nil {
		c.Fatal(err)
	}

	err = s.cl.Fetch(ob, ob.Info().Bucket(), ob.Info().Key(), nil)
	if err != ErrNotFound {
		c.Fatalf("Expected ErrNotFound; got %s", err)
	}
}
