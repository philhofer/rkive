package rkive

import (
	check "gopkg.in/check.v1"
	"time"
)

func (s *riakSuite) TestCounter(c *check.C) {
	startt := time.Now()

	var ct *Counter
	var err error
	ct, err = s.cl.Bucket("testbucket").NewCounter("test-counter", 0)
	if err != nil {
		c.Fatal(err)
	}

	start := ct.Val()

	err = ct.Add(5)
	if err != nil {
		c.Error(err)
	}

	if ct.Val() != start+5 {
		c.Errorf("Expected value %d; got %d", start+5, ct.Val())
	}

	err = ct.Refresh()
	if err != nil {
		c.Error(err)
	}

	if ct.Val() != start+5 {
		c.Errorf("Expected value %d; got %d", start+5, ct.Val())
	}

	nct, err := s.cl.Bucket("testbucket").GetCounter("test-counter")
	if err != nil {
		c.Fatal(err)
	}

	if nct.Val() != start+5 {
		c.Errorf("Expected value %d; got %d", start+5, nct.Val())
	}

	err = ct.Destroy()
	if err != nil {
		c.Error(err)
	}

	s.runtime += time.Since(startt)
}
