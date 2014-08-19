// +build riak

package rkive

import (
	check "gopkg.in/check.v1"
)

func (s *riakSuite) TestIndexLookup(c *check.C) {
	ob := &TestObject{
		info: &Info{},
		Data: []byte("Hello world!"),
	}

	ob.Info().AddIndex("testIdx", "myValue")

	bucket := s.cl.Bucket("testbucket")

	err := bucket.New(ob, nil)
	if err != nil {
		c.Fatal(err)
	}

	res, err := bucket.IndexLookup("testIdx", "myValue")
	if err != nil {
		c.Fatal(err)
	}

	if res.Len() < 1 {
		c.Fatalf("Expected multiple keys; got %d.", res.Len())
	}
	if !res.Contains(ob.Info().Key()) {
		c.Errorf("Response doesn't contain original key...?")
	}

	hasCorrectIndex := func(o Object) bool {
		val := o.Info().GetIndex("testIdx")
		if val != "myValue" {
			c.Logf("Found incorrect: %v", o.Info().idxs)
			return false
		}
		return true
	}

	ncorrect, err := res.Which(ob, hasCorrectIndex)
	if err != nil {
		c.Fatal(err)
	}
	if len(ncorrect) != res.Len() {
		c.Errorf("Ncorrect is %d; response length is %d", len(ncorrect), res.Len())
	}

	c.Logf("Found %d keys.", res.Len())
}

func (s *riakSuite) TestIndexRange(c *check.C) {
	ob := &TestObject{
		info: &Info{},
		Data: []byte("Hello world!"),
	}

	ob.Info().AddIndexInt("testNum", 35)

	err := s.cl.New(ob, "testbucket", nil, nil)
	if err != nil {
		c.Fatal(err)
	}

	res, err := s.cl.Bucket("testbucket").IndexRange("testNum", 30, 40)
	if err != nil {
		c.Fatal(err)
	}

	if res.Len() < 1 {
		c.Fatalf("Expected multiple keys; got %d", res.Len())
	}

	if !res.Contains(ob.Info().Key()) {
		c.Errorf("Response doesn't contain original key...?")
	}
	c.Logf("Found %d keys.", res.Len())
}
